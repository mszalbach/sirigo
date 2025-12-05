// Package siri contains everything needed to handle SIRI communication
package siri

import (
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"
	"time"
)

// Client contains everything needed for a SIRI client
type Client struct {
	clientAddress       string
	ClientRef           string
	ServerURL           string
	ServerRequest       <-chan ServerRequest
	serverRequestWriter chan ServerRequest
	AutoClientResponse  *AutoClientResponse
}

// AutoClientResponse should be used to answer server requests
type AutoClientResponse struct {
	Body   string
	Status int
}

// ServerResponse returned when sending requests to the server
type ServerResponse struct {
	Body     string
	Status   int
	Language string
}

// ServerRequest represents the requests the SIRI server sends to the client, such as DataReady requests
type ServerRequest struct {
	RemoteAddress string
	URL           string
	Body          string
	Language      string
}

// NewClient creates a new Client to interact with a SIRI server
func NewClient(clientRef string, serverURL string, address string) Client {
	serverRequest := make(chan ServerRequest, 5)
	return Client{
		ClientRef:           clientRef,
		ServerURL:           serverURL,
		clientAddress:       address,
		ServerRequest:       serverRequest,
		serverRequestWriter: serverRequest,
		AutoClientResponse: &AutoClientResponse{
			Body:   "",
			Status: http.StatusOK,
		},
	}
}

var httpclient = http.Client{Timeout: 10 * time.Second}

// Send sends a message to the SIRI server
func (c Client) Send(url string, body string) (ServerResponse, error) {
	executedBody, err := executeTemplate(body, data{Now: time.Now(), ClientRef: c.ClientRef})
	if err != nil {
		return ServerResponse{}, err
	}

	res, err := httpclient.Post(url, "application/xml", strings.NewReader(executedBody))
	if err != nil {
		return ServerResponse{}, err
	}
	defer res.Body.Close()
	bytesBody, err := io.ReadAll(res.Body)
	if err != nil {
		return ServerResponse{}, err
	}
	return ServerResponse{
		Body:     string(bytesBody),
		Status:   res.StatusCode,
		Language: getLanguage(res.Header.Get("Content-Type")),
	}, nil
}

// ListenAndServe starts the HTTP server needed to listen for SIRI server requests such as DataReady requests
func (c Client) ListenAndServe() error {
	server := &http.Server{
		Addr:              c.clientAddress,
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           c.createHandler(),
	}
	return server.ListenAndServe()
}

func (c Client) createHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", c.handleServerRequests)
	return mux
}

func (c Client) handleServerRequests(w http.ResponseWriter, r *http.Request) {
	bytesBody, err := io.ReadAll(r.Body)
	if err != nil {
		request := ServerRequest{
			RemoteAddress: r.RemoteAddr,
			URL:           r.URL.RequestURI(),
			Body:          err.Error(),
			Language:      "plaintext",
		}
		c.serverRequestWriter <- request
		return
	}

	request := ServerRequest{
		RemoteAddress: r.RemoteAddr,
		URL:           r.URL.RequestURI(),
		Body:          string(bytesBody),
		Language:      getLanguage(r.Header.Get("Content-Type")),
	}

	c.serverRequestWriter <- request

	responseBody, err := executeTemplate(
		c.AutoClientResponse.Body,
		data{Now: time.Now(), ClientRef: c.ClientRef},
	)
	if err != nil {
		slog.Error("Could not execute template for autoresponse", slog.Any("error", err))
		return
	}
	w.WriteHeader(c.AutoClientResponse.Status)
	_, ferr := fmt.Fprint(w, responseBody)
	if ferr != nil {
		slog.Error("Could not write auto response", slog.Any("error", ferr))
	}
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
