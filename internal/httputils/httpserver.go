package httputils

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// LoggingMuxServer is a simple HTTP server wrapper
type LoggingMuxServer struct {
	*http.Server
	mux *http.ServeMux
}

// NewLoggingMuxServer creates a new HTTPServer with default settings
func NewLoggingMuxServer(address string, writer io.Writer) *LoggingMuxServer {
	mux := http.NewServeMux()
	return &LoggingMuxServer{
		Server: &http.Server{
			Addr:              address,
			ReadHeaderTimeout: 5 * time.Second,
			Handler:           loggingMiddleware(mux, writer),
		},
		mux: mux,
	}
}

// HandleFunc registers the handler function for the given pattern
func (hs *LoggingMuxServer) HandleFunc(pattern string, handleFunc func(http.ResponseWriter, *http.Request)) {
	hs.mux.HandleFunc(pattern, handleFunc)
}

func loggingMiddleware(next http.Handler, writer io.Writer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggingResponseWriter := &loggingResponseWriter{
			wrappedWriter: w,
		}

		bytesBody, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Warn("Could not read body from incoming server request", slog.Any("error", err.Error()))
		}
		// restore body because you can read only once
		r.Body = io.NopCloser(bytes.NewBuffer(bytesBody))

		// call original handler
		next.ServeHTTP(loggingResponseWriter, r)

		// log request and response
		logRequest(writer, "Incoming Request:", r, bytesBody)
		logResponse(
			writer,
			"Outgoing Response:",
			loggingResponseWriter.statusCode,
			loggingResponseWriter.Header(),
			loggingResponseWriter.body,
		)
	})
}

type loggingResponseWriter struct {
	wrappedWriter http.ResponseWriter
	statusCode    int
	body          []byte
}

func (lw *loggingResponseWriter) Header() http.Header {
	return lw.wrappedWriter.Header()
}

func (lw *loggingResponseWriter) WriteHeader(statusCode int) {
	lw.statusCode = statusCode
	lw.wrappedWriter.WriteHeader(statusCode)
}

func (lw *loggingResponseWriter) Write(body []byte) (int, error) {
	lw.body = append(lw.body, body...)
	return lw.wrappedWriter.Write(body)
}
