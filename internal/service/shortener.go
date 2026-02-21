package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
	"strings"
	"time"
)

type ShortenerService struct {
	repo       repository.Repository
	deleteChan chan deleteTask
}

type deleteTask struct {
	userID string
	ids    []string
}

func NewShortenerService(repo repository.Repository) *ShortenerService {
	s := &ShortenerService{
		repo:       repo,
		deleteChan: make(chan deleteTask, 100),
	}
	go s.deleteWorker()

	return s
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

func (s *ShortenerService) DeleteUserURLs(
	userID string,
	shortIDs []string,
) {
	if len(shortIDs) == 0 {
		return
	}

	s.deleteChan <- deleteTask{
		userID: userID,
		ids:    shortIDs,
	}
}

func (s *ShortenerService) deleteWorker() {
	const batchSize = 100

	for {
		task, ok := <-s.deleteChan
		if !ok {
			return
		}
		batch := []deleteTask{task}

	Loop:
		for len(batch) < batchSize {
			select {
			case t, ok := <-s.deleteChan:
				if !ok {
					break Loop
				}
				batch = append(batch, t)
			default:
				break Loop
			}
		}

		grouped := make(map[string][]string)
		for _, t := range batch {
			grouped[t.userID] = append(grouped[t.userID], t.ids...)
		}

		for userID, ids := range grouped {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			_ = s.repo.MarkDeleted(ctx, userID, ids)
			cancel()
		}
	}
}
