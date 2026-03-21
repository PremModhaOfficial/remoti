package protocol

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"remoti/pkg/executor"
)

var (
	ErrUnknownCommand  = errors.New("unknown command")
	ErrMissingArgument = errors.New("missing argument")
)

// Parser defines the interface for translating an input stream into executable commands.
type Parser interface {
	// ParseStream reads lines from the reader and continuously triggers commands on the Executor.
	ParseStream(r io.Reader, exec executor.Executor) error
}

type textParser struct{}

// NewTextParser creates a new Redis-like plain-text protocol parser.
func NewTextParser() Parser {
	return &textParser{}
}

// ParseStream reads the stream line-by-line and triggers the executor.
func (p *textParser) ParseStream(r io.Reader, exec executor.Executor) error {
	scanner := bufio.NewScanner(r)

	// Type-assert once for mouse support
	mouseExec, hasMouseSupport := exec.(executor.MouseExecutor)

	for scanner.Scan() {
		line := scanner.Bytes()
		// Remove carriage return if present
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}

		if len(line) == 0 {
			continue // skip empty lines
		}

		cmd := line[0]

		switch cmd {
		case 'T':
			if len(line) < 3 || line[1] != ' ' {
				if len(line) > 1 && line[1] != ' ' {
					return fmt.Errorf("%w: expected space after T", ErrUnknownCommand)
				}
				continue
			}
			text := string(line[2:])
			if err := exec.Type(text); err != nil {
				return err
			}
		case 'C':
			if len(line) < 3 || line[1] != ' ' {
				return fmt.Errorf("%w: expected space after C", ErrUnknownCommand)
			}
			keys := strings.Fields(string(line[2:]))
			if len(keys) == 0 {
				return fmt.Errorf("%w: missing keys for C", ErrMissingArgument)
			}
			if err := exec.Combo(keys...); err != nil {
				return err
			}
		case 'D':
			if len(line) < 3 || line[1] != ' ' {
				return fmt.Errorf("%w: expected space after D", ErrUnknownCommand)
			}
			key := strings.TrimSpace(string(line[2:]))
			if key == "" {
				return fmt.Errorf("%w: missing key for D", ErrMissingArgument)
			}
			if err := exec.KeyDown(key); err != nil {
				return err
			}
		case 'U':
			if len(line) < 3 || line[1] != ' ' {
				return fmt.Errorf("%w: expected space after U", ErrUnknownCommand)
			}
			key := strings.TrimSpace(string(line[2:]))
			if key == "" {
				return fmt.Errorf("%w: missing key for U", ErrMissingArgument)
			}
			if err := exec.KeyUp(key); err != nil {
				return err
			}
		case 'R':
			if err := exec.Reset(); err != nil {
				return err
			}
		case 'P':
			// Heartbeat/ping — no-op, used by clients to check connection
			continue
		case 'M':
			if !hasMouseSupport {
				return fmt.Errorf("%w: mouse commands not supported by executor", ErrUnknownCommand)
			}
			if err := parseMouseCommand(string(line), mouseExec); err != nil {
				return err
			}
		default:
			return fmt.Errorf("%w: %c", ErrUnknownCommand, cmd)
		}
	}

	return scanner.Err()
}

// parseMouseCommand handles M sub-commands.
// Format: M <sub> [args...]
func parseMouseCommand(line string, mouse executor.MouseExecutor) error {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return fmt.Errorf("%w: missing mouse sub-command", ErrMissingArgument)
	}

	sub := fields[1]
	switch sub {
	case "M": // M M <x> <y> — move
		if len(fields) < 4 {
			return fmt.Errorf("%w: M M requires <x> <y>", ErrMissingArgument)
		}
		x, y, err := parseCoords(fields[2], fields[3])
		if err != nil {
			return err
		}
		return mouse.MoveTo(x, y)

	case "C": // M C <x> <y> — left click
		if len(fields) < 4 {
			return fmt.Errorf("%w: M C requires <x> <y>", ErrMissingArgument)
		}
		x, y, err := parseCoords(fields[2], fields[3])
		if err != nil {
			return err
		}
		return mouse.LeftClick(x, y)

	case "R": // M R <x> <y> — right click
		if len(fields) < 4 {
			return fmt.Errorf("%w: M R requires <x> <y>", ErrMissingArgument)
		}
		x, y, err := parseCoords(fields[2], fields[3])
		if err != nil {
			return err
		}
		return mouse.RightClick(x, y)

	case "K": // M K <x> <y> — middle click
		if len(fields) < 4 {
			return fmt.Errorf("%w: M K requires <x> <y>", ErrMissingArgument)
		}
		x, y, err := parseCoords(fields[2], fields[3])
		if err != nil {
			return err
		}
		return mouse.MiddleClick(x, y)

	case "D": // M D <x> <y> [button] — mouse down
		if len(fields) < 4 {
			return fmt.Errorf("%w: M D requires <x> <y>", ErrMissingArgument)
		}
		x, y, err := parseCoords(fields[2], fields[3])
		if err != nil {
			return err
		}
		button := "left"
		if len(fields) >= 5 {
			button = fields[4]
		}
		return mouse.MouseDown(x, y, button)

	case "U": // M U [button] — mouse up
		button := "left"
		if len(fields) >= 3 {
			button = fields[2]
		}
		return mouse.MouseUp(button)

	case "S": // M S <dx> <dy> — scroll
		if len(fields) < 4 {
			return fmt.Errorf("%w: M S requires <dx> <dy>", ErrMissingArgument)
		}
		dx, dy, err := parseCoords(fields[2], fields[3])
		if err != nil {
			return err
		}
		return mouse.Scroll(dx, dy)

	default:
		return fmt.Errorf("%w: unknown mouse sub-command: %s", ErrUnknownCommand, sub)
	}
}

func parseCoords(xs, ys string) (int32, int32, error) {
	x, err := strconv.ParseInt(xs, 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid x coordinate %q: %w", xs, err)
	}
	y, err := strconv.ParseInt(ys, 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid y coordinate %q: %w", ys, err)
	}
	return int32(x), int32(y), nil
}
