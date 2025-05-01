package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/stretchr/testify/require"
)

func TestLoginIntegration(t *testing.T) {
	tests := []struct {
		name                 string
		input                LoginInput
		xDeviceInfo          DeviceInfo
		hasErr               bool
		expectedErrSubstring string
	}{
		{
			name: "sucessful login, device found",
			input: LoginInput{
				Email:    "user@example.com",
				Password: "Validpass1!",
			},
			xDeviceInfo: DeviceInfo{
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
			input: LoginInput{
				Email:    "user@example.com",
				Password: "Validpass1!",
			},
			xDeviceInfo: DeviceInfo{
				DeviceType:     "Desktop",
				Browser:        "Chrome",
				BrowserVersion: "100.0",
				Os:             "Windows",
				OsVersion:      "11",
			},
			hasErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			db, err := sql.Open("sqlite3", ":memory:")
			require.NoError(t, err)
			defer db.Close()
			_, err = db.Exec("PRAGMA foreign_keys = ON")
			require.NoError(t, err)

			createSchema(t, db)
			transactionalQ := database.NewRealTransactionalQuerier(database.New(db))
			tokenQ := database.NewRealTokenQuerier(transactionalQ)
			sqlTxQ := database.NewRealSqlTxQuerier(transactionalQ)
			userQ := database.NewRealUserQuerier(transactionalQ)
			tokeGen := NewRealTokenGenerator("tokensecret", TokenType("unity-wealth"))
			pwdHasher := NewRealPwdHasher()
			userID := seedTestUser(t, pwdHasher, userQ)

			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("X-Device-Info", fmt.Sprintf("os=%s; os_version=%s; device_type=%s; browser=%s; browser_version=%s",
				tc.xDeviceInfo.Os,
				tc.xDeviceInfo.OsVersion,
				tc.xDeviceInfo.DeviceType,
				tc.xDeviceInfo.Browser,
				tc.xDeviceInfo.BrowserVersion,
			))
			ctx := context.WithValue(req.Context(), constants.RequestKey, req)

			svc := NewAuthService(sqlTxQ, userQ, tokeGen, nil, pwdHasher)
			response, err := svc.Login(ctx, tc.input)
			require.NoError(t, err)
			if diff := cmp.Diff(userID, response.UserID); diff != "" {
				t.Errorf("response mismatch (-want +got)\n%s", diff)
			}
			require.NotEmpty(t, response.JWT)
			require.NotEmpty(t, response.RefreshToken)
			deviceID, err := transactionalQ.GetDeviceInfoByUser(ctx, database.GetDeviceInfoByUserParams{
				UserID:         userID.String(),
				DeviceType:     tc.xDeviceInfo.DeviceType,
				Browser:        tc.xDeviceInfo.Browser,
				BrowserVersion: tc.xDeviceInfo.BrowserVersion,
				Os:             tc.xDeviceInfo.Os,
				OsVersion:      tc.xDeviceInfo.OsVersion,
			})
			require.NoError(t, err)
			getRefreshTokenEntry, err := tokenQ.GetRefreshByUserAndDevice(ctx, database.GetRefreshByUserAndDeviceParams{
				UserID:       userID.String(),
				DeviceInfoID: deviceID,
			})

			require.NoError(t, err)
			require.NotNil(t, getRefreshTokenEntry)
			err = svc.pwdHasher.CheckPasswordHash(response.RefreshToken, getRefreshTokenEntry.TokenHash)
			require.NoError(t, err)
			_, err = svc.tokenGen.ValidateJWT(response.JWT)
			require.NoError(t, err)

		})
	}
}

// Helpers
func createSchema(t *testing.T, db *sql.DB) {
	_, err := db.Exec(constants.CreateUsersTable)
	require.NoError(t, err)
	_, err = db.Exec(constants.CreateRefrTokenTable)
	require.NoError(t, err)

	_, err = db.Exec(constants.CreateDeviceInfoTable)
	require.NoError(t, err)
}

func seedTestUser(t *testing.T, hasher PasswordHasher, userQ database.UserQuerier) uuid.UUID {
	password := "Validpass1!"
	email := "user@example.com"
	userID := uuid.New()
	hashedPwd, err := hasher.HashPassword(password)
	require.NoError(t, err)

	err = userQ.CreateUser(context.Background(), database.CreateUserParams{
		ID:             userID.String(),
		Email:          email,
		HashedPassword: hashedPwd,
	})
	require.NoError(t, err)
	return userID
}
