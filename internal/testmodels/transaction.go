package testmodels

import (
	"github.com/google/uuid"
)

type BaseHTTPTestCase struct {
	Name               string
	UserID             uuid.UUID
	UserIDErr          error
	ExpectedError      string
	ExpectedStatusCode int
	ExpectedResponse   map[string]interface{}
}

type CreateTxTestCase struct {
	BaseHTTPTestCase
	ReqBody     string
	CreateTxErr error
}

func (c CreateTxTestCase) BaseAccess() BaseHTTPTestCase {
	return c.BaseHTTPTestCase
}

type GetTxTestCase struct {
	BaseHTTPTestCase
	TxID    uuid.UUID
	TxErr   error
	TxIDRaw string
}

func (g GetTxTestCase) BaseAccess() BaseHTTPTestCase {
	return g.BaseHTTPTestCase
}

type UpdateTxTestCase struct {
	GetTxTestCase
	ReqBody     string
	UpdateTxErr error
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
	NextCursor        string
	PageSize          int
	GetFirstPageErr   error
	GetTxPaginatedErr error
	FirstPageTest     bool
	MoreData          bool
}

func (allTx GetAllTxByUserIDTestCase) BaseAccess() BaseHTTPTestCase {
	return allTx.BaseHTTPTestCase
}
