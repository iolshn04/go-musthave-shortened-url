package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
)

type ShortenerService struct {
	repo repository.Repository
}

func NewShortenerService(repo repository.Repository) *ShortenerService {
	return &ShortenerService{repo: repo}
}

func (s *ShortenerService) Shorten(ctx context.Context, userID, original string) (string, error) {
	shortID := generateID()
	err := s.repo.Save(ctx, userID, shortID, original)
	if err != nil {
		var existErr repository.ErrAlreadyExistsWithID
		if errors.As(err, &existErr) {
			return "", existErr
		}
		return "", err
	}
	return shortID, nil
}

func (s *ShortenerService) GetOriginal(ctx context.Context, shortID string) (string, error) {
	return s.repo.Get(ctx, shortID)
}

func generateID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	id := base64.URLEncoding.EncodeToString(b)
	return strings.TrimRight(id, "=")
}

func (s *ShortenerService) ShortenBatch(
	ctx context.Context,
	userID string,
	input map[string]string,
) (map[string]string, error) {

	result := make(map[string]string, len(input))
	toSave := make(map[string]string, len(input))

	for cid, original := range input {
		id := generateID()
		result[cid] = id
		toSave[id] = original
	}

	if err := s.repo.SaveBatch(ctx, userID, toSave); err != nil {
		return nil, err
	}

	return result, nil
}
