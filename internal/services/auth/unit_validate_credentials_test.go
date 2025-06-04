package auth_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
	authmocks "github.com/seanhuebl/unity-wealth/internal/mocks/auth"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
	"github.com/stretchr/testify/require"
)

func TestValidateCredentials(t *testing.T) {
	ctx := context.Background()

	dummyUser := database.GetUserByEmailRow{
		ID:             uuid.NewString(),
		HashedPassword: "hashedpassword",
	}

	tests := []struct {
		name                   string
		input                  models.LoginInput
		getUserErr             error
		pwdHasherErr           error
		expectedErrorSubstring string
		expectedUserID         uuid.UUID
	}{
		{
			name: "successful credentials",
			input: models.LoginInput{
				Email:    "user@example.com",
				Password: "correctpassword",
			},
			getUserErr:             nil,
			pwdHasherErr:           nil,
			expectedErrorSubstring: "",
			expectedUserID:         uuid.MustParse(dummyUser.ID),
		},
		{
			name: "user not found",
			input: models.LoginInput{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			getUserErr:             sql.ErrNoRows,
			pwdHasherErr:           nil,
			expectedErrorSubstring: "invalid email / password",
		},
		{
			name: "error fetching user",
			input: models.LoginInput{
				Email:    "error@example.com",
				Password: "password123",
			},
			getUserErr:             errors.New("db error"),
			pwdHasherErr:           nil,
			expectedErrorSubstring: "failed to fetch user",
		},
		{
			name: "invalid password",
			input: models.LoginInput{
				Email:    "user@example.com",
				Password: "wrongpassword",
			},
			getUserErr:             nil,
			pwdHasherErr:           errors.New("password mismatch"),
			expectedErrorSubstring: "invalid email / password",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockUserQ := dbmocks.NewUserQuerier(t)
			mockPwdHasher := authmocks.NewPasswordHasher(t)

			mockUserQ.On("GetUserByEmail", ctx, tc.input.Email).Return(func(ctx context.Context, email string) database.GetUserByEmailRow {
				return dummyUser
			}, tc.getUserErr)

			if tc.getUserErr == nil {
				mockPwdHasher.On("CheckPasswordHash", tc.input.Password, dummyUser.HashedPassword).Return(tc.pwdHasherErr)
			}
			authSvc := auth.NewAuthService(nil, mockUserQ, nil, nil, mockPwdHasher)

			userID, err := authSvc.ValidateCredentials(ctx, tc.input)
			if tc.expectedErrorSubstring != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrorSubstring)
				require.Equal(t, uuid.Nil, userID)
			} else {
				require.NoError(t, err)
				if diff := cmp.Diff(tc.expectedUserID, userID); diff != "" {
					t.Errorf("validateCredentials() mismatch (-want +got)\n%s", diff)
				}
			}
			mockUserQ.AssertExpectations(t)
			mockPwdHasher.AssertExpectations(t)
		})
	}
}
