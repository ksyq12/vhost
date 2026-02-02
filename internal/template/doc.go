// Package template provides rendering of web server configuration files
// from embedded Go templates.
//
// The template package contains pre-built configuration templates for Nginx,
// Apache, and Caddy web servers, covering five different virtual host types.
// Templates are embedded in the binary using go:embed directives.
//
// # Template Organization
//
// Templates are organized by driver and type:
//
//	nginx/static.tmpl
//	nginx/php.tmpl
//	nginx/proxy.tmpl
//	nginx/laravel.tmpl
//	nginx/wordpress.tmpl
//	apache/ (same structure)
//	caddy/ (same structure)
//
// # Rendering Templates
//
// To render a configuration file:
//
//	vhost := &config.VHost{
//	    Domain:     "example.com",
//	    Type:       "static",
//	    Root:       "/var/www/html",
//	    SSL:        true,
//	    SSLCert:    "/etc/letsencrypt/live/example.com/fullchain.pem",
//	    SSLKey:     "/etc/letsencrypt/live/example.com/privkey.pem",
//	}
//
//	content, err := template.Render("nginx", vhost)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Write content to web server config file
//
// # Template Data
//
// Templates receive the VHost struct fields:
//   - Domain: The domain name
//   - Root: Document root path
//   - ProxyPass: Proxy backend URL
//   - PHPVersion: PHP-FPM version
//   - SSL: Whether HTTPS is enabled
//   - SSLCert: Path to certificate
//   - SSLKey: Path to private key
//
// # Custom Functions
//
// Templates have access to these functions:
//   - replace: strings.ReplaceAll for string manipulation
//
// # Adding New Templates
//
// To add a new template:
//  1. Create the .tmpl file in the appropriate driver directory
//  2. Rebuild the binary to embed the new template
//  3. Add the type to config.ValidTypes() if it's a new type
package template
