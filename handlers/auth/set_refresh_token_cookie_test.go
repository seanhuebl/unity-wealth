package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetRefreshTokenCookie(t *testing.T) {
	// Initialize Gin in test mode
	gin.SetMode(gin.TestMode)

	// Define test cases
	tests := []struct {
		name             string
		env              string
		cookieDomainEnv  string
		refreshToken     string
		expectedDomain   string
		expectedSecure   bool
		expectedHttpOnly bool
	}{
		{
			name:             "Production environment",
			env:              "prod",
			cookieDomainEnv:  "example.com",
			refreshToken:     "prod-token",
			expectedDomain:   "example.com",
			expectedSecure:   true,
			expectedHttpOnly: true,
		},
		{
			name:             "Non-production environment",
			env:              "dev",
			cookieDomainEnv:  "",
			refreshToken:     "dev-token",
			expectedDomain:   "localhost",
			expectedSecure:   false,
			expectedHttpOnly: true,
		},
		{
			name:             "Custom cookie domain",
			env:              "prod",
			cookieDomainEnv:  "custom.com",
			refreshToken:     "custom-token",
			expectedDomain:   "custom.com",
			expectedSecure:   true,
			expectedHttpOnly: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Setenv("ENV", tt.env)
			os.Setenv("COOKIE_DOMAIN", tt.cookieDomainEnv)


			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)


			SetRefreshTokenCookie(ctx, tt.refreshToken)


			result := recorder.Result()
			cookies := result.Cookies()

			assert.Len(t, cookies, 1)
			cookie := cookies[0]

			assert.Equal(t, "refresh_token", cookie.Name)
			assert.Equal(t, tt.refreshToken, cookie.Value)
			assert.Equal(t, tt.expectedDomain, cookie.Domain)
			assert.Equal(t, "/", cookie.Path)
			assert.Equal(t, tt.expectedSecure, cookie.Secure)
			assert.Equal(t, tt.expectedHttpOnly, cookie.HttpOnly)
			assert.WithinDuration(t, time.Now().Add(7*24*time.Hour), cookie.Expires, time.Second)
			assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
		})
	}
}
