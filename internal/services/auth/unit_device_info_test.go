package auth_test

import (
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
	"github.com/stretchr/testify/require"
)

func TestParseUserAgent(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expected  models.DeviceInfo
	}{
		{
			name: "Desktop Chrome on Windows",
			// A common Chrome UA on Windows 10.
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36",
			expected: models.DeviceInfo{
				DeviceType:     "Desktop",
				Browser:        "Chrome",
				BrowserVersion: "90.0.4430.212",
				Os:             "Windows 10",
				OsVersion:      "10",
			},
		},
		{
			name: "Mobile Safari on iPhone",
			// A typical UA for an iPhone.
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
			expected: models.DeviceInfo{
				DeviceType:     "Mobile",
				Browser:        "Safari",
				BrowserVersion: "14.0",
				Os:             "CPU iPhone OS 14_6 like Mac OS X",
				OsVersion:      "14.6",
			},
		},
		{
			name:      "empty user agent",
			userAgent: "",
			expected: models.DeviceInfo{
				DeviceType:     "Desktop",
				Browser:        "",
				BrowserVersion: "",
				Os:             "",
				OsVersion:      "",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := auth.ParseUserAgent(tc.userAgent)
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Errorf("ParseUserAgent(%q) mismatch (-want +got):\n%s", tc.userAgent, diff)
			}
		})
	}
}

func TestParseDeviceInfoFromHeader(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected models.DeviceInfo
	}{
		{
			name:   "all fields provided",
			header: "os=Android; os_version=11; device_type=Mobile; browser=Chrome; browser_version=100.0",
			expected: models.DeviceInfo{
				Os:             "Android",
				OsVersion:      "11",
				DeviceType:     "Mobile",
				Browser:        "Chrome",
				BrowserVersion: "100.0",
			},
		},
		{
			name:   "extra whitespace and mixed case keys",
			header: " OS = android ; Os_Version = 10 ; device_type = Tablet ; Browser = Firefox ; Browser_Version = 85.0 ",
			expected: models.DeviceInfo{
				Os:             "android",
				OsVersion:      "10",
				DeviceType:     "Tablet",
				Browser:        "Firefox",
				BrowserVersion: "85.0",
			},
		},
		{
			name:   "missing some keys",
			header: "os=IOS; device_type=Mobile",
			expected: models.DeviceInfo{
				Os:         "IOS",
				DeviceType: "Mobile",
			},
		},
		{
			name:     "empty header",
			header:   "",
			expected: models.DeviceInfo{},
		},
		{
			name:   "irrelevant key is ignored",
			header: "os=Linux; extra=value; browser=Opera",
			expected: models.DeviceInfo{
				Os:      "Linux",
				Browser: "Opera",
			},
		},
		{
			name:   "improper format is skipped",
			header: "os=Windows; badformat; browser=Edge",
			expected: models.DeviceInfo{
				Os:      "Windows",
				Browser: "Edge",
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := auth.ParseDeviceInfoFromHeader(tc.header)
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Errorf("ParseDeviceInfoFromHeader(%q) mismatch (-want +got)\n%s", tc.header, diff)
			}
		})
	}
}

