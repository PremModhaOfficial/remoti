package main

import (
	"flag"
	"log"

	"remoti/pkg/executor"
	"remoti/pkg/protocol"
	"remoti/pkg/server"
)

func main() {
	var addr string
	var screenWidth, screenHeight int
	flag.StringVar(&addr, "addr", "127.0.0.1:8080", "Address to listen on (e.g., 127.0.0.1:8080)")
	flag.IntVar(&screenWidth, "screen-width", 1920, "Screen width in pixels for mouse input")
	flag.IntVar(&screenHeight, "screen-height", 1080, "Screen height in pixels for mouse input")
	flag.Parse()

	// 1. Initialize Executor (Linux uinput)
	log.Println("Initializing uinput executor...")
	exec, err := executor.NewLinuxExecutor(int32(screenWidth), int32(screenHeight))
	if err != nil {
		log.Fatalf("Failed to initialize executor: %v\nNote: uinput requires root privileges. Try running with sudo.", err)
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
