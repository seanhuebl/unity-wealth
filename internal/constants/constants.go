package constants

type contextKey string

const (
	ClaimsKey  = contextKey("claims")
	UserIDKey  = contextKey("userID")
	RequestKey = contextKey("httpRequest")
)
