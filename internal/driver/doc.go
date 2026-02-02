// Package driver provides abstractions for managing virtual host configurations
// across different web servers (Nginx, Apache, Caddy).
//
// The driver package implements a unified interface for web server operations,
// allowing the vhost tool to support multiple web server backends without
// code duplication.
//
// # Supported Web Servers
//
//   - Nginx: Standard sites-available/sites-enabled pattern
//   - Apache: .conf extension with symlink activation
//   - Caddy: Caddyfile-based configuration
//
// # Basic Usage
//
// Create a driver instance with platform-specific paths:
//
//	import "github.com/ksyq12/vhost/internal/driver"
//
//	// Create Nginx driver with custom paths
//	drv := driver.NewNginxWithPaths(
//	    "/etc/nginx/sites-available",
//	    "/etc/nginx/sites-enabled",
//	)
//
//	// Add a virtual host
//	vhost := &config.VHost{
//	    Domain: "example.com",
//	    Type:   "static",
//	    Root:   "/var/www/html",
//	}
//	err := drv.Add(vhost, configContent)
//
// # Driver Selection
//
// Use the NewXxxWithPaths constructors for platform-aware driver creation:
//
//	// Nginx
//	drv := driver.NewNginxWithPaths(availablePath, enabledPath)
//
//	// Apache
//	drv := driver.NewApacheWithPaths(availablePath, enabledPath)
//
//	// Caddy
//	drv := driver.NewCaddyWithPaths(availablePath, enabledPath)
//
// # Testing
//
// Each driver implementation provides a WithExecutor constructor that accepts
// a mock executor.CommandExecutor for testing without actual system calls:
//
//	mockExec := &executor.MockExecutor{}
//	drv := driver.NewNginxWithExecutor(availablePath, enabledPath, mockExec)
//
// # Error Handling
//
// All driver methods return descriptive errors that include context about
// the operation that failed. Errors are wrapped using fmt.Errorf with %w
// to maintain the error chain.
package driver
