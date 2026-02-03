package cli

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/driver"
)

func TestRunAdd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupFlags  func()
		setupDeps   func(*testing.T, *driver.MockDriver) *Dependencies
		wantErr     bool
		errContains string
		validate    func(*testing.T, *config.Config, *driver.MockDriver)
	}{
		{
			name: "add static vhost successfully",
			args: []string{"example.com"},
			setupFlags: func() {
				vhostType = "static"
				vhostRoot = "/var/www/html"
				proxyPass = ""
				phpVersion = ""
				withSSL = false
				noReload = false
			},
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) *Dependencies {
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build()
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				// Verify vhost was added
				if _, exists := cfg.VHosts["example.com"]; !exists {
					t.Error("vhost not added to config")
				}
				// Verify driver calls
				if len(mockDrv.AddCalls) != 1 {
					t.Errorf("expected 1 Add call, got %d", len(mockDrv.AddCalls))
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
			name: "add php vhost with version",
			args: []string{"php.example.com"},
			setupFlags: func() {
				vhostType = "php"
				vhostRoot = "/var/www/php"
				proxyPass = ""
				phpVersion = "8.2"
				withSSL = false
				noReload = false
			},
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) *Dependencies {
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build()
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				vhost := cfg.VHosts["php.example.com"]
				if vhost == nil {
					t.Fatal("vhost not found in config")
				}
				if vhost.PHPVersion != "8.2" {
					t.Errorf("expected PHP 8.2, got %s", vhost.PHPVersion)
				}
				if vhost.Type != "php" {
					t.Errorf("expected type php, got %s", vhost.Type)
				}
			},
		},
		{
			name: "add proxy vhost",
			args: []string{"proxy.example.com"},
			setupFlags: func() {
				vhostType = "proxy"
				vhostRoot = ""
				proxyPass = "http://localhost:3000"
				phpVersion = ""
				withSSL = false
				noReload = false
			},
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) *Dependencies {
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build()
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				vhost := cfg.VHosts["proxy.example.com"]
				if vhost == nil {
					t.Fatal("vhost not found in config")
				}
				if vhost.ProxyPass != "http://localhost:3000" {
					t.Errorf("expected proxy http://localhost:3000, got %s", vhost.ProxyPass)
				}
			},
		},
		{
			name: "add laravel vhost uses default PHP",
			args: []string{"laravel.example.com"},
			setupFlags: func() {
				vhostType = "laravel"
				vhostRoot = "/var/www/laravel"
				proxyPass = ""
				phpVersion = "" // Should use default
				withSSL = false
				noReload = false
			},
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) *Dependencies {
				cfg := config.New()
				cfg.DefaultPHP = "8.1"
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build()
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				vhost := cfg.VHosts["laravel.example.com"]
				if vhost == nil {
					t.Fatal("vhost not found")
				}
				if vhost.PHPVersion != "8.1" {
					t.Errorf("expected PHP 8.1 (default), got %s", vhost.PHPVersion)
				}
			},
		},
		{
			name: "add vhost without root privilege fails",
			args: []string{"test.com"},
			setupFlags: func() {
				vhostType = "static"
				vhostRoot = "/var/www/test"
				proxyPass = ""
				phpVersion = ""
				withSSL = false
				noReload = false
			},
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) *Dependencies {
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(false).
					Build()
			},
			wantErr:     true,
			errContains: "root privileges",
		},
		{
			name: "add duplicate vhost fails",
			args: []string{"existing.com"},
			setupFlags: func() {
				vhostType = "static"
				vhostRoot = "/var/www/existing"
				proxyPass = ""
				phpVersion = ""
				withSSL = false
				noReload = false
			},
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) *Dependencies {
				cfg := config.New()
				cfg.VHosts["existing.com"] = &config.VHost{
					Domain: "existing.com",
					Type:   "static",
				}
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build()
			},
			wantErr:     true,
			errContains: "already exists",
		},
		{
			name: "invalid domain fails validation",
			args: []string{"invalid domain.com"},
			setupFlags: func() {
				vhostType = "static"
				vhostRoot = "/var/www/invalid"
				proxyPass = ""
				phpVersion = ""
				withSSL = false
				noReload = false
			},
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) *Dependencies {
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build()
			},
			wantErr:     true,
			errContains: "spaces",
		},
		{
			name: "missing root for static type fails",
			args: []string{"test.com"},
			setupFlags: func() {
				vhostType = "static"
				vhostRoot = "" // Missing
				proxyPass = ""
				phpVersion = ""
				withSSL = false
				noReload = false
			},
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) *Dependencies {
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build()
			},
			wantErr:     true,
			errContains: "--root is required",
		},
		{
			name: "missing proxy for proxy type fails",
			args: []string{"test.com"},
			setupFlags: func() {
				vhostType = "proxy"
				vhostRoot = ""
				proxyPass = "" // Missing
				phpVersion = ""
				withSSL = false
				noReload = false
			},
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) *Dependencies {
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build()
			},
			wantErr:     true,
			errContains: "--proxy is required",
		},
		{
			name: "invalid vhost type fails",
			args: []string{"test.com"},
			setupFlags: func() {
				vhostType = "invalid_type"
				vhostRoot = "/var/www/test"
				proxyPass = ""
				phpVersion = ""
				withSSL = false
				noReload = false
			},
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) *Dependencies {
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build()
			},
			wantErr:     true,
			errContains: "invalid type",
		},
		{
			name: "no-reload flag skips reload",
			args: []string{"noreload.com"},
			setupFlags: func() {
				vhostType = "static"
				vhostRoot = "/var/www/noreload"
				proxyPass = ""
				phpVersion = ""
				withSSL = false
				noReload = true
			},
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) *Dependencies {
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build()
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				if mockDrv.ReloadCalls != 0 {
					t.Errorf("expected no Reload calls with --no-reload, got %d", mockDrv.ReloadCalls)
				}
				if mockDrv.TestCalls != 1 {
					t.Errorf("Test should still be called, got %d", mockDrv.TestCalls)
				}
			},
		},
		{
			name: "rollback on test failure",
			args: []string{"test-fail.com"},
			setupFlags: func() {
				vhostType = "static"
				vhostRoot = "/var/www/test-fail"
				proxyPass = ""
				phpVersion = ""
				withSSL = false
				noReload = false
			},
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) *Dependencies {
				// Make Test fail
				mockDrv.TestFunc = func() error {
					return errors.New("config test failed")
				}
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build()
			},
			wantErr:     true,
			errContains: "configuration test failed",
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				// Verify rollback was called
				if len(mockDrv.DisableCalls) != 1 {
					t.Errorf("expected 1 Disable call for rollback, got %d", len(mockDrv.DisableCalls))
				}
				if len(mockDrv.RemoveCalls) != 1 {
					t.Errorf("expected 1 Remove call for rollback, got %d", len(mockDrv.RemoveCalls))
				}
			},
		},
		{
			name: "rollback on enable failure",
			args: []string{"enable-fail.com"},
			setupFlags: func() {
				vhostType = "static"
				vhostRoot = "/var/www/enable-fail"
				proxyPass = ""
				phpVersion = ""
				withSSL = false
				noReload = false
			},
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) *Dependencies {
				// Make Enable fail
				mockDrv.EnableFunc = func(domain string) error {
					return errors.New("enable failed")
				}
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build()
			},
			wantErr:     true,
			errContains: "enable",
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				// Verify rollback Remove was called after Enable failure
				if len(mockDrv.RemoveCalls) != 1 {
					t.Errorf("expected 1 Remove call for rollback, got %d", len(mockDrv.RemoveCalls))
				}
			},
		},
		{
			name: "add with ssl flag",
			args: []string{"ssl.example.com"},
			setupFlags: func() {
				vhostType = "static"
				vhostRoot = "/var/www/ssl"
				proxyPass = ""
				phpVersion = ""
				withSSL = true
				noReload = false
			},
			setupDeps: func(t *testing.T, mockDrv *driver.MockDriver) *Dependencies {
				cfg := config.New()
				return NewMockDeps().
					WithConfig(cfg).
					WithDriver(mockDrv).
					WithRootAccess(true).
					Build()
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config, mockDrv *driver.MockDriver) {
				vhost := cfg.VHosts["ssl.example.com"]
				if vhost == nil {
					t.Fatal("vhost not found")
				}
				if !vhost.SSL {
					t.Error("SSL should be enabled")
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
			tt.setupFlags()

			// Setup and inject dependencies
			oldDeps := deps
			mockDepsObj := tt.setupDeps(t, mockDrv)
			deps = mockDepsObj
			defer func() { deps = oldDeps }()

			// Execute
			err := runAdd(nil, tt.args)

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
				cfg, _ := mockDepsObj.ConfigLoader.Load()
				tt.validate(t, cfg, mockDrv)
			}
		})
	}
}

