package cli

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/driver"
)

func TestRunEnable(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		noReload    bool
		setupDeps   func(*testing.T, *driver.MockDriver) (*Dependencies, *config.Config)
		wantErr     bool
		errContains string
		validate    func(*testing.T, *config.Config, *driver.MockDriver)
	}{
		{
			name:     "enable vhost successfully",
			domain:   "test.com",
			noReload: false,
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				cfg.VHosts["test.com"] = &config.VHost{
					Domain:  "test.com",
					Type:    "static",
					Enabled: false,
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build(), cfg
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				vhost := cfg.VHosts["test.com"]
				if vhost == nil {
					t.Fatal("vhost not found")
				}
				if !vhost.Enabled {
					t.Error("vhost should be enabled in config")
				}
				if len(mockDrv.EnableCalls) != 1 {
					t.Errorf("expected 1 Enable call, got %d", len(mockDrv.EnableCalls))
				}
				if mockDrv.TestCalls != 1 {
					t.Errorf("expected 1 Test call, got %d", mockDrv.TestCalls)
				}
				if mockDrv.ReloadCalls != 1 {
					t.Errorf("expected 1 Reload call, got %d", mockDrv.ReloadCalls)
				}
			},
		},
		{
			name:     "enable with no-reload flag",
			domain:   "noreload.com",
			noReload: true,
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				cfg.VHosts["noreload.com"] = &config.VHost{
					Domain:  "noreload.com",
					Type:    "static",
					Enabled: false,
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build(), cfg
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				if mockDrv.ReloadCalls != 0 {
					t.Errorf("expected 0 Reload calls with no-reload, got %d", mockDrv.ReloadCalls)
				}
				if mockDrv.TestCalls != 1 {
					t.Errorf("Test should still be called, got %d", mockDrv.TestCalls)
				}
			},
		},
		{
			name:     "enable without root privilege fails",
			domain:   "noroot.com",
			noReload: false,
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(false).
					Build(), cfg
			},
			wantErr:     true,
			errContains: "root privileges",
		},
		{
			name:     "invalid domain fails",
			domain:   "invalid domain",
			noReload: false,
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build(), cfg
			},
			wantErr:     true,
			errContains: "spaces",
		},
		{
			name:     "enable driver error",
			domain:   "error.com",
			noReload: false,
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				mockDrv.EnableFunc = func(domain string) error {
					return errors.New("enable failed")
				}
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build(), cfg
			},
			wantErr:     true,
			errContains: "failed to enable",
		},
		{
			name:     "rollback on test failure",
			domain:   "rollback.com",
			noReload: false,
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				mockDrv.TestFunc = func() error {
					return errors.New("test failed")
				}
				cfg := config.New()
				cfg.VHosts["rollback.com"] = &config.VHost{
					Domain:  "rollback.com",
					Type:    "static",
					Enabled: false,
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build(), cfg
			},
			wantErr:     true,
			errContains: "configuration test failed",
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				// Verify rollback was called
				if len(mockDrv.DisableCalls) != 1 {
					t.Errorf("expected 1 Disable call for rollback, got %d", len(mockDrv.DisableCalls))
				}
			},
		},
		{
			name:     "enable vhost not in config still works",
			domain:   "notinconfig.com",
			noReload: false,
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				// Not adding vhost to config
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build(), cfg
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				// Enable should still be called
				if len(mockDrv.EnableCalls) != 1 {
					t.Errorf("expected 1 Enable call, got %d", len(mockDrv.EnableCalls))
				}
			},
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

			// Setup flag
			noReload = tt.noReload

			// Setup and inject dependencies
			oldDeps := deps
			mockDepsObj, cfg := tt.setupDeps(t, mockDrv)
			deps = mockDepsObj
			defer func() { deps = oldDeps }()

			// Execute
			err := runEnable(nil, []string{tt.domain})

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

func TestRunEnableDryRun(t *testing.T) {
	tests := []struct {
		name      string
		domain    string
		setupDeps func(*testing.T, *driver.MockDriver) (*Dependencies, *config.Config)
		validate  func(*testing.T, *config.Config, *driver.MockDriver)
	}{
		{
			name:   "dry-run enable shows operations without changes",
			domain: "dryrun-enable.com",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				cfg.VHosts["dryrun-enable.com"] = &config.VHost{
					Domain:  "dryrun-enable.com",
					Type:    "static",
					Enabled: false,
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build(), cfg
			},
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				// Verify NO driver operations were called
				if len(mockDrv.EnableCalls) != 0 {
					t.Errorf("expected 0 Enable calls in dry-run, got %d", len(mockDrv.EnableCalls))
				}
				if mockDrv.TestCalls != 0 {
					t.Errorf("expected 0 Test calls in dry-run, got %d", mockDrv.TestCalls)
				}
				if mockDrv.ReloadCalls != 0 {
					t.Errorf("expected 0 Reload calls in dry-run, got %d", mockDrv.ReloadCalls)
				}
				// Verify vhost.Enabled was NOT changed
				vhost := cfg.VHosts["dryrun-enable.com"]
				if vhost != nil && vhost.Enabled {
					t.Error("vhost.Enabled should NOT be changed in dry-run mode")
				}
			},
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
			noReload = false
			dryRun = true
			defer func() { dryRun = false }()

			// Setup and inject dependencies
			oldDeps := deps
			mockDepsObj, cfg := tt.setupDeps(t, mockDrv)
			deps = mockDepsObj
			defer func() { deps = oldDeps }()

			// Execute
			err := runEnable(nil, []string{tt.domain})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Validate
			if tt.validate != nil {
				tt.validate(t, cfg, mockDrv)
			}
		})
	}
}
