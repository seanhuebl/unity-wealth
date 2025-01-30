package handlers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/unity-wealth/handlers"
)

func TestIsValidDeviceInfo(t *testing.T) {
	tests := []struct {
		name     string
		info     handlers.DeviceInfo
		expected bool
	}{
		{
			name: "Valid Desktop Device",
			info: handlers.DeviceInfo{
				DeviceType:     "Desktop",
				Browser:        "Chrome",
				BrowserVersion: "99.0",
				Os:             "Windows",
				OsVersion:      "10.0",
			},
			expected: true,
		},
		{
			name: "Valid Mobile Device",
			info: handlers.DeviceInfo{
				DeviceType:     "Mobile",
				Browser:        "Safari",
				BrowserVersion: "14.1",
				Os:             "iOS",
				OsVersion:      "14.5",
			},
			expected: true,
		},
		{
			name: "Invalid DeviceType",
			info: handlers.DeviceInfo{
				DeviceType:     "Tablet",
				Browser:        "Chrome",
				BrowserVersion: "99.0",
				Os:             "Android",
				OsVersion:      "12.0",
			},
			expected: false,
		},
		{
			name: "Missing OS",
			info: handlers.DeviceInfo{
				DeviceType:     "Desktop",
				Browser:        "Firefox",
				BrowserVersion: "91.0",
				Os:             "",
				OsVersion:      "11.0",
			},
			expected: false,
		},
		{
			name: "Missing Browser",
			info: handlers.DeviceInfo{
				DeviceType:     "Desktop",
				Browser:        "",
				BrowserVersion: "95.0",
				Os:             "Windows",
				OsVersion:      "10.0",
			},
			expected: false,
		},
		{
			name: "Invalid BrowserVersion",
			info: handlers.DeviceInfo{
				DeviceType:     "Mobile",
				Browser:        "Chrome",
				BrowserVersion: "95.x", // not valid per isValidVersion
				Os:             "Android",
				OsVersion:      "12.1",
			},
			expected: false,
		},
		{
			name: "Invalid OSVersion",
			info: handlers.DeviceInfo{
				DeviceType:     "Mobile",
				Browser:        "Chrome",
				BrowserVersion: "95.0",
				Os:             "Android",
				OsVersion:      "12x", // not valid
			},
			expected: false,
		},
		{
			name: "Empty Fields",
			info: handlers.DeviceInfo{
				DeviceType:     "Desktop",
				Browser:        "",
				BrowserVersion: "",
				Os:             "",
				OsVersion:      "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handlers.IsValidDeviceInfo(tt.info)

			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("isValidDeviceInfo(%+v) mismatch (-want +got):\n%s", tt.info, diff)
			}
		})
	}
}
