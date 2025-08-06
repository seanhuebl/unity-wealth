package auth_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/stretchr/testify/require"

	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
)

func TestIntegrationLogin(t *testing.T) {
	t.Parallel()
	userID := uuid.New()
	tests := []struct {
		name               string
		reqBody            string
		deviceFound        bool
		expErrSubstr       string
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name:               "success: existing device",
			reqBody:            `{"email": "user@example.com", "password": "Validpass1!"}`,
			deviceFound:        true,
			expErrSubstr:       "",
			expectedStatusCode: http.StatusOK,
			expectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"message": "login successful",
				},
			},
		},
		{
			name:               "success: new device",
			reqBody:            `{"email": "user@example.com", "password": "Validpass1!"}`,
			deviceFound:        false,
			expErrSubstr:       "",
			expectedStatusCode: http.StatusOK,
			expectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"message": "login successful",
				},
			},
		},
		{
			name:               "invalid password",
			reqBody:            `{"email": "user@example.com", "password": "Invalidpass1!"}`,
			deviceFound:        false,
			expErrSubstr:       "login failed",
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"error": "login failed",
				},
			},
		},
		{
			name:               "malformed JSON",
			reqBody:            `{"email": "user@example.com", "password": "Validpass1!"`,
			deviceFound:        false,
			expErrSubstr:       "invalid request body",
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"error": "invalid request body",
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			env := testhelpers.SetupTestEnv(t)
			t.Cleanup(func() { env.Db.Close() })

			testhelpers.SeedTestUser(t, env.UserQ, userID, true)

			if tc.deviceFound {
				testhelpers.SeedTestDeviceInfo(t, env.DeviceQ, userID)
			}

			w := httptest.NewRecorder()

			req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(tc.reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Device-Info", "os=Android; os_version=11; device_type=Mobile; browser=Chrome; browser_version=100.0")

			req = req.WithContext(context.WithValue(req.Context(), constants.RequestKey, req))

			env.Router.POST("/login", env.Handlers.AuthHandler.Login)
			env.Router.ServeHTTP(w, req)

			testhelpers.CheckHTTPResponse(t, w, tc.expErrSubstr, tc.expectedStatusCode, tc.expectedResponse, testhelpers.ProcessResponse(w, t))

			if tc.expectedStatusCode == http.StatusOK {
				if !tc.deviceFound {
					deviceId, err := env.DeviceQ.GetDeviceInfoByUser(context.Background(), database.GetDeviceInfoByUserParams{
						UserID:         userID,
						DeviceType:     "Mobile",
						Browser:        "Chrome",
						BrowserVersion: "100.0",
						Os:             "Android",
						OsVersion:      "11",
					})
					require.NoError(t, err)
					require.NotEmpty(t, deviceId, "expected device ID to be not empty")
				}
				result := w.Result()
				defer result.Body.Close()

				var refToken *http.Cookie
				for _, cookie := range result.Cookies() {
					if cookie.Name == "refresh_token" {
						refToken = cookie
						break
					}
				}
				if refToken == nil {
					t.Fatalf("expected refresh_token cookie, got nil")
				}
				if refToken.Value == "" {
					t.Fatalf("expected refresh_token cookie to have a value, got empty")
				}
			}
		})
	}
	t.Run("req not in context", func(t *testing.T) {
		t.Parallel()
		env := testhelpers.SetupTestEnv(t)
		defer env.Db.Close()

		testhelpers.SeedTestUser(t, env.UserQ, userID, true)
		w := httptest.NewRecorder()

		req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(`{"email": "user@example.com", "password": "Validpass1!"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Device-Info", "os=Android; os_version=11; device_type=Mobile; browser=Chrome; browser_version=100.0")

		env.Router.POST("/login", env.Handlers.AuthHandler.Login)
		env.Router.ServeHTTP(w, req)

		testhelpers.CheckHTTPResponse(
			t,
			w,
			"login failed",
			http.StatusUnauthorized,
			map[string]interface{}{
				"data": map[string]interface{}{
					"error": "login failed",
				},
			},
			testhelpers.ProcessResponse(w, t))
	})
}
