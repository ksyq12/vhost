package cli

import (
	"testing"
)

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		wantErr bool
	}{
		{"valid simple domain", "example.com", false},
		{"valid subdomain", "www.example.com", false},
		{"valid deep subdomain", "api.v2.example.com", false},
		{"valid with hyphen", "my-site.example.com", false},
		{"valid with numbers", "api123.example.com", false},
		{"empty domain", "", true},
		{"domain with space", "example .com", true},
		{"domain with spaces", "my domain.com", true},
		{"starts with hyphen", "-example.com", true},
		{"ends with hyphen", "example.com-", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDomain(tt.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDomain(%q) error = %v, wantErr %v", tt.domain, err, tt.wantErr)
			}
		})
	}
}

func TestValidateRoot(t *testing.T) {
	tests := []struct {
		name    string
		root    string
		wantErr bool
	}{
		{"absolute path", "/var/www/html", false},
		{"root path", "/", false},
		{"home directory", "/home/user/www", false},
		{"empty (allowed)", "", false},
		{"relative path", "var/www/html", true},
		{"relative dot path", "./www", true},
		{"relative parent path", "../www", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRoot(tt.root)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRoot(%q) error = %v, wantErr %v", tt.root, err, tt.wantErr)
			}
		})
	}
}

func TestValidateProxyURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid http url", "http://localhost:3000", false},
		{"valid https url", "https://api.example.com", false},
		{"valid http with path", "http://localhost:8080/api", false},
		{"host:port without scheme", "localhost:3000", false},
		{"ip:port without scheme", "127.0.0.1:8080", false},
		{"empty (allowed)", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProxyURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateProxyURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestNewSuccessResult(t *testing.T) {
	result := newSuccessResult("example.com", "added")

	if !result.Success {
		t.Error("expected Success to be true")
	}
	if result.Domain != "example.com" {
		t.Errorf("expected domain example.com, got %s", result.Domain)
	}
	if result.Action != "added" {
		t.Errorf("expected action added, got %s", result.Action)
	}
}

func TestCommandResult(t *testing.T) {
	t.Run("full result", func(t *testing.T) {
		result := CommandResult{
			Success: true,
			Domain:  "test.com",
			Action:  "created",
			Message: "VHost created successfully",
		}

		if !result.Success {
			t.Error("expected Success to be true")
		}
		if result.Domain != "test.com" {
			t.Errorf("expected domain test.com, got %s", result.Domain)
		}
		if result.Action != "created" {
			t.Errorf("expected action created, got %s", result.Action)
		}
		if result.Message != "VHost created successfully" {
			t.Errorf("unexpected message: %s", result.Message)
		}
	})

	t.Run("minimal result", func(t *testing.T) {
		result := CommandResult{
			Success: false,
			Domain:  "fail.com",
		}

		if result.Success {
			t.Error("expected Success to be false")
		}
		if result.Action != "" {
			t.Errorf("expected empty action, got %s", result.Action)
		}
		if result.Message != "" {
			t.Errorf("expected empty message, got %s", result.Message)
		}
	})
}
