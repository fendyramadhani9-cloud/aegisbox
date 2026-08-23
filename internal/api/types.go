package api

import (
	"github.com/aegisbox/aegisbox/internal/executor"
)

// ExecuteRequestPayload represents JSON body for POST /execute.
type ExecuteRequestPayload struct {
	Language     string            `json:"language"`
	Code         string            `json:"code"`
	TimeoutMs    int64             `json:"timeout_ms"`
	MaxMemMB     int64             `json:"max_mem_mb"`
	MaxProcesses int64             `json:"max_processes,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	Args         []string          `json:"args,omitempty"`
}

// ToDomain converts API payload into executor domain model.
func (p *ExecuteRequestPayload) ToDomain() *executor.ExecutionRequest {
	return &executor.ExecutionRequest{
		Language:     p.Language,
		Code:         p.Code,
		TimeoutMs:    p.TimeoutMs,
		MaxMemMB:     p.MaxMemMB,
		MaxProcesses: p.MaxProcesses,
		Env:          p.Env,
		Args:         p.Args,
	}
}

// ErrorResponse represents standard error payload.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// HealthResponse represents system health status and build metadata.
type HealthResponse struct {
	Status             string   `json:"status"`
	Version            string   `json:"version"`
	GitCommit          string   `json:"git_commit,omitempty"`
	BuildTime          string   `json:"build_time,omitempty"`
	OS                 string   `json:"os"`
	Arch               string   `json:"arch"`
	SupportedLanguages []string `json:"supported_languages"`
}
