package mcp

import (
	"fmt"
	"strings"

	gomcp "github.com/mark3labs/mcp-go/mcp"
)

// splitCombo splits "ctrl+shift+a" into ["ctrl", "shift", "a"].
func splitCombo(keys string) []string {
	parts := strings.Split(keys, "+")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// requireCoords extracts x,y from a tool request.
func requireCoords(req gomcp.CallToolRequest) (int32, int32, error) {
	x, err := req.RequireFloat("x")
	if err != nil {
		return 0, 0, fmt.Errorf("x: %w", err)
	}
	y, err := req.RequireFloat("y")
	if err != nil {
		return 0, 0, fmt.Errorf("y: %w", err)
	}
	return int32(x), int32(y), nil
}
