package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Driver     string            `yaml:"driver"`
	DefaultPHP string            `yaml:"default_php"`
	VHosts     map[string]*VHost `yaml:"vhosts"`
}

// configDir is the default config directory
const configDir = ".config/vhost"
const configFile = "config.yaml"

// New creates a new Config with default values
func New() *Config {
	return &Config{
		Driver:     "nginx",
		DefaultPHP: "8.2",
		VHosts:     make(map[string]*VHost),
	}
}

// ConfigDir returns the config directory path
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, configDir), nil
}

// ConfigPath returns the config file path
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFile), nil
}

// Load reads the config from disk
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, return default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return New(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := New()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Initialize VHosts map if nil
	if cfg.VHosts == nil {
		cfg.VHosts = make(map[string]*VHost)
	}

	return cfg, nil
}

// Save writes the config to disk
func (c *Config) Save() error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// AddVHost adds a vhost to the config
func (c *Config) AddVHost(vhost *VHost) error {
	if _, exists := c.VHosts[vhost.Domain]; exists {
		return fmt.Errorf("vhost %s already exists", vhost.Domain)
	}
	c.VHosts[vhost.Domain] = vhost
	return nil
}

// GetVHost returns a vhost by domain
func (c *Config) GetVHost(domain string) (*VHost, error) {
	vhost, exists := c.VHosts[domain]
	if !exists {
		return nil, fmt.Errorf("vhost %s not found", domain)
	}
	return vhost, nil
}

// RemoveVHost removes a vhost from the config
func (c *Config) RemoveVHost(domain string) error {
	if _, exists := c.VHosts[domain]; !exists {
		return fmt.Errorf("vhost %s not found", domain)
	}
	delete(c.VHosts, domain)
	return nil
}

// ListVHosts returns all vhosts
func (c *Config) ListVHosts() []*VHost {
	vhosts := make([]*VHost, 0, len(c.VHosts))
	for _, v := range c.VHosts {
		vhosts = append(vhosts, v)
	}
	return vhosts
}
