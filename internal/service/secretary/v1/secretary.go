// Package secretary provides methods for ciphering.
package secretary

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"

	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/secretary"
)

// Check interface implementation explicitly
var (
	_ secretary.Secretary = (*Secretary)(nil)
)

// Secretary defines object structure and its attributes.
type Secretary struct {
	aesgcm cipher.AEAD
	nonce  []byte
}

// NewSecretaryService initializes a secretary service with ciphering functionality.
func NewSecretaryService(c *config.SecretConfig) (*Secretary, error) {
	key := sha256.Sum256([]byte(c.UserKey))
	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}
	nonce := key[len(key)-aesgcm.NonceSize():]
	return &Secretary{
		aesgcm: aesgcm,
		nonce:  nonce,
	}, nil
}

// Encode ciphers data using the previously established cipher.
func (s *Secretary) Encode(data string) string {
	encoded := s.aesgcm.Seal(nil, s.nonce, []byte(data), nil)
	return hex.EncodeToString(encoded)
}

// Decode deciphers data using the previously established cipher.
func (s *Secretary) Decode(msg string) (string, error) {
	msgBytes, err := hex.DecodeString(msg)
	if err != nil {
		return "", err
	}
	decoded, err := s.aesgcm.Open(nil, s.nonce, msgBytes, nil)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
