package handlers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/unity-wealth/handlers"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Basic valid email",
			input:    "user@example.com",
			expected: true,
		},
		{
			name:     "Valid email with subdomain",
			input:    "user@mail.example.com",
			expected: true,
		},
		{
			name:     "Valid email with plus sign",
			input:    "user+tag@domain.co",
			expected: true,
		},
		{
			name:     "Valid email with numbers and special chars",
			input:    "my_email123-xyz@test-domain.org",
			expected: true,
		},
		{
			name:     "Missing @ symbol",
			input:    "missingatsymbol.com",
			expected: false,
		},
		{
			name:     "Missing domain",
			input:    "user@",
			expected: false,
		},
		{
			name:     "Missing TLD",
			input:    "user@domain",
			expected: false,
		},
		{
			name:     "Invalid special characters",
			input:    "user()@domain.com",
			expected: false,
		},
		{
			name:     "Leading dot",
			input:    ".user@domain.com",
			expected: false,
		},
		{
			name:     "Trailing dot in domain",
			input:    "user@domain.com.",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handlers.IsValidEmail(tt.input)

			// Use cmp.Diff to compare the booleans and print a helpful diff on mismatch
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("isValidEmail(%q) mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}
