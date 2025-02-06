// cache_test.go
package cache

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite driver

	"github.com/go-redis/redismock/v8"
	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/unity-wealth/cache"
	"github.com/seanhuebl/unity-wealth/handlers"
	"github.com/seanhuebl/unity-wealth/internal/database" // package that provides New(db) returning the querier
)

// setupSQLiteDB creates an in-memory SQLite DB and executes the provided setup function.
func setupSQLiteDB(setupFunc func(db *sql.DB) error) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}
	if err := setupFunc(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func TestWarmCategoriesCache_TableDriven(t *testing.T) {
	// Save the original redisClient and restore it after tests.
	origRedisClient := cache.RedisClient
	defer func() { cache.RedisClient = origRedisClient }()

	tests := []struct {
		name                     string
		setupSQLite              func(db *sql.DB) error
		simulateRedisPrimaryErr  error
		simulateRedisDetailedErr error
		wantErr                  bool
	}{
		{
			name: "success",
			setupSQLite: func(db *sql.DB) error {
				// Create primary_categories table using your schema.
				primaryStmt := `
					CREATE TABLE IF NOT EXISTS primary_categories (
						id INTEGER PRIMARY KEY,
						name TEXT NOT NULL
					);`
				if _, err := db.Exec(primaryStmt); err != nil {
					return err
				}
				// Insert a row into primary_categories.
				if _, err := db.Exec(`INSERT INTO primary_categories (id, name) VALUES (1, 'Primary Cat 1');`); err != nil {
					return err
				}

				// Create detailed_categories table using your schema.
				detailedStmt := `
					CREATE TABLE IF NOT EXISTS detailed_categories (
						id INTEGER PRIMARY KEY,
						name TEXT NOT NULL,
						description TEXT NOT NULL,
						primary_category_id INTEGER NOT NULL,
						FOREIGN KEY (primary_category_id) REFERENCES primary_categories (id) ON DELETE CASCADE
					);`
				if _, err := db.Exec(detailedStmt); err != nil {
					return err
				}
				// Insert a row into detailed_categories.
				_, err := db.Exec(`INSERT INTO detailed_categories (id, name, description, primary_category_id) VALUES (1, 'Detailed Cat 1', 'Some description', 1);`)
				return err
			},
			wantErr: false,
		},
		{
			name: "primary query error - missing primary_categories table",
			setupSQLite: func(db *sql.DB) error {
				// Do not create primary_categories table.
				// Create only the detailed_categories table.
				detailedStmt := `
					CREATE TABLE IF NOT EXISTS detailed_categories (
						id INTEGER PRIMARY KEY,
						name TEXT NOT NULL,
						description TEXT NOT NULL,
						primary_category_id INTEGER NOT NULL,
						FOREIGN KEY (primary_category_id) REFERENCES primary_categories (id) ON DELETE CASCADE
					);`
				if _, err := db.Exec(detailedStmt); err != nil {
					return err
				}
				// Insert a row into detailed_categories.
				_, err := db.Exec(`INSERT INTO detailed_categories (id, name, description, primary_category_id) VALUES (1, 'Detailed Cat 1', 'Some description', 1);`)
				return err
			},
			wantErr: true,
		},
		{
			name: "redis set primary error",
			setupSQLite: func(db *sql.DB) error {
				// Create both tables and insert rows as in the success case.
				primaryStmt := `
					CREATE TABLE IF NOT EXISTS primary_categories (
						id INTEGER PRIMARY KEY,
						name TEXT NOT NULL
					);`
				if _, err := db.Exec(primaryStmt); err != nil {
					return err
				}
				if _, err := db.Exec(`INSERT INTO primary_categories (id, name) VALUES (1, 'Primary Cat 1');`); err != nil {
					return err
				}
				detailedStmt := `
					CREATE TABLE IF NOT EXISTS detailed_categories (
						id INTEGER PRIMARY KEY,
						name TEXT NOT NULL,
						description TEXT NOT NULL,
						primary_category_id INTEGER NOT NULL,
						FOREIGN KEY (primary_category_id) REFERENCES primary_categories (id) ON DELETE CASCADE
					);`
				if _, err := db.Exec(detailedStmt); err != nil {
					return err
				}
				_, err := db.Exec(`INSERT INTO detailed_categories (id, name, description, primary_category_id) VALUES (1, 'Detailed Cat 1', 'Some description', 1);`)
				return err
			},
			simulateRedisPrimaryErr: errors.New("redis set primary error"),
			wantErr:                 true,
		},
		{
			name: "redis set detailed error",
			setupSQLite: func(db *sql.DB) error {
				// Create both tables and insert rows as in the success case.
				primaryStmt := `
					CREATE TABLE IF NOT EXISTS primary_categories (
						id INTEGER PRIMARY KEY,
						name TEXT NOT NULL
					);`
				if _, err := db.Exec(primaryStmt); err != nil {
					return err
				}
				if _, err := db.Exec(`INSERT INTO primary_categories (id, name) VALUES (1, 'Primary Cat 1');`); err != nil {
					return err
				}
				detailedStmt := `
					CREATE TABLE IF NOT EXISTS detailed_categories (
						id INTEGER PRIMARY KEY,
						name TEXT NOT NULL,
						description TEXT NOT NULL,
						primary_category_id INTEGER NOT NULL,
						FOREIGN KEY (primary_category_id) REFERENCES primary_categories (id) ON DELETE CASCADE
					);`
				if _, err := db.Exec(detailedStmt); err != nil {
					return err
				}
				_, err := db.Exec(`INSERT INTO detailed_categories (id, name, description, primary_category_id) VALUES (1, 'Detailed Cat 1', 'Some description', 1);`)
				return err
			},
			simulateRedisDetailedErr: errors.New("redis set detailed error"),
			wantErr:                  true,
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Set up an in-memory SQLite DB using the test-specific schema and data.
			db, err := setupSQLiteDB(tc.setupSQLite)
			if err != nil {
				t.Fatalf("failed to set up SQLite DB: %v", err)
			}
			defer db.Close()

			// Build the configuration by explicitly setting both Database and Querier.
			cfg := &handlers.ApiConfig{
				Database: db,
				Queries:  database.New(db),
			}

			// Override the global cache.RedisClient with a redismock client.
			mockRedis, mock := redismock.NewClientMock()
			cache.RedisClient = mockRedis

			ctx := context.Background()

			primaryData, primaryErr := cfg.Queries.GetPrimaryCategories(ctx)
			if primaryErr == nil {
				exp := mock.ExpectSet("primary_categories", primaryData, 0)
				if tc.simulateRedisPrimaryErr != nil {
					exp.SetErr(tc.simulateRedisPrimaryErr)
				} else {
					exp.SetVal("OK")
				}
			} else {
				// If primary query fails, skip setting any expectation for detailed categories.
				t.Log("primary query failed, skipping detailed redis expectation")
			}

			// Only set detailed expectation if primary query succeeded
			// AND no error is being simulated for the primary Redis call.
			detailedData, detailedErr := cfg.Queries.GetDetailedCategories(ctx)
			if primaryErr == nil && tc.simulateRedisPrimaryErr == nil {
				if detailedErr == nil {
					exp := mock.ExpectSet("detailed_categories", detailedData, 0)
					if tc.simulateRedisDetailedErr != nil {
						exp.SetErr(tc.simulateRedisDetailedErr)
					} else {
						exp.SetVal("OK")
					}
				}
			}

			// Call the function under test.
			err = cache.WarmCategoriesCache(cfg)
			if (err != nil) != tc.wantErr {
				t.Fatalf("expected error=%v, got error=%v", tc.wantErr, err)
			}

			// Verify that all Redis expectations were met.
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled Redis expectations: %v", err)
			}

			// Optionally, verify that the stored values match expectations using google cmp.
			if primaryErr == nil {
				expectedPrimary, _ := cfg.Queries.GetPrimaryCategories(ctx)
				if diff := cmp.Diff(expectedPrimary, primaryData); diff != "" {
					t.Errorf("mismatch in primary data (-want +got):\n%s", diff)
				}
			}
			if detailedErr == nil {
				expectedDetailed, _ := cfg.Queries.GetDetailedCategories(ctx)
				if diff := cmp.Diff(expectedDetailed, detailedData); diff != "" {
					t.Errorf("mismatch in detailed data (-want +got):\n%s", diff)
				}
			}
		})
	}
}
