package niri

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"

	"remoti/pkg/eye/sense"
)

// Source discovers windows via Niri compositor IPC.
type Source struct{}

// New creates a new Niri source.
func New() *Source { return &Source{} }

func (s *Source) Name() string  { return "niri" }
func (s *Source) Priority() int { return 10 }

func (s *Source) Available(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "niri", "msg", "version")
	return cmd.Run() == nil
}

// niriWindow represents the JSON structure from `niri msg -j windows`.
type niriWindow struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	AppID       string `json:"app_id"`
	WorkspaceID int    `json:"workspace_id"`
	IsFocused   bool   `json:"is_focused"`
	IsFloating  *bool  `json:"is_floating"`
}

// niriOutput represents the JSON structure from `niri msg -j outputs`.
type niriOutput struct {
	Name     string `json:"name"`
	LogicalX int32  `json:"x"`
	LogicalY int32  `json:"y"`
	// CurrentMode contains the current resolution info.
	CurrentMode *struct {
		Width  int32 `json:"width"`
		Height int32 `json:"height"`
	} `json:"current_mode"`
}

// niriFocusedWindow is the response from `niri msg -j focused-window`.
type niriFocusedWindow struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	AppID string `json:"app_id"`
	Size  struct {
		W int32 `json:"w"`
		H int32 `json:"h"`
	} `json:"size"`
}

func (s *Source) Find(ctx context.Context, q sense.Query) ([]sense.Match, error) {
	cmd := exec.CommandContext(ctx, "niri", "msg", "-j", "windows")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var windows []niriWindow
	if err := json.Unmarshal(out, &windows); err != nil {
		return nil, err
	}

	var matches []sense.Match
	for _, w := range windows {
		// Filter by role — niri only has windows
		if q.Role != sense.RoleAny && q.Role != sense.RoleWindow {
			continue
		}

		// Filter by name (case-insensitive substring)
		if q.Name != "" && !strings.Contains(strings.ToLower(w.Title), strings.ToLower(q.Name)) {
			continue
		}

		// Filter by app ID
		if q.AppID != "" && !strings.Contains(strings.ToLower(w.AppID), strings.ToLower(q.AppID)) {
			continue
		}

		bounds := getWindowBounds(ctx, w.ID)

		matches = append(matches, sense.Match{
			Name:   w.Title,
			Role:   sense.RoleWindow,
			AppID:  w.AppID,
			Bounds: bounds,
			Source: "niri",
		})
	}

	return matches, nil
}

// WindowPosition returns the screen position of a window by app ID or title.
// Implements sense.WindowLocator.
//
// On Niri (scrolling tiler), exact window positions require layout calculation.
// This uses a focused-window heuristic: if the target app is focused, we can
// determine its position from the output geometry. Returns (0,0) for unfocused
// windows as a safe fallback.
func (s *Source) WindowPosition(ctx context.Context, appID string, title string) (int32, int32, error) {
	// Get the focused window to check if it matches
	cmd := exec.CommandContext(ctx, "niri", "msg", "-j", "focused-window")
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	var focused niriFocusedWindow
	if err := json.Unmarshal(out, &focused); err != nil {
		return 0, 0, err
	}

	// Check if the focused window matches the requested app
	appMatch := appID != "" && strings.EqualFold(focused.AppID, appID)
	titleMatch := title != "" && strings.Contains(strings.ToLower(focused.Title), strings.ToLower(title))
	if !appMatch && !titleMatch {
		return 0, 0, nil
	}

	// Get output geometry to determine where the window is on screen.
	// Niri typically places the focused window at the output's logical position.
	cmd = exec.CommandContext(ctx, "niri", "msg", "-j", "outputs")
	out, err = cmd.Output()
	if err != nil {
		return 0, 0, nil
	}

	var outputs []niriOutput
	if err := json.Unmarshal(out, &outputs); err != nil {
		return 0, 0, nil
	}

	// Use the first output's position as the window origin.
	// For multi-monitor, a more sophisticated approach would match
	// the window's workspace to the correct output.
	if len(outputs) > 0 {
		return outputs[0].LogicalX, outputs[0].LogicalY, nil
	}

	return 0, 0, nil
}

// getWindowBounds gets window geometry via niri msg -j focused-window
// or calculates from output geometry. For now, uses a heuristic based
// on the focused window data.
func getWindowBounds(ctx context.Context, windowID int) sense.Rect {
	// Try to get focused window details
	cmd := exec.CommandContext(ctx, "niri", "msg", "-j", "focused-window")
	out, err := cmd.Output()
	if err != nil {
		return sense.Rect{}
	}

	var focused struct {
		ID   int `json:"id"`
		Size struct {
			W int32 `json:"w"`
			H int32 `json:"h"`
		} `json:"size"`
	}
	if err := json.Unmarshal(out, &focused); err != nil {
		return sense.Rect{}
	}

	// If this is the focused window, we have its size
	// Position calculation is approximate for tiled layouts
	if focused.ID == windowID {
		return sense.Rect{
			Width:  focused.Size.W,
			Height: focused.Size.H,
		}
	}

	return sense.Rect{}
}
