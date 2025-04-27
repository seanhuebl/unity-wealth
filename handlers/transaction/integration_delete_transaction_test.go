package transaction_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/models"
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
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
					Name:               "invalid txID in req",
					UserID:             uuid.New(),
					ExpectedError:      "invalid id",
					ExpectedStatusCode: http.StatusBadRequest,
					ExpectedResponse: map[string]interface{}{
						"error": "invalid id",
					},
				},
				TxID: "",
			},
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{
					Name:               "unauthorized: user ID is uuid.NIL",
					UserID:             uuid.Nil,
					UserIDErr:          errors.New("user ID not found in context"),
					ExpectedError:      "unauthorized",
					ExpectedStatusCode: http.StatusUnauthorized,
					ExpectedResponse: map[string]interface{}{
						"error": "unauthorized",
					},
				},
				TxID: uuid.NewString(),
			},
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{

					Name:               "unauthorized: user ID not UUID",
					UserID:             uuid.Nil,
					UserIDErr:          errors.New("user ID is not UUID"),
					ExpectedStatusCode: http.StatusUnauthorized,
					ExpectedResponse: map[string]interface{}{
						"error": "unauthorized",
					},
				},
				TxID: uuid.NewString(),
			},
		},
		{
			GetTxTestCase: testmodels.GetTxTestCase{
				BaseHTTPTestCase: testmodels.BaseHTTPTestCase{

					Name:               "error deleting tx",
					UserID:             uuid.New(),
					ExpectedStatusCode: http.StatusInternalServerError,
					ExpectedResponse: map[string]interface{}{
						"error": "error deleting transaction",
					},
				},
				TxID:  uuid.NewString(),
				TxErr: errors.New("error deleting transaction"),
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
				testhelpers.SeedTestTransaction(t, env.TxQ, tc.UserID, uuid.MustParse(tc.TxID), &models.NewTransactionRequest{
					Date:             "2025-03-05",
					Merchant:         "costco",
					Amount:           125.98,
					DetailedCategory: 40,
				})
			}
			w := httptest.NewRecorder()
			var req *http.Request
			if strings.Contains(tc.Name, "error deleting tx") {
				req = httptest.NewRequest("DELETE", fmt.Sprintf("/transactions/%v", ""), nil)
			} else {
				req = httptest.NewRequest("DELETE", fmt.Sprintf("/transactions/%v", tc.TxID), nil)
			}

			if tc.TxID == "" {
				c, _ := gin.CreateTestContext(w)
				c.Request = req
				c.Params = gin.Params{{Key: "id", Value: ""}}
				if tc.Name == "unauthorized: user ID not UUID" {
					c.Set(string(constants.UserIDKey), "userID")
				} else {
					c.Set(string(constants.UserIDKey), tc.UserID)
				}
				env.Handler.DeleteTransaction(c)
			} else {
				env.Router.DELETE("/transactions/:id", func(c *gin.Context) {
					if tc.Name == "unauthorized: user ID not UUID" {
						c.Set(string(constants.UserIDKey), "userID")
					} else {
						c.Set(string(constants.UserIDKey), tc.UserID)
					}
					env.Handler.DeleteTransaction(c)
				})
				env.Router.ServeHTTP(w, req)
			}

		})
	}
}
