package category

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
)

func TestGetCategories(t *testing.T) {

	tests := []struct {
		name                 string
		primaryResult        map[string]string
		primaryErr           error
		detailedResult       map[string]string
		detailedErr          error
		expectedStatus       int
		expectedResponseJSON string
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

	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			mockRedis, mock := redismock.NewClientMock()
			cache.RedisClient = mockRedis

			expPrimary := mock.ExpectHGetAll("primary_categories")
			if tt.primaryErr != nil {
				expPrimary.SetErr(tt.primaryErr)
			} else {
				expPrimary.SetVal(tt.primaryResult)

				expDetailed := mock.ExpectHGetAll("detailed_categories")
				if tt.detailedErr != nil {
					expDetailed.SetErr(tt.detailedErr)
				} else {
					expDetailed.SetVal(tt.detailedResult)
				}
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest("GET", "/categories", nil)
			c.Request = req

			h := NewHandler()

			h.GetCategories(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

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

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled Redis expectations: %v", err)
			}
		})
	}
}
