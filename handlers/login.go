package handlers

import (
	"database/sql"
	"errors"
	"net/http"
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
	user, err := cfg.Queries.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid email / password",
			})
			return
		}

		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to fetch user: " + err.Error(),
		})
		return
	}

	err = auth.CheckPasswordHash(input.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid email / password" + err.Error(),
		})
		return
	}
	userID := user.ID.(uuid.UUID)
	clientDeviceInfo := getDeviceInfo(ctx.Request)

	device, err := cfg.Queries.GetDeviceInfoByUser(ctx, database.GetDeviceInfoByUserParams{
		UserID:         userID,
		DeviceType:     clientDeviceInfo.DeviceType,
		Browser:        clientDeviceInfo.Browser,
		BrowserVersion: clientDeviceInfo.BrowserVersion,
		Os:             clientDeviceInfo.Os,
		OsVersion:      clientDeviceInfo.OsVersion,
	})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to fetch device info: " + err.Error(),
			})
			return
		}
	} else {
		err = cfg.Queries.RevokeToken(ctx, database.RevokeTokenParams{
			UserID:       userID,
			DeviceInfoID: device.(uuid.UUID),
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to revoke token: " + err.Error(),
			})
			return
		}
	}

	// Generate and pass across JWT and refresh token
	JWT, err := auth.MakeJWT(user.ID.(uuid.UUID), cfg.TokenSecret, time.Minute*15)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create JWT: " + err.Error(),
		})
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create refresh token: " + err.Error(),
		})
	}
	refreshHash, err := auth.HashPassword(refreshToken)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to hash refresh token: " + err.Error(),
		})
	}

	err = cfg.Queries.CreateRefreshToken(ctx, database.CreateRefreshTokenParams{
		TokenHash:    refreshHash,
		UserID:       userID,
		DeviceInfoID: device.(uuid.UUID),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create refresh token entry: " + err.Error(),
		})
	}
	cookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Domain:   "localhost", // Use 'localhost' for local testing
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false, // Disable 'Secure' for HTTP testing
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(ctx.Writer, &cookie)

	ctx.JSON(http.StatusOK, gin.H{
		"token": JWT,
	})
}

func getDeviceInfo(req *http.Request) DeviceInfo {

	xDeviceInfo := req.Header.Get("X-Device-Info")
	if xDeviceInfo != "" {
		return parseDeviceInfoFromHeader(xDeviceInfo)
	}

	userAgent := req.Header.Get("User-Agent")
	ua := user_agent.New(userAgent)

	deviceType := "Desktop"
	if ua.Mobile() {
		deviceType = "Mobile"
	}

	os := ua.OS()
	browser, browserVersion := ua.Browser()
	return DeviceInfo{
		DeviceType:     deviceType,
		Browser:        browser,
		BrowserVersion: browserVersion,
		Os:             os,
		OsVersion:      "",
	}
}

// X-Device-Info
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
				deviceInfo.Os = value
			case "os_version":
				deviceInfo.OsVersion = value
			case "device_type":
				deviceInfo.DeviceType = value
			case "browser":
				deviceInfo.Browser = value
			case "browser_version":
				deviceInfo.BrowserVersion = value
			}
		}
	}

	return deviceInfo
}
