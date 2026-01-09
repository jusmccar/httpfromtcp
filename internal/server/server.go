package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		return nil, err
	}

	s := &Server{
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

	headers := response.GetDefaultHeaders(0)

	err := response.WriteStatusLine(conn, response.StatusCodeOK)

	if err != nil {
		return
	}

	err = response.WriteHeaders(conn, headers)

	if err != nil {
		return
	}
}
