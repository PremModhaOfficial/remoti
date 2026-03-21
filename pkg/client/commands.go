package client

import (
	"fmt"
	"strings"
)

// Protocol command formatters — single source of truth for wire format.

func cmdType(text string) string        { return "T " + text }
func cmdCombo(keys []string) string     { return "C " + strings.Join(keys, " ") }
func cmdKeyDown(key string) string      { return "D " + key }
func cmdKeyUp(key string) string        { return "U " + key }
func cmdReset() string                  { return "R" }
func cmdPing() string                   { return "P" }
func cmdMoveTo(x, y int32) string      { return fmt.Sprintf("M M %d %d", x, y) }
func cmdClick(x, y int32) string       { return fmt.Sprintf("M C %d %d", x, y) }
func cmdRightClick(x, y int32) string  { return fmt.Sprintf("M R %d %d", x, y) }
func cmdMiddleClick(x, y int32) string { return fmt.Sprintf("M K %d %d", x, y) }
func cmdMouseDown(x, y int32, button string) string {
	return fmt.Sprintf("M D %d %d %s", x, y, button)
}
func cmdMouseUp(button string) string { return "M U " + button }
func cmdScroll(dx, dy int32) string   { return fmt.Sprintf("M S %d %d", dx, dy) }
