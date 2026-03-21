package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"remoti/pkg/eye"
	"remoti/pkg/eye/sense"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	addr := "127.0.0.1:8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	fmt.Println("=== Remoti Eye Demo ===")
	fmt.Printf("Connecting to remoti at %s...\n", addr)

	e, err := eye.Connect(addr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer e.Close()
	fmt.Println("Connected!")

	// Show available sources
	fmt.Printf("\nAvailable sources:\n")
	for _, src := range e.Finder().Sources() {
		avail := src.Available(ctx)
		fmt.Printf("  [%d] %s — available: %v\n", src.Priority(), src.Name(), avail)
	}

	// Try to find all windows
	fmt.Println("\n--- Finding all windows ---")
	windows, err := e.Find(ctx, sense.Query{Role: sense.RoleWindow})
	if err != nil {
		fmt.Printf("No windows found: %v\n", err)
	} else {
		for _, w := range windows {
			b := w.Bounds()
			fmt.Printf("  Window: %q [%s] app=%s bounds=(%d,%d %dx%d) via %s\n",
				w.Name(), w.Role(), w.AppID(), b.X, b.Y, b.Width, b.Height, w.Source())
		}
	}

	// Try to find buttons
	fmt.Println("\n--- Finding buttons ---")
	buttons, err := e.Find(ctx, sense.Query{Role: sense.RoleButton, MaxResults: 10})
	if err != nil {
		fmt.Printf("No buttons found: %v\n", err)
	} else {
		for _, btn := range buttons {
			b := btn.Bounds()
			fmt.Printf("  Button: %q bounds=(%d,%d %dx%d) via %s\n",
				btn.Name(), b.X, b.Y, b.Width, b.Height, btn.Source())
		}
	}

	// Try to find by name
	if len(os.Args) > 2 {
		name := os.Args[2]
		fmt.Printf("\n--- Finding element: %q ---\n", name)
		elem, err := e.FindOne(ctx, sense.Query{Name: name})
		if err != nil {
			fmt.Printf("Not found: %v\n", err)
		} else {
			b := elem.Bounds()
			cx, cy := b.Center()
			fmt.Printf("  Found: %q role=%s app=%s bounds=(%d,%d %dx%d) center=(%d,%d) via %s\n",
				elem.Name(), elem.Role(), elem.AppID(), b.X, b.Y, b.Width, b.Height, cx, cy, elem.Source())
			fmt.Println("  Clicking it...")
			if err := elem.Click(); err != nil {
				fmt.Printf("  Click failed: %v\n", err)
			} else {
				fmt.Println("  Clicked!")
			}
		}
	}

	fmt.Println("\n=== Demo Complete ===")
}
