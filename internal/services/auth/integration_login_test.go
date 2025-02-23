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
		{
			name: "login failed, user not found",
			input: LoginInput{
				Email:    "notfound@example.com",
				Password: "Validpass1!",
			},
			xDeviceInfo: DeviceInfo{
				DeviceType:     "Mobile",
				Browser:        "Chrome",
				BrowserVersion: "100.0",
				Os:             "Android",
				OsVersion:      "11",
			},
			hasErr:               true,
			expectedErrSubstring: "invalid email / password",
		},
		{
			name: "login failed, incorrect password",
			input: LoginInput{
				Email:    "user@example.com",
				Password: "Wrongpass1!",
			},
			xDeviceInfo: DeviceInfo{
				DeviceType:     "Mobile",
				Browser:        "Chrome",
				BrowserVersion: "100.0",
				Os:             "Android",
				OsVersion:      "11",
			},
			hasErr:               true,
			expectedErrSubstring: "invalid email / password",
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
			if !tc.hasErr {
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
			} else {
				require.Error(t, err)
				if diff := cmp.Diff(err.Error(), tc.expectedErrSubstring); diff != "" {
					t.Errorf("error mismatch (-want +got)\n%s", diff)
				}
			}
		})
	}
}

// Helpers
func createSchema(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		hashed_password TEXT NOT NULL,
		risk_preference TEXT NOT NULL DEFAULT 'LOW',
		plan_type TEXT NOT NULL DEFAULT 'FREE',
		stripe_customer_id TEXT,
		stripe_subscription_id TEXT,
		scholarship_flag INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	require.NoError(t, err)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS refresh_tokens (
		id TEXT PRIMARY KEY,
		token_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		revoked_at DATETIME,
		user_id TEXT NOT NULL,
		device_info_id TEXT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
		FOREIGN KEY (device_info_id) REFERENCES device_info_logs (id) ON DELETE CASCADE
		);
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS device_info_logs (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		device_type TEXT NOT NULL,
		browser TEXT NOT NULL,
		browser_version TEXT NOT NULL,
		os TEXT NOT NULL,
		os_version TEXT NOT NULL,
		app_info TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
		);
	`)
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
