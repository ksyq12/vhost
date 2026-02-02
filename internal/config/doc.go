// Package config manages the vhost application configuration and virtual host
// definitions stored in YAML format.
//
// The config package provides persistent storage for virtual host configurations,
// tracking domains, types, SSL settings, and metadata. Configuration is stored
// in the user's home directory at ~/.config/vhost/config.yaml.
//
// # Configuration Structure
//
// The main Config struct contains:
//   - Driver selection (nginx, apache, caddy)
//   - Default PHP version for PHP-based vhosts
//   - Custom path overrides for the web server
//   - Map of all managed virtual hosts
//
// Example config.yaml:
//
//	driver: nginx
//	default_php: "8.2"
//	paths:
//	  available: /opt/nginx/sites-available
//	  enabled: /opt/nginx/sites-enabled
//	vhosts:
//	  example.com:
//	    domain: example.com
//	    type: static
//	    root: /var/www/html
//	    ssl: false
//	    enabled: true
//	    created_at: 2026-02-01T10:00:00Z
//
// # Virtual Host Types
//
// The package defines five virtual host types:
//   - static: Static HTML/CSS/JS sites
//   - php: Generic PHP applications
//   - laravel: Laravel framework (routes through public/)
//   - wordpress: WordPress CMS with optimizations
//   - proxy: Reverse proxy to backend services
//
// Use the type constants (TypeStatic, TypePHP, etc.) instead of string literals.
//
// # Usage
//
// Loading and modifying configuration:
//
//	// Load configuration (creates default if missing)
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Add a new virtual host
//	vhost := &config.VHost{
//	    Domain:    "example.com",
//	    Type:      config.TypeStatic,
//	    Root:      "/var/www/html",
//	    SSL:       false,
//	    Enabled:   true,
//	    CreatedAt: time.Now(),
//	}
//	err = cfg.AddVHost(vhost)
//
//	// Save changes to disk
//	err = cfg.Save()
//
// # Thread Safety
//
// Config operations are NOT thread-safe. Callers must implement their own
// synchronization if accessing Config from multiple goroutines.
package config
