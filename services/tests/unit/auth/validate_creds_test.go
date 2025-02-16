// Not services_test b/c of need to test unexported helpers
package services

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/mocks"
	"github.com/seanhuebl/unity-wealth/models"
	"github.com/seanhuebl/unity-wealth/services"
	"github.com/stretchr/testify/mock"
)

// validateCredentials does not use authInterface so we don't need to bring in the full mock
type testAuthSvc struct {
	*services.AuthService

	checkPasswordHash func(password, hash string) error
}

func (t *testAuthSvc) CheckPasswordHash(password, hash string) error {
	return t.checkPasswordHash(password, hash)
}

func TestValidateCredentials(t *testing.T) {
	tests := []struct {
		name                   string
		input                  models.LoginInput
		getUserByEmailResult   database.GetUserByEmailRow
		getUserByEmailError    error
		checkPasswordHashError error
		expectedErrorSubstring string
	}{
		{
			name: "successful credentials",
			input: models.LoginInput{
				Email:    "valid@example.com",
				Password: "Validpass1!",
			},
			getUserByEmailResult: database.GetUserByEmailRow{
				ID:             uuid.NewString(),
				HashedPassword: "hashedpassword",
			},
			getUserByEmailError:    nil,
			checkPasswordHashError: nil,
			expectedErrorSubstring: "",
		},
		{
			name: "user not found",
			input: models.LoginInput{
				Email:    "valid@example.com",
				Password: "Validpass1!",
			},
			getUserByEmailError:    sql.ErrNoRows,
			expectedErrorSubstring: "invalid email / password",
		},
		{
			name: "error fetching user",
			input: models.LoginInput{
				Email:    "valid@example.com",
				Password: "Validpass1!",
			},
			getUserByEmailError:    errors.New("db error"),
			expectedErrorSubstring: "failed to fetch user",
		},
		{
			name: "password check fails",
			input: models.LoginInput{
				Email:    "valid@example.com",
				Password: "Validpass1!",
			},
			getUserByEmailResult: database.GetUserByEmailRow{
				ID:             uuid.NewString(),
				HashedPassword: "hashedpassword",
			},
			getUserByEmailError:    nil,
			checkPasswordHashError: errors.New("invalid password"),
			expectedErrorSubstring: "invalid email / password",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockQ := mocks.NewQuerier(t)

			mockQ.On("GetUserByEmail", mock.Anything, tc.input.Email).Return(tc.getUserByEmailResult, tc.getUserByEmailError)

			authSvc := services.NewAuthService("", "", mockQ)

			testAuthSvc := &testAuthSvc{
				AuthService: authSvc,
				checkPasswordHash: func(password, hash string) error {
					return tc.checkPasswordHashError
				},
			}
			userID, err := testAuthSvc.validateCredentials
		})
	}
}
