package client

import (
	"fmt"
	"time"
)

// Type sends a text string to be typed character-by-character.
// Protocol: T <text>
func (c *Client) Type(text string) error {
	return c.send(cmdType(text))
}

// TypeSlow types text with a delay between keystrokes.
func (c *Client) TypeSlow(text string, delay time.Duration) error {
	for _, ch := range text {
		if err := c.send(cmdType(fmt.Sprintf("%c", ch))); err != nil {
			return err
		}
		time.Sleep(delay)
	}
	return nil
}

// Combo presses a key combination and releases it.
// Protocol: C <key1> <key2> ...
func (c *Client) Combo(keys ...string) error {
	return c.send(cmdCombo(keys))
}

// KeyDown presses and holds a key.
// Protocol: D <key>
func (c *Client) KeyDown(key string) error {
	return c.send(cmdKeyDown(key))
}

// KeyUp releases a held key.
// Protocol: U <key>
func (c *Client) KeyUp(key string) error {
	return c.send(cmdKeyUp(key))
}

// Reset releases all held keys and mouse buttons.
// Protocol: R
func (c *Client) Reset() error {
	return c.send(cmdReset())
}

// Enter is a convenience for Combo("enter").
func (c *Client) Enter() error {
	return c.Combo("enter")
}

// Tab is a convenience for Combo("tab").
func (c *Client) Tab() error {
	return c.Combo("tab")
}

// Escape is a convenience for Combo("escape").
func (c *Client) Escape() error {
	return c.Combo("escape")
}

// Space is a convenience for Combo("space").
func (c *Client) Space() error {
	return c.Combo("space")
}

// Backspace is a convenience for Combo("backspace").
func (c *Client) Backspace() error {
	return c.Combo("backspace")
}
