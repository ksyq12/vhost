package ssl

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Cert represents an SSL certificate
type Cert struct {
	Domain   string
	CertPath string
	KeyPath  string
}

// letsencryptDir is the base directory for Let's Encrypt certificates
const letsencryptDir = "/etc/letsencrypt/live"

// IsInstalled checks if certbot is installed
func IsInstalled() bool {
	_, err := exec.LookPath("certbot")
	return err == nil
}

// runCertbot executes certbot with the given arguments
func runCertbot(args []string) error {
	if !IsInstalled() {
		return fmt.Errorf("certbot is not installed. Install it with: apt install certbot")
	}

	cmd := exec.Command("certbot", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("certbot failed: %s", string(output))
	}
	return nil
}

// GetCertPaths returns the certificate paths for a domain
func GetCertPaths(domain string) *Cert {
	return &Cert{
		Domain:   domain,
		CertPath: filepath.Join(letsencryptDir, domain, "fullchain.pem"),
		KeyPath:  filepath.Join(letsencryptDir, domain, "privkey.pem"),
	}
}

// Issue obtains a new SSL certificate using certbot webroot mode
func Issue(domain, email, webroot string) (*Cert, error) {
	args := []string{
		"certonly",
		"--webroot",
		"-w", webroot,
		"-d", domain,
		"--email", email,
		"--agree-tos",
		"--non-interactive",
	}

	if err := runCertbot(args); err != nil {
		return nil, err
	}

	return GetCertPaths(domain), nil
}

// IssueStandalone obtains a certificate using standalone mode
func IssueStandalone(domain, email string) (*Cert, error) {
	args := []string{
		"certonly",
		"--standalone",
		"-d", domain,
		"--email", email,
		"--agree-tos",
		"--non-interactive",
	}

	if err := runCertbot(args); err != nil {
		return nil, err
	}

	return GetCertPaths(domain), nil
}

// IssueNginx obtains a certificate using nginx plugin
func IssueNginx(domain, email string) (*Cert, error) {
	args := []string{
		"--nginx",
		"-d", domain,
		"--email", email,
		"--agree-tos",
		"--non-interactive",
		"--redirect",
	}

	if err := runCertbot(args); err != nil {
		return nil, err
	}

	return GetCertPaths(domain), nil
}

// Renew renews a specific certificate
func Renew(domain string) error {
	args := []string{
		"renew",
		"--cert-name", domain,
		"--non-interactive",
	}
	return runCertbot(args)
}

// RenewAll renews all certificates
func RenewAll() error {
	return runCertbot([]string{"renew", "--non-interactive"})
}

// Delete removes a certificate
func Delete(domain string) error {
	args := []string{
		"delete",
		"--cert-name", domain,
		"--non-interactive",
	}
	return runCertbot(args)
}

// List returns all managed certificates
func List() ([]string, error) {
	if !IsInstalled() {
		return nil, fmt.Errorf("certbot is not installed")
	}

	cmd := exec.Command("certbot", "certificates")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("certbot certificates failed: %s", string(output))
	}

	// Parse output to extract domain names
	var domains []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Certificate Name:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				domains = append(domains, strings.TrimSpace(parts[1]))
			}
		}
	}

	return domains, nil
}
