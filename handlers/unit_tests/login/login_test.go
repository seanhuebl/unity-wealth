package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seanhuebl/unity-wealth/handlers"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/mocks"
	"github.com/stretchr/testify/mock"
)

func TestLoginHandler(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite in-memory DB: %v", err)
	}
	defer db.Close()

	queries := database.New(db)

	_, err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			hashed_password TEXT NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	hashedPassword := "hashed_correct_password"

	_, err = db.Exec(`
		INSERT INTO users (id, email, hashed_password)
		VALUES (?, ?, ?);
	`, "11111111-2222-3333-4444-555555555555", "validuser@example.com", hashedPassword)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS device_info_logs (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			device_type TEXT NOT NULL,
			browser TEXT NOT NULL,
			browser_version TEXT NOT NULL,
			os TEXT NOT NULL,
			os_version TEXT NOT NULL,
			app_info TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_used_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create device_info_logs table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id TEXT PRIMARY KEY,
			token_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME,
			revoked_at DATETIME,
			user_id TEXT NOT NULL,
			device_info_id TEXT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
			FOREIGN KEY (device_info_id) REFERENCES device_info_logs (id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create refresh_tokens table: %v", err)
	}

	tests := []struct {
		name           string
		inputBody      map[string]string
		xDeviceInfo    string
		mockSetup      func(q *mocks.Quierier, mockAuth *mocks.AuthInterface)
		expectedStatus int
		expectedJSON   map[string]string
	}{
		{
			name: "Successful Login",
			inputBody: map[string]string{
				"email":    "validuser@example.com",
				"password": "correct_password",
			},
			xDeviceInfo: "os=iOS; os_version=14.4; device_type=Mobile; browser=Safari; browser_version=14.0",
			mockSetup: func(q *mocks.Quierier, mockAuth *mocks.AuthInterface) {

				mockAuth.On("CheckPasswordHash", "correct_password", "hashed_correct_password").
					Return(nil)

				mockAuth.On("MakeJWT", uuid.MustParse("11111111-2222-3333-4444-555555555555"),
					"test-secret", mock.AnythingOfType("time.Duration")).Return("fake-jwt-token", nil)

				mockAuth.On("MakeRefreshToken").Return("fake-refresh-token", nil)

				mockAuth.On("HashPassword", "fake-refresh-token").Return("refresh-hash", nil)

			},
			expectedStatus: http.StatusOK,
			expectedJSON: map[string]string{
				"token": "fake-jwt-token",
			},
		},

		{
			name: "User Not Found",
			inputBody: map[string]string{
				"email":    "unknownuser@example.com",
				"password": "some_password",
			},
			xDeviceInfo:    "os=Android; os_version=12; device_type=Mobile; browser=Chrome; browser_version=100.0",
			expectedStatus: http.StatusUnauthorized,
			expectedJSON: map[string]string{
				"error": "invalid email / password",
			},
		},

		{
			name: "Incorrect Password",
			inputBody: map[string]string{
				"email":    "validuser@example.com",
				"password": "wrong_password",
			},
			xDeviceInfo: "os=Windows; os_version=10; device_type=Desktop; browser=Firefox; browser_version=95.0",
			mockSetup: func(q *mocks.Quierier, mockAuth *mocks.AuthInterface) {
				mockAuth.On("CheckPasswordHash", "wrong_password", "hashed_correct_password").Return(fmt.Errorf("invalid email / password"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedJSON: map[string]string{
				"error": "invalid email / password",
			},
		},

		{
			name: "Invalid Email Format",
			inputBody: map[string]string{
				"email":    "invalid-email",
				"password": "some_password",
			},
			xDeviceInfo:    "os=Linux; os_version=Ubuntu 22.04; device_type=Desktop; browser=Brave; browser_version=1.38.0",
			expectedStatus: http.StatusBadRequest,
			expectedJSON: map[string]string{
				"error": "Invalid email format",
			},
		},

		{
			name:           "Empty Request Body",
			inputBody:      map[string]string{},
			xDeviceInfo:    "",
			expectedStatus: http.StatusBadRequest,
			expectedJSON: map[string]string{
				"error": "invalid request / data",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockQ := &mocks.Quierier{}
			mockAuth := &mocks.AuthInterface{}

			if tc.mockSetup != nil {
				tc.mockSetup(mockQ, mockAuth)
			}

			cfg := &handlers.ApiConfig{
				Port:        ":8080",
				Queries:     queries,
				TokenSecret: "test-secret",
				Database:    db,
				Auth:        mockAuth,
			}

			// Create Gin Engine
			router := gin.Default()
			router.POST("/login", cfg.Login)

			// Create Request
			jsonBody, err := json.Marshal(tc.inputBody)
			if err != nil {
				t.Fatalf("Failed to marshal input body: %v", err)
			}

			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			if tc.xDeviceInfo != "" {
				req.Header.Set("X-Device-Info", tc.xDeviceInfo)
			}

			// Record Response
			w := httptest.NewRecorder()

			// Execute Request
			router.ServeHTTP(w, req)

			// Validate Status
			if w.Code != tc.expectedStatus {
				t.Errorf("Status code mismatch. Expected %d, got %d", tc.expectedStatus, w.Code)
			}

			// Validate Response
			var got map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("Failed to unmarshal response body: %v", err)
			}

			// Compare JSON
			if diff := cmp.Diff(tc.expectedJSON, got); diff != "" {
				t.Errorf("Response body mismatch (-want +got):\n%s", diff)
			}

			// Verify Mocks
			mockQ.AssertExpectations(t)
			mockAuth.AssertExpectations(t)
		})
	}
}
