package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
)

type Server struct {
	handler  Handler
	listener net.Listener
	closed   atomic.Bool
}

type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return nil, err
	}

	s := &Server{
		handler:  handler,
		listener: listener,
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

		if err != nil {
			if s.closed.Load() {
				return
			}

			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	writer := response.NewWriter(conn)
	headers := response.GetDefaultHeaders(0)
	req, err := request.RequestFromReader(conn)

	if err != nil {
		writer.WriteStatusLine(response.StatusCodeBadRequest)
		writer.WriteHeaders(headers)

		return
	}

	s.handler(&writer, req)
}
