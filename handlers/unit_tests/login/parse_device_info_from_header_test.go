package handlers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/unity-wealth/handlers"
)

// Example table-driven tests for ParseDeviceInfoFromHeader
func TestParseDeviceInfoFromHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   handlers.DeviceInfo
	}{
		{
			name:   "Empty header",
			header: "",
			want:   handlers.DeviceInfo{},
		},
		{
			name:   "Only OS info",
			header: "os=Android",
			want: handlers.DeviceInfo{
				Os: "Android",
			},
		},
		{
			name:   "Mixed fields with whitespace",
			header: "os = iOS ; browser = Safari ; os_version=14.4  ",
			want: handlers.DeviceInfo{
				Os:             "iOS",
				Browser:        "Safari",
				OsVersion:      "14.4",
				DeviceType:     "", // not provided
				BrowserVersion: "",
			},
		},
		{
			name: "All fields, different order",
			header: "browser=Chrome; device_type=Mobile; os=Android; " +
				"browser_version=95.0; os_version=12",
			want: handlers.DeviceInfo{
				DeviceType:     "Mobile",
				Browser:        "Chrome",
				BrowserVersion: "95.0",
				Os:             "Android",
				OsVersion:      "12",
			},
		},
		{
			name: "Ignore unexpected keys",
			header: "device_type=Desktop; foo=bar; os=Windows; " +
				"browser=Edge; randomKey=value; os_version=10; browser_version=96.0",
			want: handlers.DeviceInfo{
				DeviceType:     "Desktop",
				Os:             "Windows",
				Browser:        "Edge",
				OsVersion:      "10",
				BrowserVersion: "96.0",
			},
		},
		{
			name:   "Single quotes or punctuation in values",
			header: "os='iOS'; browser='Safari'; device_type=Mobile",
			want: handlers.DeviceInfo{
				Os:         "''iOS''", // If SanitizeInput doubles single quotes
				Browser:    "''Safari''",
				DeviceType: "Mobile",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handlers.ParseDeviceInfoFromHeader(tt.header)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseDeviceInfoFromHeader(%q) mismatch (-want +got):\n%s", tt.header, diff)
			}
		})
	}
}
