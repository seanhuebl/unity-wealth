package config

import (
	"database/sql"

	"github.com/seanhuebl/unity-wealth/internal/database"
)

type ApiConfig struct {
	Port     string
	Queries  *database.Queries
	Database *sql.DB
}
