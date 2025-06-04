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
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
	"github.com/seanhuebl/unity-wealth/internal/models"
)

type AuthService struct {
	SqlTxQuerier database.SqlTxQuerier
	UserQuerier  database.UserQuerier
	TokenGen     TokenGenerator
	TokenExtract TokenExtractor
	PwdHasher    PasswordHasher
}

func NewAuthService(SqlTxQuerier database.SqlTxQuerier, UserQuerier database.UserQuerier, TokenGen TokenGenerator, tokenExtract TokenExtractor, PwdHasher PasswordHasher) *AuthService {
	return &AuthService{
		SqlTxQuerier: SqlTxQuerier,
		UserQuerier:  UserQuerier,
		TokenGen:     TokenGen,
		TokenExtract: tokenExtract,
		PwdHasher:    PwdHasher,
	}
}

// Login encapsulates the entire login process.
func (a *AuthService) Login(ctx context.Context, input models.LoginInput) (models.LoginResponse, error) {

	// 1. Validate the email format (optionally done here or in the handler)
	if !models.IsValidEmail(input.Email) {
		return models.LoginResponse{}, fmt.Errorf("invalid email format")
	}

	// 2. Validate credentials and fetch user.
	userID, err := a.ValidateCredentials(ctx, input)
	if err != nil {
		return models.LoginResponse{}, err
	}

	req, err := helpers.GetRequestFromContext(ctx)
	if err != nil {
		return models.LoginResponse{}, err
	}
	// 3. Extract device information.
	deviceInfo, err := GetDeviceInfoFromRequest(req)
	if err != nil {
		return models.LoginResponse{}, err
	}

	// 4. Start a database transaction.
	tx, err := a.SqlTxQuerier.BeginTx(ctx, nil)
	if err != nil {
		return models.LoginResponse{}, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()
	queriesTx := a.SqlTxQuerier.WithTx(tx)
	deviceQ := database.NewRealDevicequerier(queriesTx)
	tokenQ := database.NewRealTokenQuerier(queriesTx)

	// 5. Handle device information.
	deviceID, err := a.HandleDeviceInfo(ctx, deviceQ, tokenQ, userID, deviceInfo)
	if err != nil {
		return models.LoginResponse{}, err
	}

	// 6. Generate JWT and refresh token.
	jwtToken, refreshToken, err := a.GenerateTokens(userID)
	if err != nil {
		return models.LoginResponse{}, err
	}

	// 7. Hash the refresh token.
	refreshHash, err := a.PwdHasher.HashPassword(refreshToken)
	if err != nil {
		return models.LoginResponse{}, fmt.Errorf("failed to hash refresh token: %w", err)
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
		return models.LoginResponse{}, fmt.Errorf("failed to create refresh token entry: %w", err)
	}

	// 9. Commit the transaction.
	if err := tx.Commit(); err != nil {
		return models.LoginResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 10. Return a structured login response.
	return models.LoginResponse{
		UserID:       userID,
		RefreshToken: refreshToken,
		JWTToken:     jwtToken,
	}, nil
}

// Helpers
func (a *AuthService) ValidateCredentials(ctx context.Context, input models.LoginInput) (uuid.UUID, error) {
	user, err := a.UserQuerier.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, fmt.Errorf("invalid email / password")
		}
		return uuid.Nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	if err := a.PwdHasher.CheckPasswordHash(input.Password, user.HashedPassword); err != nil {
		return uuid.Nil, fmt.Errorf("invalid email / password")
	}
	return uuid.Parse(user.ID)
}

func (a *AuthService) HandleDeviceInfo(ctx context.Context, deviceQ database.DeviceQuerier, tokenQ database.TokenQuerier, userID uuid.UUID, info models.DeviceInfo) (uuid.UUID, error) {
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

func (a *AuthService) GenerateTokens(userID uuid.UUID) (string, string, error) {
	jwtToken, err := a.TokenGen.MakeJWT(userID, 15*time.Minute)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	refreshToken, err := a.TokenGen.MakeRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return jwtToken, refreshToken, nil
}

func GetDeviceInfoFromRequest(req *http.Request) (models.DeviceInfo, error) {
	// Check for the X-Device-Info header first.
	xDeviceInfo := req.Header.Get("X-Device-Info")
	if xDeviceInfo != "" {
		deviceInfo := ParseDeviceInfoFromHeader(xDeviceInfo)
		if IsValidDeviceInfo(deviceInfo) {
			return deviceInfo, nil
		}
	}
	// Fallback to parsing the User-Agent header.
	userAgent := req.Header.Get("User-Agent")
	if userAgent != "" {
		deviceInfo := ParseUserAgent(userAgent)
		if IsValidDeviceInfo(deviceInfo) {
			return deviceInfo, nil
		}
	}
	return models.DeviceInfo{}, fmt.Errorf("invalid or unknown device information")
}

// parseDeviceInfoFromHeader parses the X-Device-Info header into a models.DeviceInfo struct.
func ParseDeviceInfoFromHeader(header string) models.DeviceInfo {
	var info models.DeviceInfo
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

// ParseUserAgent parses the User-Agent header using the mssola/user_agent package.
func ParseUserAgent(userAgent string) models.DeviceInfo {
	ua := user_agent.New(userAgent)
	deviceType := "Desktop"
	if ua.Mobile() {
		deviceType = "Mobile"
	}
	browser, browserVersion := ua.Browser()
	return models.DeviceInfo{
		DeviceType:     deviceType,
		Browser:        sanitizeInput(browser),
		BrowserVersion: sanitizeInput(browserVersion),
		Os:             sanitizeInput(ua.OSInfo().FullName),
		OsVersion:      sanitizeInput(ua.OSInfo().Version),
	}
}

func IsValidDeviceInfo(info models.DeviceInfo) bool {
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
