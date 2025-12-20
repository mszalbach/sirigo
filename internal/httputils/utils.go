package httputils

import (
	"mime"
	"net/http"
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
