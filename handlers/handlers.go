package handlers

import (
	"github.com/seanhuebl/unity-wealth/handlers/auth"
	"github.com/seanhuebl/unity-wealth/handlers/category"
	"github.com/seanhuebl/unity-wealth/handlers/common"
	"github.com/seanhuebl/unity-wealth/handlers/transaction"
	"github.com/seanhuebl/unity-wealth/handlers/user"
)

type Handlers struct {
	authHandler   *auth.Handler
	catHandler    *category.Handler
	commonHandler *common.Handler
	txHandler     *transaction.Handler
	userHandler   *user.Handler
}

func NewHandlers(authHandler *auth.Handler, catHandler *category.Handler, commonHandler *common.Handler, txHandler *transaction.Handler, userHandler *user.Handler) *Handlers {
	return &Handlers{
		authHandler:   authHandler,
		catHandler:    catHandler,
		commonHandler: commonHandler,
		txHandler:     txHandler,
		userHandler:   userHandler,
	}
}
