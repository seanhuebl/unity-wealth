package server

import (
	"github.com/seanhuebl/unity-wealth/handlers/auth"
	"github.com/seanhuebl/unity-wealth/handlers/category"
	"github.com/seanhuebl/unity-wealth/handlers/common"
	"github.com/seanhuebl/unity-wealth/handlers/transaction"
	"github.com/seanhuebl/unity-wealth/handlers/user"
)

type HandlersGroup struct {
	Auth *auth.Handler
	Cat  *category.Handler
	Cmn  *common.Handler
	Tx   *transaction.Handler
	User *user.Handler
}

func NewHandlers(authHandler *auth.Handler, catHandler *category.Handler, commonHandler *common.Handler, txHandler *transaction.Handler, userHandler *user.Handler) *HandlersGroup {
	return &HandlersGroup{
		Auth: authHandler,
		Cat:  catHandler,
		Cmn:  commonHandler,
		Tx:   txHandler,
		User: userHandler,
	}
}
