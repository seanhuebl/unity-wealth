package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/auth"
	"github.com/seanhuebl/unity-wealth/internal/config"
	"github.com/seanhuebl/unity-wealth/middleware"
)

func TestUserAuthMiddleware_TableDriven(t *testing.T) {
	// Create an ApiConfig with the TokenSecret.
	cfg := &config.ApiConfig{
		TokenSecret: "dummysecret",
	}

	// Create an instance of your auth service to generate a valid token.
	authSvc := auth.NewAuthService()
	testUserID := uuid.New()
	validToken, err := authSvc.MakeJWT(testUserID, cfg.TokenSecret, time.Hour)
	if err != nil {
		t.Fatalf("failed to generate valid token: %v", err)
	}

	// Define test cases.
	tests := []struct {
		name           string
		authHeader     string // full value of the Authorization header
		expectedStatus int
		expectedSubstr string // a substring we expect to see in the JSON response
	}{
		{
			name:           "missing token",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedSubstr: `"error"`,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid.token.string",
			expectedStatus: http.StatusUnauthorized,
			expectedSubstr: `{"error":"invalid token"}`,
		},
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectedSubstr: `{"message":"passed"}`,
		},
	}

	// Run each test case.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new Gin engine for this sub-test.
			router := gin.New()

			m := middleware.NewMiddleware(cfg, authSvc)
			// Attach the middleware. (Since our middleware is defined as a method on ApiConfig,
			// it automatically uses cfg.TokenSecret.)
			router.Use(m.UserAuthMiddleware())

			// Define a dummy final handler that will return a JSON response if the request passes.
			router.GET("/test", func(c *gin.Context) {
				// Optionally check that the middleware stored the claims in the context.
				if _, exists := c.Get("claims"); !exists {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "claims not set"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "passed"})
			})

			// Create a test HTTP request.
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rr := httptest.NewRecorder()

			// Serve the HTTP request.
			router.ServeHTTP(rr, req)

			// Compare the expected status code with the actual status code.
			if diff := cmp.Diff(tc.expectedStatus, rr.Code); diff != "" {
				t.Errorf("Status code mismatch (-want +got):\n%s", diff)
			}

			// Check that the response body contains the expected substring.
			if !strings.Contains(rr.Body.String(), tc.expectedSubstr) {
				t.Errorf("Response body does not contain expected substring.\nGot: %s\nWant substring: %s",
					rr.Body.String(), tc.expectedSubstr)
			}
		})
	}
}
