package cli

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/driver"
)

func TestRunRemove(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		force       bool
		stdinInput  string
		setupDeps   func(*testing.T, *driver.MockDriver) (*Dependencies, *config.Config)
		wantErr     bool
		errContains string
		validate    func(*testing.T, *config.Config, *driver.MockDriver)
	}{
		{
			name:       "remove with force flag",
			domain:     "test.com",
			force:      true,
			stdinInput: "",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				cfg.VHosts["test.com"] = &config.VHost{
					Domain: "test.com",
					Type:   "static",
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build(), cfg
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				if _, exists := cfg.VHosts["test.com"]; exists {
					t.Error("vhost should be removed from config")
				}
				if len(mockDrv.RemoveCalls) != 1 {
					t.Errorf("expected 1 Remove call, got %d", len(mockDrv.RemoveCalls))
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
			name:       "remove with user confirmation yes",
			domain:     "confirm.com",
			force:      false,
			stdinInput: "y\n",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				cfg.VHosts["confirm.com"] = &config.VHost{
					Domain: "confirm.com",
					Type:   "static",
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					WithStdinInput("y\n").
					Build(), cfg
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				if _, exists := cfg.VHosts["confirm.com"]; exists {
					t.Error("vhost should be removed")
				}
				if len(mockDrv.RemoveCalls) != 1 {
					t.Errorf("expected 1 Remove call, got %d", len(mockDrv.RemoveCalls))
				}
			},
		},
		{
			name:       "remove with user confirmation yes (full word)",
			domain:     "yes.com",
			force:      false,
			stdinInput: "yes\n",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				cfg.VHosts["yes.com"] = &config.VHost{
					Domain: "yes.com",
					Type:   "static",
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					WithStdinInput("yes\n").
					Build(), cfg
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				if _, exists := cfg.VHosts["yes.com"]; exists {
					t.Error("vhost should be removed")
				}
			},
		},
		{
			name:       "remove cancelled by user (n)",
			domain:     "cancel.com",
			force:      false,
			stdinInput: "n\n",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				cfg.VHosts["cancel.com"] = &config.VHost{
					Domain: "cancel.com",
					Type:   "static",
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					WithStdinInput("n\n").
					Build(), cfg
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				// Vhost should still exist
				if _, exists := cfg.VHosts["cancel.com"]; !exists {
					t.Error("vhost should NOT be removed when cancelled")
				}
				if len(mockDrv.RemoveCalls) != 0 {
					t.Errorf("Remove should not be called when cancelled, got %d calls", len(mockDrv.RemoveCalls))
				}
			},
		},
		{
			name:       "remove cancelled by user (empty input)",
			domain:     "empty.com",
			force:      false,
			stdinInput: "\n",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				cfg.VHosts["empty.com"] = &config.VHost{
					Domain: "empty.com",
					Type:   "static",
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					WithStdinInput("\n").
					Build(), cfg
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				// Default to no
				if _, exists := cfg.VHosts["empty.com"]; !exists {
					t.Error("vhost should NOT be removed with empty input")
				}
			},
		},
		{
			name:   "remove without root privilege fails",
			domain: "noroot.com",
			force:  true,
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				cfg.VHosts["noroot.com"] = &config.VHost{
					Domain: "noroot.com",
					Type:   "static",
				}
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
			name:   "remove with invalid domain fails",
			domain: "invalid domain",
			force:  true,
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
			name:   "remove driver error",
			domain: "error.com",
			force:  true,
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				mockDrv.RemoveFunc = func(domain string) error {
					return errors.New("driver remove failed")
				}
				cfg := config.New()
				cfg.VHosts["error.com"] = &config.VHost{
					Domain: "error.com",
					Type:   "static",
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build(), cfg
			},
			wantErr:     true,
			errContains: "failed to remove",
		},
		{
			name:   "remove with no-reload flag",
			domain: "noreload.com",
			force:  true,
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				cfg.VHosts["noreload.com"] = &config.VHost{
					Domain: "noreload.com",
					Type:   "static",
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build(), cfg
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				// This test uses noReload = true, set before test
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
			forceRemove = tt.force
			noReload = false
			if tt.name == "remove with no-reload flag" {
				noReload = true
			}

			// Setup and inject dependencies
			oldDeps := deps
			mockDepsObj, cfg := tt.setupDeps(t, mockDrv)
			deps = mockDepsObj
			defer func() { deps = oldDeps }()

			// Execute
			err := runRemove(nil, []string{tt.domain})

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

func TestRunRemoveDryRun(t *testing.T) {
	tests := []struct {
		name      string
		domain    string
		setupDeps func(*testing.T, *driver.MockDriver) (*Dependencies, *config.Config)
		validate  func(*testing.T, *config.Config, *driver.MockDriver)
	}{
		{
			name:   "dry-run remove shows operations without changes",
			domain: "dryrun-remove.com",
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) (*Dependencies, *config.Config) {
				cfg := config.New()
				cfg.VHosts["dryrun-remove.com"] = &config.VHost{
					Domain: "dryrun-remove.com",
					Type:   "static",
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build(), cfg
			},
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				// Verify NO driver operations were called
				if len(mockDrv.RemoveCalls) != 0 {
					t.Errorf("expected 0 Remove calls in dry-run, got %d", len(mockDrv.RemoveCalls))
				}
				if mockDrv.TestCalls != 0 {
					t.Errorf("expected 0 Test calls in dry-run, got %d", mockDrv.TestCalls)
				}
				if mockDrv.ReloadCalls != 0 {
					t.Errorf("expected 0 Reload calls in dry-run, got %d", mockDrv.ReloadCalls)
				}
				// Verify vhost was NOT removed from config
				if _, exists := cfg.VHosts["dryrun-remove.com"]; !exists {
					t.Error("vhost should NOT be removed from config in dry-run mode")
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
			forceRemove = true
			noReload = false
			dryRun = true
			defer func() { dryRun = false }()

			// Setup and inject dependencies
			oldDeps := deps
			mockDepsObj, cfg := tt.setupDeps(t, mockDrv)
			deps = mockDepsObj
			defer func() { deps = oldDeps }()

			// Execute
			err := runRemove(nil, []string{tt.domain})
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
