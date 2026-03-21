package sense

import "context"

// WindowLocator provides window screen positions from the compositor.
// Different compositors implement this differently:
// - Niri: niri msg -j windows
// - Sway: swaymsg -t get_tree
// - GNOME: Mutter D-Bus introspect
// - X11: not needed (AT-SPI COORD_TYPE_SCREEN works)
type WindowLocator interface {
	// WindowPosition returns the top-left screen coordinates of a window
	// identified by app ID or title. Returns (0,0) if unknown.
	WindowPosition(ctx context.Context, appID string, title string) (x, y int32, err error)
}
