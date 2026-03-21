package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"remoti/pkg/client"
)

func main() {
	c, err := client.Connect("127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	start := time.Now()
	ms := func() string { return fmt.Sprintf("[%dms]", time.Since(start).Milliseconds()) }

	// Skip steps based on args
	skipTmux := false
	for _, arg := range os.Args[1:] {
		if arg == "--skip-tmux" {
			skipTmux = true
		}
	}

	fmt.Println("remoti-eye speed demo")

	if !skipTmux {
		// 1. sesh picker: prefix+o (Ctrl+Space, then o)
		fmt.Printf("%s prefix+o (sesh picker)\n", ms())
		c.Combo("ctrl", "space")
		time.Sleep(150 * time.Millisecond)
		c.Type("o")
		time.Sleep(800 * time.Millisecond)

		// 2. Filter for nms + select
		fmt.Printf("%s filter 'nms' + enter\n", ms())
		c.Type("nms")
		time.Sleep(300 * time.Millisecond)
		c.Enter()
		time.Sleep(1500 * time.Millisecond)
	}

	// 3. Launch claude code
	fmt.Printf("%s launch claude\n", ms())
	c.Type("claude")
	c.Enter()
	time.Sleep(3 * time.Second)

	// 4. New tmux window: prefix+c
	fmt.Printf("%s prefix+c (new window)\n", ms())
	c.Combo("ctrl", "space")
	time.Sleep(150 * time.Millisecond)
	c.Type("c")
	time.Sleep(500 * time.Millisecond)

	// 5. Launch neovim
	fmt.Printf("%s nvim .\n", ms())
	c.Type("nvim .")
	c.Enter()

	fmt.Printf("%s DONE — 1 persistent TCP conn, 0 focus switches\n", ms())
}
