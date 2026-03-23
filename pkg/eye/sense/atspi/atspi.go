package atspi

import (
	"context"
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
	"remoti/pkg/eye/sense"
)

const (
	busInterface    = "org.a11y.Bus"
	busPath         = "/org/a11y/bus"
	accessibleIface = "org.a11y.atspi.Accessible"
	componentIface  = "org.a11y.atspi.Component"
	registryBus     = "org.a11y.atspi.Registry"
	registryPath     = "/org/a11y/atspi/accessible/root"
	coordTypeWindow  = uint32(1) // relative to window, works on Wayland
)

// Source discovers UI elements via the AT-SPI accessibility tree.
type Source struct {
	conn    *dbus.Conn         // connection to a11y bus
	locator sense.WindowLocator // optional, for coordinate correction on Wayland
}

// SourceOption configures an AT-SPI Source.
type SourceOption func(*Source)

// WithLocator sets a WindowLocator for translating window-relative coordinates
// to absolute screen coordinates. Required on Wayland where AT-SPI
// COORD_TYPE_SCREEN returns (0,0).
func WithLocator(l sense.WindowLocator) SourceOption {
	return func(s *Source) { s.locator = l }
}

// New creates a new AT-SPI source. Returns nil error if D-Bus is not available.
func New(opts ...SourceOption) (*Source, error) {
	// Step 1: Connect to session bus
	sessionBus, err := dbus.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("atspi: session bus: %w", err)
	}

	// Step 2: Get a11y bus address
	obj := sessionBus.Object(busInterface, busPath)
	var addr string
	err = obj.Call(busInterface+".GetAddress", 0).Store(&addr)
	if err != nil {
		return nil, fmt.Errorf("atspi: get a11y bus address: %w", err)
	}

	// Step 3: Connect to a11y bus
	a11yConn, err := dbus.Connect(addr)
	if err != nil {
		return nil, fmt.Errorf("atspi: connect to a11y bus: %w", err)
	}

	s := &Source{conn: a11yConn}
	for _, opt := range opts {
		opt(s)
	}
	return s, nil
}

func (s *Source) Name() string  { return "atspi" }
func (s *Source) Priority() int { return 1 } // highest priority

func (s *Source) Available(ctx context.Context) bool {
	return s.conn != nil
}

func (s *Source) Find(ctx context.Context, q sense.Query) ([]sense.Match, error) {
	if s.conn == nil {
		return nil, fmt.Errorf("atspi: not connected")
	}

	var matches []sense.Match

	// Get the registry root's children (applications)
	apps, err := s.getChildren(registryBus, registryPath)
	if err != nil {
		return nil, fmt.Errorf("atspi: get apps: %w", err)
	}

	for _, app := range apps {
		if ctx.Err() != nil {
			return matches, ctx.Err()
		}

		appName, _ := s.getName(app.busName, app.path)

		// Filter by AppID if specified
		if q.AppID != "" && !strings.Contains(strings.ToLower(appName), strings.ToLower(q.AppID)) {
			continue
		}

		// Walk this application's children recursively
		s.walkTree(ctx, app.busName, app.path, appName, q, &matches, 0)
	}

	return matches, nil
}

type accessible struct {
	busName string
	path    dbus.ObjectPath
}

func (s *Source) getChildren(busName string, path dbus.ObjectPath) ([]accessible, error) {
	obj := s.conn.Object(busName, path)

	// ChildCount is a D-Bus PROPERTY, not a method
	variant, err := obj.GetProperty(accessibleIface + ".ChildCount")
	if err != nil {
		return nil, fmt.Errorf("get ChildCount property: %w", err)
	}
	count, ok := variant.Value().(int32)
	if !ok {
		return nil, fmt.Errorf("ChildCount is not int32: %T", variant.Value())
	}

	var children []accessible
	for i := int32(0); i < count; i++ {
		call := obj.Call(accessibleIface+".GetChildAtIndex", 0, i)
		if call.Err != nil {
			continue
		}
		childBus, childPath, ok := parseAccessibleRef(call.Body)
		if !ok {
			continue
		}
		children = append(children, accessible{busName: childBus, path: childPath})
	}
	return children, nil
}

func (s *Source) getName(busName string, path dbus.ObjectPath) (string, error) {
	obj := s.conn.Object(busName, path)
	variant, err := obj.GetProperty(accessibleIface + ".Name")
	if err != nil {
		return "", err
	}
	name, ok := variant.Value().(string)
	if !ok {
		return "", fmt.Errorf("name is not a string")
	}
	return name, nil
}

func (s *Source) getRole(busName string, path dbus.ObjectPath) (sense.Role, error) {
	obj := s.conn.Object(busName, path)
	var roleID uint32
	err := obj.Call(accessibleIface+".GetRole", 0).Store(&roleID)
	if err != nil {
		return sense.RoleAny, err
	}
	return mapRole(roleID), nil
}

