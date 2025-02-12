package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/handlers"
	"golang.org/x/crypto/bcrypt"

	"github.com/seanhuebl/unity-wealth/internal/config"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/mocks"
	"github.com/stretchr/testify/mock"
)

func TestValidateCredentials(t *testing.T) {
	userID := uuid.New().String()
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	testCases := []struct {
		name          string
		input         handlers.LoginInput
		mockEmail     string
		mockUser      database.GetUserByEmailRow
		mockError     error
		mockPassword  string
		mockHash      string
		mockAuthError error
		expectedUUID  string
		expectedError error
	}{
		{
			name: "Valid credentials",
			input: handlers.LoginInput{
				Email:    "valid@example.com",
				Password: "correct-password",
			},
			mockEmail:     "valid@example.com",
			mockUser:      database.GetUserByEmailRow{ID: userID, HashedPassword: string(hash)},
			mockError:     nil,
			mockPassword:  "correct-password",
			mockHash:      string(hash),
			mockAuthError: nil,
			expectedUUID:  userID,
			expectedError: nil,
		},
		{
			name: "Invalid password",
			input: handlers.LoginInput{
				Email:    "valid@example.com",
				Password: "wrong-password",
			},
			mockEmail:     "valid@example.com",
			mockUser:      database.GetUserByEmailRow{ID: userID, HashedPassword: string(hash)},
			mockError:     nil,
			mockPassword:  "wrong-password",
			mockHash:      "hashed-pass",
			mockAuthError: fmt.Errorf("invalid email / password"),
			expectedUUID:  uuid.Nil.String(),
			expectedError: fmt.Errorf("invalid email / password"),
		},
		{
			name: "User not found",
			input: handlers.LoginInput{
				Email:    "nonexistent@example.com",
				Password: "any-password",
			},
			mockEmail:     "nonexistent@example.com",
			mockUser:      database.GetUserByEmailRow{},
			mockError:     sql.ErrNoRows,
			mockPassword:  "",
			mockHash:      "",
			mockAuthError: sql.ErrNoRows,
			expectedUUID:  uuid.Nil.String(),
			expectedError: fmt.Errorf("invalid email / password"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize mocks
			mockQueries := mocks.NewQuerier(t)
			mockAuth := mocks.NewAuthInterface(t)

			// Set up mock behavior for database queries
			mockQueries.On("GetUserByEmail", mock.Anything, mock.Anything).
				Return(func(ctx context.Context, email string) (database.GetUserByEmailRow, error) {
					if email == "valid@example.com" {
						return tc.mockUser, nil
					}
					if email == "nonexistent@example.com" {
						return database.GetUserByEmailRow{}, sql.ErrNoRows
					}
					return database.GetUserByEmailRow{}, fmt.Errorf("unknown error")
				})
			if tc.name != "User not found" {

				mockAuth.On("CheckPasswordHash", tc.input.Password, string(hash)).Return(func(password string, hash string) error {
					// Dynamically handle return based on password
					if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
						return fmt.Errorf("password mismatch")
					}
					return nil
				})
			}
			// Create API config with mocks
			cfg := &config.ApiConfig{
				Queries: mockQueries,
				Auth:    mockAuth,
			}

			// Create a test gin.Context
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder() // Use httptest.ResponseRecorder as the ResponseWriter
			c, _ := gin.CreateTestContext(w)

			// Call the function under test
			resultUUID, err := handlers.ValidateCredentials(c, cfg, &tc.input)
			// Compare results using `cmp`
			if diff := cmp.Diff(tc.expectedError, err, cmp.Comparer(func(e1, e2 error) bool {
				if e1 == nil && e2 == nil {
					return true
				}
				if e1 == nil || e2 == nil {
					return false
				}
				return e1.Error() == e2.Error()
			})); diff != "" {
				t.Errorf("unexpected error (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.expectedUUID, resultUUID.String()); diff != "" {
				t.Errorf("unexpected UUID (-want +got):\n%s", diff)
			}

			// Assert expectations dynamically
			if tc.name == "User not found" {
				mockAuth.AssertNotCalled(t, "CheckPasswordHash", mock.Anything, mock.Anything)
			} else {
				mockAuth.AssertExpectations(t)
			}

			mockQueries.AssertExpectations(t)
		})
	}
}
