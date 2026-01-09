package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
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

		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			status = response.StatusCodeBadRequest
			message = response.MessageBadRequest
		case "/myproblem":
			status = response.StatusCodeInternalServerError
			message = response.MessageInternalServerError
		}

		headers := response.GetDefaultHeaders(0)
		headers.Set("Content-Type", "text/html")
		headers.Set("Content-Length", fmt.Sprintf("%d", len(message)))

		w.WriteStatusLine(status)
		w.WriteHeaders(headers)
		w.WriteBody([]byte(message))
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
