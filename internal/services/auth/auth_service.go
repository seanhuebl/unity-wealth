package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mssola/user_agent"
	"github.com/seanhuebl/unity-wealth/helpers"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/interfaces"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	tokenTypeAccess TokenType
	tokenSecret     string
	queries         interfaces.Querier
}

func NewAuthService(tokenType, tokenSecret string, queries interfaces.Querier) *AuthService {
	return &AuthService{
		tokenTypeAccess: TokenType(tokenType),
		tokenSecret:     tokenSecret,
		queries:         queries,
	}
}

type TokenType string

var ErrNoAuthHeaderIncluded = errors.New("no authorization header included")
var RandReader = rand.Read

func (a *AuthService) GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "ApiKey" {
		return "", errors.New("malformed authorization header")
	}
	if splitAuth[1] == "" {
		return "", errors.New("malformed authorization header")
	}
	return splitAuth[1], nil
}

func (a *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (a *AuthService) CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (a *AuthService) MakeJWT(userID uuid.UUID, expiresIn time.Duration) (string, error) {
	if a.tokenSecret == "" {
		return "", errors.New("tokenSecret must not be empty")
	}
	if expiresIn <= 0 {
		return "", errors.New("expiresIn must be positive")
	}
	signingKey := []byte(a.tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(a.tokenTypeAccess),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})
	return token.SignedString(signingKey)
}

func (a *AuthService) ValidateJWT(tokenString string) (*jwt.RegisteredClaims, error) {
	// Create an instance of RegisteredClaims to hold the parsed token claims.
	var claims jwt.RegisteredClaims

	// Parse the token using the claims instance.
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.tokenSecret), nil
	})
	if err != nil {
		return nil, err
	}

	// Ensure the token is valid.
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	now := time.Now().Unix()
	if claims.ExpiresAt != nil && claims.ExpiresAt.Unix() < now {
		return nil, fmt.Errorf("token expired")
	}
	if claims.NotBefore != nil && claims.NotBefore.Unix() > now {
		return nil, fmt.Errorf("token not valid yet")
	}

	// Check the issuer (this example assumes you want the issuer to equal TokenTypeAccess).
	if claims.Issuer != string(a.tokenTypeAccess) {
		return nil, errors.New("invalid issuer")
	}

	// Optionally, if you expect the Subject (user ID) to be a valid UUID, you can verify that.
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Return the complete claims struct.
	return &claims, nil
}

func (a *AuthService) GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		return "", errors.New("malformed authorization header")
	}
	if splitAuth[1] == "" {
		return "", errors.New("malformed authorization header")
	}
	return splitAuth[1], nil
}

func (a *AuthService) MakeRefreshToken() (string, error) {
	token := make([]byte, 32)
	_, err := RandReader(token)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
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
	tx, err := a.queries.BeginTx(ctx, nil)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()
	queriesTx := a.queries.WithTx(tx)

	// 5. Handle device information.
	deviceID, err := a.handleDeviceInfo(ctx, queriesTx, userID, deviceInfo)
	if err != nil {
		return LoginResponse{}, err
	}

	// 6. Generate JWT and refresh token.
	jwtToken, refreshToken, err := a.generateTokens(userID)
	if err != nil {
		return LoginResponse{}, err
	}

	// 7. Hash the refresh token.
	refreshHash, err := a.HashPassword(refreshToken)
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
	user, err := a.queries.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, fmt.Errorf("invalid email / password")
		}
		return uuid.Nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	if err := a.CheckPasswordHash(input.Password, user.HashedPassword); err != nil {
		return uuid.Nil, fmt.Errorf("invalid email / password")
	}
	return uuid.Parse(user.ID)
}

func (a *AuthService) handleDeviceInfo(ctx context.Context, queriesTx interfaces.Querier, userID uuid.UUID, info DeviceInfo) (uuid.UUID, error) {
	// Try to fetch an existing device record for this user with matching attributes.
	foundDevice, err := queriesTx.GetDeviceInfoByUser(ctx, database.GetDeviceInfoByUserParams{
		UserID:         userID.String(),
		DeviceType:     info.DeviceType,
		Browser:        info.Browser,
		BrowserVersion: info.BrowserVersion,
		Os:             info.Os,
		OsVersion:      info.OsVersion,
	})
	if err != nil {
		// If no device record exists, create one.
		if errors.Is(err, sql.ErrNoRows) {
			newDeviceID, err := queriesTx.CreateDeviceInfo(ctx, database.CreateDeviceInfoParams{
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

	// If found, parse the device ID.
	deviceID, err := uuid.Parse(foundDevice)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse device ID: %w", err)
	}

	// Revoke any existing tokens for this device.
	if err := queriesTx.RevokeToken(ctx, database.RevokeTokenParams{
		RevokedAt:    sql.NullTime{Time: time.Now(), Valid: true},
		UserID:       userID.String(),
		DeviceInfoID: deviceID.String(),
	}); err != nil {
		return uuid.Nil, fmt.Errorf("failed to revoke token: %w", err)
	}

	return deviceID, nil
}

func (a *AuthService) generateTokens(userID uuid.UUID) (string, string, error) {
	// For this example, assume AuthService has a field tokenSecret (string)
	// and that a.auth is your AuthInterface.
	jwtToken, err := a.MakeJWT(userID, 15*time.Minute)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	refreshToken, err := a.MakeRefreshToken()
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

// isValidDeviceInfo checks that the device info meets basic criteria.
func isValidDeviceInfo(info DeviceInfo) bool {
	// Define valid device types (case-insensitive).
	validDeviceTypes := map[string]bool{
		"desktop": true,
		"mobile":  true,
	}
	if !validDeviceTypes[strings.ToLower(info.DeviceType)] {
		return false
	}
	// Ensure that essential fields are not empty.
	if info.Browser == "" || info.Os == "" {
		return false
	}
	// Optionally, validate version formats.
	if info.BrowserVersion != "" && !isValidVersion(info.BrowserVersion) {
		return false
	}
	if info.OsVersion != "" && !isValidVersion(info.OsVersion) {
		return false
	}
	return true
}

// isValidVersion uses a regex to validate version strings.
func isValidVersion(version string) bool {
	versionRegex := `^\d+(\.\d+)*$`
	matched, _ := regexp.MatchString(versionRegex, version)
	return matched
}

// sanitizeInput trims whitespace and enforces length limits.
func sanitizeInput(input string) string {
	input = strings.TrimSpace(input)
	if len(input) > 100 {
		input = input[:100]
	}
	return input
}
