package auth_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	authhttp "github.com/seanhuebl/unity-wealth/handlers/auth"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	handlermocks "github.com/seanhuebl/unity-wealth/internal/mocks/handlers"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	tests := []struct {
		name               string
		deviceFound        bool
		reqBody            string
		expErrSubstr       string
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name:               "invalid request body: malformed JSON",
			reqBody:            `{"email": "user@example.com", "password": "Validpass1!"`,
			expErrSubstr:       "invalid request body",
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"error": "invalid request body",
				},
			},
		},
		{
			name:               "service error: invalid email format",
			reqBody:            `{"email": "user", "password": "Validpass1!"}`,
			expErrSubstr:       "login failed",
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"error": "login failed",
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(tc.reqBody))
			req.Header.Set("X-Device-Info", "os=Android; os_version=11; device_type=Mobile; browser=Chrome; browser_version=100.0")
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.WithValue(req.Context(), constants.RequestKey, req))
			router := gin.New()
			mockSvc := handlermocks.NewAuthService(t)

			if tc.expectedStatusCode == 401 {
				mockSvc.On("Login", mock.Anything, mock.Anything).Return(models.LoginResponse{}, errors.New(tc.expErrSubstr))
			}

			h := authhttp.NewHandler(mockSvc)
			router.POST("/login", h.Login)

			router.ServeHTTP(w, req)
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.expErrSubstr, tc.expectedStatusCode, tc.expectedResponse, actualResponse)
			mockSvc.AssertExpectations(t)
		})
	}
	t.Run("login successful", func(t *testing.T) {

		w := httptest.NewRecorder()

		req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(`{"email": "user@example.com", "password": "Validpass1!"}`))
		req.Header.Set("X-Device-Info", "os=Android; os_version=11; device_type=Mobile; browser=Chrome; browser_version=100.0")
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), constants.RequestKey, req))

		router := gin.New()
		mockSvc := handlermocks.NewAuthService(t)
		h := authhttp.NewHandler(mockSvc)
		router.POST("/login", h.Login)

		mockSvc.
			On("Login", mock.Anything, models.LoginInput{
				Email:    "user@example.com",
				Password: "Validpass1!",
			}).
			Return(models.LoginResponse{
				UserID:       uuid.New(),
				RefreshToken: "refresh",
				JWTToken:     "dummytoken",
			}, nil).Once()

		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		resp := w.Result()
		defer resp.Body.Close()

		var refreshCookie *http.Cookie
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "refresh_token" {
				refreshCookie = cookie
				break
			}
		}

		require.NotNil(t, refreshCookie)
		require.Equal(t, "refresh", refreshCookie.Value)
		mockSvc.AssertExpectations(t)
	})
}
