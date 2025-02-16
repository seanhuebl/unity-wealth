package services__test

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/unity-wealth/services"
	"golang.org/x/crypto/bcrypt"
)

// CustomHashPassword allows testing different bcrypt costs for invalid cost scenarios
func CustomHashPassword(password string, cost int) (string, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return "", bcrypt.InvalidCostError(cost)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func TestHashPassword(t *testing.T) {
	authSvc := services.NewAuthService(os.Getenv("TOKEN_TYPE"), os.Getenv("TOKEN_SECRET"), nil)
	tests := []struct {
		name     string
		password string
		cost     int
		wantErr  bool
		errType  error // Expected error type
	}{
		{
			name:     "Valid cost",
			password: "securePassword123",
			cost:     bcrypt.DefaultCost,
			wantErr:  false,
		},
		{
			name:     "Invalid cost - Exceeding max cost",
			password: "securePassword123",
			cost:     bcrypt.MaxCost + 1,
			wantErr:  true,
			errType:  bcrypt.InvalidCostError(bcrypt.MaxCost + 1),
		},
		{
			name:     "Invalid cost - Negative cost",
			password: "securePassword123",
			cost:     -1,
			wantErr:  true,
			errType:  bcrypt.InvalidCostError(-1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := CustomHashPassword(tt.password, tt.cost)

			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				// Compare error types using cmp
				if diff := cmp.Diff(tt.errType.Error(), err.Error()); diff != "" {
					t.Errorf("HashPassword() unexpected error (-want +got):\n%s", diff)
				}
			} else {
				// Ensure valid hashes are verifiable
				if err := authSvc.CheckPasswordHash(tt.password, hash); err != nil {
					t.Errorf("HashPassword() generated a hash that could not be verified: %v", err)
				}
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	authSvc := services.NewAuthService(os.Getenv("TOKEN_TYPE"), os.Getenv("TOKEN_SECRET"), nil)
	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
		errMsg   string // Expected error message
	}{
		{
			name:     "Valid password and hash",
			password: "securePassword123",
			hash:     func() string { h, _ := authSvc.HashPassword("securePassword123"); return h }(),
			wantErr:  false,
		},
		{
			name:     "Invalid password",
			password: "wrongPassword",
			hash:     func() string { h, _ := authSvc.HashPassword("securePassword123"); return h }(),
			wantErr:  true,
			errMsg:   bcrypt.ErrMismatchedHashAndPassword.Error(),
		},
		{
			name:     "Invalid hash prefix",
			password: "securePassword123",
			hash:     "2x$10$" + strings.Repeat("a", 54), // '2x' is an invalid bcrypt prefix, padding added for correct length
			wantErr:  true,
			errMsg:   "crypto/bcrypt: bcrypt hashes must start with '$', but hashedSecret started with '2'",
		},
		{
			name:     "Invalid hash format",
			password: "securePassword123",
			hash:     "$2x$10$" + strings.Repeat("a", 53), // padded to correct length
			wantErr:  true,
			errMsg:   "crypto/bcrypt: hashedPassword is not the hash of the given password",
		},
		{
			name:     "Secret too short",
			password: "securePassword123",
			hash:     "$2x$10$",
			wantErr:  true,
			errMsg:   "crypto/bcrypt: hashedSecret too short to be a bcrypted password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authSvc.CheckPasswordHash(tt.password, tt.hash)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				// Compare the actual error message with the expected message
				if diff := cmp.Diff(tt.errMsg, err.Error()); diff != "" {
					t.Errorf("CheckPasswordHash() unexpected error (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {

	authSvc := services.NewAuthService(os.Getenv("TOKEN_TYPE"), os.Getenv("TOKEN_SECRET"), nil)
	tests := []struct {
		name       string
		password   string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "Valid password",
			password:   "StrongPass1!",
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name:       "Too short",
			password:   "Short1!",
			wantErr:    true,
			wantErrMsg: "password must be at least 8 characters long",
		},
		{
			name:       "No uppercase",
			password:   "weakpass1!",
			wantErr:    true,
			wantErrMsg: "password must contain at least one uppercase letter",
		},
		{
			name:       "No lowercase",
			password:   "WEAKPASS1!",
			wantErr:    true,
			wantErrMsg: "password must contain at least one lowercase letter",
		},
		{
			name:       "No digit",
			password:   "WeakPass!",
			wantErr:    true,
			wantErrMsg: "password must contain at least one digit",
		},
		{
			name:       "No special character",
			password:   "WeakPass1",
			wantErr:    true,
			wantErrMsg: "password must contain at least one special character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authSvc.ValidatePassword(tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				if diff := cmp.Diff(tt.wantErrMsg, err.Error()); diff != "" {
					t.Errorf("Error message mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
