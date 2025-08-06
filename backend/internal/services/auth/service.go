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
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/sentinels"
	"go.uber.org/zap"
)

type AuthService struct {
	SqlTxQuerier database.SqlTxQuerier
	UserQuerier  database.UserQuerier
	TokenGen     TokenGenerator
	TokenExtract TokenExtractor
	PwdHasher    PasswordHasher
	logger       *zap.Logger
}

func NewAuthService(SqlTxQuerier database.SqlTxQuerier, UserQuerier database.UserQuerier, TokenGen TokenGenerator, tokenExtract TokenExtractor, PwdHasher PasswordHasher, logger *zap.Logger) *AuthService {
	return &AuthService{
		SqlTxQuerier: SqlTxQuerier,
		UserQuerier:  UserQuerier,
		TokenGen:     TokenGen,
		TokenExtract: tokenExtract,
		PwdHasher:    PwdHasher,
		logger:       logger,
	}
}

func (a *AuthService) Login(ctx context.Context, input models.LoginInput) (models.LoginResponse, error) {
	start := time.Now()
	reqID, _ := ctx.Value(constants.RequestIDKey).(string)

	a.logger.Info(evtLoginAttempt,
		zap.String("request_id", reqID),
		zap.String("email", input.Email),
	)

	if !models.IsValidEmail(input.Email) {
		a.logger.Warn(evtLoginInvalidEmail,
			zap.String("request_id", reqID),
			zap.String("email", input.Email),
		)
		return models.LoginResponse{}, sentinels.ErrInvalidEmail
	}

	userID, err := a.ValidateCredentials(ctx, input)
	if err != nil {
		wrapped := fmt.Errorf("login: validate credentials: %w", err)
		a.logger.Warn(evtLoginInvalidCreds,
			zap.String("request_id", reqID),
			zap.String("email", input.Email),
			zap.Error(wrapped),
		)
		return models.LoginResponse{}, wrapped
	}

	a.logger.Info(evtLoginCredsValid,
		zap.String("request_id", reqID),
		zap.String("user_id", userID.String()),
	)

	logger := a.logger.With(
		zap.String("request_id", reqID),
		zap.String("user_id", userID.String()),
	)

	req, err := helpers.GetRequestFromContext(ctx)
	if err != nil {
		wrapped := fmt.Errorf("login: get req from ctx: %w", err)
		logger.Error(evtLoginMissingReq,
			zap.Error(wrapped),
		)
		return models.LoginResponse{}, wrapped
	}

	deviceInfo, err := GetDeviceInfoFromRequest(req)
	if err != nil {
		wrapped := fmt.Errorf("login: get device info from ctx: %w", err)
		logger.Error(evtLoginInvalidDeviceInfo,
			zap.Error(wrapped),
		)
		return models.LoginResponse{}, wrapped
	}

	jwtToken, refreshToken, err := a.GenerateTokens(userID)
	if err != nil {
		wrapped := fmt.Errorf("login: generate tokens: %w", err)
		logger.Error(evtLoginGenerateTokensFailed,
			zap.Error(wrapped),
		)
		return models.LoginResponse{}, wrapped
	}

	refreshHash, err := a.PwdHasher.HashPassword(refreshToken)
	if err != nil {
		wrapped := fmt.Errorf("login: hash refreshToken: %w", err)
		logger.Error(evtLoginHashRefTokenFailed,
			zap.Error(wrapped),
		)
		return models.LoginResponse{}, wrapped
	}

	dbStart := time.Now()

	tx, err := a.SqlTxQuerier.BeginTx(ctx, nil)
	if err != nil {
		wrapped := fmt.Errorf("login: begin sql tx: %w", err)
		logger.Error(evtLoginBeginSqlTxFailed,
			zap.Error(wrapped),
		)
		return models.LoginResponse{}, wrapped
	}

	defer tx.Rollback()
	queriesTx := a.SqlTxQuerier.WithTx(tx)
	deviceQ := database.NewRealDevicequerier(queriesTx)
	tokenQ := database.NewRealTokenQuerier(queriesTx)

	deviceID, err := a.HandleDeviceInfo(ctx, deviceQ, tokenQ, userID, deviceInfo)
	if err != nil {
		wrapped := fmt.Errorf("login: device info: %w", err)
		logger.Error(evtLoginHandleDeviceInfoFailed,
			zap.Error(wrapped),
		)
		return models.LoginResponse{}, wrapped
	}

	expiration := sql.NullTime{
		Time:  time.Now().Add(60 * 24 * time.Hour),
		Valid: true,
	}
	err = queriesTx.CreateRefreshToken(ctx, database.CreateRefreshTokenParams{
		ID:           uuid.New(),
		TokenHash:    refreshHash,
		ExpiresAt:    expiration,
		UserID:       userID,
		DeviceInfoID: deviceID,
	})
	if err != nil {
		wrapped := fmt.Errorf("login: %w: %v", sentinels.ErrDBExecFailed, err)
		logger.Error(evtLoginRefTokInsertDBFailed,
			zap.Error(wrapped),
		)
		return models.LoginResponse{}, wrapped
	}

	if err := tx.Commit(); err != nil {
		wrapped := fmt.Errorf("login: sql commit tx: %w", err)
		logger.Error(evtLoginSqlCommitTxFailed,
			zap.Error(wrapped),
		)
		return models.LoginResponse{}, wrapped
	}

	dbDuration := time.Since(dbStart)
	totalDuration := time.Since(start)

	logger.Info(evtLoginSuccess,
		zap.Duration("db_duration_ms", dbDuration),
		zap.Duration("total_duration_ms", totalDuration),
	)
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
			return uuid.Nil, fmt.Errorf("%w: %v", ErrInvalidCreds, err)
		}
		return uuid.Nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	if err := a.PwdHasher.CheckPasswordHash(input.Password, user.HashedPassword); err != nil {
		return uuid.Nil, fmt.Errorf("%w: %v", ErrInvalidCreds, err)
	}

	if user.ID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("validate credentials: nil user ID: %w", err)
	}
	return user.ID, nil
}

