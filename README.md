# vhost

[![Go Version](https://img.shields.io/badge/Go-1.25.6-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A powerful CLI tool for managing virtual hosts with ease. Supports both **Nginx** and **Apache** web servers. Create, configure, and manage virtual hosts for static sites, PHP applications, Laravel, WordPress, and reverse proxies with a single command.

## Features

- **Multi-Server Support**: Works with both Nginx and Apache web servers
- **Multiple Template Types**: Support for static sites, PHP, Laravel, WordPress, and reverse proxy configurations
- **SSL/TLS Support**: Automatic Let's Encrypt certificate management via Certbot
- **Easy Management**: Add, remove, enable, disable, and list virtual hosts with simple commands
- **Safe Operations**: Built-in configuration testing and automatic rollback on failure
- **Flexible Output**: Human-readable colored output or JSON for scripting
- **Cross-Platform**: Builds for Linux and macOS (amd64/arm64)

## Requirements

- **Web Server**: Nginx or Apache installed and running
- **Root/sudo access** for modifying web server configurations
- **Go 1.25.6+** (for building from source)
- **PHP-FPM** (optional, for PHP/Laravel/WordPress sites)
- **Certbot** (optional, for SSL certificate management)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/ksyq12/vhost.git
cd vhost

# Build
make build

# Install to /usr/local/bin (requires sudo)
make install
```

### Pre-built Binaries

Download the appropriate binary for your platform from the [Releases](https://github.com/ksyq12/vhost/releases) page.

```bash
# Linux (amd64)
curl -LO https://github.com/ksyq12/vhost/releases/latest/download/vhost-linux-amd64
chmod +x vhost-linux-amd64
sudo mv vhost-linux-amd64 /usr/local/bin/vhost

# macOS (Apple Silicon)
curl -LO https://github.com/ksyq12/vhost/releases/latest/download/vhost-darwin-arm64
chmod +x vhost-darwin-arm64
sudo mv vhost-darwin-arm64 /usr/local/bin/vhost
```

## Quick Start

```bash
# Add a static website
sudo vhost add example.com --type static --root /var/www/example

# Add a PHP application
sudo vhost add myapp.com --type php --root /var/www/myapp --php 8.2

# Add a Laravel project
sudo vhost add laravel.test --type laravel --root /var/www/laravel --php 8.2

# Add a WordPress site
sudo vhost add blog.com --type wordpress --root /var/www/wordpress --php 8.2

# Add a reverse proxy (e.g., for Node.js)
sudo vhost add api.example.com --type proxy --proxy http://localhost:3000

# Enable SSL
sudo vhost ssl install example.com --email admin@example.com

# List all virtual hosts
vhost list
```

## Commands

### Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format |

### `vhost add <domain>`

Add a new virtual host.

```bash
vhost add <domain> [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--type` | `-t` | VHost type: `static`, `php`, `proxy`, `laravel`, `wordpress` (default: `static`) |
| `--root` | `-r` | Document root path (required for static, php, laravel, wordpress) |
| `--proxy` | `-p` | Proxy pass URL (required for proxy type) |
| `--php` | | PHP version (e.g., `8.2`) |
| `--ssl` | | Enable SSL (requires certbot) |
| `--no-reload` | | Don't reload Nginx after changes |

**Examples:**

```bash
# Static site
sudo vhost add example.com --type static --root /var/www/html

# PHP site with specific version
sudo vhost add app.com --type php --root /var/www/app --php 8.2

# Laravel with SSL
sudo vhost add laravel.test --type laravel --root /var/www/laravel --php 8.3 --ssl

# Reverse proxy for a Node.js app
sudo vhost add api.test --type proxy --proxy http://localhost:3000
```

### `vhost remove <domain>`

Remove a virtual host.

**Aliases:** `rm`, `delete`

```bash
vhost remove <domain> [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation prompt |
| `--no-reload` | | Don't reload Nginx after changes |

**Examples:**

```bash
# Remove with confirmation
sudo vhost remove example.com

# Force remove without confirmation
sudo vhost rm example.com --force
```

### `vhost enable <domain>`

Enable a disabled virtual host.

```bash
vhost enable <domain> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--no-reload` | Don't reload Nginx after changes |

### `vhost disable <domain>`

Disable a virtual host (keeps configuration).

```bash
vhost disable <domain> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--no-reload` | Don't reload Nginx after changes |

### `vhost list`

List all virtual hosts.

**Aliases:** `ls`

```bash
vhost list [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format |

**Example Output:**

```
DOMAIN              TYPE        ROOT/PROXY              SSL    ENABLED
example.com         static      /var/www/html           no     yes
api.example.com     proxy       http://localhost:3000   yes    yes
laravel.test        laravel     /var/www/laravel        no     no
```

### `vhost ssl install <domain>`

Install an SSL certificate using Let's Encrypt.

```bash
vhost ssl install <domain> --email <email>
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--email` | `-e` | Email for Let's Encrypt notifications (required) |

**Example:**

```bash
sudo vhost ssl install example.com --email admin@example.com
```

### `vhost ssl renew [domain]`

Renew SSL certificate(s).

```bash
vhost ssl renew [domain] [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--all` | Renew all certificates |

**Examples:**

```bash
# Renew specific domain
sudo vhost ssl renew example.com

# Renew all certificates
sudo vhost ssl renew --all
```

### `vhost ssl status`

Show SSL certificate status for all domains.

```bash
vhost ssl status
```

## Template Types

### `static`

For serving static HTML, CSS, and JavaScript files.

- Serves `index.html` and `index.htm`
- Security headers included
- Supports SSL with automatic HTTP to HTTPS redirect

```bash
sudo vhost add example.com --type static --root /var/www/html
```

### `php`

For general PHP applications.

- PHP-FPM socket connection
- Configurable PHP version
- `.htaccess` file access blocked
- URL rewriting support

```bash
sudo vhost add app.com --type php --root /var/www/app --php 8.2
```

### `laravel`

Optimized for Laravel framework.

- Automatic `public` directory routing
- Laravel-style URL rewriting
- `.well-known` directory allowed (for SSL verification)
- Hidden files blocked (except `.well-known`)

```bash
sudo vhost add laravel.test --type laravel --root /var/www/laravel --php 8.2
```

### `wordpress`

Optimized for WordPress CMS.

- WordPress permalink support
- Static file caching (js, css, images)
- Security hardening:
  - `wp-config.php` access blocked
  - `xmlrpc.php` access blocked
- 64MB upload limit
- Optimized FastCGI buffers

```bash
sudo vhost add blog.com --type wordpress --root /var/www/wordpress --php 8.2
```

### `proxy`

For reverse proxying to backend applications.

- WebSocket support
- Proper header forwarding (X-Real-IP, X-Forwarded-For, X-Forwarded-Proto)
- Long timeout for persistent connections (24 hours)

```bash
sudo vhost add api.test --type proxy --proxy http://localhost:3000
```

## SSL Certificate Management

vhost uses Certbot for Let's Encrypt SSL certificate management.

### Installing Certbot

```bash
# Ubuntu/Debian
sudo apt install certbot python3-certbot-nginx

# CentOS/RHEL
sudo dnf install certbot python3-certbot-nginx

# macOS (Homebrew)
brew install certbot
```

### Certificate Paths

- Certificate: `/etc/letsencrypt/live/<domain>/fullchain.pem`
- Private Key: `/etc/letsencrypt/live/<domain>/privkey.pem`

### Automatic Renewal

Set up a cron job or systemd timer to automatically renew certificates:

```bash
# Add to crontab (runs twice daily)
0 0,12 * * * /usr/bin/certbot renew --quiet
```

## Configuration

vhost stores its configuration in `~/.config/vhost/config.yaml`.

### Selecting a Web Server Driver

vhost supports multiple web server drivers. Set the driver in your configuration file:

```yaml
# For Nginx (default)
driver: nginx

# For Apache
driver: apache
```

### Configuration File Structure

```yaml
driver: nginx  # or "apache"
default_php: "8.2"
vhosts:
  example.com:
    domain: example.com
    type: static
    root: /var/www/html
    ssl: false
    enabled: true
    created_at: 2026-02-01T10:00:00Z
  api.example.com:
    domain: api.example.com
    type: proxy
    proxy_pass: http://localhost:3000
    ssl: true
    ssl_cert: /etc/letsencrypt/live/api.example.com/fullchain.pem
    ssl_key: /etc/letsencrypt/live/api.example.com/privkey.pem
    enabled: true
    created_at: 2026-02-01T11:00:00Z
```

### File Locations

#### Nginx

- **Available sites:** `/etc/nginx/sites-available/`
- **Enabled sites:** `/etc/nginx/sites-enabled/` (symlinks)
- **Access logs:** `/var/log/nginx/<domain>-access.log`
- **Error logs:** `/var/log/nginx/<domain>-error.log`

#### Apache

- **Available sites:** `/etc/apache2/sites-available/`
- **Enabled sites:** `/etc/apache2/sites-enabled/` (symlinks)
- **Access logs:** `${APACHE_LOG_DIR}/<domain>-access.log`
- **Error logs:** `${APACHE_LOG_DIR}/<domain>-error.log`

## Development

### Building

```bash
# Build for current platform
make build

# Build for all platforms (Linux/macOS, amd64/arm64)
make build-all
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

### Code Quality

```bash
# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Tidy dependencies
make tidy
```

### Project Structure

```
vhost/
├── cmd/
│   └── vhost/
│       └── main.go              # Entry point
├── internal/
│   ├── cli/                     # CLI commands (Cobra)
│   │   ├── root.go              # Root command
│   │   ├── add.go               # Add command
│   │   ├── remove.go            # Remove command
│   │   ├── enable.go            # Enable command
│   │   ├── disable.go           # Disable command
│   │   ├── list.go              # List command
│   │   ├── ssl.go               # SSL commands
│   │   └── common.go            # Common utilities
│   ├── config/                  # Configuration management
│   │   ├── config.go            # Config struct and methods
│   │   └── vhost.go             # VHost struct
│   ├── driver/                  # Web server drivers
│   │   ├── driver.go            # Driver interface
│   │   ├── nginx.go             # Nginx implementation
│   │   └── apache.go            # Apache implementation
│   ├── template/                # Config templates
│   │   ├── template.go          # Template rendering
│   │   ├── embedded.go          # Embedded templates
│   │   ├── nginx/               # Nginx templates
│   │   │   ├── static.tmpl
│   │   │   ├── php.tmpl
│   │   │   ├── proxy.tmpl
│   │   │   ├── laravel.tmpl
│   │   │   └── wordpress.tmpl
│   │   └── apache/              # Apache templates
│   │       ├── static.tmpl
│   │       ├── php.tmpl
│   │       ├── proxy.tmpl
│   │       ├── laravel.tmpl
│   │       └── wordpress.tmpl
│   ├── ssl/                     # SSL certificate management
│   │   └── certbot.go           # Certbot wrapper
│   └── output/                  # Output formatting
│       └── formatter.go         # Colored/JSON output
├── test/
│   └── integration/             # Integration tests
│       └── docker-compose.yml
├── Makefile
├── go.mod
└── go.sum
```

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Guidelines

- Follow Go best practices and conventions
- Add tests for new features
- Update documentation as needed
- Run `make lint` before submitting

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Made with ❤️ for the web development community