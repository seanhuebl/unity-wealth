package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	authmocks "github.com/seanhuebl/unity-wealth/internal/mocks/auth"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
	"github.com/stretchr/testify/require"
)

func TestGenerateTokens(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name                   string
		jwtToken               string
		jwtError               error
		refreshToken           string
		refreshError           error
		expectedJWT            string
		expectedRefreshToken   string
		expectedErrorSubstring string
	}{
		{
			name:                   "successful token generation",
			jwtToken:               "dummyJWT",
			jwtError:               nil,
			refreshToken:           "dummyRefreshToken",
			refreshError:           nil,
			expectedJWT:            "dummyJWT",
			expectedRefreshToken:   "dummyRefreshToken",
			expectedErrorSubstring: "",
		},
		{
			name:                   "MakeJWT fails",
			jwtToken:               "",
			jwtError:               errors.New("jwt error"),
			refreshToken:           "dummyRefreshToken",
			refreshError:           nil,
			expectedJWT:            "",
			expectedRefreshToken:   "",
			expectedErrorSubstring: "failed to generate JWT",
		},
		{
			name:                   "MakeRefreshToken fails",
			jwtToken:               "dummyJWT",
			jwtError:               nil,
			refreshToken:           "",
			refreshError:           errors.New("refresh error"),
			expectedErrorSubstring: "failed to generate refresh token",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockTokenGen := authmocks.NewTokenGenerator(t)

			mockTokenGen.On("MakeJWT", userID, 15*time.Minute).Return(tc.jwtToken, tc.jwtError)
			if tc.jwtError == nil {
				mockTokenGen.On("MakeRefreshToken").Return(tc.refreshToken, tc.refreshError)
			}

			svc := auth.NewAuthService(nil, nil, mockTokenGen, nil, nil)
			jwtToken, refreshToken, err := svc.GenerateTokens(userID)
			if tc.expectedErrorSubstring != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrorSubstring)
				require.Empty(t, jwtToken)
				require.Empty(t, refreshToken)
			} else {
				require.NoError(t, err)
				if diff := cmp.Diff(tc.expectedJWT, jwtToken); diff != "" {
					t.Errorf("jwtToken mismatch (-want +got)\n%s", diff)
				}
				if diff := cmp.Diff(tc.expectedRefreshToken, refreshToken); diff != "" {
					t.Errorf("refreshToken mismatch (-want +got)\n%s", diff)
				}

			}
			mockTokenGen.AssertExpectations(t)
		})
	}
}
