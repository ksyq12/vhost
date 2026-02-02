package platform

import (
	"runtime"
	"testing"
)

func TestDetectPaths(t *testing.T) {
	paths, err := DetectPaths()

	// Test platform-specific behavior
	switch runtime.GOOS {
	case "darwin", "linux":
		if err != nil {
			t.Logf("Detection failed (may be expected if web server not installed): %v", err)
			return
		}

		// Verify we got valid paths
		if paths.Nginx.Available == "" {
			t.Error("nginx available path is empty")
		}
		if paths.Nginx.Enabled == "" {
			t.Error("nginx enabled path is empty")
		}
		if paths.Apache.Available == "" {
			t.Error("apache available path is empty")
		}
		if paths.Caddy.Available == "" {
			t.Error("caddy available path is empty")
		}

	default:
		if err == nil {
			t.Errorf("expected error on unsupported platform %s, but got nil", runtime.GOOS)
		}
	}
}

func TestPathExists(t *testing.T) {
	// Root path should always exist
	if !pathExists("/") {
		t.Error("root path should exist")
	}

	// Non-existent path should return false
	if pathExists("/this/path/should/definitely/not/exist/anywhere") {
		t.Error("non-existent path should return false")
	}
}

func TestPlatformPathsGetPathsForDriver(t *testing.T) {
	paths := &PlatformPaths{
		Nginx: PathConfig{
			Available: "/etc/nginx/sites-available",
			Enabled:   "/etc/nginx/sites-enabled",
		},
		Apache: PathConfig{
			Available: "/etc/apache2/sites-available",
			Enabled:   "/etc/apache2/sites-enabled",
		},
		Caddy: PathConfig{
			Available: "/etc/caddy/sites-available",
			Enabled:   "/etc/caddy/sites-enabled",
		},
	}

	tests := []struct {
		driver    string
		wantAvail string
		wantErr   bool
	}{
		{"nginx", "/etc/nginx/sites-available", false},
		{"apache", "/etc/apache2/sites-available", false},
		{"caddy", "/etc/caddy/sites-available", false},
		{"unknown", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			cfg, err := paths.GetPathsForDriver(tt.driver)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if cfg.Available != tt.wantAvail {
				t.Errorf("expected available=%s, got %s", tt.wantAvail, cfg.Available)
			}
		})
	}
}

func TestPlatform(t *testing.T) {
	p := Platform()
	if p == "" {
		t.Error("Platform() should return non-empty string")
	}

	// Should contain GOOS and GOARCH
	expected := runtime.GOOS + "/" + runtime.GOARCH
	if p != expected {
		t.Errorf("expected %s, got %s", expected, p)
	}
}

func TestDetectDarwinPaths(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping macOS-specific test on non-darwin platform")
	}

	paths, err := detectDarwinPaths()
	if err != nil {
		t.Logf("Darwin detection failed (may be expected): %v", err)
		return
	}

	// On macOS with Homebrew, paths should point to Homebrew directories
	if paths.Nginx.Available == "" {
		t.Error("nginx available path should not be empty on macOS")
	}

	// Check that paths use either Apple Silicon or Intel Homebrew prefix
	if paths.Nginx.Available != "/opt/homebrew/etc/nginx/servers" &&
		paths.Nginx.Available != "/usr/local/etc/nginx/servers" {
		t.Errorf("unexpected nginx path: %s", paths.Nginx.Available)
	}
}

func TestDetectLinuxPaths(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("skipping Linux-specific test on non-linux platform")
	}

	paths, err := detectLinuxPaths()
	if err != nil {
		t.Fatalf("Linux detection should not fail: %v", err)
	}

	// On Linux, should get standard paths
	if paths.Nginx.Available == "" {
		t.Error("nginx available path should not be empty on Linux")
	}
}
