package quicknote

import (
	"testing"
)

func TestDetermineTargetEmail(t *testing.T) {

	tests := []struct {
		subject  string
		expected string
	}{
		{"w something", "work"},
		{"w test", "work"},
		{"W test", "work"},
		{"personal note", "private"},
		{"", "private"},
		{"w", "private"},
	}

	for _, tt := range tests {
		got := determineTargetEmail(tt.subject)
		if got != tt.expected {
			t.Errorf("determineTargetEmail(%q) = %q; want %q", tt.subject, got, tt.expected)
		}
	}
}
