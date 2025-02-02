package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/unity-wealth/handlers"
)

// Example table-driven tests for GetDeviceInfo.
func TestGetDeviceInfo(t *testing.T) {
	tests := []struct {
		name        string
		xDeviceInfo string
		userAgent   string
		wantDevice  handlers.DeviceInfo
		wantErr     error
	}{
		{
			name:        "Valid X-Device-Info, skip User-Agent fallback",
			xDeviceInfo: "os=Android; device_type=Mobile; browser=Chrome; browser_version=95.0",
			userAgent:   "Some UA string that won't be used",
			wantDevice: handlers.DeviceInfo{
				Os:             "Android",
				DeviceType:     "Mobile",
				Browser:        "Chrome",
				BrowserVersion: "95.0",
			},
			wantErr: nil,
		},
		{
			name:        "Invalid X-Device-Info, fallback to valid User-Agent",
			xDeviceInfo: "os=UnknownOS; device_type=Tablet", // Suppose "Tablet" is invalid
			userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/99.0.4758.80 Safari/537.36",
			wantDevice: handlers.DeviceInfo{
				DeviceType:     "Desktop", // from ParseUserAgent (assuming your library)
				Browser:        "Chrome",
				BrowserVersion: "99.0.4758.80",
				Os:             "Windows 10", // or "Windows NT 10.0", depending on library
				OsVersion:      "10",
			},
			wantErr: nil,
		},
		{
			name:        "Empty X-Device-Info, valid User-Agent",
			xDeviceInfo: "",
			userAgent:   "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148",
			wantDevice: handlers.DeviceInfo{
				DeviceType:     "Mobile",
				Browser:        "Mobile App",
				BrowserVersion: "",
				Os:             "CPU iPhone OS 14_0 like Mac OS X",
				OsVersion:      "14.0",
			},
			wantErr: nil,
		},
		{
			name:        "Invalid X-Device-Info, invalid User-Agent => error",
			xDeviceInfo: "os=??; device_type=Watch", // Suppose Watch is invalid
			userAgent:   "Some weird UA that can't parse properly",
			wantDevice:  handlers.DeviceInfo{},
			wantErr:     errors.New("invalid or unknown device information"),
		},
		{
			name:        "No headers => error",
			xDeviceInfo: "",
			userAgent:   "",
			wantDevice:  handlers.DeviceInfo{},
			wantErr:     errors.New("invalid or unknown device information"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.xDeviceInfo != "" {
				req.Header.Set("X-Device-Info", tt.xDeviceInfo)
			}
			if tt.userAgent != "" {
				req.Header.Set("User-Agent", tt.userAgent)
			}

			gotDevice, gotErr := handlers.GetDeviceInfo(req)

			// Compare handlers.DeviceInfo
			if diff := cmp.Diff(tt.wantDevice, gotDevice); diff != "" {
				t.Errorf("handlers.DeviceInfo mismatch (-want +got):\n%s", diff)
			}

			// Compare errors
			switch {
			case tt.wantErr == nil && gotErr != nil:
				t.Errorf("Expected no error, got %v", gotErr)
			case tt.wantErr != nil && gotErr == nil:
				t.Errorf("Expected error %v, got nil", tt.wantErr)
			case tt.wantErr != nil && gotErr != nil:
				// Compare error messages or use a custom comparer
				if tt.wantErr.Error() != gotErr.Error() {
					t.Errorf("Error mismatch, want %q, got %q", tt.wantErr.Error(), gotErr.Error())
				}
			}
		})
	}
}
