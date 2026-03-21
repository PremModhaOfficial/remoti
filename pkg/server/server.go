package server

import (
	"fmt"
	"log"
	"net"

	"remoti/pkg/executor"
	"remoti/pkg/protocol"
)

// Server handles incoming TCP connections and passes them to the protocol parser.
type Server struct {
	addr     string
	executor executor.Executor
	parser   protocol.Parser
}

// NewServer creates a new high-performance TCP server.
func NewServer(addr string, exec executor.Executor, parser protocol.Parser) *Server {
	return &Server{
		addr:     addr,
		executor: exec,
		parser:   parser,
	}
}

// Start begins listening for TCP connections.
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}
	defer listener.Close()

	log.Printf("Remoti server listening on %s", s.addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.SetNoDelay(true)
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Client connected: %s", conn.RemoteAddr())

	// We wrap the connection in the protocol parser, passing the shared executor.
	err := s.parser.ParseStream(conn, s.executor)
	if err != nil {
		log.Printf("Connection closed with error: %v", err)
	} else {
		log.Printf("Client disconnected naturally: %s", conn.RemoteAddr())
	}

	// Reset executor state in case the client disconnected while holding keys down
	s.executor.Reset()
}
