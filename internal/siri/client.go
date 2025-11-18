package siri

import (
	"io"
	"mime"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	address             string
	clientEndpoint      *http.ServeMux
	ServerRequest       <-chan ServerRequest
	serverRequestWriter chan ServerRequest
}

type ServerResponse struct {
	Body     string
	Status   string
	Language string
}

type ServerRequest struct {
	//TODO welche info brauch ich alles? vermutlich auch die IP der gegenseite
	Url      string
	Body     string
	Language string
}

func NewClient(address string) Client {
	serverRequest := make(chan ServerRequest, 1)
	return Client{
		address:             address,
		ServerRequest:       serverRequest,
		serverRequestWriter: serverRequest,
		clientEndpoint:      http.NewServeMux(),
	}
}

func (c Client) Send(url string, body string) ServerResponse {
	return post(url, body)
}

func (c Client) ListenAndServer() error {
	server := &http.Server{
		Addr:              c.address,
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           c.clientEndpoint,
	}
	c.clientEndpoint.HandleFunc("/", c.handleAllRequests)
	return server.ListenAndServe()
}

func (c Client) handleAllRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "SIRI Requests should be done as POST", http.StatusMethodNotAllowed)
		return
	}

	bytesBody, err := io.ReadAll(r.Body)
	if err != nil {
		request := ServerRequest{
			Url:  r.URL.RequestURI(),
			Body: err.Error(),
			//TODO oder go?
			Language: "plaintext",
		}
		c.serverRequestWriter <- request
	}

	request := ServerRequest{
		Url:      r.URL.RequestURI(),
		Body:     string(bytesBody),
		Language: getLanguage(r.Header.Get("content-type")),
	}

	c.serverRequestWriter <- request

	//TODO autoantwort fehlt
}

func getLanguage(contentType string) string {
	m, _, err := mime.ParseMediaType(contentType)

	if err != nil {
		return "plaintext"
	}

	parts := strings.Split(m, "/")
	if len(parts) < 2 {
		return "plaintext"
	}

	return parts[1]
}

var client http.Client = http.Client{Timeout: 10 * time.Second}

func post(url string, body string) ServerResponse {
	res, err := client.Post(url, "application/xml", strings.NewReader(body)) //nolint gosec

	if err != nil {
		return ServerResponse{
			Body:     err.Error(),
			Status:   res.Status,
			Language: "plaintext"}
	}
	defer res.Body.Close()

	bytesBody, err := io.ReadAll(res.Body)
	if err != nil {
		return ServerResponse{
			Body:     "",
			Status:   res.Status,
			Language: "plaintext"}
	}

	return ServerResponse{
		Body:     string(bytesBody),
		Status:   res.Status,
		Language: getLanguage(res.Header.Get("content-type"))}
}
