package auth_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
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

var pwdHasher = auth.NewRealPwdHasher()

func TestHashPassword(t *testing.T) {
	t.Parallel()
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

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			hash, err := CustomHashPassword(tc.password, tc.cost)

			if (err != nil) != tc.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if tc.wantErr {
				// Compare error types using cmp
				if diff := cmp.Diff(tc.errType.Error(), err.Error()); diff != "" {
					t.Errorf("HashPassword() unexpected error (-want +got):\n%s", diff)
				}
			} else {
				// Ensure valid hashes are verifiable
				if err := pwdHasher.CheckPasswordHash(tc.password, hash); err != nil {
					t.Errorf("HashPassword() generated a hash that could not be verified: %v", err)
				}
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	t.Parallel()
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
			hash:     func() string { h, _ := pwdHasher.HashPassword("securePassword123"); return h }(),
			wantErr:  false,
		},
		{
			name:     "Invalid password",
			password: "wrongPassword",
			hash:     func() string { h, _ := pwdHasher.HashPassword("securePassword123"); return h }(),
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

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := pwdHasher.CheckPasswordHash(tc.password, tc.hash)

			if (err != nil) != tc.wantErr {
				t.Fatalf("got err=%v, wantErr=%v", err, tc.wantErr)
			}

			if !tc.wantErr {
				return
			}

			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				return
			}

			if tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
				t.Fatalf("error mismatch:\n(-want substring +got)\n- %q\n+ %q",
					tc.errMsg, err.Error())
			}
		})
	}
}
