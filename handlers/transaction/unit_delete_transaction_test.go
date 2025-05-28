package transaction_test

import (
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

func TestDeleteTransaction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []testmodels.DeleteTxTestCase{
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
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testfixtures.EmptyTxID,
				TxID:             "",
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			w := httptest.NewRecorder()
			mockSvc := handlermocks.NewTransactionService(t)
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/transactions/%v", tc.TxID), nil)
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

			router.DELETE("/transactions/:id", h.DeleteTransaction)

			router.ServeHTTP(w, req)
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)

		})
	}
	txErrTests := []testmodels.DeleteTxTestCase{
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{

					Name:               "error deleting tx",
					UserID:             uuid.New(),
					ExpectedError:      "error deleting transaction",
					ExpectedStatusCode: http.StatusInternalServerError,
					ExpectedResponse: map[string]interface{}{
						"data": map[string]interface{}{
							"error": "error deleting transaction",
						},
					},
				},
				TxID:  uuid.NewString(),
				TxErr: errors.New("error deleting transaction"),
			},
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{

					Name:               "not found",
					UserID:             uuid.New(),
					ExpectedError:      "not found",
					ExpectedStatusCode: http.StatusNotFound,
					ExpectedResponse: map[string]interface{}{
						"data": map[string]interface{}{
							"error": "not found",
						},
					},
				},
				TxID:  uuid.NewString(),
				TxErr: errors.New("transaction not found"),
			},
		},
	}

	for _, tc := range txErrTests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			w := httptest.NewRecorder()
			mockSvc := handlermocks.NewTransactionService(t)
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/transactions/%v", tc.TxID), nil)

			mockSvc.On("DeleteTransaction", mock.Anything, tc.TxID, tc.UserID.String()).Return(tc.TxErr)

			h := htx.NewHandler(mockSvc)

			router := gin.New()

			router.Use(func(c *gin.Context) {
				c.Params = gin.Params{{Key: "id", Value: tc.TxID}}
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				c.Next()
			})
			router.DELETE("/transactions/:id", h.DeleteTransaction)
			router.ServeHTTP(w, req)

			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)

			mockSvc.AssertExpectations(t)
		})
	}
}
