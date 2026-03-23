package memory

import (
	"context"
	"github.com/Mavichy/urlShortener/internal/model"
	"github.com/Mavichy/urlShortener/internal/storage"
	"sync"
)

type Store struct {
	mu                 sync.RWMutex
	linksByCode        map[string]model.Link
	linksByOriginalURL map[string]model.Link
}

func New() *Store {
	return &Store{
		linksByCode:        make(map[string]model.Link),
		linksByOriginalURL: make(map[string]model.Link),
	}
}

func (s *Store) Save(_ context.Context, link model.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.linksByOriginalURL[link.OriginalURL]; exists {
		return storage.ErrOriginalExists
	}
	if _, exists := s.linksByCode[link.Code]; exists {
		return storage.ErrCodeExists
	}

	s.linksByCode[link.Code] = link
	s.linksByOriginalURL[link.OriginalURL] = link
	return nil
}

func (s *Store) GetByCode(_ context.Context, code string) (model.Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	link, ok := s.linksByCode[code]
	if !ok {
		return model.Link{}, storage.ErrNotFound
	}
	return link, nil
}

func (s *Store) GetByOriginalURL(_ context.Context, originalURL string) (model.Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	link, ok := s.linksByOriginalURL[originalURL]
	if !ok {
		return model.Link{}, storage.ErrNotFound
	}
	return link, nil
}

func (s *Store) Close() error {
	return nil
}
