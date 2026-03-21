package main

import (
	"flag"
	"log"
	"os"

	"remoti/pkg/executor"
	"remoti/pkg/protocol"
	"remoti/pkg/server"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8080", "Address to listen on (e.g., :8080)")
	flag.Parse()

	// 1. Initialize Executor (Linux uinput)
	log.Println("Initializing uinput executor...")
	exec, err := executor.NewLinuxExecutor()
	if err != nil {
		log.Fatalf("Failed to initialize executor: %v\nNote: uinput requires root privileges. Try running with sudo.", err)
		os.Exit(1)
	}
	defer exec.Close()

	// 2. Initialize Parser
	parser := protocol.NewTextParser()

	// 3. Initialize and Start Server
	srv := server.NewServer(addr, exec, parser)
	log.Printf("Starting TCP server on %s", addr)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
