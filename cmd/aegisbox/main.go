package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/aegisbox/aegisbox/internal/config"
)

const (
	Version   = "0.1.0"
	Milestone = "Milestone 1: Project Foundation"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Milestone string `json:"milestone"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	GoVersion string `json:"go_version"`
}

func main() {
	var (
		showVersion = flag.Bool("version", false, "Display version information and exit")
		showHealth  = flag.Bool("health", false, "Perform a basic self-check and display health info in JSON format")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("AegisBox Engine v%s (%s)\n", Version, Milestone)
		fmt.Printf("Platform: %s/%s | Go: %s\n", runtime.GOOS, runtime.GOARCH, runtime.Version())
		os.Exit(0)
	}

	if *showHealth {
		health := HealthResponse{
			Status:    "ok",
			Version:   Version,
			Milestone: Milestone,
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
			GoVersion: runtime.Version(),
		}
		data, err := json.MarshalIndent(health, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting health response: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
		os.Exit(0)
	}

	cfg := config.DefaultConfig()
	cfg.LoadFromEnv()
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== AegisBox Engine v%s ===\n", Version)
	fmt.Printf("Milestone: %s\n", Milestone)
	fmt.Printf("Server Host: %s:%d (Mode: %s)\n", cfg.Server.Host, cfg.Server.Port, cfg.Server.Mode)
	fmt.Printf("Sandbox Rootfs: %s\n", cfg.Sandbox.RootfsDir)
	fmt.Printf("Workspace Base: %s\n", cfg.Sandbox.WorkspaceBaseDir)
	fmt.Println("Run with -health or -version for quick diagnostic information.")
}
