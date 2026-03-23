package client

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Client provides a persistent connection to the remoti server.
// It is safe for concurrent use from multiple goroutines.
type Client struct {
	addr    string
	opts    Options
	conn    net.Conn
	writer  *bufio.Writer
	writeMu sync.Mutex
	closed  atomic.Bool
}

// Connect establishes a persistent connection to the remoti server.
func Connect(address string, opts ...Option) (*Client, error) {
	o := DefaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	conn, err := net.Dial(o.Network, address)
	if err != nil {
		return nil, fmt.Errorf("remoti: failed to connect to %s: %w", address, err)
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
	}

	return &Client{
		addr:   address,
		opts:   o,
		conn:   conn,
		writer: bufio.NewWriterSize(conn, 4096),
	}, nil
}

// InputClient defines the interface for sending input to the remoti server.
type InputClient interface {
	Type(text string) error
	TypeSlow(text string, delay time.Duration) error
	Combo(keys ...string) error
	KeyDown(key string) error
	KeyUp(key string) error
	Reset() error
	Enter() error
	Tab() error
	Escape() error
	Space() error
	Backspace() error
	MoveTo(x, y int32) error
	Click(x, y int32) error
	RightClick(x, y int32) error
	MiddleClick(x, y int32) error
	MouseDown(x, y int32, button string) error
	MouseUp(button string) error
	Scroll(dx, dy int32) error
	Drag(x1, y1, x2, y2 int32) error
	CtrlClick(x, y int32) error
	ShiftClick(x, y int32) error
	DoubleClick(x, y int32) error
	Batch(fn func(b *Batch)) error
	Ping() error
	Close() error
}

// Verify Client satisfies InputClient at compile time
var _ InputClient = (*Client)(nil)

// Close sends a Reset command and closes the connection.
func (c *Client) Close() error {
	if c.closed.Swap(true) {
		return nil // already closed
	}
	c.writeMu.Lock()
	c.writer.WriteString("R\n")
	c.writer.Flush()
	err := c.conn.Close()
	c.writeMu.Unlock()
	return err
}

// reconnect re-establishes the TCP connection.
func (c *Client) reconnect() error {
	conn, err := net.Dial(c.opts.Network, c.addr)
	if err != nil {
		return fmt.Errorf("remoti: reconnect failed: %w", err)
	}
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
	}
	c.conn = conn
	c.writer = bufio.NewWriterSize(conn, 4096)
	return nil
}

// send writes a protocol line to the server.
// It acquires the write mutex, writes the line + newline, and flushes.
// Automatically reconnects on broken pipe.
func (c *Client) send(line string) error {
	if c.closed.Load() {
		return fmt.Errorf("remoti: client is closed")
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_, err := c.writer.WriteString(line + "\n")
	if err == nil {
		err = c.writer.Flush()
	}
	if err != nil {
		// Try reconnect once
		if reconErr := c.reconnect(); reconErr != nil {
			return fmt.Errorf("remoti: write failed and reconnect failed: %w", err)
		}
		_, err = c.writer.WriteString(line + "\n")
		if err == nil {
			err = c.writer.Flush()
		}
		if err != nil {
			return fmt.Errorf("remoti: write failed after reconnect: %w", err)
		}
	}
	return nil
}

// sendBatch writes multiple protocol lines atomically (under one lock).
// Automatically reconnects on broken pipe, then retries the entire batch.
func (c *Client) sendBatch(lines []string) error {
	if c.closed.Load() {
		return fmt.Errorf("remoti: client is closed")
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	writeAll := func() error {
		for _, line := range lines {
			if _, err := c.writer.WriteString(line + "\n"); err != nil {
				return err
			}
		}
		return c.writer.Flush()
	}

	if err := writeAll(); err != nil {
		// Try reconnect once, then retry the whole batch
		if reconErr := c.reconnect(); reconErr != nil {
			return fmt.Errorf("remoti: batch write failed and reconnect failed: %w", err)
		}
		if err := writeAll(); err != nil {
			return fmt.Errorf("remoti: batch write failed after reconnect: %w", err)
		}
	}
	return nil
}

// Ping sends a no-op heartbeat to verify the connection is alive.
func (c *Client) Ping() error {
	return c.send("P")
}
