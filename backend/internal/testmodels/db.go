package testmodels

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/handlers/auth"
	"github.com/seanhuebl/unity-wealth/handlers/transaction"
	"github.com/seanhuebl/unity-wealth/handlers/user"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/middleware"
	authSvc "github.com/seanhuebl/unity-wealth/internal/services/auth"
	txSvc "github.com/seanhuebl/unity-wealth/internal/services/transaction"
	userSvc "github.com/seanhuebl/unity-wealth/internal/services/user"
	"go.uber.org/zap"
)

type TestEnv struct {
	Db             *sql.DB
	Router         *gin.Engine
	UserQ          database.UserQuerier
	TxQ            database.TransactionQuerier
	TokenQ         database.TokenQuerier
	DeviceQ        database.DeviceQuerier
	SqlTxQ         database.SqlTxQuerier
	TransactionalQ database.SqlTransactionalQuerier
	Logger         *zap.Logger
	Services       *Services
	Middleware     *middleware.Middleware
	Handlers       *Handlers
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
