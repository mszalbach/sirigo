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
	ServerRequest       <-chan ServerRequest
	serverRequestWriter chan ServerRequest
	AutoClientResponse  *AutoClientResponse
}

type AutoClientResponse struct {
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
	Url           string
	Body          string
	Language      string
}

func NewClient(address string) Client {
	serverRequest := make(chan ServerRequest, 1)
	return Client{
		address:             address,
		ServerRequest:       serverRequest,
		serverRequestWriter: serverRequest,
		AutoClientResponse: &AutoClientResponse{
			Body:   "",
			Status: http.StatusOK,
		},
	}
}

var client http.Client = http.Client{Timeout: 10 * time.Second}

func (c Client) Send(url string, body string) ServerResponse {
	res, err := client.Post(url, "application/xml", strings.NewReader(body))
	if err != nil {
		return ServerResponse{Body: err.Error(), Status: res.StatusCode, Language: "plaintext"}
	}
	defer res.Body.Close()
	bytesBody, err := io.ReadAll(res.Body)
	if err != nil {
		return ServerResponse{Body: "", Status: res.StatusCode, Language: "plaintext"}
	}
	return ServerResponse{Body: string(bytesBody), Status: res.StatusCode, Language: getLanguage(res.Header.Get("content-type"))}
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
			Url:           r.URL.RequestURI(),
			Body:          err.Error(),
			Language:      "plaintext",
		}
		c.serverRequestWriter <- request
	}

	request := ServerRequest{
		RemoteAddress: r.RemoteAddr,
		Url:           r.URL.RequestURI(),
		Body:          string(bytesBody),
		Language:      getLanguage(r.Header.Get("content-type")),
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
