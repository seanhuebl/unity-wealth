package user_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	httpuser "github.com/seanhuebl/unity-wealth/handlers/user"
	"github.com/seanhuebl/unity-wealth/internal/database"
	authmocks "github.com/seanhuebl/unity-wealth/internal/mocks/auth"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
	"github.com/seanhuebl/unity-wealth/internal/services/user"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAddUserHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		reqBody            string
		validPasswordError error
		hashPasswordOutput string
		hashPasswordError  error
		createUserError    error
		expectedError      string
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name:               "successful sign up",
			reqBody:            `{"email": "valid@example.com", "password": "Validpass1!"}`,
			validPasswordError: nil,
			hashPasswordOutput: "hashedpassword",
			hashPasswordError:  nil,
			createUserError:    nil,
			expectedError:      "",
			expectedStatusCode: http.StatusCreated,
			expectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"message": "sign up successful!",
					"email":   "valid@example.com",
				},
			},
		},
		{
			name:               "invalid req body",
			reqBody:            `{"email": "valid@example.com", "password": "ValidPass1!"`,
			expectedError:      "invalid request",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "invalid email",
			reqBody:            `{"email": "invalid", "password": "Validpass1!"}`,
			expectedError:      "invalid email",
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:               "invalid password",
			reqBody:            `{"email": "valid@example.com", "password": "invalid"}`,
			validPasswordError: errors.New("passwords must contain at least one uppercase letter"),
			expectedError:      "invalid password",
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:               "hash password error",
			reqBody:            `{"email": "valid@example.com", "password": "Validpass1!"}`,
			validPasswordError: nil,
			hashPasswordError:  errors.New("hash error"),
			expectedError:      "internal server error",
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:               "create user error",
			reqBody:            `{"email": "valid@example.com", "password": "Validpass1!"}`,
			validPasswordError: nil,
			hashPasswordOutput: "hashedpassword",
			hashPasswordError:  nil,
			createUserError:    errors.New("db error"),
			expectedError:      "internal server error",
			expectedStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockUserQ := dbmocks.NewUserQuerier(t)
			mockPwdHasher := authmocks.NewPasswordHasher(t)
			if json.Valid([]byte(tc.reqBody)) {
				if auth.IsValidEmail(testhelpers.GetEmailFromBody(tc.reqBody)) {
					if tc.validPasswordError == nil {
						mockPwdHasher.On("HashPassword", testhelpers.GetPasswordFromBody(tc.reqBody)).Return(tc.hashPasswordOutput, tc.hashPasswordError)
						if tc.hashPasswordError == nil {
							mockUserQ.On("CreateUser", mock.Anything, mock.MatchedBy(func(params database.CreateUserParams) bool {
								expected := database.CreateUserParams{
									Email:          testhelpers.GetEmailFromBody(tc.reqBody),
									HashedPassword: tc.hashPasswordOutput,
								}

								diff := cmp.Diff(expected, params, cmpopts.IgnoreFields(database.CreateUserParams{}, "ID"))
								if diff != "" {
									t.Logf("CreateUserParams mismatch (-want, +got):\n%s", diff)
									return false
								}
								return params.ID != ""
							})).Return(tc.createUserError)
						}
					}
				}
			}
			userSvc := user.NewUserService(mockUserQ, mockPwdHasher)
			h := httpuser.NewHandler(userSvc)
			req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(tc.reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router := gin.New()
			router.POST("/signup", h.SignUp)
			router.ServeHTTP(w, req)

			var actualResponse map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
			require.NoError(t, err)
			if tc.expectedError != "" {
				data := actualResponse["data"].(map[string]interface{})
				require.Contains(t, data["error"].(string), tc.expectedError)
			} else {
				expected := tc.expectedResponse
				if diff := cmp.Diff(expected, actualResponse); diff != "" {
					t.Errorf("response mismatch (-want, +got):\n%s", diff)
				}
			}
			mockUserQ.AssertExpectations(t)
			mockPwdHasher.AssertExpectations(t)
		})
	}
}
