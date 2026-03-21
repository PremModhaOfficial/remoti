package executor

import (
	"fmt"
	"strings"

	"github.com/bendahl/uinput"
)

// LinuxExecutor implements the Executor interface using /dev/uinput.
type LinuxExecutor struct {
	vkb     uinput.Keyboard
	pressed map[int]bool // keep track of down keys to reset them
}

// NewLinuxExecutor creates a new Executor backed by /dev/uinput.
func NewLinuxExecutor() (*LinuxExecutor, error) {
	vkb, err := uinput.CreateKeyboard("/dev/uinput", []byte("Remoti Virtual Keyboard"))
	if err != nil {
		return nil, fmt.Errorf("failed to create virtual keyboard: %w", err)
	}

	return &LinuxExecutor{
		vkb:     vkb,
		pressed: make(map[int]bool),
	}, nil
}

// Close closes the virtual keyboard and releases any resources.
func (e *LinuxExecutor) Close() error {
	e.Reset()
	return e.vkb.Close()
}

func (e *LinuxExecutor) Type(text string) error {
	for _, char := range text {
		keyCode, shift := runeToKey(char)

		if shift {
			e.vkb.KeyDown(uinput.KeyLeftshift)
		}

		err := e.vkb.KeyPress(keyCode)

		if shift {
			e.vkb.KeyUp(uinput.KeyLeftshift)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func (e *LinuxExecutor) Combo(keys ...string) error {
	keyCodes := make([]int, 0, len(keys))
	for _, k := range keys {
		kc, err := parseKeyName(k)
		if err != nil {
			return err
		}
		keyCodes = append(keyCodes, kc)
	}

	// Press all down
	for _, kc := range keyCodes {
		if err := e.vkb.KeyDown(kc); err != nil {
			return err
		}
	}

	// Release all in reverse order
	for i := len(keyCodes) - 1; i >= 0; i-- {
		e.vkb.KeyUp(keyCodes[i])
	}
	return nil
}

func (e *LinuxExecutor) KeyDown(key string) error {
	kc, err := parseKeyName(key)
	if err != nil {
		return err
	}
	e.pressed[kc] = true
	return e.vkb.KeyDown(kc)
}

func (e *LinuxExecutor) KeyUp(key string) error {
	kc, err := parseKeyName(key)
	if err != nil {
		return err
	}
	delete(e.pressed, kc)
	return e.vkb.KeyUp(kc)
}

func (e *LinuxExecutor) Reset() error {
	for kc := range e.pressed {
		e.vkb.KeyUp(kc)
		delete(e.pressed, kc)
	}
	return nil
}

// Helper to parse key names into uinput key codes
func parseKeyName(name string) (int, error) {
	name = strings.ToLower(name)
	switch name {
	case "ctrl", "leftctrl":
		return uinput.KeyLeftctrl, nil
	case "rightctrl":
		return uinput.KeyRightctrl, nil
	case "alt", "leftalt":
		return uinput.KeyLeftalt, nil
	case "rightalt":
		return uinput.KeyRightalt, nil
	case "shift", "leftshift":
		return uinput.KeyLeftshift, nil
	case "rightshift":
		return uinput.KeyRightshift, nil
	case "meta", "leftmeta", "win", "cmd":
		return uinput.KeyLeftmeta, nil
	case "rightmeta":
		return uinput.KeyRightmeta, nil
	case "enter", "return":
		return uinput.KeyEnter, nil
	case "space":
		return uinput.KeySpace, nil
	case "tab":
		return uinput.KeyTab, nil
	case "backspace":
		return uinput.KeyBackspace, nil
	case "escape", "esc":
		return uinput.KeyEsc, nil
	case "delete", "del":
		return uinput.KeyDelete, nil
	case "up":
		return uinput.KeyUp, nil
	case "down":
		return uinput.KeyDown, nil
	case "left":
		return uinput.KeyLeft, nil
	case "right":
		return uinput.KeyRight, nil
	case "f1":
		return uinput.KeyF1, nil
	case "f2":
		return uinput.KeyF2, nil
	case "f3":
		return uinput.KeyF3, nil
	case "f4":
		return uinput.KeyF4, nil
	case "f5":
		return uinput.KeyF5, nil
	case "f6":
		return uinput.KeyF6, nil
	case "f7":
		return uinput.KeyF7, nil
	case "f8":
		return uinput.KeyF8, nil
	case "f9":
		return uinput.KeyF9, nil
	case "f10":
		return uinput.KeyF10, nil
	case "f11":
		return uinput.KeyF11, nil
	case "f12":
		return uinput.KeyF12, nil
	default:
		if len(name) == 1 {
			kc, _ := runeToKey(rune(name[0]))
			if kc != 0 {
				return kc, nil
			}
		}
		return 0, fmt.Errorf("unknown key name: %s", name)
	}
}

// Simple map for standard ASCII characters to uinput KeyCodes
func runeToKey(r rune) (int, bool) {
	switch r {
	case 'a':
		return uinput.KeyA, false
	case 'b':
		return uinput.KeyB, false
	case 'c':
		return uinput.KeyC, false
	case 'd':
		return uinput.KeyD, false
	case 'e':
		return uinput.KeyE, false
	case 'f':
		return uinput.KeyF, false
	case 'g':
		return uinput.KeyG, false
	case 'h':
		return uinput.KeyH, false
	case 'i':
		return uinput.KeyI, false
	case 'j':
		return uinput.KeyJ, false
	case 'k':
		return uinput.KeyK, false
	case 'l':
		return uinput.KeyL, false
	case 'm':
		return uinput.KeyM, false
	case 'n':
		return uinput.KeyN, false
	case 'o':
		return uinput.KeyO, false
	case 'p':
		return uinput.KeyP, false
	case 'q':
		return uinput.KeyQ, false
	case 'r':
		return uinput.KeyR, false
	case 's':
		return uinput.KeyS, false
	case 't':
		return uinput.KeyT, false
	case 'u':
		return uinput.KeyU, false
	case 'v':
		return uinput.KeyV, false
	case 'w':
		return uinput.KeyW, false
	case 'x':
		return uinput.KeyX, false
	case 'y':
		return uinput.KeyY, false
	case 'z':
		return uinput.KeyZ, false

	case 'A':
		return uinput.KeyA, true
	case 'B':
		return uinput.KeyB, true
	case 'C':
		return uinput.KeyC, true
	case 'D':
		return uinput.KeyD, true
	case 'E':
		return uinput.KeyE, true
	case 'F':
		return uinput.KeyF, true
	case 'G':
		return uinput.KeyG, true
	case 'H':
		return uinput.KeyH, true
	case 'I':
		return uinput.KeyI, true
	case 'J':
		return uinput.KeyJ, true
	case 'K':
		return uinput.KeyK, true
	case 'L':
		return uinput.KeyL, true
	case 'M':
		return uinput.KeyM, true
	case 'N':
		return uinput.KeyN, true
	case 'O':
		return uinput.KeyO, true
	case 'P':
		return uinput.KeyP, true
	case 'Q':
		return uinput.KeyQ, true
	case 'R':
		return uinput.KeyR, true
	case 'S':
		return uinput.KeyS, true
	case 'T':
		return uinput.KeyT, true
	case 'U':
		return uinput.KeyU, true
	case 'V':
		return uinput.KeyV, true
	case 'W':
		return uinput.KeyW, true
	case 'X':
		return uinput.KeyX, true
	case 'Y':
		return uinput.KeyY, true
	case 'Z':
		return uinput.KeyZ, true

	case '0':
		return uinput.Key0, false
	case '1':
		return uinput.Key1, false
	case '2':
		return uinput.Key2, false
	case '3':
		return uinput.Key3, false
	case '4':
		return uinput.Key4, false
	case '5':
		return uinput.Key5, false
	case '6':
		return uinput.Key6, false
	case '7':
		return uinput.Key7, false
	case '8':
		return uinput.Key8, false
	case '9':
		return uinput.Key9, false

	case '!':
		return uinput.Key1, true
	case '@':
		return uinput.Key2, true
	case '#':
		return uinput.Key3, true
	case '$':
		return uinput.Key4, true
	case '%':
		return uinput.Key5, true
	case '^':
		return uinput.Key6, true
	case '&':
		return uinput.Key7, true
	case '*':
		return uinput.Key8, true
	case '(':
		return uinput.Key9, true
	case ')':
		return uinput.Key0, true

	case ' ':
		return uinput.KeySpace, false
	case '\n':
		return uinput.KeyEnter, false
	case '\t':
		return uinput.KeyTab, false

	case '-':
		return uinput.KeyMinus, false
	case '_':
		return uinput.KeyMinus, true
	case '=':
		return uinput.KeyEqual, false
	case '+':
		return uinput.KeyEqual, true

	case '[':
		return uinput.KeyLeftbrace, false
	case '{':
		return uinput.KeyLeftbrace, true
	case ']':
		return uinput.KeyRightbrace, false
	case '}':
		return uinput.KeyRightbrace, true

	case '\\':
		return uinput.KeyBackslash, false
	case '|':
		return uinput.KeyBackslash, true

	case ';':
		return uinput.KeySemicolon, false
	case ':':
		return uinput.KeySemicolon, true

	case '\'':
		return uinput.KeyApostrophe, false
	case '"':
		return uinput.KeyApostrophe, true

	case ',':
		return uinput.KeyComma, false
	case '<':
		return uinput.KeyComma, true
	case '.':
		return uinput.KeyDot, false
	case '>':
		return uinput.KeyDot, true

	case '/':
		return uinput.KeySlash, false
	case '?':
		return uinput.KeySlash, true

	case '`':
		return uinput.KeyGrave, false
	case '~':
		return uinput.KeyGrave, true
	}
	return 0, false
}
