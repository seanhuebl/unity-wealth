package handlers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/unity-wealth/handlers"
)

func TestIsValidVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "Simple integer version",
			version:  "1",
			expected: true,
		},
		{
			name:     "Multiple segments",
			version:  "10.0.1",
			expected: true,
		},
		{
			name:     "Browser-like version",
			version:  "95.0.4638.69",
			expected: true,
		},
		{
			name:     "Leading zero is allowed",
			version:  "01.02.003",
			expected: true,
		},
		{
			name:     "Empty string",
			version:  "",
			expected: false,
		},
		{
			name:     "Trailing dot",
			version:  "1.0.",
			expected: false,
		},
		{
			name:     "Non-numeric segment",
			version:  "1.0a.3",
			expected: false,
		},
		{
			name:     "Spaces are invalid",
			version:  "1. 2.3",
			expected: false,
		},
		{
			name:     "Completely non-numeric",
			version:  "abc",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handlers.IsValidVersion(tt.version)

			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("isValidVersion(%q) mismatch (-want +got):\n%s", tt.version, diff)
			}
		})
	}
}
