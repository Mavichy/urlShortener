package shortcode

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

const (
	Length   = 10
	Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
)

type Generator struct{}

func NewGenerator() Generator { return Generator{} }

func (Generator) Generate(originalUrl string, attempt int) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s#%d", originalUrl, attempt)))
	n := binary.BigEndian.Uint64(sum[:8])
	base := uint64(len(Alphabet))
	code := make([]byte, Length)

	for i := Length - 1; i >= 0; i-- {
		code[i] = Alphabet[n%base]
		n /= base
	}

	return string(code)
}
