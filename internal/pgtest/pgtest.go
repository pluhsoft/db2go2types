// Package pgtest gives tests a fresh PostgreSQL database.
package pgtest

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EnvDSN names the variable with the connection string of the test server.
// The user must be allowed to create databases.
const EnvDSN = "DB2GO_TEST_DATABASE_URL"

// New creates an empty database, runs the SQL file in it and returns a pool
// connected to it. The database is dropped when the test ends. The test is
// skipped when DB2GO_TEST_DATABASE_URL is not set.
func New(t testing.TB, sqlFile string) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv(EnvDSN)
	if dsn == "" {
		t.Skipf("%s is not set; see CONTRIBUTING.md", EnvDSN)
	}
	ctx := context.Background()

	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect to %s: %v", EnvDSN, err)
	}
	defer admin.Close(ctx)
	name := fmt.Sprintf("db2go_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		admin, err := pgx.Connect(ctx, dsn)
		if err != nil {
			t.Error(err)
			return
		}
		defer admin.Close(ctx)
		if _, err := admin.Exec(ctx, "DROP DATABASE "+name+" WITH (FORCE)"); err != nil {
			t.Error(err)
		}
	})

	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + name
	pool, err := pgxpool.New(ctx, u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	sql, err := os.ReadFile(sqlFile)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, string(sql)); err != nil {
		t.Fatalf("%s: %v", sqlFile, err)
	}
	return pool
}
