package testmodels

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/handlers/transaction"
	"github.com/seanhuebl/unity-wealth/internal/database"
)

type TestEnv struct {
	Db      *sql.DB
	Router  *gin.Engine
	UserQ   database.UserQuerier
	TxQ     database.TransactionQuerier
	Handler *transaction.Handler
}
