package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

// ServerConfig holds HTTP/API listener configurations.
type ServerConfig struct {
	Host string `json:"host" yaml:"host"`
	Port int    `json:"port" yaml:"port"`
	Mode string `json:"mode" yaml:"mode"`
}

// SandboxConfig holds default limits and filesystem path constraints for execution environments.
type SandboxConfig struct {
	RootfsDir        string `json:"rootfs_dir" yaml:"rootfs_dir"`
	WorkspaceBaseDir string `json:"workspace_base_dir" yaml:"workspace_base_dir"`
	DefaultTimeoutMs int64  `json:"default_timeout_ms" yaml:"default_timeout_ms"`
	MaxTimeoutMs     int64  `json:"max_timeout_ms" yaml:"max_timeout_ms"`
	DefaultMemoryMB  int64  `json:"default_memory_mb" yaml:"default_memory_mb"`
	MaxMemoryMB      int64  `json:"max_memory_mb" yaml:"max_memory_mb"`
	DefaultProcesses int64  `json:"default_processes" yaml:"default_processes"`
	MaxProcesses     int64  `json:"max_processes" yaml:"max_processes"`
	CgroupParent     string `json:"cgroup_parent" yaml:"cgroup_parent"`
}

// Config represents the root configuration model for AegisBox.
type Config struct {
	Server  ServerConfig  `json:"server" yaml:"server"`
	Sandbox SandboxConfig `json:"sandbox" yaml:"sandbox"`
}

// DefaultConfig returns safe and sensible defaults for development and testing.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
			Mode: "development",
		},
		Sandbox: SandboxConfig{
			RootfsDir:        "/opt/aegisbox/rootfs",
			WorkspaceBaseDir: "/tmp/aegisbox/workspaces",
			DefaultTimeoutMs: 1000,
			MaxTimeoutMs:     30000,
			DefaultMemoryMB:  64,
			MaxMemoryMB:      512,
			DefaultProcesses: 10,
			MaxProcesses:     64,
			CgroupParent:     "/sys/fs/cgroup/aegisbox",
		},
	}
}

// Validate validates that configuration parameters fall within safe operating boundaries.
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d (must be between 1 and 65535)", c.Server.Port)
	}

	if c.Sandbox.DefaultTimeoutMs <= 0 || c.Sandbox.MaxTimeoutMs < c.Sandbox.DefaultTimeoutMs {
		return errors.New("invalid sandbox timeout configuration: max_timeout_ms must be >= default_timeout_ms and > 0")
	}

	if c.Sandbox.DefaultMemoryMB <= 0 || c.Sandbox.MaxMemoryMB < c.Sandbox.DefaultMemoryMB {
		return errors.New("invalid sandbox memory configuration: max_memory_mb must be >= default_memory_mb and > 0")
	}

	if c.Sandbox.DefaultProcesses <= 0 || c.Sandbox.MaxProcesses < c.Sandbox.DefaultProcesses {
		return errors.New("invalid sandbox process configuration: max_processes must be >= default_processes and > 0")
	}

	return nil
}

// LoadFromEnv overrides default configuration values from environment variables if present.
func (c *Config) LoadFromEnv() {
	if host := os.Getenv("AEGISBOX_SERVER_HOST"); host != "" {
		c.Server.Host = host
	}
	if portStr := os.Getenv("AEGISBOX_SERVER_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
			c.Server.Port = port
		}
	}
	if rootfs := os.Getenv("AEGISBOX_ROOTFS_DIR"); rootfs != "" {
		c.Sandbox.RootfsDir = rootfs
	}
	if workspace := os.Getenv("AEGISBOX_WORKSPACE_DIR"); workspace != "" {
		c.Sandbox.WorkspaceBaseDir = workspace
	}
}
