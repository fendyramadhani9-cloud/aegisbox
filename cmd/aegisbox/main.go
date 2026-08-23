package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/aegisbox/aegisbox/internal/api"
	"github.com/aegisbox/aegisbox/internal/config"
	"github.com/aegisbox/aegisbox/internal/executor"
	"github.com/aegisbox/aegisbox/internal/logging"
	aegisruntime "github.com/aegisbox/aegisbox/internal/runtime"
	"github.com/aegisbox/aegisbox/internal/sandbox"
)

var (
	Version   = "0.1.0"
	GitCommit = "dev"
	BuildTime = "unknown"
	Milestone = "AegisBox Production Isolation Engine"
)

func init() {
	api.SetVersionMetadata(Version, GitCommit, BuildTime)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "__init__":
		sandbox.RunChildInit()
		os.Exit(0)
	case "server":
		runServer(os.Args[2:])
	case "execute":
		runExecute(os.Args[2:])
	case "health", "-health", "--health":
		runHealth()
	case "version", "-version", "--version":
		runVersion()
	case "help", "-help", "--help":
		printUsage()
	default:
		// Check for flags passed directly
		if os.Args[1] == "-v" || os.Args[1] == "-version" {
			runVersion()
		} else if os.Args[1] == "-h" || os.Args[1] == "-health" {
			runHealth()
		} else {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
			printUsage()
			os.Exit(1)
		}
	}
}

func printUsage() {
	fmt.Println("AegisBox — Secure Ephemeral Code Execution Engine")
	fmt.Println("\nUsage:")
	fmt.Println("  aegisbox server   [options]   Start the REST API daemon")
	fmt.Println("  aegisbox execute  [options]   Execute code directly via CLI")
	fmt.Println("  aegisbox health               Perform local diagnostics")
	fmt.Println("  aegisbox version              Show version and build information")
	fmt.Println("\nServer Options:")
	fmt.Println("  -port int     Port to listen on (default 8080)")
	fmt.Println("  -host string  Host to bind to (default 0.0.0.0)")
	fmt.Println("\nExecute Options:")
	fmt.Println("  -language string  Language (default: python)")
	fmt.Println("  -code string      Source code string to execute")
	fmt.Println("  -timeout int64    Timeout in milliseconds (default: 1000)")
	fmt.Println("  -memory int64     Memory limit in MB (default: 64)")
	fmt.Println("  -processes int64  Process limit (default: 10)")
}

func runVersion() {
	fmt.Printf("AegisBox Engine v%s (%s)\n", Version, Milestone)
	fmt.Printf("Commit: %s | Built: %s\n", GitCommit, BuildTime)
	fmt.Printf("Platform: %s/%s | Go: %s\n", runtime.GOOS, runtime.GOARCH, runtime.Version())
}

func runHealth() {
	health := api.HealthResponse{
		Status:             "ok",
		Version:            Version,
		GitCommit:          GitCommit,
		BuildTime:          BuildTime,
		OS:                 runtime.GOOS,
		Arch:               runtime.GOARCH,
		SupportedLanguages: aegisruntime.DefaultRegistry().SupportedLanguages(),
	}
	data, _ := json.MarshalIndent(health, "", "  ")
	fmt.Println(string(data))
}

func runServer(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	cfg := config.DefaultConfig()
	cfg.LoadFromEnv()

	port := fs.Int("port", cfg.Server.Port, "Port to listen on")
	host := fs.String("host", cfg.Server.Host, "Host to bind to")
	_ = fs.Parse(args)

	cfg.Server.Port = *port
	cfg.Server.Host = *host

	logger := logging.Default()
	resolver := aegisruntime.DefaultRegistry()
	sbxAdapter := sandbox.NewSandboxAdapter(cfg, logger)
	execMgr := executor.NewExecutionManager(cfg, resolver, sbxAdapter, logger)

	router := api.NewRouter(execMgr, resolver, logger)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	logger.Info("starting aegisbox server", map[string]interface{}{
		"addr":       addr,
		"version":    Version,
		"git_commit": GitCommit,
		"build_time": BuildTime,
		"os":         runtime.GOOS,
	})

	// Graceful shutdown channel
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed to start", map[string]interface{}{"error": err.Error()})
			os.Exit(1)
		}
	}()

	fmt.Printf("AegisBox API Server listening at http://%s (commit: %s)\n", addr, GitCommit)
	<-stopChan

	logger.Info("shutting down server gracefully...", nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	logger.Info("server exited cleanly", nil)
}

func runExecute(args []string) {
	fs := flag.NewFlagSet("execute", flag.ExitOnError)
	cfg := config.DefaultConfig()
	cfg.LoadFromEnv()

	lang := fs.String("language", "python", "Language to execute")
	code := fs.String("code", "", "Source code to execute")
	timeout := fs.Int64("timeout", cfg.Sandbox.DefaultTimeoutMs, "Timeout in ms")
	mem := fs.Int64("memory", cfg.Sandbox.DefaultMemoryMB, "Memory limit in MB")
	pids := fs.Int64("processes", cfg.Sandbox.DefaultProcesses, "Max processes")
	_ = fs.Parse(args)

	if *code == "" {
		fmt.Fprintln(os.Stderr, "Error: --code flag cannot be empty")
		fs.Usage()
		os.Exit(1)
	}

	logger := logging.Default()
	resolver := aegisruntime.DefaultRegistry()
	sbxAdapter := sandbox.NewSandboxAdapter(cfg, logger)
	execMgr := executor.NewExecutionManager(cfg, resolver, sbxAdapter, logger)

	req := &executor.ExecutionRequest{
		Language:     *lang,
		Code:         *code,
		TimeoutMs:    *timeout,
		MaxMemMB:     *mem,
		MaxProcesses: *pids,
	}

	result := execMgr.Execute(context.Background(), req)
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))

	if result.Status != executor.StatusCompleted {
		os.Exit(result.ExitCode)
	}
}
