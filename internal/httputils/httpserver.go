package httputils

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
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
		fmt.Fprintf(writer, "Incoming Request:\n")
		fmt.Fprintf(writer, "Server %s\n", r.RemoteAddr)
		fmt.Fprintf(writer, "%s %s \n", r.Method, r.URL.RequestURI())
		fmt.Fprintf(writer, "Host %s\n", r.Host)
		fmt.Fprintf(writer, "Content-Type %s\n\n", r.Header.Get(HeaderContentType))
		fmt.Fprintf(writer, "%s\n\n", string(bytesBody))

		fmt.Fprintf(writer, "Outgoing Response:\n")
		fmt.Fprintf(writer, "%s\n", strconv.Itoa(loggingResponseWriter.statusCode))
		fmt.Fprintf(writer, "Content-Type %s\n\n", loggingResponseWriter.Header().Get(HeaderContentType))
		fmt.Fprintf(writer, "%s\n\n", string(loggingResponseWriter.body))
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
