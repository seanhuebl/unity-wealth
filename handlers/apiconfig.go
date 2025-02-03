package handlers

import (
	"database/sql"

	"github.com/seanhuebl/unity-wealth/internal/auth"
)

type ApiConfig struct {
	Port        string
	Queries     Quierier
	Database    *sql.DB
	TokenSecret string
	Auth        auth.AuthInterface
}