func TestIsValidDeviceInfo(t *testing.T) {
	tests := []struct {
		name     string
		info     models.DeviceInfo
		expected bool
	}{
		{
			name: "valid desktop device with versions",
			info: models.DeviceInfo{
				DeviceType:     "Desktop",
				Browser:        "Chrome",
				BrowserVersion: "90.0.4430.212",
				Os:             "Windows",
				OsVersion:      "10.0",
			},
			expected: true,
		},
		{
			name: "valid mobile device with empty versions",
			info: models.DeviceInfo{
				DeviceType: "Mobile",
				Browser:    "Safari",
				Os:         "iOS",
			},
			expected: true,
		},
		{
			name: "invalid device type",
			info: models.DeviceInfo{
				DeviceType: "Tablet",
				Browser:    "Chrome",
				Os:         "Android",
			},
			expected: false,
		},
		{
			name: "missing browser",
			info: models.DeviceInfo{
				DeviceType: "Desktop",
				Browser:    "",
				Os:         "Linux",
			},
			expected: false,
		},
		{
			name: "missing OS",
			info: models.DeviceInfo{
				DeviceType: "Mobile",
				Browser:    "Firefox",
				Os:         "",
			},
			expected: false,
		},
		{
			name: "invalid browser version",
			info: models.DeviceInfo{
				DeviceType:     "Desktop",
				Browser:        "Chrome",
				BrowserVersion: "90.0a",
				Os:             "Windows",
				OsVersion:      "10",
			},
			expected: false,
		},
		{
			name: "invalid OS version",
			info: models.DeviceInfo{
				DeviceType:     "Mobile",
				Browser:        "Safari",
				BrowserVersion: "14.0",
				Os:             "iOS",
				OsVersion:      "14a",
			},
			expected: false,
		},
		{
			name: "valid device with mixed case type",
			info: models.DeviceInfo{
				DeviceType:     "DeSkToP",
				Browser:        "Edge",
				BrowserVersion: "91.0",
				Os:             "Windows",
				OsVersion:      "10",
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			result := auth.IsValidDeviceInfo(tc.info)
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Errorf("IsValidDeviceInfo() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetDeviceInfoFromRequest(t *testing.T) {
	tests := []struct {
		name              string
		headers           map[string]string
		expected          models.DeviceInfo
		expectedErrSubstr string
	}{
		{
			name: "valid X-Device-Info header",
			headers: map[string]string{
				"X-Device-Info": "os=Android; os_version=11; device_type=Mobile; browser=Chrome; browser_version=100.0",
			},
			expected: models.DeviceInfo{
				Os:             "Android",
				OsVersion:      "11",
				DeviceType:     "Mobile",
				Browser:        "Chrome",
				BrowserVersion: "100.0",
			},
			expectedErrSubstr: "",
		},
		{
			name: "invalid X-Device-Info header, fallback to valid User-Agent header",
			headers: map[string]string{
				// This header is invalid or incomplete so that isValidDeviceInfo returns false.
				"X-Device-Info": "invalid",
				"User-Agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) " +
					"AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
			},
			// Expected values depend on how parseUserAgent interprets this UA.
			// Adjust these values based on the actual behavior of mssola/user_agent.
			expected: models.DeviceInfo{
				DeviceType:     "Mobile",
				Browser:        "Safari",
				BrowserVersion: "14.0",
				Os:             "CPU iPhone OS 14_6 like Mac OS X",
				OsVersion:      "14.6",
			},
			expectedErrSubstr: "",
		},
		{
			name: "no headers provided",
			headers: map[string]string{
				"X-Device-Info": "",
				"User-Agent":    "",
			},
			expected:          models.DeviceInfo{},
			expectedErrSubstr: "invalid or unknown device information",
		},
		{
			name: "invalid X-Device-Info and invalid User-Agent",
			headers: map[string]string{
				"X-Device-Info": "os=Android", // incomplete information
				"User-Agent":    "invalid user agent",
			},
			expected:          models.DeviceInfo{},
			expectedErrSubstr: "invalid or unknown device information",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Create a new HTTP request using httptest.
			req := httptest.NewRequest("GET", "/", nil)
			// Set headers according to the test case.
			for key, value := range tc.headers {
				req.Header.Set(key, value)
			}
			// Call the function under test.
			deviceInfo, err := auth.GetDeviceInfoFromRequest(req)
			if tc.expectedErrSubstr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrSubstr)
			} else {
				require.NoError(t, err)
				if diff := cmp.Diff(tc.expected, deviceInfo); diff != "" {
					t.Errorf("GetDeviceInfoFromRequest() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
