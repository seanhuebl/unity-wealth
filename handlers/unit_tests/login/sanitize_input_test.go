package handlers

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/unity-wealth/handlers"
)

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No change for typical input",
			input:    "HelloWorld",
			expected: "HelloWorld",
		},
		{
			name:     "Trim leading and trailing whitespace",
			input:    "   Hello World   ",
			expected: "Hello World",
		},
		{
			name:     "Replace single quotes",
			input:    "O'Reilly",
			expected: "O''Reilly",
		},
		{
			name:     "Replace multiple single quotes",
			input:    "'Hello' 'World'",
			expected: "''Hello'' ''World''",
		},
		{
			name:     "Truncate over 100 characters",
			input:    strings.Repeat("A", 105), // 105 'A's
			expected: strings.Repeat("A", 100), // truncated to 100
		},
		{
			name:     "Empty input stays empty",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handlers.SanitizeInput(tt.input)

			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("sanitizeInput(%q) mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}
