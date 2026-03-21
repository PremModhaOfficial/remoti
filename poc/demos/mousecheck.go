package main

import (
	"fmt"
	"time"

	"github.com/bendahl/uinput"
)

func main() {
	fmt.Println("Creating virtual touchpad...")
	tp, err := uinput.CreateTouchPad("/dev/uinput", []byte("Remoti Test TouchPad"), 0, 1920, 0, 1080)
	if err != nil {
		fmt.Printf("ERROR: Failed to create touchpad: %v\n", err)
		fmt.Println("Note: requires root privileges (sudo)")
		return
	}
	defer tp.Close()

	fmt.Println("TouchPad created successfully!")
	fmt.Println("Moving cursor to center (960, 540) in 2 seconds...")
	time.Sleep(2 * time.Second)

	err = tp.MoveTo(960, 540)
	if err != nil {
		fmt.Printf("ERROR: MoveTo failed: %v\n", err)
		return
	}
	fmt.Println("MoveTo(960, 540) — did the cursor move to center?")

	time.Sleep(1 * time.Second)

	fmt.Println("Moving to top-left (100, 100)...")
	err = tp.MoveTo(100, 100)
	if err != nil {
		fmt.Printf("ERROR: MoveTo failed: %v\n", err)
		return
	}
	fmt.Println("MoveTo(100, 100) — did the cursor move to top-left?")

	time.Sleep(1 * time.Second)

	fmt.Println("Moving to bottom-right (1800, 1000)...")
	err = tp.MoveTo(1800, 1000)
	if err != nil {
		fmt.Printf("ERROR: MoveTo failed: %v\n", err)
		return
	}
	fmt.Println("MoveTo(1800, 1000) — did the cursor move to bottom-right?")

	time.Sleep(1 * time.Second)

	fmt.Println("Testing left click at center...")
	tp.MoveTo(960, 540)
	time.Sleep(200 * time.Millisecond)
	err = tp.LeftClick()
	if err != nil {
		fmt.Printf("ERROR: LeftClick failed: %v\n", err)
		return
	}
	fmt.Println("LeftClick() — did a click register?")

	time.Sleep(1 * time.Second)

	fmt.Println("Testing right click...")
	err = tp.RightClick()
	if err != nil {
		fmt.Printf("ERROR: RightClick failed: %v\n", err)
		return
	}
	fmt.Println("RightClick() — did a context menu appear?")

	fmt.Println("\n=== ALL TESTS PASSED (no errors) ===")
	fmt.Println("If the cursor moved and clicks registered, uinput mouse works on Niri/Wayland!")
}
