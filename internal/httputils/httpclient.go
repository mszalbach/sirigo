// Package httputils provides HTTP client functionality and does most of the error handling
package httputils

import (
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPClient is a simple HTTP client wrapper
type HTTPClient struct {
	client http.Client
}

// Response represents an HTTP response
type Response struct {
	Body       string
	StatusCode int
	Header     http.Header
}

// NewHTTPClient creates a new HTTPClient with default settings
func NewHTTPClient() HTTPClient {
	return HTTPClient{
		client: http.Client{Timeout: 10 * time.Second},
	}
}

// PostXML sends a POST request with XML content to the specified URL
func (hc HTTPClient) PostXML(url string, body string) (Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	req.Header.Set(HeaderContentType, ContentTypeXML)
	return hc.Do(req)
}

// Do sends an HTTP request and returns the response
func (hc HTTPClient) Do(req *http.Request) (Response, error) {
	res, err := hc.client.Do(req)
	if err != nil {
		return Response{}, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Response{}, err
	}

	return Response{
		Body:       string(body),
		StatusCode: res.StatusCode,
		Header:     res.Header,
	}, nil
}
