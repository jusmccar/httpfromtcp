package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	handler  Handler
	closed   atomic.Bool
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func Serve(port int, handler Handler) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: listener,
		handler:  handler,
	}

	go s.listen()

	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)

	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()

		if s.closed.Load() {
			return
		}

		if err != nil {
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)

	if err != nil {
		writeHandlerError(conn, &HandlerError{
			StatusCode: response.StatusCodeBadRequest,
			Message:    "Bad Request\n",
		})

		return
	}

	var w bytes.Buffer
	handlerError := s.handler(&w, req)

	if handlerError != nil {
		writeHandlerError(conn, handlerError)

		return
	}

	err = response.WriteStatusLine(conn, response.StatusCodeOK)

	if err != nil {
		return
	}

	headers := response.GetDefaultHeaders(w.Len())
	err = response.WriteHeaders(conn, headers)

	if err != nil {
		return
	}

	fmt.Fprintf(conn, "%s", w.Bytes())
}

func writeHandlerError(w io.Writer, handlerError *HandlerError) {
	headers := response.GetDefaultHeaders(len(handlerError.Message))

	response.WriteStatusLine(w, handlerError.StatusCode)
	response.WriteHeaders(w, headers)

	fmt.Fprintf(w, "%s", handlerError.Message)
}
