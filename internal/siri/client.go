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

type client struct {
	address             string
	ServerRequest       <-chan ServerRequest
	serverRequestWriter chan ServerRequest
	AutoClientResponse  *autoClientResponse
}

type autoClientResponse struct {
	Body   string
	Status int
}

type serverResponse struct {
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

func NewClient(address string) client {
	serverRequest := make(chan ServerRequest, 5)
	return client{
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

func (c client) Send(url string, body string) serverResponse {
	res, err := httpclient.Post(url, "application/xml", strings.NewReader(body))
	if err != nil {
		return serverResponse{Body: err.Error(), Status: http.StatusBadRequest, Language: "plaintext"}
	}
	defer res.Body.Close()
	bytesBody, err := io.ReadAll(res.Body)
	if err != nil {
		return serverResponse{Body: "", Status: res.StatusCode, Language: "plaintext"}
	}
	return serverResponse{Body: string(bytesBody), Status: res.StatusCode, Language: getLanguage(res.Header.Get("Content-Type"))}
}

func (c client) ListenAndServe() error {
	server := &http.Server{
		Addr:              c.address,
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           c.createHandler(),
	}
	return server.ListenAndServe()
}

func (c client) createHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", c.handleServerRequests)
	return mux
}

func (c client) handleServerRequests(w http.ResponseWriter, r *http.Request) {
	bytesBody, err := io.ReadAll(r.Body)
	if err != nil {
		request := ServerRequest{
			RemoteAddress: r.RemoteAddr,
			Url:           r.URL.RequestURI(),
			Body:          err.Error(),
			Language:      "plaintext",
		}
		c.serverRequestWriter <- request
		return
	}

	request := ServerRequest{
		RemoteAddress: r.RemoteAddr,
		Url:           r.URL.RequestURI(),
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
