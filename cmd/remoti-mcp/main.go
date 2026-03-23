package main

import (
	"fmt"
	"log"
	"os"

	"remoti/pkg/mcp"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	addr := "127.0.0.1:8080"
	if a := os.Getenv("REMOTI_ADDR"); a != "" {
		addr = a
	}

	s, err := mcp.NewServer(addr)
	if err != nil {
		log.Fatalf("remoti-mcp: init failed: %v", err)
	}
	defer s.Close()

	log.Println("remoti-mcp: serving on stdio")
	if err := server.ServeStdio(s.MCPServer()); err != nil {
		fmt.Fprintf(os.Stderr, "remoti-mcp: %v\n", err)
		os.Exit(1)
	}
}
