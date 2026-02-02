package driver

import "github.com/ksyq12/vhost/internal/config"

// Driver is the interface that all web server drivers must implement
type Driver interface {
	// Name returns the driver name (nginx, apache)
	Name() string

	// Add creates and enables a vhost config
	Add(vhost *config.VHost, configContent string) error

	// Remove deletes a vhost config
	Remove(domain string) error

	// Enable activates a vhost
	Enable(domain string) error

	// Disable deactivates a vhost
	Disable(domain string) error

	// List returns all vhost domains from the web server config
	List() ([]string, error)

	// IsEnabled checks if a vhost is enabled
	IsEnabled(domain string) (bool, error)

	// Test validates the web server config syntax
	Test() error

	// Reload reloads the web server
	Reload() error

	// Paths returns the driver's config paths
	Paths() Paths
}

// Paths contains the web server config directory paths
type Paths struct {
	Available string // config available directory
	Enabled   string // config enabled directory
}

// registry holds all registered drivers
// Deprecated: The global registry is kept for backward compatibility.
// New code should use NewNginxWithPaths, NewApacheWithPaths, or NewCaddyWithPaths
// to create drivers with platform-specific paths.
var registry = make(map[string]Driver)

// Register adds a driver to the registry
// Deprecated: Use NewXxxWithPaths constructors instead for platform-aware driver creation.
func Register(d Driver) {
	registry[d.Name()] = d
}

// Get returns a driver by name
// Deprecated: Returns a driver with default Linux paths. Use NewXxxWithPaths constructors
// with platform-detected paths for cross-platform compatibility.
func Get(name string) (Driver, bool) {
	d, ok := registry[name]
	return d, ok
}

// Available returns all registered driver names
func Available() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
