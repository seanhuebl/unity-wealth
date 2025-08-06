package user_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/seanhuebl/unity-wealth/internal/services/user"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
	"github.com/stretchr/testify/require"
)

func TestIntSignup(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			env := testhelpers.SetupTestEnv(t)
			defer env.Db.Close()
			req := httptest.NewRequest("POST", "/signup", bytes.NewBufferString(tc.ReqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			env.Router.POST("/signup", env.Handlers.UserHandler.SignUp)
			env.Router.ServeHTTP(w, req)
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.WantErrSubstr, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
		})
	}
	t.Run("create user failure", func(t *testing.T) {
		t.Parallel()
		env := testhelpers.SetupTestEnv(t)
		tc := testmodels.SignUpTest{
			ReqBody:            `{"email": "dupe@example.com", "password": "Validpass1!"}`,
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
		require.NoError(t, env.Services.UserService.SignUp(context.Background(), input))
		w := httptest.NewRecorder()

		env.Router.POST("/signup", env.Handlers.UserHandler.SignUp)
		env.Router.ServeHTTP(w, req)
		actualResponse := testhelpers.ProcessResponse(w, t)
		testhelpers.CheckHTTPResponse(t, w, tc.WantErrSubstr, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
	})
}
