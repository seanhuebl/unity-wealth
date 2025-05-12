package testmodels

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/handlers/auth"
	authSvc "github.com/seanhuebl/unity-wealth/internal/services/auth"
	"github.com/seanhuebl/unity-wealth/handlers/transaction"
	txSvc "github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/seanhuebl/unity-wealth/handlers/user"
	userSvc "github.com/seanhuebl/unity-wealth/internal/services/user"
	"github.com/seanhuebl/unity-wealth/internal/database"
)

type TestEnv struct {
	Db       *sql.DB
	Router   *gin.Engine
	UserQ    database.UserQuerier
	TxQ      database.TransactionQuerier
	TokenQ   database.TokenQuerier
	DeviceQ  database.DeviceQuerier
	Services *Services
	Handlers *Handlers
}

type Services struct {
	AuthService *authSvc.AuthService
	TxService   *txSvc.TransactionService
	UserService *userSvc.UserService
}

type Handlers struct {
	AuthHandler *auth.Handler
	TxHandler   *transaction.Handler
	UserHandler *user.Handler
}
