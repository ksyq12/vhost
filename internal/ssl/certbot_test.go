package ssl

import (
	"errors"
	"testing"

	"github.com/ksyq12/vhost/internal/executor"
)

func TestIsInstalled(t *testing.T) {
	t.Run("certbot installed", func(t *testing.T) {
		mock := &executor.MockExecutor{
			LookPathFunc: func(file string) (string, error) {
				if file == "certbot" {
					return "/usr/bin/certbot", nil
				}
				return "", errors.New("not found")
			},
		}
		SetExecutor(mock)
		defer ResetExecutor()

		if !IsInstalled() {
			t.Error("IsInstalled should return true")
		}
	})

	t.Run("certbot not installed", func(t *testing.T) {
		mock := &executor.MockExecutor{
			LookPathFunc: func(file string) (string, error) {
				return "", errors.New("not found")
			},
		}
		SetExecutor(mock)
		defer ResetExecutor()

		if IsInstalled() {
			t.Error("IsInstalled should return false")
		}
	})
}

func TestGetCertPaths(t *testing.T) {
	cert := GetCertPaths("example.com")

	if cert.Domain != "example.com" {
		t.Errorf("expected domain example.com, got %s", cert.Domain)
	}
	if cert.CertPath != "/etc/letsencrypt/live/example.com/fullchain.pem" {
		t.Errorf("unexpected cert path: %s", cert.CertPath)
	}
	if cert.KeyPath != "/etc/letsencrypt/live/example.com/privkey.pem" {
		t.Errorf("unexpected key path: %s", cert.KeyPath)
	}
}

func TestIssue(t *testing.T) {
	t.Run("successful issue", func(t *testing.T) {
		mock := &executor.MockExecutor{
			LookPathFunc: func(file string) (string, error) {
				return "/usr/bin/" + file, nil
			},
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				if name == "certbot" {
					// Verify correct arguments
					hasWebroot := false
					for _, arg := range args {
						if arg == "--webroot" {
							hasWebroot = true
						}
					}
					if !hasWebroot {
						return nil, errors.New("expected --webroot flag")
					}
					return []byte("Successfully received certificate"), nil
				}
				return nil, errors.New("unexpected command")
			},
		}
		SetExecutor(mock)
		defer ResetExecutor()

		cert, err := Issue("example.com", "admin@example.com", "/var/www/html")
		if err != nil {
			t.Fatalf("Issue failed: %v", err)
		}
		if cert.Domain != "example.com" {
			t.Errorf("expected domain example.com, got %s", cert.Domain)
		}
	})

	t.Run("certbot not installed", func(t *testing.T) {
		mock := &executor.MockExecutor{
			LookPathFunc: func(file string) (string, error) {
				return "", errors.New("not found")
			},
		}
		SetExecutor(mock)
		defer ResetExecutor()

		_, err := Issue("example.com", "admin@example.com", "/var/www/html")
		if err == nil {
			t.Error("Issue should fail when certbot not installed")
		}
	})

	t.Run("certbot execution fails", func(t *testing.T) {
		mock := &executor.MockExecutor{
			LookPathFunc: func(file string) (string, error) {
				return "/usr/bin/" + file, nil
			},
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				return []byte("Rate limit exceeded"), errors.New("exit status 1")
			},
		}
		SetExecutor(mock)
		defer ResetExecutor()

		_, err := Issue("example.com", "admin@example.com", "/var/www/html")
		if err == nil {
			t.Error("Issue should fail when certbot fails")
		}
	})
}

func TestIssueStandalone(t *testing.T) {
	t.Run("successful issue", func(t *testing.T) {
		mock := &executor.MockExecutor{
			LookPathFunc: func(file string) (string, error) {
				return "/usr/bin/" + file, nil
			},
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				if name == "certbot" {
					hasStandalone := false
					for _, arg := range args {
						if arg == "--standalone" {
							hasStandalone = true
						}
					}
					if !hasStandalone {
						return nil, errors.New("expected --standalone flag")
					}
					return []byte("Success"), nil
				}
				return nil, errors.New("unexpected command")
			},
		}
		SetExecutor(mock)
		defer ResetExecutor()

		cert, err := IssueStandalone("example.com", "admin@example.com")
		if err != nil {
			t.Fatalf("IssueStandalone failed: %v", err)
		}
		if cert.Domain != "example.com" {
			t.Errorf("expected domain example.com, got %s", cert.Domain)
		}
	})
}

func TestIssueNginx(t *testing.T) {
	t.Run("successful issue", func(t *testing.T) {
		mock := &executor.MockExecutor{
			LookPathFunc: func(file string) (string, error) {
				return "/usr/bin/" + file, nil
			},
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				if name == "certbot" {
					hasNginx := false
					for _, arg := range args {
						if arg == "--nginx" {
							hasNginx = true
						}
					}
					if !hasNginx {
						return nil, errors.New("expected --nginx flag")
					}
					return []byte("Success"), nil
				}
				return nil, errors.New("unexpected command")
			},
		}
		SetExecutor(mock)
		defer ResetExecutor()

		cert, err := IssueNginx("example.com", "admin@example.com")
		if err != nil {
			t.Fatalf("IssueNginx failed: %v", err)
		}
		if cert.Domain != "example.com" {
			t.Errorf("expected domain example.com, got %s", cert.Domain)
		}
	})
}

