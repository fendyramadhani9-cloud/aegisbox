package config_test

import (
	"os"
	"testing"

	"github.com/aegisbox/aegisbox/internal/config"
)

func TestDefaultConfig_Valid(t *testing.T) {
	cfg := config.DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default config should be valid, got: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Sandbox.DefaultTimeoutMs != 1000 {
		t.Errorf("expected default timeout 1000ms, got %d", cfg.Sandbox.DefaultTimeoutMs)
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		modify      func(*config.Config)
		expectError bool
	}{
		{
			name: "Invalid server port",
			modify: func(c *config.Config) {
				c.Server.Port = 70000
			},
			expectError: true,
		},
		{
			name: "Invalid timeout bounds",
			modify: func(c *config.Config) {
				c.Sandbox.MaxTimeoutMs = 500
				c.Sandbox.DefaultTimeoutMs = 1000
			},
			expectError: true,
		},
		{
			name: "Invalid memory bounds",
			modify: func(c *config.Config) {
				c.Sandbox.MaxMemoryMB = 32
				c.Sandbox.DefaultMemoryMB = 64
			},
			expectError: true,
		},
		{
			name: "Invalid process bounds",
			modify: func(c *config.Config) {
				c.Sandbox.MaxProcesses = 5
				c.Sandbox.DefaultProcesses = 10
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			tc.modify(cfg)
			err := cfg.Validate()
			if tc.expectError && err == nil {
				t.Fatalf("expected validation error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("expected valid config, got: %v", err)
			}
		})
	}
}

func TestConfig_LoadFromEnv(t *testing.T) {
	os.Setenv("AEGISBOX_SERVER_HOST", "127.0.0.1")
	os.Setenv("AEGISBOX_SERVER_PORT", "9090")
	os.Setenv("AEGISBOX_ROOTFS_DIR", "/custom/rootfs")
	os.Setenv("AEGISBOX_WORKSPACE_DIR", "/custom/workspaces")
	defer func() {
		os.Unsetenv("AEGISBOX_SERVER_HOST")
		os.Unsetenv("AEGISBOX_SERVER_PORT")
		os.Unsetenv("AEGISBOX_ROOTFS_DIR")
		os.Unsetenv("AEGISBOX_WORKSPACE_DIR")
	}()

	cfg := config.DefaultConfig()
	cfg.LoadFromEnv()

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Sandbox.RootfsDir != "/custom/rootfs" {
		t.Errorf("expected rootfs /custom/rootfs, got %s", cfg.Sandbox.RootfsDir)
	}
	if cfg.Sandbox.WorkspaceBaseDir != "/custom/workspaces" {
		t.Errorf("expected workspace /custom/workspaces, got %s", cfg.Sandbox.WorkspaceBaseDir)
	}
}
