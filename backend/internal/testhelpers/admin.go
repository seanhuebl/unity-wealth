package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	migrations "github.com/seanhuebl/unity-wealth/sql"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	once        sync.Once
	pgContainer *postgres.PostgresContainer
	adminDSN    string
	templateDSN string
)

func Main(m *testing.M) int {
	once.Do(startContainerAndTemplate)
	code := m.Run()
	terminateContainer()
	return code
}

func NewDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()

	admin, _ := sql.Open("pgx", adminDSN)

	dbName := "t_" + uuid.NewString()
	_, err := admin.ExecContext(ctx, `CREATE DATABASE`+pgQuoteIdent(dbName)+`TEMPLATE test_template`)
	if err != nil {
		t.Fatalf("clone db error: %v", err)
	}

	testDSN := swapDBName(adminDSN, dbName)
	db, _ := sql.Open("pgx", testDSN)

	t.Cleanup(func() {
		db.Close()
		admin.ExecContext(ctx, `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = `+pgQuoteLiteral(dbName))
		admin.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pgQuoteIdent(dbName))
	})
	return db
}

func startContainerAndTemplate() {
	ctx := context.Background()

	var err error
	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("secret"),
	)
	if err != nil {
		log.Fatalf("postgres run: %v", err)
	}
	adminDSN, _ = container.ConnectionString(ctx, "sslmode=disable")

	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		log.Fatalf("open admin: %v", err)
	}
	if err = waitForPing(ctx, admin, 20, 300*time.Millisecond); err != nil {
		log.Fatalf("postgres never became ready")
	}
	if _, err := admin.ExecContext(ctx, `DROP DATABASE IF EXISTS test_template`); err != nil {
		log.Fatalf("drop template: %v", err)
	}
	if _, err := admin.ExecContext(ctx, `CREATE DATABASE test_template`); err != nil {
		log.Fatalf("create template: %v", err)
	}
	admin.Close()

	templateDSN = swapDBName(adminDSN, "test_template")
	templateDB, err := sql.Open("pgx", templateDSN)
	if err != nil {
		log.Fatalf("open template: %v", err)
	}
	if err = waitForPing(ctx, templateDB, 10, 300*time.Millisecond); err != nil {
		log.Fatalf("template never became ready")
	}
	migrationsDir := "schema"
	goose.SetBaseFS(migrations.FS)
	if err = goose.UpContext(ctx, templateDB, migrationsDir); err != nil {
		log.Fatalf("migrate template: %v", err)
	}
	templateDB.Close()
}

func terminateContainer() {
	if pgContainer != nil {
		_ = pgContainer.Terminate(context.Background())
	}
}

func waitForPing(ctx context.Context, db *sql.DB, retries int, delay time.Duration) error {
	for range retries {
		if err := db.PingContext(ctx); err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("gave up after %v retries", retries)
}
func swapDBName(dsn, dbName string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		log.Fatalf("parse dsn: %v", err)
	}
	u.Path = "/" + dbName
	return u.String()
}
func pgQuoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func pgQuoteLiteral(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `''`) + `'`
}
