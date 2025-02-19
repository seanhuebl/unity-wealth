package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mssola/user_agent"
	"github.com/seanhuebl/unity-wealth/helpers"
	"github.com/seanhuebl/unity-wealth/internal/database"
)

type AuthService struct {
	sqlTxQuerier database.SqlTxQuerier
	userQuerier  database.UserQuerier
	tokenGen     TokenGenerator
	tokenExtract TokenExtractor
	pwdHasher    PasswordHasher
}

func NewAuthService(sqlTxQuerier database.SqlTxQuerier, userQuerier database.UserQuerier, tokenGen TokenGenerator, tokenExtract TokenExtractor, pwdHasher PasswordHasher) *AuthService {
	return &AuthService{
		sqlTxQuerier: sqlTxQuerier,
		userQuerier:  userQuerier,
		tokenGen:     tokenGen,
		tokenExtract: tokenExtract,
		pwdHasher:    pwdHasher,
	}
}

// Login encapsulates the entire login process.
func (a *AuthService) Login(ctx context.Context, input LoginInput) (LoginResponse, error) {

	// 1. Validate the email format (optionally done here or in the handler)
	if !IsValidEmail(input.Email) {
		return LoginResponse{}, fmt.Errorf("invalid email format")
	}

	// 2. Validate credentials and fetch user.
	userID, err := a.validateCredentials(ctx, input)
	if err != nil {
		return LoginResponse{}, err
	}

	req, err := helpers.GetRequestFromContext(ctx)
	if err != nil {
		return LoginResponse{}, err
	}
	// 3. Extract device information.
	deviceInfo, err := getDeviceInfoFromRequest(req)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("device information could not be verified")
	}

	// 4. Start a database transaction.
	tx, err := a.sqlTxQuerier.BeginTx(ctx, nil)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()
	queriesTx := a.sqlTxQuerier.WithTx(tx)
	deviceQ := database.NewRealDevicequerier(queriesTx)
	tokenQ := database.NewRealTokenQuerier(queriesTx)

	// 5. Handle device information.
	deviceID, err := a.handleDeviceInfo(ctx, deviceQ, tokenQ, userID, deviceInfo)
	if err != nil {
		return LoginResponse{}, err
	}

	// 6. Generate JWT and refresh token.
	jwtToken, refreshToken, err := a.generateTokens(userID)
	if err != nil {
		return LoginResponse{}, err
	}

	// 7. Hash the refresh token.
	refreshHash, err := a.pwdHasher.HashPassword(refreshToken)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("failed to hash refresh token: %w", err)
	}

	// 8. Create a refresh token record.
	expiration := sql.NullTime{
		Time:  time.Now().Add(60 * 24 * time.Hour),
		Valid: true,
	}
	err = queriesTx.CreateRefreshToken(ctx, database.CreateRefreshTokenParams{
		ID:           uuid.NewString(),
		TokenHash:    refreshHash,
		ExpiresAt:    expiration,
		UserID:       userID.String(),
		DeviceInfoID: deviceID.String(),
	})
	if err != nil {
		return LoginResponse{}, fmt.Errorf("failed to create refresh token entry: %w", err)
	}

	// 9. Commit the transaction.
	if err := tx.Commit(); err != nil {
		return LoginResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 10. Return a structured login response.
	return LoginResponse{
		UserID:       userID,
		JWT:          jwtToken,
		RefreshToken: refreshToken,
	}, nil
}

// Helpers
func (a *AuthService) validateCredentials(ctx context.Context, input LoginInput) (uuid.UUID, error) {
	user, err := a.userQuerier.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, fmt.Errorf("invalid email / password")
		}
		return uuid.Nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	if err := a.pwdHasher.CheckPasswordHash(input.Password, user.HashedPassword); err != nil {
		return uuid.Nil, fmt.Errorf("invalid email / password")
	}
	return uuid.Parse(user.ID)
}

