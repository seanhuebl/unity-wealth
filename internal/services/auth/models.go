package auth

import (
	"regexp"

	"github.com/google/uuid"
)

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	UserID       uuid.UUID
	JWT          string
	RefreshToken string
}

type DeviceInfo struct {
	DeviceType     string `json:"device_type"`
	Browser        string `json:"browser"`
	BrowserVersion string `json:"browser_version"`
	Os             string `json:"os"`
	OsVersion      string `json:"os_version"`
}

func IsValidEmail(email string) bool {
	// Define a regex pattern for validating email
	emailRegex := `^[a-zA-Z0-9_%+\-][a-zA-Z0-9._%+\-]*@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`

	// Compile the regex and match the email
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}
