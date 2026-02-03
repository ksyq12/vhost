package cli

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/driver"
)

func TestRunDisable(t *testing.T) {
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
			name:     "disable vhost successfully",
			domain:   "test.com",
			noReload: false,
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				cfg.VHosts["test.com"] = &config.VHost{
					Domain:  "test.com",
					Type:    "static",
					Enabled: true,
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
				if vhost.Enabled {
					t.Error("vhost should be disabled in config")
				}
				if len(mockDrv.DisableCalls) != 1 {
					t.Errorf("expected 1 Disable call, got %d", len(mockDrv.DisableCalls))
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
			name:     "disable with no-reload flag",
			domain:   "noreload.com",
			noReload: true,
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				cfg.VHosts["noreload.com"] = &config.VHost{
					Domain:  "noreload.com",
					Type:    "static",
					Enabled: true,
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
			name:     "disable without root privilege fails",
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
			name:     "disable driver error",
			domain:   "error.com",
			noReload: false,
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				mockDrv.DisableFunc = func(domain string) error {
					return errors.New("disable failed")
				}
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build(), cfg
			},
			wantErr:     true,
			errContains: "failed to disable",
		},
		{
			name:     "disable continues on test failure (no rollback)",
			domain:   "testfail.com",
			noReload: false,
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				mockDrv.TestFunc = func() error {
					return errors.New("test failed")
				}
				cfg := config.New()
				cfg.VHosts["testfail.com"] = &config.VHost{
					Domain:  "testfail.com",
					Type:    "static",
					Enabled: true,
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build(), cfg
			},
			wantErr: false, // Disable continues even on test failure
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				// Disable should have been called
				if len(mockDrv.DisableCalls) != 1 {
					t.Errorf("expected 1 Disable call, got %d", len(mockDrv.DisableCalls))
				}
				// No rollback for disable (unlike enable)
				if len(mockDrv.EnableCalls) != 0 {
					t.Errorf("expected 0 Enable calls (no rollback), got %d", len(mockDrv.EnableCalls))
				}
			},
		},
		{
			name:     "disable vhost not in config still works",
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
				// Disable should still be called
				if len(mockDrv.DisableCalls) != 1 {
					t.Errorf("expected 1 Disable call, got %d", len(mockDrv.DisableCalls))
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
			err := runDisable(nil, []string{tt.domain})

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
