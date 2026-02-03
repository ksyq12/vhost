package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/driver"
	"github.com/ksyq12/vhost/internal/executor"
)

func TestCheckSystemRequirements(t *testing.T) {
	tests := []struct {
		name           string
		driverName     string
		setupExecutor  func() *executor.MockExecutor
		setupConfig    func() *config.Config
		checkResults   func(*testing.T, []CheckResult)
	}{
		{
			name:       "all requirements satisfied",
			driverName: "nginx",
			setupExecutor: func() *executor.MockExecutor {
				return &executor.MockExecutor{
					LookPathFunc: func(file string) (string, error) {
						return "/usr/bin/" + file, nil
					},
					ExecuteFunc: func(name string, args ...string) ([]byte, error) {
						if name == "nginx" && len(args) > 0 && args[0] == "-v" {
							return []byte("nginx version: nginx/1.20.0"), nil
						}
						if name == "systemctl" && len(args) >= 2 && args[0] == "is-active" {
							return []byte("active"), nil
						}
						return []byte(""), nil
					},
				}
			},
			setupConfig: func() *config.Config {
				cfg := config.New()
				cfg.Driver = "nginx"
				return cfg
			},
			checkResults: func(t *testing.T, results []CheckResult) {
				// Find nginx check
				foundNginx := false
				for _, r := range results {
					if strings.Contains(r.Message, "Nginx") && r.Status == "success" {
						foundNginx = true
						if !strings.Contains(r.Message, "1.20.0") {
							t.Error("nginx version not extracted")
						}
					}
				}
				if !foundNginx {
					t.Error("nginx check not found or failed")
				}
			},
		},
		{
			name:       "nginx not installed (required)",
			driverName: "nginx",
			setupExecutor: func() *executor.MockExecutor {
				return &executor.MockExecutor{
					LookPathFunc: func(file string) (string, error) {
						if file == "nginx" {
							return "", os.ErrNotExist
						}
						return "/usr/bin/" + file, nil
					},
				}
			},
			setupConfig: func() *config.Config {
				cfg := config.New()
				cfg.Driver = "nginx"
				return cfg
			},
			checkResults: func(t *testing.T, results []CheckResult) {
				// Find nginx check - should be error since it's required
				foundNginxError := false
				for _, r := range results {
					if strings.Contains(r.Message, "Nginx") && r.Status == "error" {
						foundNginxError = true
					}
				}
				if !foundNginxError {
					t.Error("expected nginx error status for missing required server")
				}
			},
		},
		{
			name:       "apache optional when nginx is driver",
			driverName: "nginx",
			setupExecutor: func() *executor.MockExecutor {
				return &executor.MockExecutor{
					LookPathFunc: func(file string) (string, error) {
						if file == "apache2" {
							return "", os.ErrNotExist
						}
						return "/usr/bin/" + file, nil
					},
					ExecuteFunc: func(name string, args ...string) ([]byte, error) {
						return []byte(""), nil
					},
				}
			},
			setupConfig: func() *config.Config {
				cfg := config.New()
				cfg.Driver = "nginx"
				return cfg
			},
			checkResults: func(t *testing.T, results []CheckResult) {
				// Find apache check - should be warning since it's optional
				foundApacheWarning := false
				for _, r := range results {
					if strings.Contains(r.Message, "Apache") && r.Status == "warning" {
						foundApacheWarning = true
					}
				}
				if !foundApacheWarning {
					t.Error("expected apache warning status for optional server")
				}
			},
		},
		{
			name:       "php-fpm error when needed",
			driverName: "nginx",
			setupExecutor: func() *executor.MockExecutor {
				return &executor.MockExecutor{
					LookPathFunc: func(file string) (string, error) {
						return "/usr/bin/" + file, nil
					},
					ExecuteFunc: func(name string, args ...string) ([]byte, error) {
						// PHP-FPM not running
						if name == "systemctl" && strings.Contains(strings.Join(args, " "), "php") {
							return []byte("inactive"), nil
						}
						return []byte(""), nil
					},
				}
			},
			setupConfig: func() *config.Config {
				cfg := config.New()
				cfg.Driver = "nginx"
				cfg.VHosts["php.com"] = &config.VHost{
					Domain: "php.com",
					Type:   "php",
				}
				return cfg
			},
			checkResults: func(t *testing.T, results []CheckResult) {
				// PHP-FPM should be error when PHP vhost exists
				foundPHPError := false
				for _, r := range results {
					if strings.Contains(r.Message, "PHP-FPM") && r.Status == "error" {
						foundPHPError = true
					}
				}
				if !foundPHPError {
					t.Error("expected PHP-FPM error when PHP vhost exists")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec := tt.setupExecutor()
			cfg := tt.setupConfig()

			results := checkSystemRequirements(exec, cfg)

			tt.checkResults(t, results)
		})
	}
}

func TestCheckConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		setupDriver  func(string, string) *driver.MockDriver
		setupConfig  func(*testing.T) *config.Config
		checkResults func(*testing.T, []CheckResult)
	}{
		{
			name: "config valid",
			setupDriver: func(available, enabled string) *driver.MockDriver {
				drv := driver.NewMockDriver("nginx", available, enabled)
				drv.TestFunc = func() error {
					return nil
				}
				return drv
			},
			setupConfig: func(t *testing.T) *config.Config {
				cfg := config.New()
				cfg.Driver = "nginx"
				return cfg
			},
			checkResults: func(t *testing.T, results []CheckResult) {
				foundTestSuccess := false
				for _, r := range results {
					if strings.Contains(r.Message, "syntax OK") && r.Status == "success" {
						foundTestSuccess = true
					}
				}
				if !foundTestSuccess {
					t.Error("expected config test success")
				}
			},
		},
		{
			name: "config syntax error",
			setupDriver: func(available, enabled string) *driver.MockDriver {
				drv := driver.NewMockDriver("nginx", available, enabled)
				drv.TestFunc = func() error {
					return os.ErrNotExist
				}
				return drv
			},
			setupConfig: func(t *testing.T) *config.Config {
				cfg := config.New()
				cfg.Driver = "nginx"
				return cfg
			},
			checkResults: func(t *testing.T, results []CheckResult) {
				foundTestError := false
				for _, r := range results {
					if strings.Contains(r.Message, "syntax error") && r.Status == "error" {
						foundTestError = true
					}
				}
				if !foundTestError {
					t.Error("expected config test error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			available := filepath.Join(tempDir, "sites-available")
			enabled := filepath.Join(tempDir, "sites-enabled")

			drv := tt.setupDriver(available, enabled)
			cfg := tt.setupConfig(t)

			results := checkConfiguration(drv, cfg)

			tt.checkResults(t, results)
		})
	}
}

func TestCheckVHosts(t *testing.T) {
	tests := []struct {
		name         string
		setupDriver  func(string, string) *driver.MockDriver
		setupConfig  func(*testing.T) *config.Config
		checkResults func(*testing.T, []VHostStatus)
	}{
		{
			name: "vhost enabled and valid",
			setupDriver: func(available, enabled string) *driver.MockDriver {
				drv := driver.NewMockDriver("nginx", available, enabled)
				drv.IsEnabledFunc = func(domain string) (bool, error) {
					return true, nil
				}
				return drv
			},
			setupConfig: func(t *testing.T) *config.Config {
				tempDir := t.TempDir()
				// Create root directory
				rootDir := filepath.Join(tempDir, "www")
				if err := os.MkdirAll(rootDir, 0755); err != nil {
					t.Fatalf("failed to create root dir: %v", err)
				}

				cfg := config.New()
				cfg.VHosts["test.com"] = &config.VHost{
					Domain:  "test.com",
					Type:    "static",
					Root:    rootDir,
					Enabled: true,
				}
				return cfg
			},
			checkResults: func(t *testing.T, statuses []VHostStatus) {
				if len(statuses) != 1 {
					t.Fatalf("expected 1 status, got %d", len(statuses))
				}
				status := statuses[0]
				if status.Domain != "test.com" {
					t.Errorf("expected domain test.com, got %s", status.Domain)
				}
				if !status.Enabled {
					t.Error("expected enabled to be true")
				}
				// Should have success check
				hasSuccess := false
				for _, check := range status.Checks {
					if check.Status == "success" {
						hasSuccess = true
					}
				}
				if !hasSuccess {
					t.Error("expected success check")
				}
			},
		},
		{
			name: "vhost enabled status mismatch",
			setupDriver: func(available, enabled string) *driver.MockDriver {
				drv := driver.NewMockDriver("nginx", available, enabled)
				drv.IsEnabledFunc = func(domain string) (bool, error) {
					return false, nil // Driver says disabled
				}
				return drv
			},
			setupConfig: func(t *testing.T) *config.Config {
				cfg := config.New()
				cfg.VHosts["mismatch.com"] = &config.VHost{
					Domain:  "mismatch.com",
					Type:    "static",
					Root:    "/var/www/mismatch",
					Enabled: true, // Config says enabled
				}
				return cfg
			},
			checkResults: func(t *testing.T, statuses []VHostStatus) {
				if len(statuses) != 1 {
					t.Fatalf("expected 1 status, got %d", len(statuses))
				}
				status := statuses[0]
				// Should have warning about mismatch
				hasWarning := false
				for _, check := range status.Checks {
					if check.Status == "warning" && strings.Contains(check.Message, "mismatch") {
						hasWarning = true
					}
				}
				if !hasWarning {
					t.Error("expected warning about enabled status mismatch")
				}
			},
		},
		{
			name: "vhost missing root directory",
			setupDriver: func(available, enabled string) *driver.MockDriver {
				drv := driver.NewMockDriver("nginx", available, enabled)
				drv.IsEnabledFunc = func(domain string) (bool, error) {
					return true, nil
				}
				return drv
			},
			setupConfig: func(t *testing.T) *config.Config {
				cfg := config.New()
				cfg.VHosts["missing-root.com"] = &config.VHost{
					Domain:  "missing-root.com",
					Type:    "static",
					Root:    "/nonexistent/path",
					Enabled: true,
				}
				return cfg
			},
			checkResults: func(t *testing.T, statuses []VHostStatus) {
				if len(statuses) != 1 {
					t.Fatalf("expected 1 status, got %d", len(statuses))
				}
				status := statuses[0]
				// Should have warning about missing root
				hasWarning := false
				for _, check := range status.Checks {
					if check.Status == "warning" && strings.Contains(check.Message, "root directory") {
						hasWarning = true
					}
				}
				if !hasWarning {
					t.Error("expected warning about missing root directory")
				}
			},
		},
		{
			name: "vhost ssl missing certificate",
			setupDriver: func(available, enabled string) *driver.MockDriver {
				drv := driver.NewMockDriver("nginx", available, enabled)
				drv.IsEnabledFunc = func(domain string) (bool, error) {
					return true, nil
				}
				return drv
			},
			setupConfig: func(t *testing.T) *config.Config {
				cfg := config.New()
				cfg.VHosts["ssl.com"] = &config.VHost{
					Domain:  "ssl.com",
					Type:    "static",
					Root:    "/var/www/ssl",
					SSL:     true,
					SSLCert: "/nonexistent/cert.pem",
					SSLKey:  "/nonexistent/key.pem",
					Enabled: true,
				}
				return cfg
			},
			checkResults: func(t *testing.T, statuses []VHostStatus) {
				if len(statuses) != 1 {
					t.Fatalf("expected 1 status, got %d", len(statuses))
				}
				status := statuses[0]
				// Should have error about missing SSL cert
				hasCertError := false
				hasKeyError := false
				for _, check := range status.Checks {
					if check.Status == "error" && strings.Contains(check.Message, "certificate") {
						hasCertError = true
					}
					if check.Status == "error" && strings.Contains(check.Message, "key") {
						hasKeyError = true
					}
				}
				if !hasCertError {
					t.Error("expected error about missing SSL certificate")
				}
				if !hasKeyError {
					t.Error("expected error about missing SSL key")
				}
			},
		},
		{
			name: "empty vhosts config",
			setupDriver: func(available, enabled string) *driver.MockDriver {
				return driver.NewMockDriver("nginx", available, enabled)
			},
			setupConfig: func(t *testing.T) *config.Config {
				return config.New()
			},
			checkResults: func(t *testing.T, statuses []VHostStatus) {
				if len(statuses) != 0 {
					t.Errorf("expected 0 statuses, got %d", len(statuses))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			available := filepath.Join(tempDir, "sites-available")
			enabled := filepath.Join(tempDir, "sites-enabled")

			drv := tt.setupDriver(available, enabled)
			cfg := tt.setupConfig(t)

			statuses := checkVHosts(drv, cfg)

			tt.checkResults(t, statuses)
		})
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"nginx", "Nginx"},
		{"apache", "Apache"},
		{"caddy", "Caddy"},
		{"", ""},
		{"A", "A"},
		{"ABC", "ABC"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := capitalize(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
