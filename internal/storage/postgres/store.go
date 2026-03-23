package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/Mavichy/urlShortener/internal/model"
	"github.com/Mavichy/urlShortener/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

const schema = `
CREATE TABLE IF NOT EXISTS links (
    id BIGSERIAL PRIMARY KEY,
    original_url TEXT NOT NULL UNIQUE,
    code VARCHAR(10) NOT NULL UNIQUE
);`

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	if _, err := pool.Exec(ctx, schema); err != nil {
		pool.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}

	return &Store{pool: pool}, nil
}

func (s *Store) Save(ctx context.Context, link model.Link) error {
	_, err := s.pool.Exec(
		ctx,
		`INSERT INTO links (original_url, code) VALUES ($1, $2)`,
		link.OriginalURL,
		link.Code,
	)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		constraint := strings.ToLower(pgErr.ConstraintName)

		switch {
		case strings.Contains(constraint, "original"):
			return storage.ErrOriginalExists
		case strings.Contains(constraint, "code"):
			return storage.ErrCodeExists
		default:
			detail := strings.ToLower(pgErr.Detail)
			if strings.Contains(detail, "original_url") {
				return storage.ErrOriginalExists
			}
			if strings.Contains(detail, "code") {
				return storage.ErrCodeExists
			}
		}
	}

	return fmt.Errorf("insert link: %w", err)
}

func (s *Store) GetByCode(ctx context.Context, code string) (model.Link, error) {
	var link model.Link

	err := s.pool.QueryRow(
		ctx,
		`SELECT code, original_url FROM links WHERE code = $1`,
		code,
	).Scan(&link.Code, &link.OriginalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Link{}, storage.ErrNotFound
		}
		return model.Link{}, fmt.Errorf("select by code: %w", err)
	}

	return link, nil
}

func (s *Store) GetByOriginalURL(ctx context.Context, originalURL string) (model.Link, error) {
	var link model.Link

	err := s.pool.QueryRow(
		ctx,
		`SELECT code, original_url FROM links WHERE original_url = $1`,
		originalURL,
	).Scan(&link.Code, &link.OriginalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Link{}, storage.ErrNotFound
		}
		return model.Link{}, fmt.Errorf("select by original url: %w", err)
	}

	return link, nil
}

func (s *Store) Close() error {
	s.pool.Close()
	return nil
}
