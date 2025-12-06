package repository

import (
	"go.uber.org/zap"
)

func NewRepositoryFromConfig(filePath string, log *zap.Logger) Repository {
	if filePath != "" {
		fs, err := NewFileStorage(filePath)
		if err != nil {
			log.Warn("failed to init file storage, fallback to memory",
				zap.String("path", filePath),
				zap.Error(err),
			)
			return NewMemoryStorage()
		}
		log.Info("using file storage", zap.String("path", filePath))
		return fs
	}
	log.Info("using memory storage")
	return NewMemoryStorage()
}
