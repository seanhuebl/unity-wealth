package handlers

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/handlers"
	"github.com/seanhuebl/unity-wealth/mocks"
)

func TestGenerateTokens(t *testing.T) {
	tests := []struct {
		name             string
		userID           uuid.UUID
		secret           string
		mockJWTResponse  string
		mockJWTError     error
		mockRefreshToken string
		mockRefreshError error
		expectError      bool
	}{
		{
			name:             "Successful token generation",
			userID:           uuid.New(),
			secret:           "my_secret",
			mockJWTResponse:  "valid_jwt", // Placeholder for comparison
			mockJWTError:     nil,
			mockRefreshToken: "valid_refresh_token", // Placeholder for comparison
			mockRefreshError: nil,
			expectError:      false,
		},
		{
			name:             "Error in MakeJWT",
			userID:           uuid.New(),
			secret:           "my_secret",
			mockJWTResponse:  "",
			mockJWTError:     errors.New("failed to generate JWT"),
			mockRefreshToken: "",
			mockRefreshError: nil,
			expectError:      true,
		},
		{
			name:             "Error in MakeRefreshToken",
			userID:           uuid.New(),
			secret:           "my_secret",
			mockJWTResponse:  "valid_jwt", // Placeholder for comparison
			mockJWTError:     nil,
			mockRefreshToken: "",
			mockRefreshError: errors.New("failed to generate refresh token"),
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock AuthInterface
			mockAuth := mocks.NewAuthInterface(t)

			// Set up expectations for MakeJWT
			mockAuth.On("MakeJWT", tt.userID, tt.secret, time.Minute*15).Return(tt.mockJWTResponse, tt.mockJWTError)

			if tt.name != "Error in MakeJWT" {
				// Set up expectations for MakeRefreshToken
				mockAuth.On("MakeRefreshToken").Return(tt.mockRefreshToken, tt.mockRefreshError)
			}

			// Call the function
			jwt, refreshToken, err := handlers.GenerateTokens(tt.userID, mockAuth)

			// Assert error scenarios
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected an error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("did not expect an error but got: %v", err)
				}

				// Validate JWT and Refresh Token structure using cmp
				if diff := cmp.Diff(tt.mockJWTResponse, jwt); diff != "" {
					t.Errorf("JWT mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.mockRefreshToken, refreshToken); diff != "" {
					t.Errorf("Refresh Token mismatch (-want +got):\n%s", diff)
				}
			}

			// Assert that expectations were met
			mockAuth.AssertExpectations(t)
		})
	}
}
