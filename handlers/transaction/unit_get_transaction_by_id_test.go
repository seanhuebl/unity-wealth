package transaction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/database"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/stretchr/testify/require"
)

func TestGetTxByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name               string
		userID             uuid.UUID
		userIDErr          error
		txID               string
		txErr              error
		expectedErr        string
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name:               "success",
			userID:             uuid.New(),
			txID:               uuid.NewString(),
			expectedStatusCode: http.StatusOK,
			expectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"date":              "2025-03-05",
					"merchant":          "costco",
					"amount":            125.98,
					"detailed_category": 40,
				},
			},
		},
		{
			name:               "unauthorized: user ID is uuid.NIL",
			userID:             uuid.Nil,
			txID:               uuid.NewString(),
			userIDErr:          errors.New("user ID not found in context"),
			expectedErr:        "unauthorized",
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "unauthorized: user ID not UUID",
			userID:             uuid.Nil,
			userIDErr:          errors.New("user ID is not UUID"),
			txID:               uuid.NewString(),
			expectedErr:        "unauthorized",
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "error getting tx",
			userID:             uuid.New(),
			txID:               uuid.NewString(),
			txErr:              errors.New("error getting transaction"),
			expectedErr:        "unable to get transaction",
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:               "invalid txID in req",
			userID:             uuid.New(),
			txID:               "",
			expectedErr:        "invalid id",
			expectedStatusCode: http.StatusBadRequest,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockTxQ := dbmocks.NewTransactionQuerier(t)
			w := httptest.NewRecorder()
			svc := transaction.NewTransactionService(mockTxQ)
			req := httptest.NewRequest("GET", fmt.Sprintf("/transactions/%v", tc.txID), nil)

			dummyRow := database.GetUserTransactionByIDRow{
				ID:                 tc.txID,
				UserID:             tc.userID.String(),
				TransactionDate:    "2025-03-05",
				Merchant:           "costco",
				AmountCents:        12598,
				DetailedCategoryID: 40,
			}

			if tc.userIDErr == nil {
				mockTxQ.On("GetUserTransactionByID", context.Background(), database.GetUserTransactionByIDParams{
					UserID: tc.userID.String(),
					ID:     tc.txID,
				}).Return(dummyRow, tc.txErr)
			}

			h := NewHandler(svc)

			router := gin.New()
			if tc.txID == "" {
				c, _ := gin.CreateTestContext(w)
				c.Request = req
				c.Params = gin.Params{{Key: "id", Value: ""}}
				h.GetTransactionByID(c)
			} else {

				router.GET("/transactions/:id", func(c *gin.Context) {
					if tc.name == "unauthorized: user ID not UUID" {
						c.Set(string(constants.UserIDKey), "userID")
					} else {
						c.Set(string(constants.UserIDKey), tc.userID)
					}
					h.GetTransactionByID(c)
				})
			}

			router.ServeHTTP(w, req)

			var actualResponse map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
			require.NoError(t, err)

			actualResponse = convertResponseFloatToInt(actualResponse)
			if tc.expectedErr != "" {
				require.Contains(t, actualResponse["error"].(string), tc.expectedErr)
			} else {
				if diff := cmp.Diff(tc.expectedResponse, actualResponse); diff != "" {
					t.Errorf("response mismatch (-want +got)\n%s", diff)
				}
			}
			if diff := cmp.Diff(tc.expectedStatusCode, w.Code); diff != "" {
				t.Errorf("status code mismatch (-want +got)\n%s", diff)
			}
			mockTxQ.AssertExpectations(t)
		})
	}
}
