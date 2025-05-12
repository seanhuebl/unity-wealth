package transaction_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/testfixtures"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
)

func TestIntegrationDeleteTransaction(t *testing.T) {
	tests := []testmodels.DeleteTxTestCase{
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
					Name:               "success",
					UserID:             uuid.New(),
					ExpectedStatusCode: http.StatusOK,
					ExpectedResponse: map[string]interface{}{
						"data": map[string]interface{}{
							"transaction_deleted": "success",
						},
					},
				},
				TxID: uuid.NewString(),
			},
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testfixtures.InvalidTxID,
				TxID:             "",
			},
		},
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
				BaseHTTPTestCase: testfixtures.NotFound,
				TxID:             uuid.NewString(),
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {

			env := testhelpers.SetupTestEnv(t)
			defer env.Db.Close()

			if tc.TxID != "" {
				testhelpers.SeedTestUser(t, env.UserQ, tc.UserID)
				testhelpers.SeedTestCategories(t, env.Db)
				testhelpers.IsTxFound(t, tc.BaseHTTPTestCase, uuid.MustParse(tc.TxID), env)
			}
			w := httptest.NewRecorder()

			req := httptest.NewRequest("DELETE", fmt.Sprintf("/transactions/%v", tc.TxID), nil)

			if tc.TxID == "" {
				c, _ := gin.CreateTestContext(w)
				c.Request = req
				c.Params = gin.Params{{Key: "id", Value: ""}}
				testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
				env.Handlers.TxHandler.DeleteTransaction(c)
			} else {
				env.Router.DELETE("/transactions/:id", func(c *gin.Context) {
					testhelpers.CheckForUserIDIssues(tc.Name, tc.UserID, c)
					env.Handlers.TxHandler.DeleteTransaction(c)
				})
				env.Router.ServeHTTP(w, req)
			}
			actualResponse := testhelpers.ProcessResponse(w, t)
			testhelpers.CheckHTTPResponse(t, w, tc.ExpectedError, tc.ExpectedStatusCode, tc.ExpectedResponse, actualResponse)
		})
	}
}
