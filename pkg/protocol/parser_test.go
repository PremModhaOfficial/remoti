package protocol

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	"remoti/pkg/executor"
)

// mockExecutor implements both executor.Executor and executor.MouseExecutor.
type mockExecutor struct {
	calls []string
}

func (m *mockExecutor) Type(text string) error {
	m.calls = append(m.calls, "Type:"+text)
	return nil
}

func (m *mockExecutor) Combo(keys ...string) error {
	m.calls = append(m.calls, "Combo:"+strings.Join(keys, ","))
	return nil
}

func (m *mockExecutor) KeyDown(key string) error {
	m.calls = append(m.calls, "KeyDown:"+key)
	return nil
}

func (m *mockExecutor) KeyUp(key string) error {
	m.calls = append(m.calls, "KeyUp:"+key)
	return nil
}

func (m *mockExecutor) Reset() error {
	m.calls = append(m.calls, "Reset")
	return nil
}

func (m *mockExecutor) Close() error {
	m.calls = append(m.calls, "Close")
	return nil
}

func (m *mockExecutor) MoveTo(x, y int32) error {
	m.calls = append(m.calls, fmt.Sprintf("MoveTo:%d,%d", x, y))
	return nil
}

func (m *mockExecutor) LeftClick(x, y int32) error {
	m.calls = append(m.calls, fmt.Sprintf("LeftClick:%d,%d", x, y))
	return nil
}

func (m *mockExecutor) RightClick(x, y int32) error {
	m.calls = append(m.calls, fmt.Sprintf("RightClick:%d,%d", x, y))
	return nil
}

func (m *mockExecutor) MiddleClick(x, y int32) error {
	m.calls = append(m.calls, fmt.Sprintf("MiddleClick:%d,%d", x, y))
	return nil
}

func (m *mockExecutor) MouseDown(x, y int32, button string) error {
	m.calls = append(m.calls, fmt.Sprintf("MouseDown:%d,%d,%s", x, y, button))
	return nil
}

func (m *mockExecutor) MouseUp(button string) error {
	m.calls = append(m.calls, "MouseUp:"+button)
	return nil
}

func (m *mockExecutor) Scroll(dx, dy int32) error {
	m.calls = append(m.calls, fmt.Sprintf("Scroll:%d,%d", dx, dy))
	return nil
}

// mockExecutorNoMouse implements only executor.Executor — no mouse support.
type mockExecutorNoMouse struct {
	calls []string
}

func (m *mockExecutorNoMouse) Type(text string) error {
	m.calls = append(m.calls, "Type:"+text)
	return nil
}

func (m *mockExecutorNoMouse) Combo(keys ...string) error {
	m.calls = append(m.calls, "Combo:"+strings.Join(keys, ","))
	return nil
}

func (m *mockExecutorNoMouse) KeyDown(key string) error {
	m.calls = append(m.calls, "KeyDown:"+key)
	return nil
}

func (m *mockExecutorNoMouse) KeyUp(key string) error {
	m.calls = append(m.calls, "KeyUp:"+key)
	return nil
}

func (m *mockExecutorNoMouse) Reset() error {
	m.calls = append(m.calls, "Reset")
	return nil
}

func (m *mockExecutorNoMouse) Close() error {
	m.calls = append(m.calls, "Close")
	return nil
}

