package eye

import (
	"remoti/pkg/eye/sense"
)

// Element represents a found UI element that can be acted upon.
type Element struct {
	match  sense.Match
	client *Eye
}

// Name returns the element's accessible name.
func (e *Element) Name() string { return e.match.Name }

// Role returns the element's accessible role.
func (e *Element) Role() sense.Role { return e.match.Role }

// AppID returns the owning application.
func (e *Element) AppID() string { return e.match.AppID }

// Bounds returns the element's screen rectangle.
func (e *Element) Bounds() sense.Rect { return e.match.Bounds }

// Source returns which sensing source found this element.
func (e *Element) Source() string { return e.match.Source }

// Click clicks the center of this element.
func (e *Element) Click() error {
	x, y := e.match.Bounds.Center()
	return e.client.act.Click(x, y)
}

// RightClick right-clicks the center of this element.
func (e *Element) RightClick() error {
	x, y := e.match.Bounds.Center()
	return e.client.act.RightClick(x, y)
}

// DoubleClick double-clicks the center of this element.
func (e *Element) DoubleClick() error {
	x, y := e.match.Bounds.Center()
	return e.client.act.DoubleClick(x, y)
}

// Type clicks this element to focus it, then types the text.
func (e *Element) Type(text string) error {
	if err := e.Click(); err != nil {
		return err
	}
	return e.client.act.Type(text)
}

// Hover moves the mouse to the center of this element without clicking.
func (e *Element) Hover() error {
	x, y := e.match.Bounds.Center()
	return e.client.act.MoveTo(x, y)
}

// DragTo drags from this element's center to the target element's center.
func (e *Element) DragTo(target *Element) error {
	x1, y1 := e.match.Bounds.Center()
	x2, y2 := target.match.Bounds.Center()
	return e.client.act.Drag(x1, y1, x2, y2)
}
