package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mssola/user_agent"
	"github.com/seanhuebl/unity-wealth/internal/auth"
	"github.com/seanhuebl/unity-wealth/internal/database"
)

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type DeviceInfo struct {
	DeviceType     string `json:"device_type"`
	Browser        string `json:"browser"`
	BrowserVersion string `json:"browser_version"`
	Os             string `json:"os"`
	OsVersion      string `json:"os_version"`
}

func Login(ctx *gin.Context, cfg *ApiConfig) {
	var input LoginInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid input data: " + err.Error(),
		})
		return
	}

	if !isValidEmail(input.Email) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email format",
		})
		return
	}

	userID, err := ValidateCredentials(ctx, cfg, &input)
	if err != nil {
		if err.Error() != "failed to fetch user" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	clientDeviceInfo, err := getDeviceInfo(ctx.Request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "device information could not be verified",
		})
		return
	}

	tx, err := cfg.Database.BeginTx(ctx, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start transaction: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()
	queriesTx := cfg.Queries.WithTx(tx)

	deviceID, err := handleDeviceInfo(ctx, queriesTx, userID, clientDeviceInfo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	JWT, refreshToken, err := generateTokens(userID, cfg.TokenSecret)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	refreshHash, err := auth.HashPassword(refreshToken)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to hash refresh token: " + err.Error(),
		})
		return
	}

	err = queriesTx.CreateRefreshToken(ctx, database.CreateRefreshTokenParams{
		TokenHash:    refreshHash,
		UserID:       userID,
		DeviceInfoID: deviceID,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create refresh token entry: " + err.Error(),
		})
		return
	}

	if err := tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit transaction: " + err.Error(),
		})
		return
	}

	setRefreshTokenCookie(ctx, refreshToken)

	ctx.JSON(http.StatusOK, gin.H{
		"token": JWT,
	})
}

func castToUUID(value interface{}) (uuid.UUID, error) {
	switch v := value.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		return uuid.Parse(v)
	default:
		return uuid.Nil, errors.New("failed to cast to UUID")
	}
}

func getDeviceInfo(req *http.Request) (DeviceInfo, error) {
	// Check for X-Device-Info header
	xDeviceInfo := req.Header.Get("X-Device-Info")
	if xDeviceInfo != "" {
		deviceInfo := parseDeviceInfoFromHeader(xDeviceInfo)
		if isValidDeviceInfo(deviceInfo) {
			return deviceInfo, nil
		}
	}

	// Fallback to User-Agent header
	userAgent := req.Header.Get("User-Agent")
	if userAgent != "" {
		deviceInfo := parseUserAgent(userAgent)
		if isValidDeviceInfo(deviceInfo) {
			return deviceInfo, nil
		}
	}

	// If both are invalid, return an error
	return DeviceInfo{}, fmt.Errorf("invalid or unknown device information")
}

// parseDeviceInfoFromHeader parses the X-Device-Info header into a DeviceInfo struct.
func parseDeviceInfoFromHeader(header string) DeviceInfo {
	deviceInfo := DeviceInfo{}

	// Split the header into key-value pairs
	pairs := strings.Split(header, ";")
	for _, pair := range pairs {
		kv := strings.Split(strings.TrimSpace(pair), "=")
		if len(kv) == 2 {
			key, value := strings.ToLower(strings.TrimSpace(kv[0])), strings.TrimSpace(kv[1])
			switch key {
			case "os":
				deviceInfo.Os = sanitizeInput(value)
			case "os_version":
				deviceInfo.OsVersion = sanitizeInput(value)
			case "device_type":
				deviceInfo.DeviceType = sanitizeInput(value)
			case "browser":
				deviceInfo.Browser = sanitizeInput(value)
			case "browser_version":
				deviceInfo.BrowserVersion = sanitizeInput(value)
			default:
				// Ignore unexpected keys
			}
		}
	}

	return deviceInfo
}

