package apikeys

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func GenerateKey() (string, string, error) {
	raw := make([]byte, 32)

	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("generate api key: %w", err)
	}

	key := "usk_" + hex.EncodeToString(raw)

	hash := sha256.Sum256([]byte(key))
	keyHash := hex.EncodeToString(hash[:])

	return key, keyHash, nil
}
