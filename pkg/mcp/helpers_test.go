package mcp

import "testing"

func TestSplitCombo(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"ctrl+c", []string{"ctrl", "c"}},
		{"ctrl+shift+a", []string{"ctrl", "shift", "a"}},
		{"enter", []string{"enter"}},
		{"alt + tab", []string{"alt", "tab"}},
		{"super+1", []string{"super", "1"}},
	}
	for _, tt := range tests {
		got := splitCombo(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("splitCombo(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitCombo(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}
