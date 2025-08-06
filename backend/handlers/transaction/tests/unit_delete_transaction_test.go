package transaction_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	htx "github.com/seanhuebl/unity-wealth/handlers/transaction"
	handlermocks "github.com/seanhuebl/unity-wealth/internal/mocks/handlers"
	"github.com/seanhuebl/unity-wealth/internal/sentinels"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/seanhuebl/unity-wealth/internal/testfixtures"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
	"github.com/stretchr/testify/mock"
)

func TestDeleteTransaction(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	tests := []testmodels.DeleteTxTestCase{
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
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			mockSvc := handlermocks.NewTransactionService(t)
			t.Cleanup(func() { mockSvc.AssertExpectations(t) })
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/transactions/%v", tc.TxID), nil)
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
					ExpectedError:      "something went wrong",
					ExpectedStatusCode: http.StatusInternalServerError,
					ExpectedResponse: map[string]interface{}{
						"data": map[string]interface{}{
							"error": "something went wrong",
						},
					},
				},
				TxID:  uuid.New(),
				TxErr: sentinels.ErrDBExecFailed,
			},
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{

					Name:               "not found", // non empty tx ID db search returns no tx
					UserID:             uuid.New(),
					ExpectedError:      "not found",
					ExpectedStatusCode: http.StatusNotFound,
					ExpectedResponse: map[string]interface{}{
						"data": map[string]interface{}{
							"error": "not found",
						},
					},
				},
				TxID:  uuid.New(),
				TxErr: transaction.ErrTxNotFound,
			},
		},
	}

	for _, tc := range txErrTests {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			mockSvc := handlermocks.NewTransactionService(t)
			t.Cleanup(func() { mockSvc.AssertExpectations(t) })
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/transactions/%v", tc.TxID), nil)

			mockSvc.On("DeleteTransaction", mock.Anything, tc.TxID, tc.UserID).Return(tc.TxErr)

			h := htx.NewHandler(mockSvc)

			router := gin.New()

			router.Use(func(c *gin.Context) {
				c.Params = gin.Params{{Key: "id", Value: tc.TxID.String()}}
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
