package response

import (
	"fmt"
	"io"
	"strings"

	"httpfromtcp/internal/headers"
)

type Writer struct {
	writer io.Writer
}

type StatusCode int

const (
	StatusCodeOK                  StatusCode = 200
	StatusCodeBadRequest          StatusCode = 400
	StatusCodeInternalServerError StatusCode = 500
)

type Message string

const MessageOK Message = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

const MessageBadRequest Message = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

const MessageInternalServerError Message = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

var (
	ErrorUnknownStatusCode = fmt.Errorf("Unknown Status Code")
)

func NewWriter(w io.Writer) Writer {
	return Writer{writer: w}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")

	return headers
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	switch statusCode {
	case StatusCodeOK:
		fmt.Fprintf(w.writer, "HTTP/1.1 200 OK\r\n")
	case StatusCodeBadRequest:
		fmt.Fprintf(w.writer, "HTTP/1.1 400 Bad Request\r\n")
	case StatusCodeInternalServerError:
		fmt.Fprintf(w.writer, "HTTP/1.1 500 Internal Server Error\r\n")
	default:
		return ErrorUnknownStatusCode
	}

	return nil
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	for key, value := range h {
		fmt.Fprintf(w.writer, "%s: %s\r\n", key, value)
	}

	fmt.Fprintf(w.writer, "\r\n")

	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := fmt.Fprintf(w.writer, "%s", p)

	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	n, err := fmt.Fprintf(w.writer, "%X\r\n%s\r\n", len(p), p)

	return n, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	n, err := fmt.Fprintf(w.writer, "0\r\n")

	return n, err
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	trailers := strings.SplitSeq(h.Get("Trailer"), ", ")

	for trailer := range trailers {
		fmt.Fprintf(w.writer, "%s: %s\r\n", trailer, h.Get(trailer))
	}

	fmt.Fprintf(w.writer, "\r\n")

	return nil
}
