// Package platform provides platform-specific path detection for web server configurations.
package platform

import (
	"fmt"
	"os"
	"runtime"
)

// PathConfig contains the paths for a web server driver.
type PathConfig struct {
	Available string
	Enabled   string
}

// PlatformPaths contains the detected paths for all supported web servers.
type PlatformPaths struct {
	Nginx  PathConfig
	Apache PathConfig
	Caddy  PathConfig
}

// DetectPaths returns platform-specific default paths for web servers.
// It checks for common installation locations based on the OS and architecture.
func DetectPaths() (*PlatformPaths, error) {
	switch runtime.GOOS {
	case "darwin":
		return detectDarwinPaths()
	case "linux":
		return detectLinuxPaths()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// detectDarwinPaths detects paths for macOS (Homebrew installations).
func detectDarwinPaths() (*PlatformPaths, error) {
	// Check for Apple Silicon Homebrew path first
	if pathExists("/opt/homebrew") {
		return &PlatformPaths{
			Nginx: PathConfig{
				Available: "/opt/homebrew/etc/nginx/servers",
				Enabled:   "/opt/homebrew/etc/nginx/servers",
			},
			Apache: PathConfig{
				Available: "/opt/homebrew/etc/httpd/extra/vhosts",
				Enabled:   "/opt/homebrew/etc/httpd/extra/vhosts",
			},
			Caddy: PathConfig{
				Available: "/opt/homebrew/etc/caddy/sites-available",
				Enabled:   "/opt/homebrew/etc/caddy/sites-enabled",
			},
		}, nil
	}

	// Check for Intel Homebrew path
	if pathExists("/usr/local") {
		return &PlatformPaths{
			Nginx: PathConfig{
				Available: "/usr/local/etc/nginx/servers",
				Enabled:   "/usr/local/etc/nginx/servers",
			},
			Apache: PathConfig{
				Available: "/usr/local/etc/httpd/extra/vhosts",
				Enabled:   "/usr/local/etc/httpd/extra/vhosts",
			},
			Caddy: PathConfig{
				Available: "/usr/local/etc/caddy/sites-available",
				Enabled:   "/usr/local/etc/caddy/sites-enabled",
			},
		}, nil
	}

	return nil, fmt.Errorf("homebrew installation not found (checked /opt/homebrew and /usr/local)")
}

// detectLinuxPaths detects paths for Linux distributions.
func detectLinuxPaths() (*PlatformPaths, error) {
	// Try Debian/Ubuntu paths first (most common)
	if pathExists("/etc/nginx/sites-available") || pathExists("/etc/nginx") {
		return &PlatformPaths{
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
		}, nil
	}

	// Try RHEL/CentOS paths
	if pathExists("/etc/nginx/conf.d") || pathExists("/etc/httpd") {
		return &PlatformPaths{
			Nginx: PathConfig{
				Available: "/etc/nginx/conf.d",
				Enabled:   "/etc/nginx/conf.d",
			},
			Apache: PathConfig{
				Available: "/etc/httpd/conf.d",
				Enabled:   "/etc/httpd/conf.d",
			},
			Caddy: PathConfig{
				Available: "/etc/caddy/conf.d",
				Enabled:   "/etc/caddy/conf.d",
			},
		}, nil
	}

	return nil, fmt.Errorf("web server configuration paths not found (checked /etc/nginx, /etc/nginx/conf.d, /etc/httpd)")
}

// GetPathsForDriver returns the paths for a specific driver from PlatformPaths.
func (p *PlatformPaths) GetPathsForDriver(driverName string) (PathConfig, error) {
	switch driverName {
	case "nginx":
		return p.Nginx, nil
	case "apache":
		return p.Apache, nil
	case "caddy":
		return p.Caddy, nil
	default:
		return PathConfig{}, fmt.Errorf("unknown driver: %s (available: nginx, apache, caddy)", driverName)
	}
}

// pathExists checks if a path exists on the filesystem.
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Platform returns a string describing the current platform.
func Platform() string {
	return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}
