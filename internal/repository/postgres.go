package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/iolshn04/go-musthave-shortened-url/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type postgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository создаёт репозиторий,
// работающий с PostgreSQL, и выполняет миграции базы данных.
func NewPostgresRepository(dsn string) (Repository, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
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

func (p *postgresRepository) Save(ctx context.Context, userID, shortID, original string) error {
	_, err := p.db.ExecContext(ctx,
		`INSERT INTO urls (user_id, short_url, original_url)
         VALUES ($1, $2, $3)`,
		userID, shortID, original)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == pgerrcode.UniqueViolation {
			var existingID string
			getErr := p.db.GetContext(ctx, &existingID,
				`SELECT short_url FROM urls WHERE original_url=$1 AND user_id=$2`, original, userID)
			if getErr != nil {
				return fmt.Errorf("failed to get existing url after unique violation: %w", getErr)
			}
			return ErrAlreadyExistsWithID{ExistingID: existingID}
		}
		return fmt.Errorf("failed to save url: %w", err)
	}
	return nil
}

func (p *postgresRepository) SaveBatch(ctx context.Context, userID string, data map[string]string) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO urls (user_id, short_url, original_url)
        VALUES ($1, $2, $3)
        ON CONFLICT (original_url, user_id) DO NOTHING
    `)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()

	for short, original := range data {
		if _, err := stmt.ExecContext(ctx, userID, short, original); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (p *postgresRepository) Get(ctx context.Context, shortID string) (string, error) {
	var original string
	var deleted bool

	err := p.db.QueryRowContext(
		ctx,
		`SELECT original_url, is_deleted FROM urls WHERE short_url=$1`,
		shortID,
	).Scan(&original, &deleted)

	if err == sql.ErrNoRows {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if deleted {
		return "", ErrDeleted
	}

	return original, nil
}

func (p *postgresRepository) GetByUser(ctx context.Context, userID string) ([]model.UserURL, error) {
	rows := []model.UserURL{}
	err := p.db.SelectContext(ctx, &rows, `SELECT short_url, original_url FROM urls WHERE user_id=$1`, userID)
	log.Printf("GetByUser: querying for userID=%s", userID)
	log.Printf("rows returned: %+v", rows)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrNotFound
	}
	return rows, nil
}

func (p *postgresRepository) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *postgresRepository) MarkDeleted(
	ctx context.Context,
	userID string,
	shortIDs []string,
) error {
	if len(shortIDs) == 0 {
		return nil
	}

	_, err := p.db.ExecContext(
		ctx,
		`UPDATE urls
         SET is_deleted = TRUE
         WHERE user_id = $1 AND short_url = ANY($2)`,
		userID,
		pq.Array(shortIDs),
	)
	return err
}
