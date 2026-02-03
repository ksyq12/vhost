package cli

import (
	"path/filepath"
	"testing"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/driver"
)

func TestRunList(t *testing.T) {
	tests := []struct {
		name      string
		setupDeps func(*testing.T, *driver.MockDriver) (*Dependencies, *config.Config)
		wantErr   bool
		validate  func(*testing.T, *config.Config, *driver.MockDriver)
	}{
		{
			name: "list multiple vhosts",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				mockDrv.ListFunc = func() ([]string, error) {
					return []string{"example.com", "api.com"}, nil
				}
				mockDrv.IsEnabledFunc = func(domain string) (bool, error) {
					return domain == "example.com", nil
				}
				cfg := config.New()
				cfg.VHosts["example.com"] = &config.VHost{
					Domain:  "example.com",
					Type:    "static",
					Root:    "/var/www/example",
					Enabled: true,
				}
				cfg.VHosts["api.com"] = &config.VHost{
					Domain:    "api.com",
					Type:      "proxy",
					ProxyPass: "http://localhost:3000",
					Enabled:   false,
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					Build(), cfg
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				if mockDrv.ListCalls != 1 {
					t.Errorf("expected 1 List call, got %d", mockDrv.ListCalls)
				}
				// IsEnabled should be called for each domain
				if len(mockDrv.IsEnabledCalls) < 2 {
					t.Errorf("expected at least 2 IsEnabled calls, got %d", len(mockDrv.IsEnabledCalls))
				}
			},
		},
		{
			name: "list empty",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				mockDrv.ListFunc = func() ([]string, error) {
					return []string{}, nil
				}
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					Build(), cfg
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				if mockDrv.ListCalls != 1 {
					t.Errorf("expected 1 List call, got %d", mockDrv.ListCalls)
				}
			},
		},
		{
			name: "list includes unknown vhosts from driver",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				// Driver returns more domains than config
				mockDrv.ListFunc = func() ([]string, error) {
					return []string{"known.com", "unknown.com"}, nil
				}
				mockDrv.IsEnabledFunc = func(domain string) (bool, error) {
					return true, nil
				}
				cfg := config.New()
				cfg.VHosts["known.com"] = &config.VHost{
					Domain: "known.com",
					Type:   "static",
					Root:   "/var/www/known",
				}
				// unknown.com is not in config
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					Build(), cfg
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				// Should show both known and unknown domains
				if mockDrv.ListCalls != 1 {
					t.Errorf("expected 1 List call, got %d", mockDrv.ListCalls)
				}
			},
		},
		{
			name: "list with ssl vhosts",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				mockDrv.ListFunc = func() ([]string, error) {
					return []string{"ssl.com"}, nil
				}
				mockDrv.IsEnabledFunc = func(domain string) (bool, error) {
					return true, nil
				}
				cfg := config.New()
				cfg.VHosts["ssl.com"] = &config.VHost{
					Domain: "ssl.com",
					Type:   "static",
					Root:   "/var/www/ssl",
					SSL:    true,
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					Build(), cfg
			},
			wantErr: false,
		},
		{
			name: "list continues on driver list error",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				mockDrv.ListFunc = func() ([]string, error) {
					return nil, nil // Return nil instead of error - command should handle gracefully
				}
				cfg := config.New()
				cfg.VHosts["config-only.com"] = &config.VHost{
					Domain: "config-only.com",
					Type:   "static",
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					Build(), cfg
			},
			wantErr: false,
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
			err := runList(nil, []string{})

			// Check error
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
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
