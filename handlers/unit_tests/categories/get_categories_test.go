package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v8"
	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/unity-wealth/cache"
	"github.com/seanhuebl/unity-wealth/handlers"
	"github.com/seanhuebl/unity-wealth/internal/config"
)

// TestGetCategories is a table-driven test for the GetCategories handler.
func TestGetCategories(t *testing.T) {
	// Define our test cases.
	tests := []struct {
		name                 string
		primaryResult        map[string]string // simulated value returned for primary categories
		primaryErr           error             // simulated error for primary categories
		detailedResult       map[string]string // simulated value returned for detailed categories
		detailedErr          error             // simulated error for detailed categories
		expectedStatus       int
		expectedResponseJSON string // expected JSON response as a string
	}{
		{
			name: "success",
			primaryResult: map[string]string{
				"1": `{"id":1,"name":"Primary Cat 1"}`,
			},
			primaryErr: nil,
			detailedResult: map[string]string{
				"2": `{"id":2,"name":"Detailed Cat 1"}`,
			},
			detailedErr:    nil,
			expectedStatus: http.StatusOK,
			expectedResponseJSON: `{
				"primary_categories": {"1": "{\"id\":1,\"name\":\"Primary Cat 1\"}"},
				"detailed_categories": {"2": "{\"id\":2,\"name\":\"Detailed Cat 1\"}"}
			}`,
		},
		{
			name:          "primary error",
			primaryResult: nil,
			primaryErr:    errors.New("primary error"),
			detailedResult: map[string]string{
				"2": `{"id":2,"name":"Detailed Cat 1"}`,
			},
			detailedErr:          nil,
			expectedStatus:       http.StatusInternalServerError,
			expectedResponseJSON: `{"error":"unable to load primary_categories"}`,
		},
		{
			name: "detailed error",
			primaryResult: map[string]string{
				"1": `{"id":1,"name":"Primary Cat 1"}`,
			},
			primaryErr:           nil,
			detailedResult:       nil,
			detailedErr:          errors.New("detailed error"),
			expectedStatus:       http.StatusInternalServerError,
			expectedResponseJSON: `{"error":"unable to load detailed_categories"}`,
		},
	}

	// Set Gin into test mode.
	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			// Create a redismock client and override the global RedisClient.
			mockRedis, mock := redismock.NewClientMock()
			cache.RedisClient = mockRedis

			expPrimary := mock.ExpectHGetAll("primary_categories")
			if tt.primaryErr != nil {
				expPrimary.SetErr(tt.primaryErr)
			} else {
				expPrimary.SetVal(tt.primaryResult)
				// Only set detailed expectation if primary call succeeds.
				expDetailed := mock.ExpectHGetAll("detailed_categories")
				if tt.detailedErr != nil {
					expDetailed.SetErr(tt.detailedErr)
				} else {
					expDetailed.SetVal(tt.detailedResult)
				}
			}

			// Create a test HTTP request and recorder.
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest("GET", "/categories", nil)
			c.Request = req
			
			blankConfig := &config.ApiConfig{}
			h := handlers.NewHandler(blankConfig)

			// Call the handler.
			h.GetCategories(c)

			// Check the status code.
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Unmarshal the JSON response so that ordering differences do not break the test.
			var gotResp, expectedResp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &gotResp); err != nil {
				t.Fatalf("error unmarshalling got response: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.expectedResponseJSON), &expectedResp); err != nil {
				t.Fatalf("error unmarshalling expected response: %v", err)
			}

			if diff := cmp.Diff(expectedResp, gotResp); diff != "" {
				t.Errorf("response mismatch (-expected +got):\n%s", diff)
			}

			// Ensure that all expectations were met.
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled Redis expectations: %v", err)
			}
		})
	}
}
