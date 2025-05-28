package transaction_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/google/uuid"
	htx "github.com/seanhuebl/unity-wealth/handlers/transaction"
	handlermocks "github.com/seanhuebl/unity-wealth/internal/mocks/handlers"
	"github.com/seanhuebl/unity-wealth/internal/testfixtures"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
	"github.com/stretchr/testify/mock"
)

func TestUpdateTransaction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []testmodels.UpdateTxTestCase{
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testfixtures.NilUserID,
				TxID:             uuid.NewString(),
			},
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testfixtures.InvalidUserID,
				TxID:             uuid.NewString(),
			},
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testfixtures.InvalidTxID,
				TxID:             "INVALID",
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{

				BaseHTTPTestCase: testfixtures.InvalidReqBody,
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40`,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			mockSvc := handlermocks.NewTransactionService(t)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", fmt.Sprintf("/transactions/%v", tc.TxID), bytes.NewBufferString(tc.ReqBody))
			req.Header.Set("Content-Type", "application/json")
			h := htx.NewHandler(mockSvc)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Params = gin.Params{{Key: "id", Value: tc.TxID}}
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				c.Next()
			})
			router.NoRoute(func(c *gin.Context) {
				c.JSON(http.StatusNotFound, gin.H{
					"data": gin.H{"error": "not found"},
				})
			})
			router.POST("/transactions/:id", h.UpdateTransaction)
			router.ServeHTTP(w, req)
		})
	}
	txErrTests := []testmodels.UpdateTxTestCase{
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
					Name:               "failed to update tx",
					UserID:             uuid.New(),
					ExpectedError:      "failed to update transaction",
					ExpectedStatusCode: http.StatusInternalServerError,
					ExpectedResponse: map[string]interface{}{
						"data": map[string]interface{}{
							"error": "failed to update transaction",
						},
					},
				},
				TxID: uuid.NewString(),
			},
			ReqBody:     `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
			UpdateTxErr: errors.New("update err"),
		},
	}
	for _, tc := range txErrTests {
		t.Run(tc.Name, func(t *testing.T) {
			mockSvc := handlermocks.NewTransactionService(t)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", fmt.Sprintf("/transactions/%v", tc.TxID), bytes.NewBufferString(tc.ReqBody))
			req.Header.Set("Content-Type", "application/json")

			mockSvc.On("UpdateTransaction", mock.Anything, tc.TxID, tc.UserID.String(), mock.AnythingOfType("models.NewTxRequest")).Return(nil, tc.UpdateTxErr)
			h := htx.NewHandler(mockSvc)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Params = gin.Params{{Key: "id", Value: tc.TxID}}
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				c.Next()
			})

			router.POST("/transactions/:id", h.UpdateTransaction)
			router.ServeHTTP(w, req)

			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)

			mockSvc.AssertExpectations(t)
		})
	}
}
