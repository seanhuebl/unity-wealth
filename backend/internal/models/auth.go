package models

import (
	"crypto/rand"
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

var (
	uppercaseRegex = regexp.MustCompile(`[A-Z]`)
	lowercaseRegex = regexp.MustCompile(`[a-z]`)
	digitRegex     = regexp.MustCompile(`\d`)
	specialRegex   = regexp.MustCompile(`[!@#\$%\^&\*\(\)_\+\-=\[\]\{\};':"\\|,.<>\/?]`)
)

type TokenType string

var RandReader = rand.Read

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	UserID       uuid.UUID
	RefreshToken string
	JWTToken     string
}

type LoginResponseData struct {
	Message string `json:"message"`
	Token   string `json:"token"`
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

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	if !uppercaseRegex.MatchString(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !lowercaseRegex.MatchString(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !digitRegex.MatchString(password) {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !specialRegex.MatchString(password) {
		return fmt.Errorf("password must contain at least one special character")
	}
	return nil
}
