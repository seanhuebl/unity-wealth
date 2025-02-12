package config

import (
	"database/sql"

	"github.com/seanhuebl/unity-wealth/internal/interfaces"
)

type ApiConfig struct {
	Port        string
	Queries     interfaces.Querier
	Database    *sql.DB
	TokenSecret string
}
