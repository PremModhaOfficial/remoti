package protocol

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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
		default:
			return fmt.Errorf("%w: %c", ErrUnknownCommand, cmd)
		}
	}

	return scanner.Err()
}