func TestValidateAddOptions(t *testing.T) {
	tests := []struct {
		name        string
		vhostType   string
		root        string
		proxy       string
		wantErr     bool
		errContains string
	}{
		{
			name:      "static with root",
			vhostType: "static",
			root:      "/var/www/html",
			proxy:     "",
			wantErr:   false,
		},
		{
			name:        "static without root",
			vhostType:   "static",
			root:        "",
			proxy:       "",
			wantErr:     true,
			errContains: "--root is required",
		},
		{
			name:      "proxy with proxy url",
			vhostType: "proxy",
			root:      "",
			proxy:     "http://localhost:3000",
			wantErr:   false,
		},
		{
			name:        "proxy without proxy url",
			vhostType:   "proxy",
			root:        "",
			proxy:       "",
			wantErr:     true,
			errContains: "--proxy is required",
		},
		{
			name:      "php with root",
			vhostType: "php",
			root:      "/var/www/php",
			proxy:     "",
			wantErr:   false,
		},
		{
			name:      "laravel with root",
			vhostType: "laravel",
			root:      "/var/www/laravel",
			proxy:     "",
			wantErr:   false,
		},
		{
			name:      "wordpress with root",
			vhostType: "wordpress",
			root:      "/var/www/wordpress",
			proxy:     "",
			wantErr:   false,
		},
		{
			name:        "relative root path fails",
			vhostType:   "static",
			root:        "relative/path",
			proxy:       "",
			wantErr:     true,
			errContains: "absolute",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global flags
			vhostType = tt.vhostType
			vhostRoot = tt.root
			proxyPass = tt.proxy

			err := validateAddOptions()

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
