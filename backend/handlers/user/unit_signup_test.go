package user_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/handlers/user"
	handlermocks "github.com/seanhuebl/unity-wealth/internal/mocks/handlers"
	"github.com/seanhuebl/unity-wealth/internal/sentinels"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/stretchr/testify/mock"
)

func TestAddUserHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		reqBody            string
		err                error
		expectedError      string
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name:               "invalid email",
			reqBody:            `{"email": "invalid", "password": "Validpass1!"}`,
			err:                sentinels.ErrInvalidEmail,
			expectedError:      "invalid email",
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"error": "invalid email",
				},
			},
		},
		{
			name:               "invalid password",
			reqBody:            `{"email": "valid@example.com", "password": "invalid"}`,
			err:                sentinels.ErrInvalidPassword,
			expectedError:      "invalid password",
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"error": "invalid password",
				},
			},
		},
		{
			name:               "hash password error",
			reqBody:            `{"email": "valid@example.com", "password": "Validpass1!"}`,
			err:                errors.New("hash error"),
			expectedError:      "internal server error",
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"error": "internal server error",
				},
			},
		},
		{
			name:               "create user error",
			reqBody:            `{"email": "valid@example.com", "password": "Validpass1!"}`,
			err:                errors.New("create user error"),
			expectedError:      "internal server error",
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"error": "internal server error",
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := handlermocks.NewUserService(t)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/signup", bytes.NewBufferString(tc.reqBody))
			req.Header.Set("Content-Type", "application/json")

			mockSvc.On("SignUp", mock.Anything, mock.AnythingOfType("user.SignUpInput")).Return(tc.err)
			h := user.NewHandler(mockSvc)
			router := gin.New()

			router.POST("/signup", h.SignUp)
			router.ServeHTTP(w, req)

			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.expectedError, tc.expectedStatusCode, tc.expectedResponse, actualResponse)
			mockSvc.AssertExpectations(t)
		})
	}
	t.Run("invalid req body", func(t *testing.T) {
		mockSvc := handlermocks.NewUserService(t)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/signup", bytes.NewBufferString(`{"email": "valid@example.com", "password": "ValidPass1!"`))
		h := user.NewHandler(mockSvc)
		router := gin.New()
		router.POST("/signup", h.SignUp)
		router.ServeHTTP(w, req)

		actualResponse := testhelpers.ProcessResponse(w, t)
		testhelpers.CheckHTTPResponse(
			t,
			w,
			"invalid request",
			http.StatusBadRequest,
			map[string]interface{}{
				"data": map[string]interface{}{
					"error": "invalid request",
				},
			},
			actualResponse,
		)
	})
}
