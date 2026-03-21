package client

import (
	"bufio"
	"net"
	"sync"
	"testing"
	"time"
)

// --------------------------------------------------------------------------
// Mock server
// --------------------------------------------------------------------------

type mockServer struct {
	addr  string
	ln    net.Listener
	lines []string
	mu    sync.Mutex
	done  chan struct{}
}

func startMockServer(t *testing.T) *mockServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("startMockServer: %v", err)
	}
	ms := &mockServer{
		addr: ln.Addr().String(),
		ln:   ln,
		done: make(chan struct{}),
	}
	go ms.serve()
	return ms
}

func (ms *mockServer) serve() {
	defer close(ms.done)
	conn, err := ms.ln.Accept()
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		ms.mu.Lock()
		ms.lines = append(ms.lines, scanner.Text())
		ms.mu.Unlock()
	}
	conn.Close()
}

func (ms *mockServer) Lines() []string {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	cp := make([]string, len(ms.lines))
	copy(cp, ms.lines)
	return cp
}

// WaitForN blocks until at least n lines are received or timeout elapses.
func (ms *mockServer) WaitForN(n int, timeout time.Duration) []string {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if lines := ms.Lines(); len(lines) >= n {
			return lines
		}
		time.Sleep(5 * time.Millisecond)
	}
	return ms.Lines()
}

func (ms *mockServer) Close() {
	ms.ln.Close()
	<-ms.done
}

// --------------------------------------------------------------------------
// Test helpers
// --------------------------------------------------------------------------

func mustConnect(t *testing.T, addr string) *Client {
	t.Helper()
	c, err := Connect(addr)
	if err != nil {
		t.Fatalf("Connect(%q): %v", addr, err)
	}
	return c
}

// assertLines checks that got[i] == want[i] for every want entry.
// It requires len(got) >= len(want).
func assertLines(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) < len(want) {
		t.Fatalf("want at least %d lines, got %d: %v", len(want), len(got), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("line[%d]: want %q, got %q (all lines: %v)", i, w, got[i], got)
		}
	}
}

// --------------------------------------------------------------------------
// Keyboard tests
// --------------------------------------------------------------------------

