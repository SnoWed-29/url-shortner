package links

import (
	"crypto/rand"
	"fmt"
	"net/url"
	"strings"
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

func ValidateURL(rawURL string) error {
	rawURL = strings.TrimSpace(rawURL)

	if rawURL == "" {
		return fmt.Errorf("url is required")
	}

	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use http or https")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must contain a host")
	}

	return nil
}

func ValidateCustomAlias(alias string) error {
	alias = strings.TrimSpace(alias)

	if alias == "" {
		return fmt.Errorf("custom alias cannot be empty")
	}

	if len(alias) < 3 || len(alias) > 16 {
		return fmt.Errorf("custom alias must be between 3 and 16 characters")
	}

	for _, char := range alias {
		validChar := (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_'
		if !validChar {
			return fmt.Errorf("custom alias can only contain letters, numbers, '-' and '_'")
		}
	}

	return nil
}
