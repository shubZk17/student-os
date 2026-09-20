// Package migrations applies the embedded *.up.sql files in version order.
// Applied versions are recorded in schema_migrations; *.down.sql files are for manual rollback.
package migrations

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed *.up.sql
var files embed.FS

// Arbitrary constant; serializes migrations when several instances start at once.
const lockID = 727401

// Up applies all pending migrations, each in its own transaction.
func Up(ctx context.Context, pool *pgxpool.Pool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, lockID); err != nil {
		return err
	}
	defer conn.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, lockID)

	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version BIGINT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`); err != nil {
		return err
	}

	names, err := fs.Glob(files, "*.up.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)

	for _, name := range names {
		version, err := Version(name)
		if err != nil {
			return err
		}
		var applied bool
		if err := conn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&applied); err != nil {
			return err
		}
		if applied {
			continue
		}

		sql, err := files.ReadFile(name)
		if err != nil {
			return err
		}
		tx, err := conn.Begin(ctx)
		if err != nil {
			return err
		}
		// No arguments => pgx uses the simple protocol, which allows multi-statement files.
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		log.Printf("[INFO] Applied migration %s", name)
	}
	return nil
}

// Version parses the numeric prefix of a migration file name, e.g. "000002_x.up.sql" -> 2.
func Version(name string) (int64, error) {
	prefix, _, ok := strings.Cut(name, "_")
	if !ok {
		return 0, fmt.Errorf("migration %q: missing version prefix", name)
	}
	return strconv.ParseInt(prefix, 10, 64)
}
