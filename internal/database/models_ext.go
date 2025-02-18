package database

type AuthQuerierDeps struct {
	UserQuerier   UserQuerier
	DeviceQuerier DeviceQuerier
	SqlTxQuerier  SqlTxQuerier
	TokenQuerier  TokenQuerier
}

func NewAuthQuerierDeps(userQuerier UserQuerier, deviceQuerier DeviceQuerier, sqlTxQuerier SqlTxQuerier, tokenQuerier TokenQuerier) *AuthQuerierDeps {
	return &AuthQuerierDeps{
		UserQuerier:   userQuerier,
		DeviceQuerier: deviceQuerier,
		SqlTxQuerier:  sqlTxQuerier,
		TokenQuerier:  tokenQuerier,
	}
}
