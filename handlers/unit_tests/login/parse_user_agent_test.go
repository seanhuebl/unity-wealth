package handlers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/unity-wealth/handlers"
)

const (
	desktopUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4758.80 Safari/537.36"
	mobileUA  = "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148"
)

func TestParseUserAgent(t *testing.T) {
	tests := []struct {
		name       string
		userAgent  string
		wantDevice handlers.DeviceInfo
	}{
		{
			name:      "Desktop user agent",
			userAgent: desktopUA,
			wantDevice: handlers.DeviceInfo{
				DeviceType:     "Desktop",
				Browser:        "Chrome",
				BrowserVersion: "99.0.4758.80", // Example version from the UA
				Os:             "Windows 10",
				OsVersion:      "10", // Always empty in ParseUserAgent
			},
		},
		{
			name:      "Mobile user agent",
			userAgent: mobileUA,
			wantDevice: handlers.DeviceInfo{
				DeviceType:     "Mobile",
				Browser:        "Mobile App", // user_agent library commonly detects Safari on iOS
				BrowserVersion: "",
				Os:             "CPU iPhone OS 14_0 like Mac OS X",
				OsVersion:      "14.0",
			},
		},
		{
			name:      "Empty user agent falls back to Desktop",
			userAgent: "",
			wantDevice: handlers.DeviceInfo{
				DeviceType:     "Desktop", // .Mobile() will be false
				Browser:        "",        // No browser found
				BrowserVersion: "",
				Os:             "",
				OsVersion:      "",
			},
		},
		// Add more scenarios if needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handlers.ParseUserAgent(tt.userAgent)

			// Because user_agent library can produce slightly different version strings,
			// we might want to see what it yields in practice. Or use placeholders if your library is consistent.

			// For demonstration, we do a straight cmp:
			if diff := cmp.Diff(tt.wantDevice, got); diff != "" {
				t.Errorf("ParseUserAgent(%q) mismatch (-want +got):\n%s", tt.userAgent, diff)
			}
		})
	}
}
