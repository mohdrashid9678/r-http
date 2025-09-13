// Package httpserver provides a high-performance, extensible HTTP server library.
package httpserver

import (
	"log"
	"net"
	"runtime/debug"

	"github.com/mohdrashid9678/rhttp/httperrors"
	"github.com/mohdrashid9678/rhttp/request"
	"github.com/mohdrashid9678/rhttp/response"
	"github.com/mohdrashid9678/rhttp/router"
)

// Server is the core of our HTTP server.
type Server struct {
	addr   string
	router *router.Router
}

// New creates a new Server instance, ready to be configured.
func New(addr string) *Server {
	return &Server{
		addr:   addr,
		router: router.New(),
	}
}

// AddRoute now uses the Handler type defined in the router package.
func (s *Server) AddRoute(method, path string, handler router.Handler) {
	s.router.AddRoute(method, path, handler)
}

// ListenAndServe starts the TCP listener and the main server loop.
func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

// handleConnection manages the entire lifecycle of a single client connection.
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	defer s.recoverFromPanic(conn)

	req, err := request.Parse(conn)
	if err != nil {
		s.handleError(conn, err)
		return
	}

	handler, params := s.router.FindHandler(req.Method, req.Target)
	req.PathParams = params

	var resp *response.Response
	if handler != nil {
		resp, err = handler(req)
	} else {
		err = httperrors.NewNotFound(req.Target)
	}

	if err != nil {
		s.handleError(conn, err)
		return
	}

	if err := resp.Write(conn); err != nil {
		log.Printf("error writing response: %v", err)
	}
}

// handleError centralizes error response logic.
func (s *Server) handleError(conn net.Conn, err error) {
	log.Printf("handler error: %v", err)
	resp, writeErr := response.Error(err)
	if writeErr != nil {
		log.Printf("could not create error response: %v", writeErr)
		return
	}
	if err := resp.Write(conn); err != nil {
		log.Printf("error sending error response: %v", err)
	}
}

// recoverFromPanic is a middleware to prevent a single request from crashing the server.
func (s *Server) recoverFromPanic(conn net.Conn) {
	if r := recover(); r != nil {
		log.Printf("panic recovered in handleConnection: %v\n%s", r, debug.Stack())
		s.handleError(conn, httperrors.NewInternalServerError("an unexpected error occurred"))
	}
}
