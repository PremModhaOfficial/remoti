package main

import (
	"fmt"
	"log"
	"time"

	"remoti/pkg/client"
)

func main() {
	c, err := client.Connect("127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	fmt.Println("Connected — executing at full speed")
	start := time.Now()

	// Step 1: sesh filter "nms" + enter (already in sesh picker from before)
	c.Type("nms")
	time.Sleep(300 * time.Millisecond)
	c.Enter()
	time.Sleep(2 * time.Second)
	fmt.Printf("  [%v] sesh → nms selected\n", time.Since(start).Round(time.Millisecond))

	// Step 2: Launch claude code
	c.Type("claude")
	c.Enter()
	time.Sleep(3 * time.Second)
	fmt.Printf("  [%v] claude launched\n", time.Since(start).Round(time.Millisecond))

	// Step 3: New tmux window (prefix=Ctrl+Space, then c)
	c.Combo("ctrl", "space")
	time.Sleep(200 * time.Millisecond)
	c.Type("c")
	time.Sleep(1 * time.Second)
	fmt.Printf("  [%v] new tmux window\n", time.Since(start).Round(time.Millisecond))

	// Step 4: Launch nvim
	c.Type("nvim .")
	c.Enter()
	time.Sleep(2 * time.Second)
	fmt.Printf("  [%v] nvim launched\n", time.Since(start).Round(time.Millisecond))

	total := time.Since(start)
	fmt.Printf("\nDone in %v — all via persistent TCP, zero screenshots\n", total.Round(time.Millisecond))
}
