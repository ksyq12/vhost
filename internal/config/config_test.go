package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	// Create temp directory for test config
	tempDir := t.TempDir()

	// Override config path for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create the .config/vhost directory
	configDir := filepath.Join(tempDir, ".config", "vhost")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	t.Run("New", func(t *testing.T) {
		cfg := New()
		if cfg.Driver != "nginx" {
			t.Errorf("expected nginx driver, got %s", cfg.Driver)
		}
		if cfg.DefaultPHP != "8.2" {
			t.Errorf("expected 8.2 PHP, got %s", cfg.DefaultPHP)
		}
		if cfg.VHosts == nil {
			t.Error("VHosts should be initialized")
		}
	})

	t.Run("LoadNonexistent", func(t *testing.T) {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		// Should return default config when file doesn't exist
		if cfg.Driver != "nginx" {
			t.Errorf("expected nginx driver, got %s", cfg.Driver)
		}
	})

	t.Run("SaveAndLoad", func(t *testing.T) {
		cfg := New()
		cfg.VHosts["test.example.com"] = &VHost{
			Domain:    "test.example.com",
			Type:      TypeStatic,
			Root:      "/var/www/test",
			SSL:       true,
			Enabled:   true,
			CreatedAt: time.Now(),
		}

		if err := cfg.Save(); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Verify file exists
		loadedPath := filepath.Join(configDir, "config.yaml")
		if _, err := os.Stat(loadedPath); os.IsNotExist(err) {
			t.Error("config file was not created")
		}

		// Load and verify
		loaded, err := Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if loaded.Driver != "nginx" {
			t.Errorf("expected nginx driver, got %s", loaded.Driver)
		}

		vhost, exists := loaded.VHosts["test.example.com"]
		if !exists {
			t.Fatal("vhost not found")
		}
		if vhost.Domain != "test.example.com" {
			t.Errorf("expected test.example.com, got %s", vhost.Domain)
		}
		if vhost.Type != TypeStatic {
			t.Errorf("expected static, got %s", vhost.Type)
		}
		if !vhost.SSL {
			t.Error("expected SSL to be true")
		}
	})

	t.Run("AddVHost", func(t *testing.T) {
		cfg := New()
		vhost := &VHost{
			Domain: "new.example.com",
			Type:   TypePHP,
		}

		if err := cfg.AddVHost(vhost); err != nil {
			t.Fatalf("AddVHost failed: %v", err)
		}

		// Adding again should fail
		err := cfg.AddVHost(vhost)
		if err == nil {
			t.Error("expected error when adding duplicate vhost")
		}
	})

	t.Run("GetVHost", func(t *testing.T) {
		cfg := New()
		cfg.VHosts["get.example.com"] = &VHost{Domain: "get.example.com"}

		vhost, err := cfg.GetVHost("get.example.com")
		if err != nil {
			t.Fatalf("GetVHost failed: %v", err)
		}
		if vhost.Domain != "get.example.com" {
			t.Errorf("expected get.example.com, got %s", vhost.Domain)
		}

		_, err = cfg.GetVHost("nonexistent.example.com")
		if err == nil {
			t.Error("expected error for nonexistent vhost")
		}
	})

	t.Run("RemoveVHost", func(t *testing.T) {
		cfg := New()
		cfg.VHosts["remove.example.com"] = &VHost{Domain: "remove.example.com"}

		if err := cfg.RemoveVHost("remove.example.com"); err != nil {
			t.Fatalf("RemoveVHost failed: %v", err)
		}

		if _, exists := cfg.VHosts["remove.example.com"]; exists {
			t.Error("vhost should have been removed")
		}

		err := cfg.RemoveVHost("nonexistent.example.com")
		if err == nil {
			t.Error("expected error for nonexistent vhost")
		}
	})

	t.Run("ListVHosts", func(t *testing.T) {
		cfg := New()
		cfg.VHosts["a.example.com"] = &VHost{Domain: "a.example.com"}
		cfg.VHosts["b.example.com"] = &VHost{Domain: "b.example.com"}

		list := cfg.ListVHosts()
		if len(list) != 2 {
			t.Errorf("expected 2 vhosts, got %d", len(list))
		}
	})
}

func TestVHostTypes(t *testing.T) {
	t.Run("ValidTypes", func(t *testing.T) {
		types := ValidTypes()
		if len(types) != 5 {
			t.Errorf("expected 5 types, got %d", len(types))
		}
	})

	t.Run("IsValidType", func(t *testing.T) {
		if !IsValidType(TypeStatic) {
			t.Error("static should be valid")
		}
		if !IsValidType(TypePHP) {
			t.Error("php should be valid")
		}
		if !IsValidType(TypeProxy) {
			t.Error("proxy should be valid")
		}
		if !IsValidType(TypeLaravel) {
			t.Error("laravel should be valid")
		}
		if !IsValidType(TypeWordPress) {
			t.Error("wordpress should be valid")
		}
		if IsValidType("invalid") {
			t.Error("invalid should not be valid")
		}
	})
}
