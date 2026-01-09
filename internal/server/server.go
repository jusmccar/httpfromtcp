package server

import (
	"fmt"
	"net"
	"sync/atomic"
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

	response := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 13\r\n\r\nHello World!"
	conn.Write([]byte(response))
}
