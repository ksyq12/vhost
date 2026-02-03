package cli

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/driver"
)

func TestRunShow(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		setupDeps   func(*testing.T, *driver.MockDriver) (*Dependencies, *config.Config)
		wantErr     bool
		errContains string
		validate    func(*testing.T, *config.Config, *driver.MockDriver)
	}{
		{
			name:   "show existing vhost",
			domain: "test.com",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				mockDrv.IsEnabledFunc = func(domain string) (bool, error) {
					return true, nil
				}
				cfg := config.New()
				cfg.VHosts["test.com"] = &config.VHost{
					Domain:    "test.com",
					Type:      "static",
					Root:      "/var/www/test",
					Enabled:   true,
					CreatedAt: time.Now(),
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					Build(), cfg
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				if len(mockDrv.IsEnabledCalls) != 1 {
					t.Errorf("expected 1 IsEnabled call, got %d", len(mockDrv.IsEnabledCalls))
				}
			},
		},
		{
			name:   "show php vhost",
			domain: "php.com",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				mockDrv.IsEnabledFunc = func(domain string) (bool, error) {
					return true, nil
				}
				cfg := config.New()
				cfg.VHosts["php.com"] = &config.VHost{
					Domain:     "php.com",
					Type:       "php",
					Root:       "/var/www/php",
					PHPVersion: "8.2",
					Enabled:    true,
					CreatedAt:  time.Now(),
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					Build(), cfg
			},
			wantErr: false,
		},
		{
			name:   "show proxy vhost",
			domain: "proxy.com",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				mockDrv.IsEnabledFunc = func(domain string) (bool, error) {
					return false, nil
				}
				cfg := config.New()
				cfg.VHosts["proxy.com"] = &config.VHost{
					Domain:    "proxy.com",
					Type:      "proxy",
					ProxyPass: "http://localhost:3000",
					Enabled:   false,
					CreatedAt: time.Now(),
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					Build(), cfg
			},
			wantErr: false,
		},
		{
			name:   "show ssl vhost",
			domain: "ssl.com",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				mockDrv.IsEnabledFunc = func(domain string) (bool, error) {
					return true, nil
				}
				cfg := config.New()
				cfg.VHosts["ssl.com"] = &config.VHost{
					Domain:    "ssl.com",
					Type:      "static",
					Root:      "/var/www/ssl",
					SSL:       true,
					SSLCert:   "/etc/ssl/certs/ssl.com.crt",
					SSLKey:    "/etc/ssl/private/ssl.com.key",
					Enabled:   true,
					CreatedAt: time.Now(),
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					Build(), cfg
			},
			wantErr: false,
		},
		{
			name:   "show non-existent vhost",
			domain: "notfound.com",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				// Not adding vhost to config
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					Build(), cfg
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:   "show with invalid domain",
			domain: "invalid domain",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					Build(), cfg
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
			mockDrv := driver.NewMockDriver("nginx", availableDir, enabledDir)

			// Reset JSON output flag
			jsonOutput = false

			// Setup and inject dependencies
			oldDeps := deps
			mockDepsObj, cfg := tt.setupDeps(t, mockDrv)
			deps = mockDepsObj
			defer func() { deps = oldDeps }()

			// Execute
			err := runShow(nil, []string{tt.domain})

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

			// Validate
			if tt.validate != nil {
				tt.validate(t, cfg, mockDrv)
			}
		})
	}
}
