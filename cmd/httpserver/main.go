package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
)

const port = 42069

func main() {
	handler := func(w *response.Writer, req *request.Request) {
		status := response.StatusCodeOK
		message := response.MessageOK
		headers := response.GetDefaultHeaders(0)
		after, found := strings.CutPrefix(req.RequestLine.RequestTarget, "/httpbin")

		if found {
			resp, err := http.Get("https://httpbin.org" + after)

			if err != nil {
				status = response.StatusCodeInternalServerError
				message = response.MessageInternalServerError

				headers.Set("Content-Length", fmt.Sprintf("%d", len(message)))

				w.WriteStatusLine(status)
				w.WriteHeaders(headers)
				w.WriteBody([]byte(message))
			} else {
				defer resp.Body.Close()

				headers.Delete("Content-Length")
				headers.Set("Transfer-Encoding", "chunked")

				w.WriteStatusLine(status)
				w.WriteHeaders(headers)

				data := make([]byte, 1024)

				for {
					n, err := resp.Body.Read(data)

					if n > 0 {
						w.WriteChunkedBody(data[:n])
					}

					if err != nil {
						break
					}

					log.Println("Data read:", n)
				}

				w.WriteChunkedBodyDone()
			}
		} else {
			if strings.HasPrefix(req.RequestLine.RequestTarget, "/yourproblem") {
				status = response.StatusCodeBadRequest
				message = response.MessageBadRequest
			} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/myproblem") {
				status = response.StatusCodeInternalServerError
				message = response.MessageInternalServerError
			}

			headers.Set("Content-Type", "text/html")
			headers.Set("Content-Length", fmt.Sprintf("%d", len(message)))

			w.WriteStatusLine(status)
			w.WriteHeaders(headers)
			w.WriteBody([]byte(message))
		}
	}

	server, err := server.Serve(port, handler)

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	defer server.Close()

	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Server gracefully stopped")
}
