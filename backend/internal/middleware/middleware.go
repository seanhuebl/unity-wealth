package middleware

import (
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
)

type Middleware struct {
	tokenGen       auth.TokenGenerator
	tokenExtractor auth.TokenExtractor
}

func NewMiddleware(tokenGen auth.TokenGenerator, tokenExtractor auth.TokenExtractor) *Middleware {
	return &Middleware{tokenGen: tokenGen, tokenExtractor: tokenExtractor}
}
