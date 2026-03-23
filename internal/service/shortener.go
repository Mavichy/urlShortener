package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/Mavichy/urlShortener/internal/model"
	"github.com/Mavichy/urlShortener/internal/shortcode"
	"github.com/Mavichy/urlShortener/internal/storage"
	"net/url"
	"strings"
)

const maxGenerateAttempts = 128

var ErrInvalidURL = errors.New("invalid url")
var ErrInvalidCode = errors.New("invalid short code")

type CodeGenerator interface {
	Generate(originalURL string, attempt int) string
}

type Service struct {
	store     storage.Storage
	generator CodeGenerator
}

func New(store storage.Storage, generator CodeGenerator) *Service {
	return &Service{
		store:     store,
		generator: generator,
	}
}

func (s *Service) Create(ctx context.Context, originalURL string) (model.Link, error) {
	normalizedURL, err := normalizeURL(originalURL)
	if err != nil {
		return model.Link{}, err
	}

	existing, err := s.store.GetByOriginalURL(ctx, normalizedURL)
	if err == nil {
		return existing, nil
	}
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return model.Link{}, fmt.Errorf("get by original url: %w", err)
	}

	for attempt := 0; attempt < maxGenerateAttempts; attempt++ {
		candidate := model.Link{
			Code:        s.generator.Generate(normalizedURL, attempt),
			OriginalURL: normalizedURL,
		}

		err = s.store.Save(ctx, candidate)
		if err == nil {
			return s.store.GetByOriginalURL(ctx, normalizedURL)
		}

		switch {
		case errors.Is(err, storage.ErrOriginalExists):
			link, getErr := s.store.GetByOriginalURL(ctx, normalizedURL)
			if getErr != nil {
				return model.Link{}, fmt.Errorf("original url already exists, but fetch failed: %w", getErr)
			}
			return link, nil
		case errors.Is(err, storage.ErrCodeExists):
			continue
		default:
			return model.Link{}, fmt.Errorf("save link: %w", err)
		}
	}

	return model.Link{}, errors.New("failed to generate unique short code")
}

func (s *Service) Resolve(ctx context.Context, code string) (model.Link, error) {
	if !isValidCode(code) {
		return model.Link{}, ErrInvalidCode
	}

	return s.store.GetByCode(ctx, code)
}

func normalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidURL
	}

	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return "", ErrInvalidURL
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidURL
	}

	return raw, nil
}

func isValidCode(code string) bool {
	if len(code) != shortcode.Length {
		return false
	}

	for _, r := range code {
		if !strings.ContainsRune(shortcode.Alphabet, r) {
			return false
		}
	}

	return true
}
