package siri //nolint testpackage

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_siri_client_sending_to_server(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/siri/2.1/situation-exchange", req.URL.String())
		rw.Header().Set("Content-Type", "application/xml")
		rw.WriteHeader(http.StatusOK)
		_, err := fmt.Fprint(rw, `
<Siri>
	<SubscriptionResponse>
		<ResponseTimestamp>2004-12-17T09:30:47-05:00</ResponseTimestamp>
		<ResponderRef>EREWHON</ResponderRef>
		<ResponseStatus>
			<ResponseTimestamp>2004-12-17T09:30:47-05:01</ResponseTimestamp>
			<SubscriptionRef>0003456</SubscriptionRef>
			<Status>true</Status>
			<ValidUntil>2004-12-17T09:30:47-05:00</ValidUntil>
			<ShortestPossibleCycle>P1Y2M3DT10H30M</ShortestPossibleCycle>
		</ResponseStatus>
	</SubscriptionResponse>
</Siri>`)
		assert.NoError(t, err)
	}))
	defer server.Close()

	// When
	client := NewClient("IMPORTANT")
	actual := client.Send(server.URL+"/siri/2.1/situation-exchange", `
<Siri>
	<ServiceRequest>
		<RequestTimestamp>2004-12-17T09:30:47-05:00</RequestTimestamp>
		<RequestorRef>NADER</RequestorRef>
		<SituationExchangeRequest>
			<RequestTimestamp>2004-12-17T09:30:47-05:00</RequestTimestamp>
			<Scope>line</Scope>
			<LineRef>52</LineRef>
		</SituationExchangeRequest>
	</ServiceRequest>
</Siri>`)

	// Then
	expected := serverResponse{Body: `
<Siri>
	<SubscriptionResponse>
		<ResponseTimestamp>2004-12-17T09:30:47-05:00</ResponseTimestamp>
		<ResponderRef>EREWHON</ResponderRef>
		<ResponseStatus>
			<ResponseTimestamp>2004-12-17T09:30:47-05:01</ResponseTimestamp>
			<SubscriptionRef>0003456</SubscriptionRef>
			<Status>true</Status>
			<ValidUntil>2004-12-17T09:30:47-05:00</ValidUntil>
			<ShortestPossibleCycle>P1Y2M3DT10H30M</ShortestPossibleCycle>
		</ResponseStatus>
	</SubscriptionResponse>
</Siri>`,
		Language: "xml",
		Status:   http.StatusOK}
	assert.Equal(t, expected, actual)

}

func Test_siri_client_receiving_from_server(t *testing.T) {
	// Given
	client := NewClient("NOT IMPORTANT")
	client.AutoClientResponse.Body = `
<Siri>
	<DataReadyAcknowledgement>
		<ResponseTimestamp>2004-12-17T09:30:47-05:00</ResponseTimestamp>
		<ConsumerRef>NADER</ConsumerRef>
		<Status>true</Status>
	</DataReadyAcknowledgement>
</Siri>`

	// When
	serverRequest, _ := http.NewRequest(http.MethodPost, "/siri", strings.NewReader(`
<Siri>
	<DataReadyNotification>
		<RequestTimestamp>2004-12-17T09:30:47-05:00</RequestTimestamp>
		<ProducerRef>KUBRICK</ProducerRef>
	</DataReadyNotification>
</Siri>`))
	serverRequest.RemoteAddr = "196.4.4.1"
	serverRequest.Header.Set("Content-Type", "application/xml")

	response := httptest.NewRecorder()

	client.createHandler().ServeHTTP(response, serverRequest)

	// Then
	assert.Equal(t, http.StatusOK, response.Result().StatusCode)

	expectedClientResponse := `
<Siri>
	<DataReadyAcknowledgement>
		<ResponseTimestamp>2004-12-17T09:30:47-05:00</ResponseTimestamp>
		<ConsumerRef>NADER</ConsumerRef>
		<Status>true</Status>
	</DataReadyAcknowledgement>
</Siri>`
	assert.Equal(t, expectedClientResponse, response.Body.String())

	require.Len(t, client.ServerRequest, 1)
	actualServerRequest := <-client.ServerRequest

	expectedServerRequest := ServerRequest{
		RemoteAddress: "196.4.4.1",
		Url:           "/siri",
		Language:      "xml",
		Body: `
<Siri>
	<DataReadyNotification>
		<RequestTimestamp>2004-12-17T09:30:47-05:00</RequestTimestamp>
		<ProducerRef>KUBRICK</ProducerRef>
	</DataReadyNotification>
</Siri>`,
	}

	assert.Equal(t, expectedServerRequest, actualServerRequest)

}

func Test_client_send_returns_server_responses(t *testing.T) {
	testCases := []struct {
		actualStatus   int
		expectedStatus int
	}{
		{http.StatusOK, http.StatusOK},
		{http.StatusInternalServerError, http.StatusInternalServerError},
		{http.StatusNotFound, http.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(strconv.Itoa(tc.actualStatus), func(t *testing.T) {
			// Given
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				assert.Equal(t, "/siri/v2", req.URL.String())
				rw.WriteHeader(tc.actualStatus)
			}))
			defer server.Close()

			// When
			client := NewClient("LISTENER NOT IMPORTANT")
			actual := client.Send(server.URL+"/siri/v2", "IGNORE")

			// Then
			expected := serverResponse{Body: ``, Language: "plaintext", Status: tc.expectedStatus}
			assert.Equal(t, expected, actual)
		})
	}

}

func Test_client_send_understands_content_types(t *testing.T) {
	testCases := []struct {
		actualContentType string
		expectedLanguage  string
	}{
		{"application/xml", "xml"},
		{"text/xml", "xml"},
		{"application/json", "json"},
		{"text/csv", "csv"},
		{"SOMETHING-STRANGE", "plaintext"},
	}
	for _, tc := range testCases {
		t.Run(tc.actualContentType, func(t *testing.T) {
			// Given
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				assert.Equal(t, "/siri/v2", req.URL.String())
				rw.Header().Set("Content-Type", tc.actualContentType)
				rw.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			// When
			client := NewClient("NOT IMPORTANT")
			actual := client.Send(server.URL+"/siri/v2", "IGNORE")

			// Then
			expected := serverResponse{Body: ``, Language: tc.expectedLanguage, Status: http.StatusOK}
			assert.Equal(t, expected, actual)
		})
	}
}

func Test_server_does_not_work_for_non_post(t *testing.T) {
	testCases := []string{http.MethodGet, http.MethodPut, http.MethodHead}
	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			// Given
			client := NewClient("NOT IMPORTANT")

			// When
			request, _ := http.NewRequest(tc, "/players/Floyd", strings.NewReader(`
<Siri>
	<DataReadyNotification>
		<RequestTimestamp>2004-12-17T09:30:47-05:00</RequestTimestamp>
		<ProducerRef>KUBRICK</ProducerRef>
	</DataReadyNotification>
</Siri>`))
			response := httptest.NewRecorder()
			client.createHandler().ServeHTTP(response, request)

			// Then
			assert.Equal(t, http.StatusMethodNotAllowed, response.Result().StatusCode)
			assert.Empty(t, client.ServerRequest)
		})
	}

}