func (a *AuthService) handleDeviceInfo(ctx context.Context, deviceQ database.DeviceQuerier, tokenQ database.TokenQuerier, userID uuid.UUID, info DeviceInfo) (uuid.UUID, error) {
	foundDevice, err := deviceQ.GetDeviceInfoByUser(ctx, database.GetDeviceInfoByUserParams{
		UserID:         userID.String(),
		DeviceType:     info.DeviceType,
		Browser:        info.Browser,
		BrowserVersion: info.BrowserVersion,
		Os:             info.Os,
		OsVersion:      info.OsVersion,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			newDeviceID, err := deviceQ.CreateDeviceInfo(ctx, database.CreateDeviceInfoParams{
				ID:             uuid.NewString(),
				UserID:         userID.String(),
				DeviceType:     info.DeviceType,
				Browser:        info.Browser,
				BrowserVersion: info.BrowserVersion,
				Os:             info.Os,
				OsVersion:      info.OsVersion,
			})
			if err != nil {
				return uuid.Nil, fmt.Errorf("failed to create new device: %w", err)
			}
			return uuid.Parse(newDeviceID)
		}
		return uuid.Nil, fmt.Errorf("failed to fetch device info: %w", err)
	}

	deviceID, err := uuid.Parse(foundDevice)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse device ID: %w", err)
	}

	if err := tokenQ.RevokeToken(ctx, database.RevokeTokenParams{
		RevokedAt:    sql.NullTime{Time: time.Now(), Valid: true},
		UserID:       userID.String(),
		DeviceInfoID: deviceID.String(),
	}); err != nil {
		return uuid.Nil, fmt.Errorf("failed to revoke token: %w", err)
	}

	return deviceID, nil
}

func (a *AuthService) generateTokens(userID uuid.UUID) (string, string, error) {
	jwtToken, err := a.tokenGen.MakeJWT(userID, 15*time.Minute)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	refreshToken, err := a.tokenGen.MakeRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return jwtToken, refreshToken, nil
}

func getDeviceInfoFromRequest(req *http.Request) (DeviceInfo, error) {
	// Check for the X-Device-Info header first.
	xDeviceInfo := req.Header.Get("X-Device-Info")
	if xDeviceInfo != "" {
		deviceInfo := parseDeviceInfoFromHeader(xDeviceInfo)
		if isValidDeviceInfo(deviceInfo) {
			return deviceInfo, nil
		}
	}
	// Fallback to parsing the User-Agent header.
	userAgent := req.Header.Get("User-Agent")
	if userAgent != "" {
		deviceInfo := parseUserAgent(userAgent)
		if isValidDeviceInfo(deviceInfo) {
			return deviceInfo, nil
		}
	}
	return DeviceInfo{}, fmt.Errorf("invalid or unknown device information")
}

// parseDeviceInfoFromHeader parses the X-Device-Info header into a DeviceInfo struct.
func parseDeviceInfoFromHeader(header string) DeviceInfo {
	var info DeviceInfo
	pairs := strings.Split(header, ";")
	for _, pair := range pairs {
		kv := strings.Split(strings.TrimSpace(pair), "=")
		if len(kv) == 2 {
			key := strings.ToLower(strings.TrimSpace(kv[0]))
			value := sanitizeInput(strings.TrimSpace(kv[1]))
			switch key {
			case "os":
				info.Os = value
			case "os_version":
				info.OsVersion = value
			case "device_type":
				info.DeviceType = value
			case "browser":
				info.Browser = value
			case "browser_version":
				info.BrowserVersion = value
			}
		}
	}
	return info
}

// parseUserAgent parses the User-Agent header using the mssola/user_agent package.
func parseUserAgent(userAgent string) DeviceInfo {
	ua := user_agent.New(userAgent)
	deviceType := "Desktop"
	if ua.Mobile() {
		deviceType = "Mobile"
	}
	browser, browserVersion := ua.Browser()
	return DeviceInfo{
		DeviceType:     deviceType,
		Browser:        sanitizeInput(browser),
		BrowserVersion: sanitizeInput(browserVersion),
		Os:             sanitizeInput(ua.OSInfo().FullName),
		OsVersion:      sanitizeInput(ua.OSInfo().Version),
	}
}

func isValidDeviceInfo(info DeviceInfo) bool {
	validDeviceTypes := map[string]bool{
		"desktop": true,
		"mobile":  true,
	}
	if !validDeviceTypes[strings.ToLower(info.DeviceType)] {
		return false
	}
	if info.Browser == "" || info.Os == "" {
		return false
	}
	if info.BrowserVersion != "" && !isValidVersion(info.BrowserVersion) {
		return false
	}
	if info.OsVersion != "" && !isValidVersion(info.OsVersion) {
		return false
	}
	return true
}

func isValidVersion(version string) bool {
	versionRegex := `^\d+(\.\d+)*$`
	matched, _ := regexp.MatchString(versionRegex, version)
	return matched
}

func sanitizeInput(input string) string {
	input = strings.TrimSpace(input)
	if len(input) > 100 {
		input = input[:100]
	}
	return input
}
