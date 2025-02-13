package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/mocks"
	"github.com/seanhuebl/unity-wealth/models"
	"github.com/seanhuebl/unity-wealth/services"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSignup(t *testing.T) {
	tests := []struct {
		name                  string
		input                 models.SignUpInput
		validatePasswordError error
		hashPasswordOutput    string
		hashPasswordError     error
		createUserError       error
		expectedError         string
	}{
		{
			name: "success",
			input: models.SignUpInput{
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
			input: models.SignUpInput{
				Email:    "invalid",
				Password: "Validpass1!",
			},
			expectedError: "invalid email",
		},
		{
			name: "invalid password",
			input: models.SignUpInput{
				Email:    "valid@example.com",
				Password: "invalid",
			},
			validatePasswordError: errors.New("password must contain one uppercase letter"),
			expectedError:         "invalid password",
		},
		{
			name: "hashing failure",
			input: models.SignUpInput{
				Email:    "valid@example.com",
				Password: "Validpass1!",
			},
			validatePasswordError: nil,
			hashPasswordError:     errors.New("hash error"),
			expectedError:         "failed to hash password",
		},
		{
			name: "create user failure",
			input: models.SignUpInput{
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
			mockQ := mocks.NewQuerier(t)
			mockAuth := mocks.NewAuthInterface(t)

			if models.IsValidEmail(tc.input.Email) {
				mockAuth.On("ValidatePassword", tc.input.Password).Return(tc.validatePasswordError)
				
				if tc.validatePasswordError == nil {
					mockAuth.On("HashPassword", tc.input.Password).Return(tc.hashPasswordOutput, tc.hashPasswordError)
					if tc.hashPasswordError == nil {
						mockQ.On("CreateUser", mock.Anything, mock.MatchedBy(func(params database.CreateUserParams) bool {
							// Create an expected value ignoring the generated ID.
							expected := database.CreateUserParams{
								Email:          tc.input.Email,
								HashedPassword: tc.hashPasswordOutput,
							}
							// Use cmp.Diff and ignore the ID field.
							diff := cmp.Diff(expected, params, cmpopts.IgnoreFields(database.CreateUserParams{}, "ID"))
							if diff != "" {
								t.Logf("CreateUserParams mismatch (-want +got):\n%s", diff)
								return false
							}
							// Additionally, ensure that ID is not empty.
							return params.ID != ""
						})).Return(tc.createUserError)
					}				
				}				
			}

			userSvc := services.NewUserService(mockQ, mockAuth)
			err := userSvc.SignUp(context.Background(), tc.input)

			if tc.expectedError == "" {
				require.NoError(t, err, "expected no error, but got one")
			} else {
				require.Error(t, err, "expected error, but got nil")
				require.Contains(t, err.Error(), tc.expectedError)
			}

			mockQ.AssertExpectations(t)
			mockAuth.AssertExpectations(t)

		})
	}
}