// parseUserAgent parses the User-Agent header and provides fallback device info.
func parseUserAgent(userAgent string) DeviceInfo {
	userAgent = sanitizeInput(userAgent)
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
		Os:             sanitizeInput(ua.OS()),
		OsVersion:      "", // User-Agent does not typically provide OS version
	}
}
func isValidDeviceInfo(info DeviceInfo) bool {
	// Define valid device types
	validDeviceTypes := map[string]bool{
		"Desktop": true,
		"Mobile":  true,
	}

	// Validate DeviceType
	if !validDeviceTypes[info.DeviceType] {
		return false
	}

	// Validate versions using regex
	if !isValidVersion(info.OsVersion) || !isValidVersion(info.BrowserVersion) {
		return false
	}

	// Validate required fields are non-empty
	return info.Browser != "" && info.Os != ""
}

func isValidVersion(version string) bool {
	versionRegex := `^\d+(\.\d+)*$` // Matches version strings like "10.0.1" or "95.0.4638.69"
	matched, _ := regexp.MatchString(versionRegex, version)
	return matched
}

func sanitizeInput(input string) string {
	// Trim leading and trailing whitespace
	input = strings.TrimSpace(input)

	// Replace problematic characters
	// Replace single quotes with escaped versions
	input = strings.ReplaceAll(input, "'", "''")

	// Optional: Enforce length limits
	if len(input) > 100 {
		input = input[:100] // Truncate to 100 characters
	}

	return input
}

func isValidEmail(email string) bool {
	// Define a regex pattern for validating email
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	// Compile the regex and match the email
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func handleDeviceInfo(ctx context.Context, queriesTx *database.Queries, userID uuid.UUID, info DeviceInfo) (uuid.UUID, error) {
	foundDevice, err := queriesTx.GetDeviceInfoByUser(ctx, database.GetDeviceInfoByUserParams{
		UserID:         userID,
		DeviceType:     info.DeviceType,
		Browser:        info.Browser,
		BrowserVersion: info.BrowserVersion,
		Os:             info.Os,
		OsVersion:      info.OsVersion,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No device found: create a new device
			newDeviceID, err := queriesTx.CreateDeviceInfo(ctx, database.CreateDeviceInfoParams{
				UserID:         userID,
				DeviceType:     info.DeviceType,
				Browser:        info.Browser,
				BrowserVersion: info.BrowserVersion,
				Os:             info.Os,
				OsVersion:      info.OsVersion,
			})
			if err != nil {
				return uuid.UUID{}, fmt.Errorf("failed to create new device")
			}
			return castToUUID(newDeviceID)

		}
		return uuid.UUID{}, fmt.Errorf("failed to fetch device info")
	}
	// Device found: Typecast the ID
	deviceID, err := castToUUID(foundDevice)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to cast device ID")
	}

	// Revoke token for the existing device
	if err := queriesTx.RevokeToken(ctx, database.RevokeTokenParams{
		UserID:       userID,
		DeviceInfoID: deviceID,
	}); err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to revoke token")
	}
	return deviceID, nil
}

func generateTokens(userID uuid.UUID, secret string) (string, string, error) {
	JWT, err := auth.MakeJWT(userID, secret, time.Minute*15)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		return "", "", err
	}
	return JWT, refreshToken, nil
}

func setRefreshTokenCookie(ctx *gin.Context, refreshToken string) {
	isProduction := os.Getenv("ENV") == "prod"
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	if cookieDomain == "" {
		cookieDomain = "localhost"
	}

	cookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Domain:   cookieDomain, // Use 'localhost' for local testing
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   isProduction, // Disable 'Secure' for HTTP testing
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(ctx.Writer, &cookie)
}

func ValidateCredentials(ctx *gin.Context, cfg *ApiConfig, input *LoginInput) (uuid.UUID, error) {
	user, err := cfg.Queries.GetUserByEmail(ctx, input.Email)
	fmt.Printf("Error from GetUserByEmail: %v (type: %T)\n", err, err)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, fmt.Errorf("invalid email / password")
		}
		return uuid.Nil, fmt.Errorf("failed to fetch user")
	}
	err = cfg.Auth.CheckPasswordHash(input.Password, user.HashedPassword)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid email / password")
	}
	return castToUUID(user.ID)
}
