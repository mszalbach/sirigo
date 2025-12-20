package httputils

import (
	"net/http"
	"time"
)

// HTTPServer is a simple HTTP server wrapper
type HTTPServer struct {
	*http.Server
}

// NewHTTPServer creates a new HTTPServer with default settings
func NewHTTPServer(address string) *HTTPServer {
	return &HTTPServer{
		Server: &http.Server{Addr: address, ReadHeaderTimeout: 5 * time.Second},
	}
}