func TestParseStream(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		noMouse   bool   // use executor without MouseExecutor support
		wantErr   bool
		wantErrIs error  // nil means just check wantErr; non-nil checked via errors.Is
		wantCalls []string // nil = don't check; []string{} = expect zero calls
	}{
		// ------------------------------------------------------------------ //
		// Valid keyboard commands
		// ------------------------------------------------------------------ //
		{
			name:      "T types text",
			input:     "T hello world",
			wantCalls: []string{"Type:hello world"},
		},
		{
			name:      "T types single character",
			input:     "T a",
			wantCalls: []string{"Type:a"},
		},
		{
			name:      "C combo ctrl c",
			input:     "C ctrl c",
			wantCalls: []string{"Combo:ctrl,c"},
		},
		{
			name:      "C combo meta space",
			input:     "C meta space",
			wantCalls: []string{"Combo:meta,space"},
		},
		{
			name:      "C combo single key",
			input:     "C ctrl",
			wantCalls: []string{"Combo:ctrl"},
		},
		{
			name:      "C combo three keys",
			input:     "C ctrl alt del",
			wantCalls: []string{"Combo:ctrl,alt,del"},
		},
		{
			name:      "D key down",
			input:     "D shift",
			wantCalls: []string{"KeyDown:shift"},
		},
		{
			name:      "U key up",
			input:     "U shift",
			wantCalls: []string{"KeyUp:shift"},
		},
		{
			name:      "R resets all held keys",
			input:     "R",
			wantCalls: []string{"Reset"},
		},
		{
			name:      "P ping is a no-op",
			input:     "P",
			wantCalls: []string{},
		},

		// ------------------------------------------------------------------ //
		// Valid mouse commands
		// ------------------------------------------------------------------ //
		{
			name:      "M M moves mouse to coordinates",
			input:     "M M 100 200",
			wantCalls: []string{"MoveTo:100,200"},
		},
		{
			name:      "M C left clicks at coordinates",
			input:     "M C 500 300",
			wantCalls: []string{"LeftClick:500,300"},
		},
		{
			name:      "M R right clicks at coordinates",
			input:     "M R 500 300",
			wantCalls: []string{"RightClick:500,300"},
		},
		{
			name:      "M K middle clicks at coordinates",
			input:     "M K 200 300",
			wantCalls: []string{"MiddleClick:200,300"},
		},
		{
			name:      "M D mouse down with explicit left button",
			input:     "M D 100 100 left",
			wantCalls: []string{"MouseDown:100,100,left"},
		},
		{
			name:      "M D mouse down with right button",
			input:     "M D 50 75 right",
			wantCalls: []string{"MouseDown:50,75,right"},
		},
		{
			name:      "M D mouse down defaults to left button",
			input:     "M D 100 100",
			wantCalls: []string{"MouseDown:100,100,left"},
		},
		{
			name:      "M U mouse up with explicit right button",
			input:     "M U right",
			wantCalls: []string{"MouseUp:right"},
		},
		{
			name:      "M U mouse up defaults to left button",
			input:     "M U",
			wantCalls: []string{"MouseUp:left"},
		},
		{
			name:      "M S scroll with negative dy",
			input:     "M S 0 -3",
			wantCalls: []string{"Scroll:0,-3"},
		},
		{
			name:      "M S scroll with positive deltas",
			input:     "M S 2 5",
			wantCalls: []string{"Scroll:2,5"},
		},

		// ------------------------------------------------------------------ //
		// Multi-command streams
		// ------------------------------------------------------------------ //
		{
			name:      "multiple commands separated by newlines",
			input:     "T hello\nC ctrl c\nR",
			wantCalls: []string{"Type:hello", "Combo:ctrl,c", "Reset"},
		},
		{
			name:      "keyboard and mouse commands interleaved",
			input:     "T click\nM C 10 20\nR",
			wantCalls: []string{"Type:click", "LeftClick:10,20", "Reset"},
		},
		{
			name:      "CRLF line endings (Windows clients)",
			input:     "T hello\r\nC ctrl c\r\n",
			wantCalls: []string{"Type:hello", "Combo:ctrl,c"},
		},
		{
			name:      "empty lines between commands are skipped",
			input:     "T hello\n\nR\n",
			wantCalls: []string{"Type:hello", "Reset"},
		},
		{
			name:      "multiple empty lines are all skipped",
			input:     "\n\n\nT hi\n\n\n",
			wantCalls: []string{"Type:hi"},
		},
		{
			name:      "ping lines between real commands are no-ops",
			input:     "P\nT hello\nP\nR\nP",
			wantCalls: []string{"Type:hello", "Reset"},
		},

		// ------------------------------------------------------------------ //
		// Edge cases — valid
		// ------------------------------------------------------------------ //
		{
			name:      "empty stream returns nil with no calls",
			input:     "",
			wantCalls: []string{},
		},
		{
			name:      "single command with no trailing newline",
			input:     "T hello",
			wantCalls: []string{"Type:hello"},
		},
		{
			name:      "T alone is silently skipped (no text to type)",
			input:     "T",
			wantCalls: []string{},
		},
		{
			name:      "T with only a space is silently skipped",
			input:     "T ",
			wantCalls: []string{},
		},
		{
			name:      "T with tab in text passes tab through",
			input:     "T hello\tworld",
			wantCalls: []string{"Type:hello\tworld"},
		},
		{
			name:      "T with special chars in text",
			input:     "T !@#$%^&*()",
			wantCalls: []string{"Type:!@#$%^&*()"},
		},
		{
			name:  "T with very long text",
			input: "T " + strings.Repeat("x", 10000),
			wantCalls: []string{"Type:" + strings.Repeat("x", 10000)},
		},
		{
			name:      "C with extra whitespace between keys is normalized",
			input:     "C  ctrl  c",
			wantCalls: []string{"Combo:ctrl,c"},
		},
		{
			name:      "M M with zero coordinates",
			input:     "M M 0 0",
			wantCalls: []string{"MoveTo:0,0"},
		},
		{
			name:      "M M with negative coordinates",
			input:     "M M -100 -200",
			wantCalls: []string{"MoveTo:-100,-200"},
		},
		{
			name:  "M C with int32 boundary coordinates",
			input: fmt.Sprintf("M C %d %d", math.MaxInt32, math.MinInt32),
			wantCalls: []string{
				fmt.Sprintf("LeftClick:%d,%d", math.MaxInt32, math.MinInt32),
			},
		},

		// ------------------------------------------------------------------ //
		// Error cases — unknown/malformed commands
		// ------------------------------------------------------------------ //
		{
			name:      "unknown command X errors immediately",
			input:     "X hello",
			wantErr:   true,
			wantErrIs: ErrUnknownCommand,
			wantCalls: []string{},
		},
		{
			name:      "unknown command lowercase errors",
			input:     "t hello",
			wantErr:   true,
			wantErrIs: ErrUnknownCommand,
			wantCalls: []string{},
		},
		{
			name:      "T without space after T errors",
			input:     "Tabc",
			wantErr:   true,
			wantErrIs: ErrUnknownCommand,
			wantCalls: []string{},
		},
		{
			name:      "C with no args errors",
			input:     "C",
			wantErr:   true,
			wantErrIs: ErrUnknownCommand,
			wantCalls: []string{},
		},
		{
			name:      "C with space but empty keys errors",
			input:     "C   ",
			wantErr:   true,
			wantErrIs: ErrMissingArgument,
			wantCalls: []string{},
		},
		{
			name:      "D with no args errors",
			input:     "D",
			wantErr:   true,
			wantErrIs: ErrUnknownCommand,
			wantCalls: []string{},
		},
		{
			name:      "D with whitespace-only key errors",
			input:     "D   ",
			wantErr:   true,
			wantErrIs: ErrMissingArgument,
			wantCalls: []string{},
		},
		{
			name:      "U with no args errors",
			input:     "U",
			wantErr:   true,
			wantErrIs: ErrUnknownCommand,
			wantCalls: []string{},
		},
		{
			name:      "U with whitespace-only key errors",
			input:     "U   ",
			wantErr:   true,
			wantErrIs: ErrMissingArgument,
			wantCalls: []string{},
		},
		{
			name:      "M command without mouse support errors",
			input:     "M M 100 200",
			noMouse:   true,
			wantErr:   true,
			wantErrIs: ErrUnknownCommand,
			wantCalls: []string{},
		},
		{
			name:      "M with no sub-command errors",
			input:     "M",
			wantErr:   true,
			wantErrIs: ErrMissingArgument,
			wantCalls: []string{},
		},
		{
			name:      "M C with only one coordinate errors",
			input:     "M C 100",
			wantErr:   true,
			wantErrIs: ErrMissingArgument,
			wantCalls: []string{},
		},
		{
			name:      "M C with non-numeric x coordinate errors",
			input:     "M C abc def",
			wantErr:   true,
			wantCalls: []string{},
		},
		{
			name:      "M C with non-numeric y coordinate errors",
			input:     "M C 100 def",
			wantErr:   true,
			wantCalls: []string{},
		},
		{
			name:      "M C with coordinate exceeding int32 max errors",
			input:     "M C 2147483648 0",
			wantErr:   true,
			wantCalls: []string{},
		},
		{
			name:      "M with unknown sub-command errors",
			input:     "M Z 100 200",
			wantErr:   true,
			wantErrIs: ErrUnknownCommand,
			wantCalls: []string{},
		},
		{
			name:      "M M missing both coordinates errors",
			input:     "M M",
			wantErr:   true,
			wantErrIs: ErrMissingArgument,
			wantCalls: []string{},
		},
		{
			name:      "M R missing coordinates errors",
			input:     "M R 100",
			wantErr:   true,
			wantErrIs: ErrMissingArgument,
			wantCalls: []string{},
		},
		{
			name:      "M S missing scroll deltas errors",
			input:     "M S 0",
			wantErr:   true,
			wantErrIs: ErrMissingArgument,
			wantCalls: []string{},
		},
		{
			name:      "M D missing coordinates errors",
			input:     "M D 100",
			wantErr:   true,
			wantErrIs: ErrMissingArgument,
			wantCalls: []string{},
		},

		// ------------------------------------------------------------------ //
		// Error halts the stream
		// ------------------------------------------------------------------ //
		{
			name:      "error on first command halts entire stream",
			input:     "X bad\nT good",
			wantErr:   true,
			wantErrIs: ErrUnknownCommand,
			wantCalls: []string{},
		},
		{
			name:      "commands before error are executed",
			input:     "T hello\nX bad\nT never",
			wantErr:   true,
			wantErrIs: ErrUnknownCommand,
			wantCalls: []string{"Type:hello"},
		},
		{
			name:      "multiple valid commands then error",
			input:     "T hello\nC ctrl c\nR\nM Z 0 0",
			wantErr:   true,
			wantErrIs: ErrUnknownCommand,
			wantCalls: []string{"Type:hello", "Combo:ctrl,c", "Reset"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				exec      executor.Executor
				getCalls  func() []string
			)

			if tt.noMouse {
				m := &mockExecutorNoMouse{}
				exec = m
				getCalls = func() []string { return m.calls }
			} else {
				m := &mockExecutor{}
				exec = m
				getCalls = func() []string { return m.calls }
			}

			p := NewTextParser()
			err := p.ParseStream(strings.NewReader(tt.input), exec)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseStream() expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("ParseStream() error = %v, want errors.Is(%v)", err, tt.wantErrIs)
				}
			} else {
				if err != nil {
					t.Fatalf("ParseStream() unexpected error: %v", err)
				}
			}

			if tt.wantCalls != nil {
				calls := getCalls()
				if len(calls) != len(tt.wantCalls) {
					t.Fatalf("executor calls = %v (len %d), want %v (len %d)",
						calls, len(calls), tt.wantCalls, len(tt.wantCalls))
				}
				for i, got := range calls {
					if got != tt.wantCalls[i] {
						t.Errorf("calls[%d] = %q, want %q", i, got, tt.wantCalls[i])
					}
				}
			}
		})
	}
}
