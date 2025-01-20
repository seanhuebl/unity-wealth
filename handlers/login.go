package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
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
			"error": "invalid email / password",
		})
		return
	}
	userID := user.ID.(uuid.UUID)
	clientDeviceInfo := getDeviceInfo(ctx.Request)
	var deviceID uuid.UUID
	var ok bool

	foundDevice, err := cfg.Queries.GetDeviceInfoByUser(ctx, database.GetDeviceInfoByUserParams{
		UserID:         userID,
		DeviceType:     clientDeviceInfo.DeviceType,
		Browser:        clientDeviceInfo.Browser,
		BrowserVersion: clientDeviceInfo.BrowserVersion,
		Os:             clientDeviceInfo.Os,
		OsVersion:      clientDeviceInfo.OsVersion,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No device found: create a new device
			newDeviceID, err := cfg.Queries.CreateDeviceInfo(ctx, database.CreateDeviceInfoParams{
				UserID:         userID,
				DeviceType:     clientDeviceInfo.DeviceType,
				Browser:        clientDeviceInfo.Browser,
				BrowserVersion: clientDeviceInfo.BrowserVersion,
				Os:             clientDeviceInfo.Os,
				OsVersion:      clientDeviceInfo.OsVersion,
			})
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error": "failed to create new device info: " + err.Error(),
				})
				return
			}

			// Typecast new device ID
			deviceID, ok = newDeviceID.(uuid.UUID)
			if !ok {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error": "failed to retrieve new device ID as UUID",
				})
				return
			}
		} else {
			// Critical error
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to fetch device info: " + err.Error(),
			})
			return
		}
	} else {
		// Device found: Typecast the ID
		deviceID, ok := foundDevice.(uuid.UUID)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to retrieve device ID as UUID",
			})
			return
		}

		// Revoke token for the existing device
		err = cfg.Queries.RevokeToken(ctx, database.RevokeTokenParams{
			UserID:       userID,
			DeviceInfoID: deviceID,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to revoke token: " + err.Error(),
			})
			return
		}
	}

	// Generate and pass across JWT and refresh token
	JWT, err := auth.MakeJWT(userID, cfg.TokenSecret, time.Minute*15)
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
		return
	}
	refreshHash, err := auth.HashPassword(refreshToken)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to hash refresh token: " + err.Error(),
		})
		return
	}

	err = cfg.Queries.CreateRefreshToken(ctx, database.CreateRefreshTokenParams{
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
