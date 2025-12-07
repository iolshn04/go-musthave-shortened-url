package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strings"

	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
)

type ShortenerService struct {
	repo repository.Repository
}

func NewShortenerService(repo repository.Repository) *ShortenerService {
	return &ShortenerService{repo: repo}
}

func (s *ShortenerService) Shorten(ctx context.Context, original string) (string, error) {
	id := generateID()
	if err := s.repo.Save(ctx, id, original); err != nil {
		return "", err
	}
	return id, nil
}

func (s *ShortenerService) GetOriginal(ctx context.Context, id string) (string, error) {
	return s.repo.Get(ctx, id)
}

func generateID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	id := base64.URLEncoding.EncodeToString(b)
	return strings.TrimRight(id, "=")
}
