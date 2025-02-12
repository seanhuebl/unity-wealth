package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/unity-wealth/cache"
	"github.com/seanhuebl/unity-wealth/handlers"
)

func TestGetPrimaryCategoryByID_WithRedismock(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a redismock client.
	client, mock := redismock.NewClientMock()
	origRedisClient := cache.RedisClient
	defer func() { cache.RedisClient = origRedisClient }()
	cache.RedisClient = client

	tests := []struct {
		name           string
		id             string
		hgetResult     string
		hgetError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "category found",
			id:             "123",
			hgetResult:     "CategoryA",
			hgetError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"primary_category": "CategoryA"},
		},
		{
			name:           "category not found",
			id:             "456",
			hgetResult:     "",
			hgetError:      redis.Nil,
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]interface{}{"error": "primary category not found"},
		},
		{
			name:           "redis error",
			id:             "789",
			hgetResult:     "",
			hgetError:      redis.TxFailedErr, // or any other error
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]interface{}{"error": "unable to load primary category"},
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Setup the redismock expectation:
			cmd := mock.ExpectHGet("primary_categories", tc.id)
			if tc.hgetError != nil {
				cmd.SetErr(tc.hgetError)
			} else {
				cmd.SetVal(tc.hgetResult)
			}

			// Prepare the Gin context.
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			// Set a valid HTTP request on the Gin context.
			c.Request = httptest.NewRequest(http.MethodGet, "/primary_categories/"+tc.id, nil)
			c.Params = gin.Params{{Key: "id", Value: tc.id}}

			h := handlers.NewHandler(nil, nil, nil)

			// Call the handler.
			h.GetPrimaryCategoryByID(c)

			// Assert the HTTP status.
			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			// Parse and compare the response body.
			var body map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("failed to unmarshal response body: %v", err)
			}
			if diff := cmp.Diff(tc.expectedBody, body); diff != "" {
				t.Errorf("response body mismatch (-want +got):\n%s", diff)
			}

			// Verify that all expectations were met.
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
