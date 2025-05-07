package user_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	httpuser "github.com/seanhuebl/unity-wealth/handlers/user"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
	"github.com/seanhuebl/unity-wealth/internal/services/user"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
	"github.com/stretchr/testify/require"
)

func TestIntSignup(t *testing.T) {
	tests := []testmodels.SignUpTest{
		{
			Name:               "success",
			ReqBody:            `{"email": "valid@example.com", "password": "Validpass1!"}`,
			ExpectedStatusCode: http.StatusCreated,
			ExpectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"message": "sign up successful!",
					"email":   "valid@example.com",
				},
			},
		},
		{
			Name:               "invalid req body",
			ReqBody:            `{"email": "valid@example.com", "password": "ValidPass1!"`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"error": "invalid request",
				},
			},
		},
		{
			Name:               "invalid email",
			ReqBody:            `{"email": "invalid", "password": "Validpass1!"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"error": "invalid email",
				},
			},
		},
		{
			Name:               "invalid password",
			ReqBody:            `{"email": "valid@example.com", "password": "invalid"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"error": "invalid password",
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			env := testhelpers.SetupTestEnv(t)
			userSvc := user.NewUserService(env.UserQ, auth.NewRealPwdHasher())
			req := httptest.NewRequest("POST", "/signup", bytes.NewBufferString(tc.ReqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h := httpuser.NewHandler(userSvc)
			env.Router.POST("/signup", h.SignUp)
			env.Router.ServeHTTP(w, req)
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckUserHTTPResponse(t, w, tc, actualResponse)
		})
	}
	t.Run("create user failure", func(t *testing.T) {
		env := testhelpers.SetupTestEnv(t)
		userSvc := user.NewUserService(env.UserQ, auth.NewRealPwdHasher())
		tc := testmodels.SignUpTest{
			Name:               "create user failure",
			ReqBody: `{"email": "dupe@example.com", "password": "Validpass1!"}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"error": "internal server error",
				},
			},
		}
		req := httptest.NewRequest("POST", "/signup", bytes.NewBufferString(tc.ReqBody))
		req.Header.Set("Content-Type", "application/json")

		input := user.SignUpInput{Email: "dupe@example.com", Password: "Validpass1!"}
		require.NoError(t, userSvc.SignUp(context.Background(), input))
		w := httptest.NewRecorder()
		h := httpuser.NewHandler(userSvc)
		env.Router.POST("/signup", h.SignUp)
		env.Router.ServeHTTP(w, req)
		actualResponse := testhelpers.ProcessResponse(w, t)
		testhelpers.CheckUserHTTPResponse(t, w, tc, actualResponse)
	})
}
