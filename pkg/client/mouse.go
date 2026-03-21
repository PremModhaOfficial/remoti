package client

import "time"

// MoveTo moves the mouse cursor to absolute coordinates.
// Protocol: M M <x> <y>
func (c *Client) MoveTo(x, y int32) error {
	return c.send(cmdMoveTo(x, y))
}

// Click moves to (x,y) and performs a left click.
// Protocol: M C <x> <y>
func (c *Client) Click(x, y int32) error {
	return c.send(cmdClick(x, y))
}

// RightClick moves to (x,y) and performs a right click.
// Protocol: M R <x> <y>
func (c *Client) RightClick(x, y int32) error {
	return c.send(cmdRightClick(x, y))
}

// MiddleClick moves to (x,y) and performs a middle click.
// Protocol: M K <x> <y>
func (c *Client) MiddleClick(x, y int32) error {
	return c.send(cmdMiddleClick(x, y))
}

// MouseDown presses a mouse button at (x,y) without releasing.
// Button can be "left", "right", or "middle". Default: "left".
// Protocol: M D <x> <y> [button]
func (c *Client) MouseDown(x, y int32, button string) error {
	if button == "" {
		button = "left"
	}
	return c.send(cmdMouseDown(x, y, button))
}

// MouseUp releases a mouse button.
// Protocol: M U [button]
func (c *Client) MouseUp(button string) error {
	if button == "" {
		button = "left"
	}
	return c.send(cmdMouseUp(button))
}

// Scroll sends a scroll event.
// Protocol: M S <dx> <dy>
func (c *Client) Scroll(dx, dy int32) error {
	return c.send(cmdScroll(dx, dy))
}

// Drag performs a drag operation from (x1,y1) to (x2,y2).
// It holds the left mouse button, moves, then releases.
func (c *Client) Drag(x1, y1, x2, y2 int32) error {
	// Small delay between drag steps for uinput to register
	if err := c.send(cmdMouseDown(x1, y1, "left")); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)
	if err := c.send(cmdMoveTo(x2, y2)); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)
	return c.send(cmdMouseUp("left"))
}

// CtrlClick holds Ctrl, clicks at (x,y), releases Ctrl.
func (c *Client) CtrlClick(x, y int32) error {
	return c.sendBatch([]string{
		cmdKeyDown("ctrl"),
		cmdClick(x, y),
		cmdKeyUp("ctrl"),
	})
}

// ShiftClick holds Shift, clicks at (x,y), releases Shift.
func (c *Client) ShiftClick(x, y int32) error {
	return c.sendBatch([]string{
		cmdKeyDown("shift"),
		cmdClick(x, y),
		cmdKeyUp("shift"),
	})
}

// DoubleClick performs two rapid left clicks at (x,y).
func (c *Client) DoubleClick(x, y int32) error {
	cmd := cmdClick(x, y)
	if err := c.send(cmd); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return c.send(cmd)
}
