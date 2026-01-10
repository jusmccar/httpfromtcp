package main

import (
	"crypto/sha256"
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
		h := response.GetDefaultHeaders(0)
		after, found := strings.CutPrefix(req.RequestLine.RequestTarget, "/httpbin")

		if found {
			resp, err := http.Get("https://httpbin.org" + after)

			if err != nil {
				status = response.StatusCodeInternalServerError
				message = response.MessageInternalServerError

				h.Set("Content-Length", fmt.Sprintf("%d", len(message)))

				w.WriteStatusLine(status)
				w.WriteHeaders(h)
				w.WriteBody([]byte(message))
			} else {
				defer resp.Body.Close()

				h.Delete("Content-Length")
				h.Set("Transfer-Encoding", "chunked")
				h.Append("Trailer", "X-Content-SHA256")
				h.Append("Trailer", "X-Content-Length")

				w.WriteStatusLine(status)
				w.WriteHeaders(h)

				data := make([]byte, 1024)
				fullData := []byte{}

				for {
					n, err := resp.Body.Read(data)

					if n > 0 {
						w.WriteChunkedBody(data[:n])
					}

					fullData = append(fullData, data[:n]...)

					if err != nil {
						break
					}

					log.Println("Data read:", n)
				}

				w.WriteChunkedBodyDone()

				h.Set("X-Content-SHA256", fmt.Sprintf("%x", sha256.Sum256(fullData)))
				h.Set("X-Content-Length", fmt.Sprintf("%d", len(fullData)))

				w.WriteTrailers(h)

			}
		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/video") {
			contents, err := os.ReadFile("./assets/vim.mp4")

			if err != nil {
				status = response.StatusCodeInternalServerError
				message = response.MessageInternalServerError
			} else {
				message = response.Message(contents)

				h.Set("Content-Type", "video/mp4")
			}

			h.Set("Content-Length", fmt.Sprintf("%d", len(message)))

			w.WriteStatusLine(status)
			w.WriteHeaders(h)
			w.WriteBody([]byte(message))
		} else {
			if strings.HasPrefix(req.RequestLine.RequestTarget, "/yourproblem") {
				status = response.StatusCodeBadRequest
				message = response.MessageBadRequest
			} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/myproblem") {
				status = response.StatusCodeInternalServerError
				message = response.MessageInternalServerError
			}

			h.Set("Content-Type", "text/html")
			h.Set("Content-Length", fmt.Sprintf("%d", len(message)))

			w.WriteStatusLine(status)
			w.WriteHeaders(h)
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
