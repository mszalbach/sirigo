package communication

import (
	"io"
	"mime"
	"net/http"
	"strings"
)

type HttpResponse struct {
	Body     string
	Status   string
	Language string
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

func Get(url string) HttpResponse {
	res, err := http.Get(url) //nolint gosec

	if err != nil {
		return HttpResponse{
			Body:     "",
			Status:   err.Error(),
			Language: "plaintext"}
	}
	defer res.Body.Close()

	bytesBody, err := io.ReadAll(res.Body)
	if err != nil {
		return HttpResponse{
			Body:     "",
			Status:   res.Status,
			Language: "plaintext"}
	}

	return HttpResponse{
		Body:     string(bytesBody),
		Status:   res.Status,
		Language: getLanguage(res.Header.Get("content-type"))}
}
