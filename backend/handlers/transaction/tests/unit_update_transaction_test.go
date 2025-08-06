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
	t.Parallel()
	gin.SetMode(gin.TestMode)
	tests := []testmodels.UpdateTxTestCase{
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testfixtures.NilUserID(),
				TxID:             uuid.New(),
			},
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testfixtures.InvalidUserID(),
				TxID:             uuid.New(),
			},
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testfixtures.InvalidTxID(),
				TxIDRaw:          "INVALID",
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{

				BaseHTTPTestCase: testfixtures.InvalidReqBody(),
			},
			ReqBody: `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			mockSvc := handlermocks.NewTransactionService(t)
			t.Cleanup(func() { mockSvc.AssertExpectations(t) })
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", fmt.Sprintf("/transactions/%v", tc.TxID), bytes.NewBufferString(tc.ReqBody))
			req.Header.Set("Content-Type", "application/json")
			h := htx.NewHandler(mockSvc)

			idVal := testhelpers.PrepareTxID(tc.TxID, tc.TxIDRaw)

			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Params = gin.Params{{Key: "id", Value: idVal}}
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				c.Next()
			})
			router.POST("/transactions/:id", h.UpdateTransaction)
			router.ServeHTTP(w, req)
		})
	}

	txErrTest := testmodels.UpdateTxTestCase{
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
			TxID: uuid.New(),
		},
		ReqBody:     `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
		UpdateTxErr: errors.New("update err"),
	}

	t.Run(txErrTest.Name, func(t *testing.T) {
		t.Parallel()
		mockSvc := handlermocks.NewTransactionService(t)
		t.Cleanup(func() { mockSvc.AssertExpectations(t) })
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/transactions/%v", txErrTest.TxID), bytes.NewBufferString(txErrTest.ReqBody))
		req.Header.Set("Content-Type", "application/json")

		mockSvc.On("UpdateTransaction", mock.Anything, txErrTest.TxID, txErrTest.UserID, mock.AnythingOfType("models.NewTxRequest")).Return(nil, txErrTest.UpdateTxErr)
		h := htx.NewHandler(mockSvc)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Params = gin.Params{{Key: "id", Value: txErrTest.TxID.String()}}
			testhelpers.CheckForUserIDIssues(txErrTest.Name, txErrTest.UserID, c)
			c.Next()
		})

		router.POST("/transactions/:id", h.UpdateTransaction)
		router.ServeHTTP(w, req)

		actualResponse := testhelpers.ProcessResponse(w, t)
		testhelpers.CheckHTTPResponse(t, w, txErrTest.ExpectedError, txErrTest.ExpectedStatusCode, txErrTest.ExpectedResponse, actualResponse)

		mockSvc.AssertExpectations(t)
	})
	t.Run("tx not found", func(t *testing.T) {
		t.Parallel()
		tc := testmodels.UpdateTxTestCase{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testfixtures.NotFound(),
				TxID:             uuid.New(),
			},
			ReqBody:     `{"date": "2025-03-05", "merchant": "costco", "amount": 125.98, "detailed_category": 40}`,
			UpdateTxErr: errors.New("update err"),
		}
		mockSvc := handlermocks.NewTransactionService(t)
		t.Cleanup(func() { mockSvc.AssertExpectations(t) })
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/transactions/%v", ""), bytes.NewBufferString(tc.ReqBody))
		req.Header.Set("Content-Type", "application/json")
		h := htx.NewHandler(mockSvc)

		idVal := testhelpers.PrepareTxID(tc.TxID, tc.TxIDRaw)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Params = gin.Params{{Key: "id", Value: idVal}}
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
