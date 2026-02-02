// Package ssl provides SSL/TLS certificate management through Certbot
// and Let's Encrypt.
//
// The ssl package wraps Certbot commands for obtaining, renewing, and
// managing SSL certificates. It provides a Go-friendly API for Let's Encrypt
// certificate operations without requiring direct Certbot CLI usage.
//
// # Prerequisites
//
// Certbot must be installed on the system:
//
//	# Ubuntu/Debian
//	sudo apt install certbot python3-certbot-nginx
//
//	# CentOS/RHEL
//	sudo dnf install certbot python3-certbot-nginx
//
//	# macOS
//	brew install certbot
//
// # Basic Usage
//
// Check if certbot is installed:
//
//	if !ssl.IsInstalled() {
//	    log.Fatal("certbot is not installed")
//	}
//
// Get certificate paths for a domain:
//
//	paths := ssl.GetCertPaths("example.com")
//	fmt.Println(paths.CertPath) // /etc/letsencrypt/live/example.com/fullchain.pem
//	fmt.Println(paths.KeyPath)  // /etc/letsencrypt/live/example.com/privkey.pem
//
// # Certificate Issuance
//
// Issue a certificate using the webroot method:
//
//	err := ssl.Issue("example.com", "admin@example.com", "/var/www/html")
//
// # Certificate Renewal
//
// Renew a specific certificate:
//
//	err := ssl.Renew("example.com")
//
// Renew all managed certificates:
//
//	err := ssl.RenewAll()
//
// # Certificate Paths
//
// Certificates are stored in Let's Encrypt's standard directory:
//
//	/etc/letsencrypt/live/{domain}/fullchain.pem  (certificate chain)
//	/etc/letsencrypt/live/{domain}/privkey.pem    (private key)
//
// # Testing
//
// The package uses a global executor that can be replaced for testing:
//
//	mockExec := &executor.MockExecutor{}
//	ssl.SetExecutor(mockExec)
//	defer ssl.ResetExecutor()
//
//	// Now all Certbot calls use the mock
//
// # Error Handling
//
// All functions return descriptive errors that include Certbot's output
// when commands fail. Common error scenarios:
//   - Certbot not installed: check with IsInstalled() first
//   - Port 80 in use: stop web server or use webroot method
//   - Rate limiting: Let's Encrypt has strict limits
//   - DNS not configured: ensure domain points to server
package ssl
