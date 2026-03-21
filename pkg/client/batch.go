package client

import "time"

// Batch accumulates commands for atomic execution.
type Batch struct {
	commands []batchCmd
}

type batchCmd struct {
	line  string
	delay time.Duration // client-side delay AFTER this command
}

// Batch executes multiple commands as a single buffered write.
// Commands within a batch are NOT coalesced.
func (c *Client) Batch(fn func(b *Batch)) error {
	b := &Batch{}
	fn(b)

	// Separate commands that need delays from pure batch
	var currentBatch []string
	for _, cmd := range b.commands {
		currentBatch = append(currentBatch, cmd.line)
		if cmd.delay > 0 {
			// Flush current batch, then sleep
			if err := c.sendBatch(currentBatch); err != nil {
				return err
			}
			currentBatch = nil
			time.Sleep(cmd.delay)
		}
	}
	// Flush remaining
	if len(currentBatch) > 0 {
		return c.sendBatch(currentBatch)
	}
	return nil
}

// Type adds a type command to the batch.
func (b *Batch) Type(text string) {
	b.commands = append(b.commands, batchCmd{line: cmdType(text)})
}

// Combo adds a key combo to the batch.
func (b *Batch) Combo(keys ...string) {
	b.commands = append(b.commands, batchCmd{line: cmdCombo(keys)})
}

// KeyDown adds a key-down to the batch.
func (b *Batch) KeyDown(key string) {
	b.commands = append(b.commands, batchCmd{line: cmdKeyDown(key)})
}

// KeyUp adds a key-up to the batch.
func (b *Batch) KeyUp(key string) {
	b.commands = append(b.commands, batchCmd{line: cmdKeyUp(key)})
}

// Reset adds a reset to the batch.
func (b *Batch) Reset() {
	b.commands = append(b.commands, batchCmd{line: cmdReset()})
}

// Enter adds an enter keypress to the batch.
func (b *Batch) Enter() {
	b.Combo("enter")
}

// MoveTo adds a mouse move to the batch.
func (b *Batch) MoveTo(x, y int32) {
	b.commands = append(b.commands, batchCmd{line: cmdMoveTo(x, y)})
}

// Click adds a left click to the batch.
func (b *Batch) Click(x, y int32) {
	b.commands = append(b.commands, batchCmd{line: cmdClick(x, y)})
}

// RightClick adds a right click to the batch.
func (b *Batch) RightClick(x, y int32) {
	b.commands = append(b.commands, batchCmd{line: cmdRightClick(x, y)})
}

// MiddleClick adds a middle click to the batch.
func (b *Batch) MiddleClick(x, y int32) {
	b.commands = append(b.commands, batchCmd{line: cmdMiddleClick(x, y)})
}

// MouseDown adds a mouse button press to the batch.
func (b *Batch) MouseDown(x, y int32, button string) {
	if button == "" {
		button = "left"
	}
	b.commands = append(b.commands, batchCmd{line: cmdMouseDown(x, y, button)})
}

// MouseUp adds a mouse button release to the batch.
func (b *Batch) MouseUp(button string) {
	if button == "" {
		button = "left"
	}
	b.commands = append(b.commands, batchCmd{line: cmdMouseUp(button)})
}

// Scroll adds a scroll event to the batch.
func (b *Batch) Scroll(dx, dy int32) {
	b.commands = append(b.commands, batchCmd{line: cmdScroll(dx, dy)})
}

// Sleep inserts a client-side delay between commands.
func (b *Batch) Sleep(d time.Duration) {
	if len(b.commands) > 0 {
		b.commands[len(b.commands)-1].delay = d
	}
}

// Raw adds an arbitrary protocol line to the batch.
func (b *Batch) Raw(line string) {
	b.commands = append(b.commands, batchCmd{line: line})
}

// Ping adds a heartbeat to the batch.
func (b *Batch) Ping() {
	b.commands = append(b.commands, batchCmd{line: cmdPing()})
}
