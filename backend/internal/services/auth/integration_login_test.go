package auth_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/stretchr/testify/require"
)

func TestLoginIntegration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		input                models.LoginInput
		xDeviceInfo          models.DeviceInfo
		hasErr               bool
		expectedErrSubstring string
		ctxErr               bool
	}{
		{
			name: "sucessful login, device found",
			input: models.LoginInput{
				Email:    "user@example.com",
				Password: "Validpass1!",
			},
			xDeviceInfo: models.DeviceInfo{
				DeviceType:     "Mobile",
				Browser:        "Chrome",
				BrowserVersion: "100.0",
				Os:             "Android",
				OsVersion:      "11",
			},
			hasErr: false,
		},
		{
			name: "successful login, device not found",
			input: models.LoginInput{
				Email:    "user@example.com",
				Password: "Validpass1!",
			},
			xDeviceInfo: models.DeviceInfo{
				DeviceType:     "Desktop",
				Browser:        "Chrome",
				BrowserVersion: "100.0",
				Os:             "Windows",
				OsVersion:      "11",
			},
			hasErr: false,
		},
		{
			name: "failed login, invalid password",
			input: models.LoginInput{
				Email:    "user@example.com",
				Password: "Invalidpass1!",
			},
			xDeviceInfo: models.DeviceInfo{
				DeviceType:     "Mobile",
				Browser:        "Chrome",
				BrowserVersion: "100.0",
				Os:             "Android",
				OsVersion:      "11",
			},
			hasErr:               true,
			expectedErrSubstring: "invalid email or password",
		},
		{
			name: "failed login, request not in context",
			input: models.LoginInput{
				Email:    "user@example.com",
				Password: "Validpass1!",
			},
			xDeviceInfo: models.DeviceInfo{
				DeviceType:     "Mobile",
				Browser:        "Chrome",
				BrowserVersion: "100.0",
				Os:             "Android",
				OsVersion:      "11",
			},
			hasErr:               true,
			expectedErrSubstring: "request not found in context",
			ctxErr:               true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			env := testhelpers.SetupTestEnv(t)
			t.Cleanup(func() { env.Db.Close() })
			userID := seedTestUserForAuth(t, env.Services.AuthService.PwdHasher, env.UserQ)
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("X-Device-Info", fmt.Sprintf("os=%s; os_version=%s; device_type=%s; browser=%s; browser_version=%s",
				tc.xDeviceInfo.Os,
				tc.xDeviceInfo.OsVersion,
				tc.xDeviceInfo.DeviceType,
				tc.xDeviceInfo.Browser,
				tc.xDeviceInfo.BrowserVersion,
			))
			var ctx context.Context

			if tc.ctxErr {
				ctx = req.Context()
			} else {
				ctx = context.WithValue(req.Context(), constants.RequestKey, req)
			}

			response, err := env.Services.AuthService.Login(ctx, tc.input)
			if tc.hasErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrSubstring)
				return
			}
			require.NoError(t, err)
			if diff := cmp.Diff(userID, response.UserID); diff != "" {
				t.Errorf("response mismatch (-want +got)\n%s", diff)
			}
			require.NotEmpty(t, response.JWTToken)
			require.NotEmpty(t, response.RefreshToken)
			deviceID, err := env.TransactionalQ.GetDeviceInfoByUser(ctx, database.GetDeviceInfoByUserParams{
				UserID:         userID,
				DeviceType:     tc.xDeviceInfo.DeviceType,
				Browser:        tc.xDeviceInfo.Browser,
				BrowserVersion: tc.xDeviceInfo.BrowserVersion,
				Os:             tc.xDeviceInfo.Os,
				OsVersion:      tc.xDeviceInfo.OsVersion,
			})
			require.NoError(t, err)
			getRefreshTokenEntry, err := env.TokenQ.GetRefreshByUserAndDevice(ctx, database.GetRefreshByUserAndDeviceParams{
				UserID:       userID,
				DeviceInfoID: deviceID,
			})

			require.NoError(t, err)
			require.NotNil(t, getRefreshTokenEntry)
			err = env.Services.AuthService.PwdHasher.CheckPasswordHash(response.RefreshToken, getRefreshTokenEntry.TokenHash)
			require.NoError(t, err)
			_, err = env.Services.AuthService.TokenGen.ValidateJWT(response.JWTToken)
			require.NoError(t, err)

		})
	}
}

// Helpers
func seedTestUserForAuth(t *testing.T, hasher auth.PasswordHasher, userQ database.UserQuerier) uuid.UUID {
	password := "Validpass1!"
	email := "user@example.com"
	userID := uuid.New()
	hashedPwd, err := hasher.HashPassword(password)
	require.NoError(t, err)

	err = userQ.CreateUser(context.Background(), database.CreateUserParams{
		ID:             userID,
		Email:          email,
		HashedPassword: hashedPwd,
	})
	require.NoError(t, err)
	return userID
}
