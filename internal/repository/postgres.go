package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type postgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(dsn string) (Repository, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	repo := &postgresRepository{db: db}
	return repo, nil
}

func runMigrations(dsn string) error {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

func (p *postgresRepository) Save(ctx context.Context, id, original string) error {
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO urls (short_url, original_url) VALUES ($1, $2)
    ON CONFLICT (short_url) DO NOTHING`,
		id, original)
	if err != nil {
		return fmt.Errorf("failed to save url %s: %w", id, err)
	}
	return nil
}

func (p *postgresRepository) Get(ctx context.Context, id string) (string, error) {
	var original string
	err := p.db.GetContext(ctx, &original, `SELECT original_url FROM urls WHERE short_url=$1`, id)
	if err != nil {
		return "", ErrNotFound
	}
	return original, nil
}

func (p *postgresRepository) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}
