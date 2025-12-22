// Package httputils provides HTTP client functionality and does most of the error handling
package httputils

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"
)

// LoggingClient is a simple HTTP client wrapper which logs requests and responses
type LoggingClient struct {
	client http.Client
	writer io.Writer
}

// Response represents an HTTP response
type Response struct {
	Body       string
	StatusCode int
	Header     http.Header
}

// NewLoggingClient creates a new LoggingClient with default settings
func NewLoggingClient(writer io.Writer) LoggingClient {
	return LoggingClient{
		client: http.Client{Timeout: 10 * time.Second},
		writer: writer,
	}
}

// PostXML sends a POST request with XML content to the specified URL
func (hc LoggingClient) PostXML(url string, body string) (Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	req.Header.Set(HeaderContentType, ContentTypeXML)
	return hc.Do(req)
}

// Do sends an HTTP request and returns the response
func (hc LoggingClient) Do(req *http.Request) (Response, error) {
	bytesBody, err := io.ReadAll(req.Body)
	if err != nil {
		return Response{}, err
	}
	// restore body because you can read only once
	req.Body = io.NopCloser(bytes.NewBuffer(bytesBody))

	res, err := hc.client.Do(req)
	if err != nil {
		return Response{}, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Response{}, err
	}

	logRequest(hc.writer, "Outgoing Request:", req, bytesBody)
	logResponse(hc.writer, "Incoming Response:", res.StatusCode, res.Header, body)

	return Response{
		Body:       string(body),
		StatusCode: res.StatusCode,
		Header:     res.Header,
	}, nil
}
