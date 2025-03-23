package transaction

import "github.com/google/uuid"

type BaseHTTPTestCase struct {
	name               string
	userID             uuid.UUID
	userIDErr          error
	expectedError      string
	expectedStatusCode int
	expectedResponse   map[string]interface{}
}

type CreateTxTestCase struct {
	BaseHTTPTestCase
	reqBody     string
	createTxErr error
}

func (c CreateTxTestCase) BaseAccess() BaseHTTPTestCase {
	return c.BaseHTTPTestCase
}

type GetTxTestCase struct {
	BaseHTTPTestCase
	txID  string
	txErr error
}

func (g GetTxTestCase) BaseAccess() BaseHTTPTestCase {
	return g.BaseHTTPTestCase
}

type UpdateTxTestCase struct {
	GetTxTestCase
	reqBody     string
	updateTxErr error
}

func (u UpdateTxTestCase) BaseAccess() BaseHTTPTestCase {
	return u.BaseHTTPTestCase
}

type DeleteTxTestCase struct {
	GetTxTestCase
}

func (d DeleteTxTestCase) BaseAccess() BaseHTTPTestCase {
	return d.BaseHTTPTestCase
}

type GetAllTxByUserIDTestCase struct {
	BaseHTTPTestCase
	cursorDate        string
	cursorID          string
	pageSize          int
	getFirstPageErr   error
	getTxPaginatedErr error
	firstPageTest     bool
	moreData          bool
}

func (allTx GetAllTxByUserIDTestCase) BaseAccess() BaseHTTPTestCase {
	return allTx.BaseHTTPTestCase
}
