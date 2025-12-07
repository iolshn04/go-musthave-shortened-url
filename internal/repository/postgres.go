package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type postgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(dsn string) (Repository, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	repo := &postgresRepository{db: db}

	if err := repo.ensureTable(context.Background()); err != nil {
		return nil, err
	}

	return repo, nil
}

func (p *postgresRepository) ensureTable(ctx context.Context) error {
	schema := `
	  CREATE TABLE IF NOT EXISTS urls (
		short_url TEXT PRIMARY KEY,
		original_url TEXT NOT NULL
	  );`
	_, err := p.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create table urls: %w", err)
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
