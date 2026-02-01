package config

import "time"

// VHost represents a virtual host configuration
type VHost struct {
	Domain     string            `yaml:"domain"`
	Type       string            `yaml:"type"` // static, php, proxy, laravel, wordpress
	Root       string            `yaml:"root,omitempty"`
	ProxyPass  string            `yaml:"proxy_pass,omitempty"`
	PHPVersion string            `yaml:"php_version,omitempty"`
	SSL        bool              `yaml:"ssl"`
	SSLCert    string            `yaml:"ssl_cert,omitempty"`
	SSLKey     string            `yaml:"ssl_key,omitempty"`
	Enabled    bool              `yaml:"enabled"`
	Extra      map[string]string `yaml:"extra,omitempty"`
	CreatedAt  time.Time         `yaml:"created_at"`
}

// VHostType constants
const (
	TypeStatic    = "static"
	TypePHP       = "php"
	TypeProxy     = "proxy"
	TypeLaravel   = "laravel"
	TypeWordPress = "wordpress"
)

// ValidTypes returns all valid vhost types
func ValidTypes() []string {
	return []string{TypeStatic, TypePHP, TypeProxy, TypeLaravel, TypeWordPress}
}

// IsValidType checks if the given type is valid
func IsValidType(t string) bool {
	for _, valid := range ValidTypes() {
		if t == valid {
			return true
		}
	}
	return false
}