func (s *Source) getExtents(busName string, path dbus.ObjectPath) (sense.Rect, error) {
	obj := s.conn.Object(busName, path)
	call := obj.Call(componentIface+".GetExtents", 0, coordTypeWindow)
	if call.Err != nil {
		return sense.Rect{}, call.Err
	}
	// GetExtents returns a struct (iiii) as a single body element
	if len(call.Body) == 1 {
		if slice, ok := call.Body[0].([]interface{}); ok && len(slice) == 4 {
			x, _ := toInt32(slice[0])
			y, _ := toInt32(slice[1])
			w, _ := toInt32(slice[2])
			h, _ := toInt32(slice[3])
			return sense.Rect{X: x, Y: y, Width: w, Height: h}, nil
		}
	}
	// Fallback: try direct store (some D-Bus impls return flat values)
	var x, y, w, h int32
	if err := call.Store(&x, &y, &w, &h); err != nil {
		return sense.Rect{}, err
	}
	return sense.Rect{X: x, Y: y, Width: w, Height: h}, nil
}

func toInt32(v interface{}) (int32, bool) {
	switch n := v.(type) {
	case int32:
		return n, true
	case int:
		return int32(n), true
	case int64:
		return int32(n), true
	case uint32:
		return int32(n), true
	default:
		return 0, false
	}
}

func (s *Source) walkTree(ctx context.Context, busName string, path dbus.ObjectPath, appName string, q sense.Query, matches *[]sense.Match, depth int) {
	if ctx.Err() != nil {
		return
	}
	// Limit recursion depth to prevent infinite loops
	if depth > 20 {
		return
	}

	name, _ := s.getName(busName, path)
	role, _ := s.getRole(busName, path)

	// Check if this element matches the query
	if matchesQuery(name, role, q) {
		bounds, _ := s.getExtents(busName, path)
		// Offset by window position if locator available (Wayland fix)
		if s.locator != nil && bounds.Width > 0 {
			wx, wy, err := s.locator.WindowPosition(ctx, appName, "")
			if err == nil {
				bounds.X += wx
				bounds.Y += wy
			}
		}
		*matches = append(*matches, sense.Match{
			Name:   name,
			Role:   role,
			AppID:  appName,
			Bounds: bounds,
			Source: "atspi",
		})

		// If we have enough matches, stop
		if q.MaxResults > 0 && len(*matches) >= q.MaxResults {
			return
		}
	}

	// Recurse into children
	children, err := s.getChildren(busName, path)
	if err != nil {
		return
	}
	for _, child := range children {
		if q.MaxResults > 0 && len(*matches) >= q.MaxResults {
			return
		}
		s.walkTree(ctx, child.busName, child.path, appName, q, matches, depth+1)
	}
}

func matchesQuery(name string, role sense.Role, q sense.Query) bool {
	// Name filter (case-insensitive substring)
	if q.Name != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(q.Name)) {
		return false
	}

	// Role filter
	if q.Role != sense.RoleAny && q.Role != role {
		return false
	}

	// Must have a name or match by role at minimum
	if q.Name == "" && q.Role == sense.RoleAny {
		return false
	}

	return true
}

// parseAccessibleRef extracts (busName, objectPath) from an AT-SPI D-Bus struct.
// GetChildAtIndex returns a single struct (so) containing [busName, objectPath].
func parseAccessibleRef(body []interface{}) (string, dbus.ObjectPath, bool) {
	if len(body) != 1 {
		return "", "", false
	}
	slice, ok := body[0].([]interface{})
	if !ok || len(slice) != 2 {
		return "", "", false
	}
	busName, ok1 := slice[0].(string)
	pathStr, ok2 := slice[1].(dbus.ObjectPath)
	if !ok1 || !ok2 {
		return "", "", false
	}
	return busName, pathStr, true
}

// mapRole maps AT-SPI role IDs to sense.Role values.
// See: https://docs.gtk.org/atspi2/enum.Role.html
func mapRole(id uint32) sense.Role {
	switch id {
	case 22: // ATSPI_ROLE_PUSH_BUTTON
		return sense.RoleButton
	case 60: // ATSPI_ROLE_TEXT
		return sense.RoleTextBox
	case 11: // ATSPI_ROLE_MENU
		return sense.RoleMenu
	case 12: // ATSPI_ROLE_MENU_ITEM
		return sense.RoleMenuItem
	case 33: // ATSPI_ROLE_FRAME (window)
		return sense.RoleWindow
	case 24: // ATSPI_ROLE_LABEL
		return sense.RoleLabel
	case 7: // ATSPI_ROLE_CHECK_BOX
		return sense.RoleCheckBox
	case 8: // ATSPI_ROLE_COMBO_BOX
		return sense.RoleComboBox
	case 36: // ATSPI_ROLE_SLIDER
		return sense.RoleSlider
	case 37: // ATSPI_ROLE_PAGE_TAB
		return sense.RoleTab
	case 57: // ATSPI_ROLE_TOOL_BAR
		return sense.RoleToolBar
	case 101: // ATSPI_ROLE_LINK
		return sense.RoleLink
	case 26: // ATSPI_ROLE_IMAGE
		return sense.RoleImage
	default:
		return sense.RoleAny
	}
}

// Close closes the D-Bus connection.
func (s *Source) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
