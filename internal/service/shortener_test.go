package service

import (
	"context"
	"errors"
	"github.com/Mavichy/urlShortener/internal/model"
	"github.com/Mavichy/urlShortener/internal/shortcode"
	"testing"

	memorystore "github.com/Mavichy/urlShortener/internal/storage/memory"
)

func TestCreateReturnsSameLinkForSameOriginalURL(t *testing.T) {
	store := memorystore.New()
	svc := New(store, shortcode.NewGenerator())
	ctx := context.Background()

	first, err := svc.Create(ctx, "https://example.com/abc")
	if err != nil {
		t.Fatalf("create first link: %v", err)
	}

	second, err := svc.Create(ctx, "https://example.com/abc")
	if err != nil {
		t.Fatalf("create second link: %v", err)
	}

	if first.Code != second.Code {
		t.Fatalf("expected same code, got %q and %q", first.Code, second.Code)
	}

	if first.OriginalURL != second.OriginalURL {
		t.Fatalf("expected same original url, got %q and %q", first.OriginalURL, second.OriginalURL)
	}
}

func TestCreateRejectsInvalidURL(t *testing.T) {
	svc := New(memorystore.New(), shortcode.NewGenerator())

	_, err := svc.Create(context.Background(), "not-a-url")
	if !errors.Is(err, ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL, got %v", err)
	}
}

func TestCreateTrimsSpaces(t *testing.T) {
	svc := New(memorystore.New(), shortcode.NewGenerator())

	link, err := svc.Create(context.Background(), "   https://example.com/abc   ")
	if err != nil {
		t.Fatalf("create link: %v", err)
	}

	if link.OriginalURL != "https://example.com/abc" {
		t.Fatalf("expected trimmed url, got %q", link.OriginalURL)
	}
}

func TestResolveReturnsLinkByCode(t *testing.T) {
	store := memorystore.New()
	svc := New(store, shortcode.NewGenerator())
	ctx := context.Background()

	created, err := svc.Create(ctx, "https://example.com/resolve")
	if err != nil {
		t.Fatalf("create link: %v", err)
	}

	resolved, err := svc.Resolve(ctx, created.Code)
	if err != nil {
		t.Fatalf("resolve link: %v", err)
	}

	if resolved.Code != created.Code {
		t.Fatalf("expected code %q, got %q", created.Code, resolved.Code)
	}

	if resolved.OriginalURL != created.OriginalURL {
		t.Fatalf("expected original url %q, got %q", created.OriginalURL, resolved.OriginalURL)
	}
}

func TestResolveRejectsInvalidCode(t *testing.T) {
	svc := New(memorystore.New(), shortcode.NewGenerator())

	_, err := svc.Resolve(context.Background(), "bad")
	if !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("expected ErrInvalidCode, got %v", err)
	}
}

func TestCreateRetriesOnCodeCollision(t *testing.T) {
	store := memorystore.New()

	err := store.Save(context.Background(), model.Link{
		Code:        "AAAAAAAAAA",
		OriginalURL: "https://already-exists.com",
	})
	if err != nil {
		t.Fatalf("prepare store: %v", err)
	}

	svc := New(store, stubGenerator{
		codes: []string{
			"AAAAAAAAAA",
			"BBBBBBBBBB",
		},
	})

	link, err := svc.Create(context.Background(), "https://new-url.com")
	if err != nil {
		t.Fatalf("create link after collision: %v", err)
	}

	if link.Code != "BBBBBBBBBB" {
		t.Fatalf("expected fallback code %q, got %q", "BBBBBBBBBB", link.Code)
	}
}

type stubGenerator struct {
	codes []string
}

func (g stubGenerator) Generate(_ string, attempt int) string {
	if attempt < len(g.codes) {
		return g.codes[attempt]
	}
	return "CCCCCCCCCC"
}
