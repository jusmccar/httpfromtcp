package response

import (
	"fmt"
	"io"

	"httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusCodeOK                  StatusCode = 200
	StatusCodeBadRequest          StatusCode = 400
	StatusCodeInternalServerError StatusCode = 500
)

var (
	ErrorUnknownStatusCode = fmt.Errorf("Unknown Status Code")
)

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers["content-length"] = fmt.Sprintf("%d", contentLen)
	headers["connection"] = "close"
	headers["content-type"] = "text/plain"

	return headers
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case StatusCodeOK:
		fmt.Fprintf(w, "HTTP/1.1 200 OK\r\n")
	case StatusCodeBadRequest:
		fmt.Fprintf(w, "HTTP/1.1 400 Bad Request\r\n")
	case StatusCodeInternalServerError:
		fmt.Fprintf(w, "HTTP/1.1 500 Internal Server Error\r\n")
	default:
		return ErrorUnknownStatusCode
	}

	return nil
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	fmt.Fprintf(w, "Content-Length: %s\r\n", headers.Get("Content-Length"))
	fmt.Fprintf(w, "Connection: %s\r\n", headers.Get("Connection"))
	fmt.Fprintf(w, "Content-Type: %s\r\n", headers.Get("Content-Type"))
	fmt.Fprintf(w, "\r\n")

	return nil
}
