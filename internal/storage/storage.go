package storage

import (
	"context"
	"errors"
	"github.com/Mavichy/urlShortener/internal/model"
)

var (
	ErrNotFound       = errors.New("link not found")
	ErrOriginalExists = errors.New("original url already exists")
	ErrCodeExists     = errors.New("short code already exists")
)

type Storage interface {
	Save(ctx context.Context, link model.Link) error
	GetByCode(ctx context.Context, code string) (model.Link, error)
	GetByOriginalURL(ctx context.Context, originalURL string) (model.Link, error)
	Close() error
}
