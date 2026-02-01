package template

import (
	"strings"
	"testing"

	"github.com/ksyq12/vhost/internal/config"
)

func TestRender(t *testing.T) {
	testCases := []struct {
		name     string
		vhost    *config.VHost
		contains []string
	}{
		{
			name: "static",
			vhost: &config.VHost{
				Domain: "static.example.com",
				Type:   config.TypeStatic,
				Root:   "/var/www/static",
			},
			contains: []string{
				"server_name static.example.com",
				"root /var/www/static",
				"index index.html",
			},
		},
		{
			name: "php",
			vhost: &config.VHost{
				Domain:     "php.example.com",
				Type:       config.TypePHP,
				Root:       "/var/www/php",
				PHPVersion: "8.2",
			},
			contains: []string{
				"server_name php.example.com",
				"root /var/www/php",
				"fastcgi_pass unix:/run/php/php8.2-fpm.sock",
				"index index.php",
			},
		},
		{
			name: "proxy",
			vhost: &config.VHost{
				Domain:    "proxy.example.com",
				Type:      config.TypeProxy,
				ProxyPass: "127.0.0.1:3000",
			},
			contains: []string{
				"server_name proxy.example.com",
				"proxy_pass",
				"proxy_set_header",
			},
		},
		{
			name: "laravel",
			vhost: &config.VHost{
				Domain:     "laravel.example.com",
				Type:       config.TypeLaravel,
				Root:       "/var/www/laravel",
				PHPVersion: "8.2",
			},
			contains: []string{
				"server_name laravel.example.com",
				"root /var/www/laravel/public",
				"fastcgi_pass unix:/run/php/php8.2-fpm.sock",
				"try_files $uri $uri/ /index.php",
			},
		},
		{
			name: "wordpress",
			vhost: &config.VHost{
				Domain:     "wp.example.com",
				Type:       config.TypeWordPress,
				Root:       "/var/www/wordpress",
				PHPVersion: "8.2",
			},
			contains: []string{
				"server_name wp.example.com",
				"root /var/www/wordpress",
				"fastcgi_pass unix:/run/php/php8.2-fpm.sock",
				"wp-config.php",
			},
		},
		{
			name: "with SSL",
			vhost: &config.VHost{
				Domain:  "ssl.example.com",
				Type:    config.TypeStatic,
				Root:    "/var/www/ssl",
				SSL:     true,
				SSLCert: "/etc/ssl/cert.pem",
				SSLKey:  "/etc/ssl/key.pem",
			},
			contains: []string{
				"listen 443 ssl",
				"ssl_certificate /etc/ssl/cert.pem",
				"ssl_certificate_key /etc/ssl/key.pem",
				"return 301 https://",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Render("nginx", tc.vhost)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}

			for _, expected := range tc.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected output to contain %q", expected)
				}
			}
		})
	}
}

func TestRenderInvalidType(t *testing.T) {
	vhost := &config.VHost{
		Domain: "invalid.example.com",
		Type:   "nonexistent",
	}

	_, err := Render("nginx", vhost)
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestRenderInvalidDriver(t *testing.T) {
	vhost := &config.VHost{
		Domain: "test.example.com",
		Type:   config.TypeStatic,
	}

	_, err := Render("nonexistent", vhost)
	if err == nil {
		t.Error("expected error for invalid driver")
	}
}

func TestDefaultPHPVersion(t *testing.T) {
	vhost := &config.VHost{
		Domain: "php.example.com",
		Type:   config.TypePHP,
		Root:   "/var/www/php",
		// PHPVersion not set
	}

	result, err := Render("nginx", vhost)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Should use default PHP version 8.2
	if !strings.Contains(result, "php8.2-fpm.sock") {
		t.Error("expected default PHP version 8.2 in output")
	}
}
