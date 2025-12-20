// Package siri contains everything needed to handle SIRI communication
package siri

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/mszalbach/sirigo/internal/httputils"
)

// Client contains everything needed for a SIRI client
type Client struct {
	clientAddress       string
	ClientRef           string
	ServerURL           string
	ServerRequest       <-chan ServerRequest
	AutoClientResponse  *AutoClientResponse
	server              *http.Server
	serverRequestWriter chan ServerRequest
	httpclient          httputils.HTTPClient
}

// ClientRequest represents a request sent by the SIRI client to the server
type ClientRequest struct {
	URL  string
	Body string
}

// AutoClientResponse represents the automatic response sent by the client to the SIRI server
// used for requests such as DataReady requests
// there is currently only one automatic response for all requests
type AutoClientResponse struct {
	Body   string
	Status int
}

// ServerResponse represents the response from the SIRI server to a client request
type ServerResponse struct {
	Body     string
	Status   int
	Language string
}

// ServerRequest represents a request sent by the SIRI server to the client
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
		httpclient: httputils.NewHTTPClient(),
		server: &http.Server{
			Addr:              address,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

// Send sends a message to the SIRI server
func (c Client) Send(clientRequest ClientRequest) (ServerResponse, error) {
	executedBody, err := executeTemplate(clientRequest.Body, data{Now: time.Now(), ClientRef: c.ClientRef})
	if err != nil {
		return ServerResponse{}, err
	}
	res, err := c.httpclient.PostXML(clientRequest.URL, executedBody)
	if err != nil {
		return ServerResponse{}, err
	}
	return ServerResponse{
		Body:     res.Body,
		Status:   res.StatusCode,
		Language: getLanguage(httputils.GetHeaderValue(res.Header, httputils.HeaderContentType)),
	}, nil
}

// ListenAndServe starts the HTTP server needed to listen for SIRI server requests such as DataReady requests
func (c Client) ListenAndServe() error {
	c.server.Handler = c.createHandler()
	return c.server.ListenAndServe()
}

// Stop stops the http server for the given context. Uses to be able to correctly stop the server from somewhere else
func (c Client) Stop(ctx context.Context) error {
	return c.server.Shutdown(ctx)
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
