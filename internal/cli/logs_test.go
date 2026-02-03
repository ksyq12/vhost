package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/driver"
)

func TestRunLogs(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		setupFlags  func()
		setupDeps   func(*testing.T, *driver.MockDriver, string) *Dependencies
		wantErr     bool
		errContains string
	}{
		{
			name:   "logs with no log files found",
			domain: "nologs.com",
			setupFlags: func() {
				logsAccess = false
				logsError = false
				logsFollow = false
				logsLines = 20
			},
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver, availableDir string) *Dependencies {
				// Create config file but no log files
				if err := os.MkdirAll(availableDir, 0755); err != nil {
					t.Fatalf("failed to create dir: %v", err)
				}
				configPath := filepath.Join(availableDir, "nologs.com")
				if err := os.WriteFile(configPath, []byte("server {}"), 0644); err != nil {
					t.Fatalf("failed to create config: %v", err)
				}

				cfg := config.New()
				cfg.VHosts["nologs.com"] = &config.VHost{
					Domain: "nologs.com",
					Type:   "static",
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					Build()
			},
			wantErr:     true,
			errContains: "no log files found",
		},
		{
			name:   "logs with invalid domain",
			domain: "invalid domain",
			setupFlags: func() {
				logsAccess = false
				logsError = false
				logsFollow = false
				logsLines = 20
			},
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
		{
			name:   "logs with vhost not in config (warning only)",
			domain: "notinconfig.com",
			setupFlags: func() {
				logsAccess = false
				logsError = false
				logsFollow = false
				logsLines = 20
			},
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver, availableDir string) *Dependencies {
				// Create config file
				if err := os.MkdirAll(availableDir, 0755); err != nil {
					t.Fatalf("failed to create dir: %v", err)
				}
				configPath := filepath.Join(availableDir, "notinconfig.com")
				if err := os.WriteFile(configPath, []byte("server {}"), 0644); err != nil {
					t.Fatalf("failed to create config: %v", err)
				}

				cfg := config.New()
				// Not adding vhost to config
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					Build()
			},
			wantErr:     true, // Still fails because log files don't exist
			errContains: "no log files found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temp directories
			tempDir := t.TempDir()
			availableDir := filepath.Join(tempDir, "sites-available")
			enabledDir := filepath.Join(tempDir, "sites-enabled")

			// Create mock driver
			mockDrv := driver.NewMockDriver("nginx", availableDir, enabledDir)

			// Setup flags
			tt.setupFlags()

			// Setup and inject dependencies
			oldDeps := deps
			mockDepsObj := tt.setupDeps(t, mockDrv, availableDir)
			deps = mockDepsObj
			defer func() { deps = oldDeps }()

			// Execute
			err := runLogs(nil, []string{tt.domain})

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

func TestParseNginxLogPath(t *testing.T) {
	tests := []struct {
		name      string
		config    string
		directive string
		expected  string
	}{
		{
			name:      "access log with combined format",
			config:    "access_log /var/log/nginx/access.log combined;",
			directive: "access_log",
			expected:  "/var/log/nginx/access.log",
		},
		{
			name:      "error log simple",
			config:    "error_log /var/log/nginx/error.log;",
			directive: "error_log",
			expected:  "/var/log/nginx/error.log",
		},
		{
			name:      "custom path",
			config:    "access_log /custom/path/site-access.log;",
			directive: "access_log",
			expected:  "/custom/path/site-access.log",
		},
		{
			name:      "directive not found",
			config:    "server_name example.com;",
			directive: "access_log",
			expected:  "",
		},
		{
			name:      "access log off",
			config:    "access_log off;",
			directive: "access_log",
			expected:  "off",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseNginxLogPath(tt.config, tt.directive)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParseApacheLogPath(t *testing.T) {
	tests := []struct {
		name      string
		config    string
		directive string
		expected  string
	}{
		{
			name:      "custom log with format",
			config:    `CustomLog /var/log/apache2/access.log combined`,
			directive: "CustomLog",
			expected:  "/var/log/apache2/access.log",
		},
		{
			name:      "error log",
			config:    "ErrorLog /var/log/apache2/error.log",
			directive: "ErrorLog",
			expected:  "/var/log/apache2/error.log",
		},
		{
			name:      "with apache variable",
			config:    "ErrorLog ${APACHE_LOG_DIR}/error.log",
			directive: "ErrorLog",
			expected:  "/var/log/apache2/error.log",
		},
		{
			name:      "directive not found",
			config:    "ServerName example.com",
			directive: "ErrorLog",
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseApacheLogPath(tt.config, tt.directive)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParseCaddyLogPath(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		expected string
	}{
		{
			name: "output file",
			config: `log {
				output file /var/log/caddy/access.log
			}`,
			expected: "/var/log/caddy/access.log",
		},
		{
			name:     "no log directive",
			config:   "example.com {\n  root * /var/www\n}",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCaddyLogPath(tt.config)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetDefaultLogPath(t *testing.T) {
	tests := []struct {
		name       string
		driverName string
		domain     string
		logType    string
		expected   string
	}{
		{
			name:       "nginx access",
			driverName: "nginx",
			domain:     "example.com",
			logType:    "access",
			expected:   "/var/log/nginx/example.com-access.log",
		},
		{
			name:       "nginx error",
			driverName: "nginx",
			domain:     "example.com",
			logType:    "error",
			expected:   "/var/log/nginx/example.com-error.log",
		},
		{
			name:       "apache access",
			driverName: "apache",
			domain:     "example.com",
			logType:    "access",
			expected:   "/var/log/apache2/example.com-access.log",
		},
		{
			name:       "apache error",
			driverName: "apache",
			domain:     "example.com",
			logType:    "error",
			expected:   "/var/log/apache2/example.com-error.log",
		},
		{
			name:       "caddy",
			driverName: "caddy",
			domain:     "example.com",
			logType:    "access",
			expected:   "/var/log/caddy/example.com.log",
		},
		{
			name:       "unknown driver",
			driverName: "unknown",
			domain:     "example.com",
			logType:    "access",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDefaultLogPath(tt.driverName, tt.domain, tt.logType)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