func TestRenew(t *testing.T) {
	t.Run("successful renew", func(t *testing.T) {
		mock := &executor.MockExecutor{
			LookPathFunc: func(file string) (string, error) {
				return "/usr/bin/" + file, nil
			},
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				if name == "certbot" {
					return []byte("Certificate renewed"), nil
				}
				return nil, errors.New("unexpected command")
			},
		}
		SetExecutor(mock)
		defer ResetExecutor()

		err := Renew("example.com")
		if err != nil {
			t.Fatalf("Renew failed: %v", err)
		}
	})

	t.Run("renew fails", func(t *testing.T) {
		mock := &executor.MockExecutor{
			LookPathFunc: func(file string) (string, error) {
				return "/usr/bin/" + file, nil
			},
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				return []byte("Renewal failed"), errors.New("exit status 1")
			},
		}
		SetExecutor(mock)
		defer ResetExecutor()

		err := Renew("example.com")
		if err == nil {
			t.Error("Renew should fail when certbot fails")
		}
	})
}

func TestRenewAll(t *testing.T) {
	t.Run("successful renew all", func(t *testing.T) {
		mock := &executor.MockExecutor{
			LookPathFunc: func(file string) (string, error) {
				return "/usr/bin/" + file, nil
			},
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				if name == "certbot" {
					return []byte("All certificates renewed"), nil
				}
				return nil, errors.New("unexpected command")
			},
		}
		SetExecutor(mock)
		defer ResetExecutor()

		err := RenewAll()
		if err != nil {
			t.Fatalf("RenewAll failed: %v", err)
		}
	})
}

func TestDelete(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		mock := &executor.MockExecutor{
			LookPathFunc: func(file string) (string, error) {
				return "/usr/bin/" + file, nil
			},
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				if name == "certbot" {
					return []byte("Certificate deleted"), nil
				}
				return nil, errors.New("unexpected command")
			},
		}
		SetExecutor(mock)
		defer ResetExecutor()

		err := Delete("example.com")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
	})

	t.Run("delete fails", func(t *testing.T) {
		mock := &executor.MockExecutor{
			LookPathFunc: func(file string) (string, error) {
				return "/usr/bin/" + file, nil
			},
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				return []byte("Delete failed"), errors.New("exit status 1")
			},
		}
		SetExecutor(mock)
		defer ResetExecutor()

		err := Delete("example.com")
		if err == nil {
			t.Error("Delete should fail when certbot fails")
		}
	})
}

func TestList(t *testing.T) {
	t.Run("list certificates", func(t *testing.T) {
		mock := &executor.MockExecutor{
			LookPathFunc: func(file string) (string, error) {
				return "/usr/bin/" + file, nil
			},
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				if name == "certbot" {
					output := `Found the following certificates:
  Certificate Name: example.com
    Domains: example.com www.example.com
    Expiry Date: 2024-05-15 (VALID: 89 days)
  Certificate Name: test.com
    Domains: test.com
    Expiry Date: 2024-04-20 (VALID: 64 days)`
					return []byte(output), nil
				}
				return nil, errors.New("unexpected command")
			},
		}
		SetExecutor(mock)
		defer ResetExecutor()

		domains, err := List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(domains) != 2 {
			t.Errorf("expected 2 domains, got %d", len(domains))
		}
		if domains[0] != "example.com" {
			t.Errorf("expected example.com, got %s", domains[0])
		}
		if domains[1] != "test.com" {
			t.Errorf("expected test.com, got %s", domains[1])
		}
	})

	t.Run("certbot not installed", func(t *testing.T) {
		mock := &executor.MockExecutor{
			LookPathFunc: func(file string) (string, error) {
				return "", errors.New("not found")
			},
		}
		SetExecutor(mock)
		defer ResetExecutor()

		_, err := List()
		if err == nil {
			t.Error("List should fail when certbot not installed")
		}
	})

	t.Run("no certificates", func(t *testing.T) {
		mock := &executor.MockExecutor{
			LookPathFunc: func(file string) (string, error) {
				return "/usr/bin/" + file, nil
			},
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				return []byte("No certificates found."), nil
			},
		}
		SetExecutor(mock)
		defer ResetExecutor()

		domains, err := List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(domains) != 0 {
			t.Errorf("expected 0 domains, got %d", len(domains))
		}
	})

	t.Run("list fails", func(t *testing.T) {
		mock := &executor.MockExecutor{
			LookPathFunc: func(file string) (string, error) {
				return "/usr/bin/" + file, nil
			},
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				return []byte("Error reading certificates"), errors.New("exit status 1")
			},
		}
		SetExecutor(mock)
		defer ResetExecutor()

		_, err := List()
		if err == nil {
			t.Error("List should fail when certbot fails")
		}
	})
}
