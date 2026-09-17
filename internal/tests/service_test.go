package tests

import (
	"github.com/SnoWed-29/url-shortener/internal"
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid https URL",
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "valid http URL",
			url:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "missing scheme",
			url:     "example.com",
			wantErr: true,
		},
		{
			name:    "unsupported scheme",
			url:     "ftp://example.com",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "javascript URL",
			url:     "javascript:alert(1)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := internal.ValidateURL(tt.url)

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"ValidateURL() error = %v, wantErr = %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestGenerateShortCode(t *testing.T) {
	code, err := internal.GenerateShortCode(6)

	if err != nil {
		t.Fatalf("GenerateShortCode() error = %v", err)
	}

	if len(code) != 6 {
		t.Fatalf(
			"GenerateShortCode() length = %d, want 6",
			len(code),
		)
	}
}
