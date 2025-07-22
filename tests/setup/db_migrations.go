package setup

import (
	"database/sql"
	"fmt"

	"echo-app/migrations"

	"github.com/pressly/goose/v3"
)

func MigrateDB(db *sql.DB) error {
	goose.SetBaseFS(migrations.EmbedMigrations)

	if err := goose.SetDialect(string(goose.DialectPostgres)); err != nil {
		return fmt.Errorf("set migrations dialect as postgres: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("up migrations: %w", err)
	}

	return nil
}
