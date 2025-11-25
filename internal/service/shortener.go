package service

import (
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

func (s *ShortenerService) Shorten(original string) (string, error) {
	id := generateID()
	if err := s.repo.Save(id, original); err != nil {
		return "", err
	}
	return id, nil
}

func (s *ShortenerService) GetOriginal(id string) (string, error) {
	return s.repo.Get(id)
}

func generateID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	id := base64.URLEncoding.EncodeToString(b)
	return strings.TrimRight(id, "=")
}
