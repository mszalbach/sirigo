// Package siri contains everything to handle SIRI communication
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

type Client struct {
	address             string
	ClientRef           string
	ServerURL           string
	ServerRequest       <-chan ServerRequest
	serverRequestWriter chan ServerRequest
	AutoClientResponse  *autoClientResponse
}

type autoClientResponse struct {
	Body   string
	Status int
}

type ServerResponse struct {
	Body     string
	Status   int
	Language string
}

type ServerRequest struct {
	RemoteAddress string
	URL           string
	Body          string
	Language      string
}

func NewClient(clientRef string, serverURL string, address string) Client {
	serverRequest := make(chan ServerRequest, 5)
	return Client{
		ClientRef:           clientRef,
		ServerURL:           serverURL,
		address:             address,
		ServerRequest:       serverRequest,
		serverRequestWriter: serverRequest,
		AutoClientResponse: &autoClientResponse{
			Body:   "",
			Status: http.StatusOK,
		},
	}
}

var httpclient http.Client = http.Client{Timeout: 10 * time.Second}

func (c Client) Send(url string, body string) (ServerResponse, error) {
	res, err := httpclient.Post(url, "application/xml", strings.NewReader(body))
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

func (c Client) ListenAndServe() error {
	server := &http.Server{
		Addr:              c.address,
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

	w.WriteHeader(c.AutoClientResponse.Status)
	_, ferr := fmt.Fprint(w, c.AutoClientResponse.Body)
	if ferr != nil {
		slog.Error("Could not write auto response", slog.Any("error", err))
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
