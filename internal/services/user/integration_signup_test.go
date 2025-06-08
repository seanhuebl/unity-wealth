package user_test

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
	"github.com/seanhuebl/unity-wealth/internal/services/user"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/stretchr/testify/require"
)

func TestIntSignup(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name          string
		input         user.SignUpInput
		wantErrSubstr string
	}{
		{
			name: "success",
			input: user.SignUpInput{
				Email:    "valid@example.com",
				Password: "Validpass1!",
			},
			wantErrSubstr: "",
		},
		{
			name: "invalid email",
			input: user.SignUpInput{
				Email:    "invalid",
				Password: "Validpass1!",
			},
			wantErrSubstr: "invalid email",
		},
		{
			name: "invalid password",
			input: user.SignUpInput{
				Email:    "valid@example.com",
				Password: "invalid",
			},
			wantErrSubstr: "invalid password",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			env := testhelpers.SetupTestEnv(t)
			userSvc := user.NewUserService(env.UserQ, auth.NewRealPwdHasher(), env.Logger)
			err := userSvc.SignUp(ctx, tc.input)
			if tc.wantErrSubstr != "" {
				if tc.wantErrSubstr == auth.ErrInvalidEmail.Error() {
					require.ErrorIs(t, err, auth.ErrInvalidEmail)
				} else if tc.wantErrSubstr == auth.ErrInvalidPassword.Error() {
					require.ErrorIs(t, err, auth.ErrInvalidPassword)
				} else {
					require.ErrorContains(t, err, tc.wantErrSubstr)
				}
			} else {
				require.NoError(t, err)
				userRecord, err := env.UserQ.GetUserByEmail(ctx, tc.input.Email)
				require.NoError(t, err)
				require.NotEmpty(t, userRecord.ID)
				require.NoError(t, auth.NewRealPwdHasher().CheckPasswordHash(tc.input.Password, userRecord.HashedPassword))
			}
		})

	}
	t.Run("create user failure", func(t *testing.T) {
		env := testhelpers.SetupTestEnv(t)
		svc := user.NewUserService(env.UserQ, auth.NewRealPwdHasher(), env.Logger)
		input := user.SignUpInput{"duplicate@example.com", "Validpass1!"}

		require.NoError(t, svc.SignUp(ctx, input))

		err := svc.SignUp(ctx, input)
		require.ErrorContains(t, err, "unable to create user")

	})
}
