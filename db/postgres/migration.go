package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func Up(ctx context.Context, url string) (err error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return fmt.Errorf("failed to create migration DB: %w", err)
	}

	defer func() { err = errors.Join(err, db.Close()) }()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		if !errors.Is(err, goose.ErrAlreadyApplied) {
			return fmt.Errorf("failed to apply transaction: %w", err)
		}
	}

	return nil
}

func Down(ctx context.Context, url string) (err error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return fmt.Errorf("failed to create migration DB: %w", err)
	}

	defer func() { err = errors.Join(err, db.Close()) }()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	if err := goose.DownContext(ctx, db, "migrations"); err != nil {
		if !errors.Is(err, goose.ErrAlreadyApplied) {
			return fmt.Errorf("failed to apply transaction: %w", err)
		}
	}

	return nil
}
