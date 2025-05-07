package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/seanhuebl/unity-wealth/internal/database"
	authmocks "github.com/seanhuebl/unity-wealth/internal/mocks/auth"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
	"github.com/seanhuebl/unity-wealth/internal/services/user"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSignup(t *testing.T) {
	tests := []struct {
		name                  string
		input                 user.SignUpInput
		validatePasswordError error
		hashPasswordOutput    string
		hashPasswordError     error
		createUserError       error
		expectedError         string
	}{
		{
			name: "success",
			input: user.SignUpInput{
				Email:    "valid@example.com",
				Password: "Validpass1!",
			},
			validatePasswordError: nil,
			hashPasswordOutput:    "hashedpassword",
			hashPasswordError:     nil,
			createUserError:       nil,
			expectedError:         "",
		},
		{
			name: "invalid email",
			input: user.SignUpInput{
				Email:    "invalid",
				Password: "Validpass1!",
			},
			expectedError: auth.ErrInvalidEmail.Error(),
		},
		{
			name: "invalid password",
			input: user.SignUpInput{
				Email:    "valid@example.com",
				Password: "invalid",
			},
			validatePasswordError: errors.New("password must contain one uppercase letter"),
			expectedError:         auth.ErrInvalidPassword.Error(),
		},
		{
			name: "hashing failure",
			input: user.SignUpInput{
				Email:    "valid@example.com",
				Password: "Validpass1!",
			},
			validatePasswordError: nil,
			hashPasswordError:     errors.New("hash error"),
			expectedError:         "failed to hash password",
		},
		{
			name: "create user failure",
			input: user.SignUpInput{
				Email:    "valid@example.com",
				Password: "Validpass1!",
			},
			validatePasswordError: nil,
			hashPasswordOutput:    "hashedpassword",
			hashPasswordError:     nil,
			createUserError:       errors.New("db error"),
			expectedError:         "unable to create user",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockUserQ := dbmocks.NewUserQuerier(t)
			mockPwdHasher := authmocks.NewPasswordHasher(t)
			userSvc := user.NewUserService(mockUserQ, mockPwdHasher)
			if auth.IsValidEmail(tc.input.Email) {
				err := auth.ValidatePassword(tc.input.Password)

				if err == nil {
					mockPwdHasher.On("HashPassword", tc.input.Password).Return(tc.hashPasswordOutput, tc.hashPasswordError)
					if tc.hashPasswordError == nil {
						mockUserQ.On("CreateUser", mock.Anything, mock.MatchedBy(func(params database.CreateUserParams) bool {
							expected := database.CreateUserParams{
								Email:          tc.input.Email,
								HashedPassword: tc.hashPasswordOutput,
							}
							diff := cmp.Diff(expected, params, cmpopts.IgnoreFields(database.CreateUserParams{}, "ID"))
							if diff != "" {
								t.Logf("CreateUserParams mismatch (-want +got):\n%s", diff)
								return false
							}
							return params.ID != ""
						})).Return(tc.createUserError)
					}
				}
			}

			err := userSvc.SignUp(context.Background(), tc.input)

			if tc.expectedError == "" {
				if tc.expectedError == auth.ErrInvalidEmail.Error() {
					require.ErrorIs(t, err, auth.ErrInvalidEmail)
				} else if tc.expectedError == auth.ErrInvalidPassword.Error() {
					require.ErrorIs(t, err, auth.ErrInvalidPassword)
				} else {
					require.NoError(t, err, "expected no error, but got one")
				}
			} else {
				require.Error(t, err, "expected error, but got nil")
				require.ErrorContains(t, err, tc.expectedError)
			}

			mockUserQ.AssertExpectations(t)
			mockPwdHasher.AssertExpectations(t)

		})
	}
}
