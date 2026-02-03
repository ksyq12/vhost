package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/driver"
)

func TestRunEdit(t *testing.T) {
	// Set EDITOR to a non-existent command to prevent actual editor execution
	originalEditor := os.Getenv("EDITOR")
	os.Setenv("EDITOR", "/nonexistent/editor/for/testing")
	defer func() {
		if originalEditor != "" {
			os.Setenv("EDITOR", originalEditor)
		} else {
			os.Unsetenv("EDITOR")
		}
	}()

	tests := []struct {
		name        string
		domain      string
		driverName  string
		setupDeps   func(*testing.T, *driver.MockDriver, string) *Dependencies
		wantErr     bool
		errContains string
	}{
		{
			name:       "edit existing nginx vhost",
			domain:     "test.com",
			driverName: "nginx",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver, availableDir string) *Dependencies {
				// Create config file
				configPath := filepath.Join(availableDir, "test.com")
				if err := os.MkdirAll(availableDir, 0755); err != nil {
					t.Fatalf("failed to create dir: %v", err)
				}
				if err := os.WriteFile(configPath, []byte("server {}"), 0644); err != nil {
					t.Fatalf("failed to create config file: %v", err)
				}

				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					Build()
			},
			wantErr:     true,
			errContains: "editor not found", // Non-existent editor
		},
		{
			name:       "edit existing apache vhost (adds .conf suffix)",
			domain:     "test.com",
			driverName: "apache",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver, availableDir string) *Dependencies {
				// Create config file with .conf suffix for apache
				configPath := filepath.Join(availableDir, "test.com.conf")
				if err := os.MkdirAll(availableDir, 0755); err != nil {
					t.Fatalf("failed to create dir: %v", err)
				}
				if err := os.WriteFile(configPath, []byte("<VirtualHost>"), 0644); err != nil {
					t.Fatalf("failed to create config file: %v", err)
				}

				cfg := config.New()
				// Use apache driver
				apacheDrv := driver.NewMockDriver("apache", availableDir, filepath.Join(filepath.Dir(availableDir), "sites-enabled"))
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(apacheDrv).
					Build()
			},
			wantErr:     true,
			errContains: "editor not found", // Non-existent editor
		},
		{
			name:       "edit config file not found",
			domain:     "missing.com",
			driverName: "nginx",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver, availableDir string) *Dependencies {
				// Don't create config file
				if err := os.MkdirAll(availableDir, 0755); err != nil {
					t.Fatalf("failed to create dir: %v", err)
				}

				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					Build()
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:       "edit with invalid domain",
			domain:     "invalid domain",
			driverName: "nginx",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver, availableDir string) *Dependencies {
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					Build()
			},
			wantErr:     true,
			errContains: "spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temp directories
			tempDir := t.TempDir()
			availableDir := filepath.Join(tempDir, "sites-available")
			enabledDir := filepath.Join(tempDir, "sites-enabled")

			// Create mock driver
			mockDrv := driver.NewMockDriver(tt.driverName, availableDir, enabledDir)

			// Setup and inject dependencies
			oldDeps := deps
			mockDepsObj := tt.setupDeps(t, mockDrv, availableDir)
			deps = mockDepsObj
			defer func() { deps = oldDeps }()

			// Execute
			err := runEdit(nil, []string{tt.domain})

			// Check error
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGetEditor(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "custom editor from env",
			envValue: "nano",
			expected: "nano",
		},
		{
			name:     "default to vi",
			envValue: "",
			expected: "vi",
		},
		{
			name:     "emacs from env",
			envValue: "emacs",
			expected: "emacs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			original := os.Getenv("EDITOR")
			defer func() {
				if original != "" {
					os.Setenv("EDITOR", original)
				} else {
					os.Unsetenv("EDITOR")
				}
			}()

			// Set test value
			if tt.envValue != "" {
				os.Setenv("EDITOR", tt.envValue)
			} else {
				os.Unsetenv("EDITOR")
			}

			result := getEditor()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
