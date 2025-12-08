package repository

import "context"

type Repository interface {
	Save(ctx context.Context, id, original string) error
	Get(ctx context.Context, id string) (string, error)
	Ping(ctx context.Context) error
	SaveBatch(ctx context.Context, data map[string]string) error
}
