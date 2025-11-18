package siri_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/mszalbach/sirigo/internal/siri"
	"github.com/stretchr/testify/assert"
)

func Test_siri_communcation(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/siri/v2/situation-exchange", req.URL.String())
		rw.Header().Set("content-type", "application/xml")
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

	//When
	client := siri.NewClient("LISTENER NOT IMPORTANT")
	actual := client.Send(server.URL+"/siri/v2/situation-exchange", `
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

	//Then
	expected := siri.ServerResponse{Body: `
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
		Status:   "200 OK"}
	assert.Equal(t, expected, actual)

	//TODO server sends data ready and it is received
}

func Test_client_send_returns_server_responses(t *testing.T) {
	testCases := []struct {
		actualStatus   int
		expectedStatus string
	}{
		{http.StatusOK, "200 OK"},
		{http.StatusInternalServerError, "500 Internal Server Error"},
		{http.StatusNotFound, "404 Not Found"},
	}

	for _, tc := range testCases {
		t.Run(strconv.Itoa(tc.actualStatus), func(t *testing.T) {
			// Given
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				assert.Equal(t, "/siri/v2", req.URL.String())
				rw.WriteHeader(tc.actualStatus)
			}))
			defer server.Close()

			//When
			client := siri.NewClient("LISTENER NOT IMPORTANT")
			actual := client.Send(server.URL+"/siri/v2", "IGNORE")

			//Then
			expected := siri.ServerResponse{Body: ``, Language: "plaintext", Status: tc.expectedStatus}
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
				rw.Header().Set("content-type", tc.actualContentType)
				rw.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			//When
			client := siri.NewClient("LISTENER NOT IMPORTANT")
			actual := client.Send(server.URL+"/siri/v2", "IGNORE")

			//Then
			expected := siri.ServerResponse{Body: ``, Language: tc.expectedLanguage, Status: "200 OK"}
			assert.Equal(t, expected, actual)
		})
	}
}
