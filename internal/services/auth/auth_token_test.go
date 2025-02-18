package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {

	tests := []struct {
		name         string
		userID       uuid.UUID
		tokenSecret  string
		expiresIn    time.Duration
		wantErr      bool
		verifyClaims jwt.RegisteredClaims
	}{
		{
			name:        "Valid token",
			userID:      uuid.New(),
			tokenSecret: "testsecret",
			expiresIn:   time.Hour,
			wantErr:     false,
			verifyClaims: jwt.RegisteredClaims{
				Issuer:  "testaccess",
				Subject: "",
			},
		},
		{
			name:        "Invalid secret",
			userID:      uuid.New(),
			tokenSecret: "",
			expiresIn:   time.Hour,
			wantErr:     true,
		},
		{
			name:        "Expired token",
			userID:      uuid.New(),
			tokenSecret: "testsecret",
			expiresIn:   -time.Hour,
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tokenGen := NewRealTokenGenerator(tc.tokenSecret, "testaccess")
			token, err := tokenGen.MakeJWT(tc.userID, tc.expiresIn)

			if (err != nil) != tc.wantErr {
				t.Errorf("MakeJWT() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr {
				parsedToken, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte(tc.tokenSecret), nil
				})
				if err != nil {
					t.Errorf("Failed to parse token: %v", err)
					return
				}

				claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
				if !ok || !parsedToken.Valid {
					t.Errorf("Token is invalid")
					return
				}

				if diff := cmp.Diff(tc.verifyClaims.Issuer, claims.Issuer); diff != "" {
					t.Errorf("Issuer mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tc.userID.String(), claims.Subject); diff != "" {
					t.Errorf("Subject mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenGen := NewRealTokenGenerator("testsecret", "testaccess")
	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantUUID    uuid.UUID
		wantErr     bool
	}{
		{
			name: "Valid token",
			tokenString: func() string {
				tokenGen := NewRealTokenGenerator("testsecret", "testaccess")
				token, _ := tokenGen.MakeJWT(userID, time.Hour)
				return token
			}(),
			tokenSecret: "testsecret",
			wantUUID:    userID,
			wantErr:     false,
		},
		{
			name: "Invalid token secret",
			tokenString: func() string {
				tokenGen := NewRealTokenGenerator("wrongsecret", "testaccess")
				token, _ := tokenGen.MakeJWT(uuid.New(), time.Hour)
				return token
			}(),
			tokenSecret: "wrongsecret",
			wantUUID:    uuid.Nil,
			wantErr:     true,
		},
		{
			name:        "Malformed token",
			tokenString: "malformed.token.string",
			tokenSecret: "testsecret",
			wantUUID:    uuid.Nil,
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := tokenGen.ValidateJWT(tc.tokenString)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if err != nil {
				// If an error is expected, there's no need to check further.
				return
			}
			// Extract the subject from the claims and parse it as a UUID.
			gotUUID, err := uuid.Parse(claims.Subject)
			if err != nil {
				t.Errorf("failed to parse subject as uuid: %v", err)
				return
			}
			if diff := cmp.Diff(tc.wantUUID, gotUUID); diff != "" {
				t.Errorf("UUID mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	tokenExtractor := NewRealTokenExtractor()
	tests := map[string]struct {
		input         http.Header
		expectedValue string
	}{
		"simple":                 {input: http.Header{"Authorization": []string{"Bearer 1234"}}, expectedValue: "1234"},
		"wrong auth header":      {input: http.Header{"Authorization": []string{"ApiKey 1234"}}, expectedValue: "malformed authorization header"},
		"incomplete auth header": {input: http.Header{"Authorization": []string{"Bearer "}}, expectedValue: "malformed authorization header"},
		"no auth header":         {input: http.Header{"Authorization": []string{""}}, expectedValue: fmt.Sprint(ErrNoAuthHeaderIncluded)},
	}

	for test, tc := range tests {
		t.Run(test, func(t *testing.T) {
			receivedValue, err := tokenExtractor.GetBearerToken(tc.input)
			var diff string
			if err != nil {
				diff = cmp.Diff(tc.expectedValue, fmt.Sprint(err))
			} else {
				diff = cmp.Diff(tc.expectedValue, receivedValue)
			}
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestMakeRefreshToken(t *testing.T) {
	tokenGen := NewRealTokenGenerator("testsecret", "testaccess")
	tests := []struct {
		name     string
		mockRand func([]byte) (int, error)
		wantErr  bool
	}{
		{
			name: "Valid refresh token",
			mockRand: func(b []byte) (int, error) {
				return rand.Read(b)
			},
			wantErr: false,
		},
		{
			name: "Error generating refresh token",
			mockRand: func(b []byte) (int, error) {
				return 0, errors.New("simulated error")
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Override randReader temporarily for this test
			origRandReader := RandReader
			RandReader = tc.mockRand
			defer func() { RandReader = origRandReader }()

			token, err := tokenGen.MakeRefreshToken()

			if (err != nil) != tc.wantErr {
				t.Errorf("MakeRefreshToken() error = %v, wantErr %v", err, tc.wantErr)
			}

			if !tc.wantErr {
				if len(token) != 64 {
					t.Errorf("Expected token length 64, got %d", len(token))
				}
			}
		})
	}
}
