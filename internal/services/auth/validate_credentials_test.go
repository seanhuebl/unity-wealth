package auth

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
)

func TestValidateCredentials(t *testing.T) {
	ctx := context.Background()

	dummyUser := database.GetUserByEmailRow{
		ID:             uuid.NewString(),
		HashedPassword: "hashedpassword",
	}

	tests := []struct {
		name                   string
		input                  LoginInput
		getUserErr             error
		pwdHasherErr           error
		expectedErrorSubstring string
		expectedUserID         uuid.UUID
	}{
		{
			name: "successful credentials",
			input: LoginInput{
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
			input: LoginInput{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			getUserErr:             sql.ErrNoRows,
			pwdHasherErr:           nil,
			expectedErrorSubstring: "invalid email / password",
		},
		{
			name: "error fetching user",
			input: LoginInput{
				Email:    "error@example.com",
				Password: "password123",
			},
			getUserErr:             errors.New("db error"),
			pwdHasherErr:           nil,
			expectedErrorSubstring: "failed to fetch user",
		},
		{
			name: "invalid password",
			input: LoginInput{
				Email:    "user@example.com",
				Password: "wrongpassword",
			},
			getUserErr:             nil,
			pwdHasherErr:           errors.New("password mismatch"),
			expectedErrorSubstring: "invalid email / password",
		},
	}
}
