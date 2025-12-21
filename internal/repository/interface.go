package repository

import (
	"context"
	"github.com/iolshn04/go-musthave-shortened-url/internal/model"
)

type Repository interface {
	Save(ctx context.Context, userID, shortID, original string) error
	SaveBatch(ctx context.Context, userID string, data map[string]string) error
	Get(ctx context.Context, shortID string) (string, error)
	GetByUser(ctx context.Context, userID string) ([]model.UserURL, error)
	Ping(ctx context.Context) error
}
