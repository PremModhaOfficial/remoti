package executor

import (
	"testing"

	"github.com/bendahl/uinput"
)

func TestParseKeyName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		// Modifier keys — case-insensitive
		{name: "ctrl lowercase", input: "ctrl", want: uinput.KeyLeftctrl},
		{name: "ctrl uppercase", input: "CTRL", want: uinput.KeyLeftctrl},
		{name: "ctrl mixed", input: "Ctrl", want: uinput.KeyLeftctrl},
		{name: "leftctrl", input: "leftctrl", want: uinput.KeyLeftctrl},
		{name: "rightctrl", input: "rightctrl", want: uinput.KeyRightctrl},
		{name: "alt", input: "alt", want: uinput.KeyLeftalt},
		{name: "leftalt", input: "leftalt", want: uinput.KeyLeftalt},
		{name: "rightalt", input: "rightalt", want: uinput.KeyRightalt},
		{name: "shift", input: "shift", want: uinput.KeyLeftshift},
		{name: "leftshift", input: "leftshift", want: uinput.KeyLeftshift},
		{name: "rightshift", input: "rightshift", want: uinput.KeyRightshift},
		{name: "meta", input: "meta", want: uinput.KeyLeftmeta},
		{name: "leftmeta", input: "leftmeta", want: uinput.KeyLeftmeta},
		{name: "win", input: "win", want: uinput.KeyLeftmeta},
		{name: "cmd", input: "cmd", want: uinput.KeyLeftmeta},
		{name: "rightmeta", input: "rightmeta", want: uinput.KeyRightmeta},

		// Common keys
		{name: "enter", input: "enter", want: uinput.KeyEnter},
		{name: "return", input: "return", want: uinput.KeyEnter},
		{name: "space", input: "space", want: uinput.KeySpace},
		{name: "tab", input: "tab", want: uinput.KeyTab},
		{name: "backspace", input: "backspace", want: uinput.KeyBackspace},
		{name: "escape", input: "escape", want: uinput.KeyEsc},
		{name: "esc", input: "esc", want: uinput.KeyEsc},
		{name: "delete", input: "delete", want: uinput.KeyDelete},
		{name: "del", input: "del", want: uinput.KeyDelete},

		// Arrow keys
		{name: "up", input: "up", want: uinput.KeyUp},
		{name: "down", input: "down", want: uinput.KeyDown},
		{name: "left", input: "left", want: uinput.KeyLeft},
		{name: "right", input: "right", want: uinput.KeyRight},

		// Function keys
		{name: "f1", input: "f1", want: uinput.KeyF1},
		{name: "f2", input: "f2", want: uinput.KeyF2},
		{name: "f3", input: "f3", want: uinput.KeyF3},
		{name: "f4", input: "f4", want: uinput.KeyF4},
		{name: "f5", input: "f5", want: uinput.KeyF5},
		{name: "f6", input: "f6", want: uinput.KeyF6},
		{name: "f7", input: "f7", want: uinput.KeyF7},
		{name: "f8", input: "f8", want: uinput.KeyF8},
		{name: "f9", input: "f9", want: uinput.KeyF9},
		{name: "f10", input: "f10", want: uinput.KeyF10},
		{name: "f11", input: "f11", want: uinput.KeyF11},
		{name: "f12", input: "f12", want: uinput.KeyF12},

		// Single character keys (via runeToKey fallthrough)
		{name: "single letter a", input: "a", want: uinput.KeyA},
		{name: "single letter z", input: "z", want: uinput.KeyZ},
		{name: "single digit 0", input: "0", want: uinput.Key0},

		// Error cases
		{name: "empty string", input: "", wantErr: true},
		{name: "nonexistent key", input: "nonexistent", wantErr: true},
		{name: "combo notation rejected", input: "ctrl+c", wantErr: true},
		{name: "two chars not a key name", input: "ab", wantErr: true},
		{name: "uppercase A not a valid name (resolves as lowercase)", input: "A", want: uinput.KeyA}, // uppercase single char goes through toLower -> 'a' -> KeyA
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseKeyName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseKeyName(%q) = %d, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("parseKeyName(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("parseKeyName(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestRuneToKey(t *testing.T) {
	tests := []struct {
		name      string
		input     rune
		wantKey   int
		wantShift bool
	}{
		// Lowercase letters
		{name: "lowercase a", input: 'a', wantKey: uinput.KeyA, wantShift: false},
		{name: "lowercase b", input: 'b', wantKey: uinput.KeyB, wantShift: false},
		{name: "lowercase c", input: 'c', wantKey: uinput.KeyC, wantShift: false},
		{name: "lowercase d", input: 'd', wantKey: uinput.KeyD, wantShift: false},
		{name: "lowercase e", input: 'e', wantKey: uinput.KeyE, wantShift: false},
		{name: "lowercase f", input: 'f', wantKey: uinput.KeyF, wantShift: false},
		{name: "lowercase g", input: 'g', wantKey: uinput.KeyG, wantShift: false},
		{name: "lowercase h", input: 'h', wantKey: uinput.KeyH, wantShift: false},
		{name: "lowercase i", input: 'i', wantKey: uinput.KeyI, wantShift: false},
		{name: "lowercase j", input: 'j', wantKey: uinput.KeyJ, wantShift: false},
		{name: "lowercase k", input: 'k', wantKey: uinput.KeyK, wantShift: false},
		{name: "lowercase l", input: 'l', wantKey: uinput.KeyL, wantShift: false},
		{name: "lowercase m", input: 'm', wantKey: uinput.KeyM, wantShift: false},
		{name: "lowercase n", input: 'n', wantKey: uinput.KeyN, wantShift: false},
		{name: "lowercase o", input: 'o', wantKey: uinput.KeyO, wantShift: false},
		{name: "lowercase p", input: 'p', wantKey: uinput.KeyP, wantShift: false},
		{name: "lowercase q", input: 'q', wantKey: uinput.KeyQ, wantShift: false},
		{name: "lowercase r", input: 'r', wantKey: uinput.KeyR, wantShift: false},
		{name: "lowercase s", input: 's', wantKey: uinput.KeyS, wantShift: false},
		{name: "lowercase t", input: 't', wantKey: uinput.KeyT, wantShift: false},
		{name: "lowercase u", input: 'u', wantKey: uinput.KeyU, wantShift: false},
		{name: "lowercase v", input: 'v', wantKey: uinput.KeyV, wantShift: false},
		{name: "lowercase w", input: 'w', wantKey: uinput.KeyW, wantShift: false},
		{name: "lowercase x", input: 'x', wantKey: uinput.KeyX, wantShift: false},
		{name: "lowercase y", input: 'y', wantKey: uinput.KeyY, wantShift: false},
		{name: "lowercase z", input: 'z', wantKey: uinput.KeyZ, wantShift: false},

		// Uppercase letters — all require shift
		{name: "uppercase A", input: 'A', wantKey: uinput.KeyA, wantShift: true},
		{name: "uppercase B", input: 'B', wantKey: uinput.KeyB, wantShift: true},
		{name: "uppercase C", input: 'C', wantKey: uinput.KeyC, wantShift: true},
		{name: "uppercase D", input: 'D', wantKey: uinput.KeyD, wantShift: true},
		{name: "uppercase E", input: 'E', wantKey: uinput.KeyE, wantShift: true},
		{name: "uppercase F", input: 'F', wantKey: uinput.KeyF, wantShift: true},
		{name: "uppercase G", input: 'G', wantKey: uinput.KeyG, wantShift: true},
		{name: "uppercase H", input: 'H', wantKey: uinput.KeyH, wantShift: true},
		{name: "uppercase I", input: 'I', wantKey: uinput.KeyI, wantShift: true},
		{name: "uppercase J", input: 'J', wantKey: uinput.KeyJ, wantShift: true},
		{name: "uppercase K", input: 'K', wantKey: uinput.KeyK, wantShift: true},
		{name: "uppercase L", input: 'L', wantKey: uinput.KeyL, wantShift: true},
		{name: "uppercase M", input: 'M', wantKey: uinput.KeyM, wantShift: true},
		{name: "uppercase N", input: 'N', wantKey: uinput.KeyN, wantShift: true},
		{name: "uppercase O", input: 'O', wantKey: uinput.KeyO, wantShift: true},
		{name: "uppercase P", input: 'P', wantKey: uinput.KeyP, wantShift: true},
		{name: "uppercase Q", input: 'Q', wantKey: uinput.KeyQ, wantShift: true},
		{name: "uppercase R", input: 'R', wantKey: uinput.KeyR, wantShift: true},
		{name: "uppercase S", input: 'S', wantKey: uinput.KeyS, wantShift: true},
		{name: "uppercase T", input: 'T', wantKey: uinput.KeyT, wantShift: true},
		{name: "uppercase U", input: 'U', wantKey: uinput.KeyU, wantShift: true},
		{name: "uppercase V", input: 'V', wantKey: uinput.KeyV, wantShift: true},
		{name: "uppercase W", input: 'W', wantKey: uinput.KeyW, wantShift: true},
		{name: "uppercase X", input: 'X', wantKey: uinput.KeyX, wantShift: true},
		{name: "uppercase Y", input: 'Y', wantKey: uinput.KeyY, wantShift: true},
		{name: "uppercase Z", input: 'Z', wantKey: uinput.KeyZ, wantShift: true},

		// Digits
		{name: "digit 0", input: '0', wantKey: uinput.Key0, wantShift: false},
		{name: "digit 1", input: '1', wantKey: uinput.Key1, wantShift: false},
		{name: "digit 2", input: '2', wantKey: uinput.Key2, wantShift: false},
		{name: "digit 3", input: '3', wantKey: uinput.Key3, wantShift: false},
		{name: "digit 4", input: '4', wantKey: uinput.Key4, wantShift: false},
		{name: "digit 5", input: '5', wantKey: uinput.Key5, wantShift: false},
		{name: "digit 6", input: '6', wantKey: uinput.Key6, wantShift: false},
		{name: "digit 7", input: '7', wantKey: uinput.Key7, wantShift: false},
		{name: "digit 8", input: '8', wantKey: uinput.Key8, wantShift: false},
		{name: "digit 9", input: '9', wantKey: uinput.Key9, wantShift: false},

		// Shifted digit symbols
		{name: "! is shift+1", input: '!', wantKey: uinput.Key1, wantShift: true},
		{name: "@ is shift+2", input: '@', wantKey: uinput.Key2, wantShift: true},
		{name: "# is shift+3", input: '#', wantKey: uinput.Key3, wantShift: true},
		{name: "$ is shift+4", input: '$', wantKey: uinput.Key4, wantShift: true},
		{name: "% is shift+5", input: '%', wantKey: uinput.Key5, wantShift: true},
		{name: "^ is shift+6", input: '^', wantKey: uinput.Key6, wantShift: true},
		{name: "& is shift+7", input: '&', wantKey: uinput.Key7, wantShift: true},
		{name: "* is shift+8", input: '*', wantKey: uinput.Key8, wantShift: true},
		{name: "( is shift+9", input: '(', wantKey: uinput.Key9, wantShift: true},
		{name: ") is shift+0", input: ')', wantKey: uinput.Key0, wantShift: true},

		// Whitespace / control runes
		{name: "space", input: ' ', wantKey: uinput.KeySpace, wantShift: false},
		{name: "newline", input: '\n', wantKey: uinput.KeyEnter, wantShift: false},
		{name: "tab", input: '\t', wantKey: uinput.KeyTab, wantShift: false},

		// Punctuation — unshifted
		{name: "minus", input: '-', wantKey: uinput.KeyMinus, wantShift: false},
		{name: "equals", input: '=', wantKey: uinput.KeyEqual, wantShift: false},
		{name: "left bracket", input: '[', wantKey: uinput.KeyLeftbrace, wantShift: false},
		{name: "right bracket", input: ']', wantKey: uinput.KeyRightbrace, wantShift: false},
		{name: "backslash", input: '\\', wantKey: uinput.KeyBackslash, wantShift: false},
		{name: "semicolon", input: ';', wantKey: uinput.KeySemicolon, wantShift: false},
		{name: "apostrophe", input: '\'', wantKey: uinput.KeyApostrophe, wantShift: false},
		{name: "comma", input: ',', wantKey: uinput.KeyComma, wantShift: false},
		{name: "dot", input: '.', wantKey: uinput.KeyDot, wantShift: false},
		{name: "slash", input: '/', wantKey: uinput.KeySlash, wantShift: false},
		{name: "grave", input: '`', wantKey: uinput.KeyGrave, wantShift: false},

		// Punctuation — shifted variants
		{name: "underscore is shift+minus", input: '_', wantKey: uinput.KeyMinus, wantShift: true},
		{name: "plus is shift+equals", input: '+', wantKey: uinput.KeyEqual, wantShift: true},
		{name: "left brace is shift+[", input: '{', wantKey: uinput.KeyLeftbrace, wantShift: true},
		{name: "right brace is shift+]", input: '}', wantKey: uinput.KeyRightbrace, wantShift: true},
		{name: "pipe is shift+backslash", input: '|', wantKey: uinput.KeyBackslash, wantShift: true},
		{name: "colon is shift+semicolon", input: ':', wantKey: uinput.KeySemicolon, wantShift: true},
		{name: "double-quote is shift+apostrophe", input: '"', wantKey: uinput.KeyApostrophe, wantShift: true},
		{name: "less-than is shift+comma", input: '<', wantKey: uinput.KeyComma, wantShift: true},
		{name: "greater-than is shift+dot", input: '>', wantKey: uinput.KeyDot, wantShift: true},
		{name: "question mark is shift+slash", input: '?', wantKey: uinput.KeySlash, wantShift: true},
		{name: "tilde is shift+grave", input: '~', wantKey: uinput.KeyGrave, wantShift: true},

		// Unknown runes — must return 0, false
		{name: "euro sign unknown", input: '€', wantKey: 0, wantShift: false},
		{name: "emoji unknown", input: '🎉', wantKey: 0, wantShift: false},
		{name: "chinese character unknown", input: '中', wantKey: 0, wantShift: false},
		{name: "null byte unknown", input: 0, wantKey: 0, wantShift: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotShift := runeToKey(tt.input)
			if gotKey != tt.wantKey || gotShift != tt.wantShift {
				t.Errorf("runeToKey(%q) = (%d, %v), want (%d, %v)",
					tt.input, gotKey, gotShift, tt.wantKey, tt.wantShift)
			}
		})
	}
}

// TestParseKeyNameRoundTrip verifies that single-char keys accessible via runeToKey
// are also reachable through parseKeyName with the same key code.
func TestParseKeyNameRoundTrip(t *testing.T) {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	for _, r := range chars {
		r := r
		t.Run(string(r), func(t *testing.T) {
			wantKey, _ := runeToKey(r)
			gotKey, err := parseKeyName(string(r))
			if err != nil {
				t.Errorf("parseKeyName(%q) unexpected error: %v", string(r), err)
				return
			}
			if gotKey != wantKey {
				t.Errorf("parseKeyName(%q) = %d, runeToKey returns %d — mismatch", string(r), gotKey, wantKey)
			}
		})
	}
}
