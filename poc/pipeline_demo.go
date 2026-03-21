package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"remoti/pkg/client"
)

// step runs an action then waits for a condition (polling niri/tmux)
type step struct {
	name   string
	action func(c *client.Client) error
	verify func(ctx context.Context) bool // return true when ready
	delay  time.Duration                  // min delay after action
}

func main() {
	c, err := client.Connect("127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	fmt.Println("=== Pipeline Demo: Full Speed with Verification ===\n")
	start := time.Now()

	steps := []step{
		{
			name: "Launch Alacritty",
			action: func(c *client.Client) error {
				return exec.Command("niri", "msg", "action", "spawn", "--", "alacritty").Run()
			},
			verify: func(ctx context.Context) bool {
				out, _ := exec.CommandContext(ctx, "niri", "msg", "-j", "windows").Output()
				return strings.Contains(string(out), "Alacritty")
			},
			delay: 500 * time.Millisecond,
		},
		{
			name: "Focus Alacritty",
			action: func(c *client.Client) error {
				out, _ := exec.Command("niri", "msg", "-j", "windows").Output()
				// Find Alacritty window ID
				for _, line := range strings.Split(string(out), "\n") {
					if strings.Contains(line, "Alacritty") {
						// Extract ID from JSON — simple parse
					}
				}
				return exec.Command("bash", "poc/send.sh", "focus", "Alacritty").Run()
			},
			verify: func(ctx context.Context) bool { return true },
			delay:  300 * time.Millisecond,
		},
		{
			name: "tmux attach",
			action: func(c *client.Client) error {
				c.Type("tmux a")
				return c.Enter()
			},
			verify: func(ctx context.Context) bool {
				// Check if tmux is running
				out, _ := exec.CommandContext(ctx, "tmux", "list-sessions").Output()
				return len(out) > 0
			},
			delay: 500 * time.Millisecond,
		},
		{
			name: "Open sesh picker (prefix+o)",
			action: func(c *client.Client) error {
				c.Combo("ctrl", "space")
				time.Sleep(200 * time.Millisecond)
				return c.Type("o")
			},
			verify: func(ctx context.Context) bool {
				// sesh picker opens as fzf — we can't easily verify, just wait
				time.Sleep(500 * time.Millisecond)
				return true
			},
			delay: 500 * time.Millisecond,
		},
		{
			name: "Select NMS Lite project",
			action: func(c *client.Client) error {
				c.Type("nms")
				time.Sleep(300 * time.Millisecond)
				return c.Enter()
			},
			verify: func(ctx context.Context) bool {
				// Verify we're in the right tmux session
				out, _ := exec.CommandContext(ctx, "tmux", "display-message", "-p", "#S").Output()
				return strings.Contains(strings.ToLower(string(out)), "nms")
			},
			delay: 1 * time.Second,
		},
		{
			name: "Launch Claude Code",
			action: func(c *client.Client) error {
				c.Type("claude")
				return c.Enter()
			},
			verify: func(ctx context.Context) bool {
				// Wait for claude to start by checking process
				time.Sleep(2 * time.Second)
				out, _ := exec.CommandContext(ctx, "pgrep", "-f", "claude").Output()
				return len(out) > 0
			},
			delay: 1 * time.Second,
		},
		{
			name: "New tmux window (prefix+c)",
			action: func(c *client.Client) error {
				c.Combo("ctrl", "space")
				time.Sleep(200 * time.Millisecond)
				return c.Type("c")
			},
			verify: func(ctx context.Context) bool { return true },
			delay: 500 * time.Millisecond,
		},
		{
			name: "Launch Neovim",
			action: func(c *client.Client) error {
				c.Type("nvim .")
				return c.Enter()
			},
			verify: func(ctx context.Context) bool {
				time.Sleep(1 * time.Second)
				return true
			},
			delay: 500 * time.Millisecond,
		},
	}

	ctx := context.Background()

	for i, s := range steps {
		stepStart := time.Now()

		// Execute action
		if err := s.action(c); err != nil {
			fmt.Printf("  [FAIL] Step %d: %s — %v\n", i+1, s.name, err)
			continue
		}

		// Wait minimum delay
		time.Sleep(s.delay)

		// Verify (poll up to 5s)
		verified := false
		for attempt := 0; attempt < 10; attempt++ {
			if s.verify(ctx) {
				verified = true
				break
			}
			time.Sleep(500 * time.Millisecond)
		}

		elapsed := time.Since(stepStart).Round(time.Millisecond)
		status := "OK"
		if !verified {
			status = "UNVERIFIED"
		}
		fmt.Printf("  [%s] Step %d: %s (%v)\n", status, i+1, s.name, elapsed)
	}

	total := time.Since(start).Round(time.Millisecond)
	fmt.Printf("\n=== Pipeline complete in %v ===\n", total)
}
