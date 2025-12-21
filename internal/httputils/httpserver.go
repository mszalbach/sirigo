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
		Server: &http.Server{Addr: address, ReadHeaderTimeout: 5 * time.Second, Handler: http.NewServeMux()},
	}
}

// HandleFunc registers the handler function for the given pattern
func (hs *HTTPServer) HandleFunc(pattern string, handleFunc func(http.ResponseWriter, *http.Request)) {
	hs.Handler.(*http.ServeMux).HandleFunc(pattern, handleFunc)
}
