package shortcode

import (
	"testing"
)

func TestGeneratorGenerate_LengthAndAlphabet(t *testing.T) {
	g := NewGenerator()
	code := g.Generate("https://example.com/path", 0)

	if len(code) != Length {
		t.Fatalf("expected code length %d, got %d", Length, len(code))
	}

	for _, r := range code {
		if !containsRune(Alphabet, r) {
			t.Fatalf("unexpected rune %q in code %q", r, code)
		}
	}
}

func TestGeneratorGenerate_DeterministicForSameInput(t *testing.T) {
	g := NewGenerator()

	left := g.Generate("https://example.com/path", 0)
	right := g.Generate("https://example.com/path", 0)

	if left != right {
		t.Fatalf("expected deterministic generation, got %q and %q", left, right)
	}
}

func TestGeneratorGenerate_DifferentAttemptsProduceDifferentCodes(t *testing.T) {
	g := NewGenerator()

	first := g.Generate("https://example.com/path", 0)
	second := g.Generate("https://example.com/path", 1)

	if first == second {
		t.Fatalf("expected different codes for different attempts, got same code %q", first)
	}
}

func containsRune(s string, target rune) bool {
	for _, r := range s {
		if r == target {
			return true
		}
	}
	return false
}