func TestKeyboard_Type_hello(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Type("hello"); err != nil {
		t.Fatalf("Type: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "T hello")
}

func TestKeyboard_Type_empty_sends_T_space(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Type(""); err != nil {
		t.Fatalf("Type: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "T ")
}

func TestKeyboard_Combo_ctrl_c(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Combo("ctrl", "c"); err != nil {
		t.Fatalf("Combo: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "C ctrl c")
}

func TestKeyboard_Combo_meta_space(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Combo("meta", "space"); err != nil {
		t.Fatalf("Combo: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "C meta space")
}

func TestKeyboard_KeyDown_shift(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.KeyDown("shift"); err != nil {
		t.Fatalf("KeyDown: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "D shift")
}

func TestKeyboard_KeyUp_shift(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.KeyUp("shift"); err != nil {
		t.Fatalf("KeyUp: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "U shift")
}

func TestKeyboard_Reset(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "R")
}

func TestKeyboard_Enter(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Enter(); err != nil {
		t.Fatalf("Enter: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "C enter")
}

func TestKeyboard_Tab(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Tab(); err != nil {
		t.Fatalf("Tab: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "C tab")
}

func TestKeyboard_Escape(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Escape(); err != nil {
		t.Fatalf("Escape: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "C escape")
}

func TestKeyboard_Space(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Space(); err != nil {
		t.Fatalf("Space: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "C space")
}

func TestKeyboard_Backspace(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Backspace(); err != nil {
		t.Fatalf("Backspace: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "C backspace")
}

func TestKeyboard_TypeSlow_sendsOneLinePerCharacter(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.TypeSlow("ab", 10*time.Millisecond); err != nil {
		t.Fatalf("TypeSlow: %v", err)
	}
	assertLines(t, ms.WaitForN(2, time.Second), "T a", "T b")
}

// --------------------------------------------------------------------------
// Mouse tests
// --------------------------------------------------------------------------

func TestMouse_Click(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Click(500, 300); err != nil {
		t.Fatalf("Click: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "M C 500 300")
}

func TestMouse_RightClick(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.RightClick(100, 200); err != nil {
		t.Fatalf("RightClick: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "M R 100 200")
}

func TestMouse_MiddleClick(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.MiddleClick(400, 300); err != nil {
		t.Fatalf("MiddleClick: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "M K 400 300")
}

func TestMouse_MoveTo(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.MoveTo(960, 540); err != nil {
		t.Fatalf("MoveTo: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "M M 960 540")
}

func TestMouse_MouseDown_left(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.MouseDown(100, 100, "left"); err != nil {
		t.Fatalf("MouseDown: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "M D 100 100 left")
}

func TestMouse_MouseDown_emptyButton_defaultsToLeft(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.MouseDown(50, 60, ""); err != nil {
		t.Fatalf("MouseDown: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "M D 50 60 left")
}

func TestMouse_MouseUp_right(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.MouseUp("right"); err != nil {
		t.Fatalf("MouseUp: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "M U right")
}

func TestMouse_MouseUp_emptyButton_defaultsToLeft(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.MouseUp(""); err != nil {
		t.Fatalf("MouseUp: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "M U left")
}

func TestMouse_Scroll(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Scroll(0, -3); err != nil {
		t.Fatalf("Scroll: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "M S 0 -3")
}

func TestMouse_DoubleClick_sendsTwoClicks(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.DoubleClick(500, 300); err != nil {
		t.Fatalf("DoubleClick: %v", err)
	}
	// DoubleClick has a 50 ms sleep between the two sends.
	lines := ms.WaitForN(2, 500*time.Millisecond)
	if len(lines) < 2 {
		t.Fatalf("DoubleClick: expected 2 lines, got %d: %v", len(lines), lines)
	}
	assertLines(t, lines, "M C 500 300", "M C 500 300")
}

func TestMouse_CtrlClick_sendsThreeLinesAtomically(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.CtrlClick(500, 300); err != nil {
		t.Fatalf("CtrlClick: %v", err)
	}
	assertLines(t, ms.WaitForN(3, time.Second), "D ctrl", "M C 500 300", "U ctrl")
}

func TestMouse_ShiftClick_sendsThreeLinesAtomically(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.ShiftClick(500, 300); err != nil {
		t.Fatalf("ShiftClick: %v", err)
	}
	assertLines(t, ms.WaitForN(3, time.Second), "D shift", "M C 500 300", "U shift")
}

func TestMouse_Click_negativeCoords(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Click(-100, -200); err != nil {
		t.Fatalf("Click: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "M C -100 -200")
}

func TestMouse_Drag_sendsMouseDownMoveTo_MouseUp(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Drag(10, 20, 100, 200); err != nil {
		t.Fatalf("Drag: %v", err)
	}
	// Drag uses individual sends with 10 ms sleeps, not sendBatch.
	lines := ms.WaitForN(3, 500*time.Millisecond)
	assertLines(t, lines, "M D 10 20 left", "M M 100 200", "M U left")
}

// --------------------------------------------------------------------------
// Batch tests
// --------------------------------------------------------------------------

func TestBatch_TypeAndEnter(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Batch(func(b *Batch) {
		b.Type("hello")
		b.Enter()
	}); err != nil {
		t.Fatalf("Batch: %v", err)
	}
	assertLines(t, ms.WaitForN(2, time.Second), "T hello", "C enter")
}

func TestBatch_KeyDownClickKeyUp(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Batch(func(b *Batch) {
		b.KeyDown("ctrl")
		b.Click(200, 300)
		b.KeyUp("ctrl")
	}); err != nil {
		t.Fatalf("Batch: %v", err)
	}
	assertLines(t, ms.WaitForN(3, time.Second), "D ctrl", "M C 200 300", "U ctrl")
}

func TestBatch_Raw_sendsArbitraryLine(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Batch(func(b *Batch) {
		b.Raw("custom command")
	}); err != nil {
		t.Fatalf("Batch: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "custom command")
}

func TestBatch_Empty_sendsNothing(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Batch(func(b *Batch) {}); err != nil {
		t.Fatalf("Batch (empty): %v", err)
	}
	// Send a known marker to confirm the empty batch produced no lines.
	if err := c.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	lines := ms.WaitForN(1, time.Second)
	if len(lines) == 0 || lines[0] != "P" {
		t.Errorf("expected first line to be P after empty batch, got %v", lines)
	}
}

func TestBatch_Sleep_flushesChunks(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Batch(func(b *Batch) {
		b.Type("first")
		b.Sleep(20 * time.Millisecond)
		b.Type("second")
	}); err != nil {
		t.Fatalf("Batch: %v", err)
	}
	assertLines(t, ms.WaitForN(2, time.Second), "T first", "T second")
}

// --------------------------------------------------------------------------
// Connection tests
// --------------------------------------------------------------------------

func TestConnect_validAddress_succeeds(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()

	c, err := Connect(ms.addr)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	c.Close()
}

func TestConnect_invalidAddress_returnsError(t *testing.T) {
	// Bind then immediately close to get a port we know won't accept.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("could not allocate port for test")
	}
	addr := ln.Addr().String()
	ln.Close()

	_, err = Connect(addr)
	if err == nil {
		t.Fatal("Connect to closed port: expected error, got nil")
	}
}

func TestClose_sendsResetToServer(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)

	if err := c.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	lines := ms.WaitForN(1, time.Second)
	if len(lines) == 0 || lines[0] != "R" {
		t.Errorf("Close: expected first line to be R, got %v", lines)
	}
}

func TestClose_idempotent(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)

	if err := c.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestPing(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)
	defer c.Close()

	if err := c.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	assertLines(t, ms.WaitForN(1, time.Second), "P")
}

func TestOperationsAfterClose_returnError(t *testing.T) {
	ms := startMockServer(t)
	defer ms.Close()
	c := mustConnect(t, ms.addr)

	if err := c.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	ops := []struct {
		name string
		fn   func() error
	}{
		{"Type", func() error { return c.Type("x") }},
		{"Combo", func() error { return c.Combo("ctrl", "c") }},
		{"KeyDown", func() error { return c.KeyDown("shift") }},
		{"KeyUp", func() error { return c.KeyUp("shift") }},
		{"Reset", func() error { return c.Reset() }},
		{"Click", func() error { return c.Click(0, 0) }},
		{"RightClick", func() error { return c.RightClick(0, 0) }},
		{"Ping", func() error { return c.Ping() }},
		{"Batch", func() error {
			return c.Batch(func(b *Batch) { b.Type("x") })
		}},
	}
	for _, op := range ops {
		if err := op.fn(); err == nil {
			t.Errorf("%s after Close: expected error, got nil", op.name)
		}
	}
}

// --------------------------------------------------------------------------
// InputClient interface compliance (compile-time check documented as a test)
// --------------------------------------------------------------------------

func TestInputClient_compiletimeInterfaceCheck(t *testing.T) {
	// The real check is at package compile time via:
	//   var _ InputClient = (*Client)(nil)
	// in client.go. This test documents and surfaces that constraint.
	var _ InputClient = (*Client)(nil)
}
