package internal
import (
	"crypto/rand"
	"fmt"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" 

 func GenerateShortCode(lenght int) (string, error) {
	if lenght <= 0 {
		return "", fmt.Errorf("short code length must be greater than zero")
	}

	result := make([]byte, lenght)

	randomBytes := make([]byte, lenght)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	
	for i := range result {
		result[i] = alphabet[int(randomBytes[i])%len(alphabet)]
	}

	return string(result), nil
 }