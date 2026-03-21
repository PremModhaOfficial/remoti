package sense

import "context"

// Role represents an accessible UI element role.
type Role string

const (
	RoleButton   Role = "button"
	RoleTextBox  Role = "text"
	RoleMenu     Role = "menu"
	RoleMenuItem Role = "menu_item"
	RoleWindow   Role = "window"
	RoleLabel    Role = "label"
	RoleCheckBox Role = "check_box"
	RoleComboBox Role = "combo_box"
	RoleSlider   Role = "slider"
	RoleTab      Role = "tab"
	RoleToolBar  Role = "tool_bar"
	RoleLink     Role = "link"
	RoleImage    Role = "image"
	RoleAny      Role = ""
)

// Rect represents a bounding rectangle in screen coordinates.
type Rect struct {
	X, Y          int32
	Width, Height int32
}

// Center returns the center point of the rectangle.
func (r Rect) Center() (int32, int32) {
	return r.X + r.Width/2, r.Y + r.Height/2
}

// Contains reports whether point (px, py) is inside the rectangle.
func (r Rect) Contains(px, py int32) bool {
	return px >= r.X && px < r.X+r.Width && py >= r.Y && py < r.Y+r.Height
}

// Query describes what to search for.
type Query struct {
	Name       string // element name or label text (substring match)
	Role       Role   // accessible role filter (empty = any)
	AppID      string // application name or ID filter
	MaxResults int    // 0 = unlimited
}

// Match represents a found UI element from a source.
type Match struct {
	Name   string // element name/label
	Role   Role   // accessible role
	AppID  string // owning application
	Bounds Rect   // screen coordinates
	Source string // which source found this ("atspi", "niri")
}

// Source is a pluggable backend for finding UI elements.
type Source interface {
	// Name returns the source identifier (e.g. "atspi", "niri").
	Name() string

	// Available reports whether this source can operate in the current environment.
	Available(ctx context.Context) bool

	// Find searches for elements matching the query.
	Find(ctx context.Context, q Query) ([]Match, error)

	// Priority returns the source's default priority (lower = tried first).
	Priority() int
}
