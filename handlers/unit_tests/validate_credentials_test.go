package handlers

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/handlers"

	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestValidateCredentials(t *testing.T) {
	testCases := []struct {
		name          string
		input         handlers.LoginInput
		mockEmail     string
		mockUser      database.GetUserByEmailRow
		mockError     error
		mockPassword  string
		mockHash      string
		mockAuthError error
		expectedUUID  string
		expectedError string
	}{
		{
			name: "Valid credentials",
			input: handlers.LoginInput{
				Email:    "valid@example.com",
				Password: "correct-password",
			},
			mockEmail:     "valid@example.com",
			mockUser:      database.GetUserByEmailRow{ID: "123", HashedPassword: "hashed-pass"},
			mockError:     nil,
			mockPassword:  "correct-password",
			mockHash:      "hashed-pass",
			mockAuthError: nil,
			expectedUUID:  "123",
			expectedError: "",
		},
		{
			name: "Invalid password",
			input: handlers.LoginInput{
				Email:    "valid@example.com",
				Password: "wrong-password",
			},
			mockEmail:     "valid@example.com",
			mockUser:      database.GetUserByEmailRow{ID: "123", HashedPassword: "hashed-pass"},
			mockError:     nil,
			mockPassword:  "wrong-password",
			mockHash:      "hashed-pass",
			mockAuthError: fmt.Errorf("password mismatch"),
			expectedUUID:  "",
			expectedError: "invalid email / password",
		},
		{
			name: "User not found",
			input: handlers.LoginInput{
				Email:    "nonexistent@example.com",
				Password: "any-password",
			},
			mockEmail:     "nonexistent@example.com",
			mockUser:      database.GetUserByEmailRow{},
			mockError:     fmt.Errorf("user not found"),
			mockPassword:  "",
			mockHash:      "",
			mockAuthError: nil,
			expectedUUID:  "",
			expectedError: "invalid email / password",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize mocks
			mockQueries := mocks.NewQuierier(t)
			mockAuth := mocks.NewAuthInterface(t)

			// Set up mock behavior for database queries
			mockQueries.On("GetUserByEmail", mock.Anything, tc.mockEmail).
				Return(tc.mockUser, tc.mockError)

			// Set up mock behavior for password check
			if tc.mockPassword != "" && tc.mockHash != "" {
				mockAuth.On("CheckPasswordHash", tc.mockPassword, tc.mockHash).
					Return(tc.mockAuthError)
			}

			// Create API config with mocks
			cfg := &handlers.ApiConfig{
				Queries: mockQueries,
				Auth:    mockAuth,
			}

			// Call the function under test
			resultUUID, err := handlers.ValidateCredentials(gin.Context(), cfg, &tc.input)

			// Assertions
			if tc.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
				require.Equal(t, "", resultUUID.String())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedUUID, resultUUID.String())
			}

			// Verify mock expectations
			mockQueries.AssertExpectations(t)
			mockAuth.AssertExpectations(t)
		})
	}
}
