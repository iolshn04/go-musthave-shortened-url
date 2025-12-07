package repository

import (
	"fmt"
	"go.uber.org/zap"
)

func NewRepositoryFromConfig(dsn, filePath string, log *zap.Logger) (Repository, error) {
	if dsn != "" {
		db, err := NewPostgresRepository(dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to init database %s: %w", filePath, err)
		}
		log.Info("using database storage", zap.String("dsn", dsn))
		return db, nil
	}

	if filePath != "" {
		fs, err := NewFileStorage(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to init file storage %s: %w", filePath, err)
		}
		log.Info("using file storage", zap.String("path", filePath))
		return fs, nil
	}

	log.Info("using memory storage")
	return NewMemoryStorage(), nil
}
