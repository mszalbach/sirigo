package httputils

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
)

const (
	// ContentTypeXML is the MIME type for XML content
	ContentTypeXML = "application/xml"
	// HeaderContentType is the HTTP header key for Content-Type
	HeaderContentType = "Content-Type"
)

// GetLanguage extracts the language from the Content-Type header
func GetLanguage(header http.Header) string {
	contentType := header.Get(HeaderContentType)
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

func logRequest(writer io.Writer, logHeading string, request *http.Request, body []byte) {
	fmt.Fprintf(writer, "%s\n", logHeading)
	fmt.Fprintf(writer, "IP %s\n", request.RemoteAddr)
	fmt.Fprintf(writer, "%s %s%s \n", request.Method, request.Host, request.URL.RequestURI())
	fmt.Fprintf(writer, "Content-Type %s\n\n", request.Header.Get(HeaderContentType))
	fmt.Fprintf(writer, "%s\n\n", string(body))
}

func logResponse(writer io.Writer, logHeading string, statusCode int, header http.Header, body []byte) {
	fmt.Fprintf(writer, "%s\n", logHeading)
	fmt.Fprintf(writer, "%s\n", strconv.Itoa(statusCode))
	fmt.Fprintf(writer, "Content-Type %s\n\n", header.Get(HeaderContentType))
	fmt.Fprintf(writer, "%s\n\n", string(body))
}