func (a *AuthService) HandleDeviceInfo(ctx context.Context, deviceQ database.DeviceQuerier, tokenQ database.TokenQuerier, userID uuid.UUID, info models.DeviceInfo) (uuid.UUID, error) {
	foundDevice, err := deviceQ.GetDeviceInfoByUser(ctx, database.GetDeviceInfoByUserParams{
		UserID:         userID,
		DeviceType:     info.DeviceType,
		Browser:        info.Browser,
		BrowserVersion: info.BrowserVersion,
		Os:             info.Os,
		OsVersion:      info.OsVersion,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			newDeviceID, err := deviceQ.CreateDeviceInfo(ctx, database.CreateDeviceInfoParams{
				ID:             uuid.New(),
				UserID:         userID,
				DeviceType:     info.DeviceType,
				Browser:        info.Browser,
				BrowserVersion: info.BrowserVersion,
				Os:             info.Os,
				OsVersion:      info.OsVersion,
			})
			if err != nil {
				return uuid.Nil, fmt.Errorf("failed to create new device: %w", err)
			}
			return newDeviceID, nil
		}
		return uuid.Nil, fmt.Errorf("failed to fetch device info: %w", err)
	}

	if foundDevice == uuid.Nil {
		return uuid.Nil, fmt.Errorf("nil device ID: %w", err)

	}

	if err := tokenQ.RevokeToken(ctx, database.RevokeTokenParams{
		RevokedAt:    sql.NullTime{Time: time.Now(), Valid: true},
		UserID:       userID,
		DeviceInfoID: foundDevice,
	}); err != nil {
		return uuid.Nil, fmt.Errorf("failed to revoke token: %w", err)
	}

	return foundDevice, nil
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
	return models.DeviceInfo{}, ErrInvalidDeviceInfo
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
