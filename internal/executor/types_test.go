package executor_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/aegisbox/aegisbox/internal/executor"
)

func TestExecutionRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		req         executor.ExecutionRequest
		expectError bool
		errContains string
	}{
		{
			name: "Valid request",
			req: executor.ExecutionRequest{
				Language:  "python",
				Code:      "print('Hello, AegisBox!')",
				TimeoutMs: 1000,
				MaxMemMB:  64,
			},
			expectError: false,
		},
		{
			name: "Empty language",
			req: executor.ExecutionRequest{
				Language:  "",
				Code:      "print(1)",
				TimeoutMs: 1000,
				MaxMemMB:  64,
			},
			expectError: true,
			errContains: "language cannot be empty",
		},
		{
			name: "Empty code",
			req: executor.ExecutionRequest{
				Language:  "python",
				Code:      "   ",
				TimeoutMs: 1000,
				MaxMemMB:  64,
			},
			expectError: true,
			errContains: "code cannot be empty",
		},
		{
			name: "Timeout too low",
			req: executor.ExecutionRequest{
				Language:  "python",
				Code:      "print(1)",
				TimeoutMs: 50,
				MaxMemMB:  64,
			},
			expectError: true,
			errContains: "timeout_ms must be between",
		},
		{
			name: "Timeout too high",
			req: executor.ExecutionRequest{
				Language:  "python",
				Code:      "print(1)",
				TimeoutMs: 70000,
				MaxMemMB:  64,
			},
			expectError: true,
			errContains: "timeout_ms must be between",
		},
		{
			name: "Memory too low",
			req: executor.ExecutionRequest{
				Language:  "python",
				Code:      "print(1)",
				TimeoutMs: 1000,
				MaxMemMB:  8,
			},
			expectError: true,
			errContains: "max_mem_mb must be between",
		},
		{
			name: "Memory too high",
			req: executor.ExecutionRequest{
				Language:  "python",
				Code:      "print(1)",
				TimeoutMs: 1000,
				MaxMemMB:  2048,
			},
			expectError: true,
			errContains: "max_mem_mb must be between",
		},
		{
			name: "Code payload exceeds max size",
			req: executor.ExecutionRequest{
				Language:  "python",
				Code:      strings.Repeat("a", executor.MaxCodeSizeBytes+1),
				TimeoutMs: 1000,
				MaxMemMB:  64,
			},
			expectError: true,
			errContains: "exceeds maximum allowed size",
		},
		{
			name: "Valid request with process limit",
			req: executor.ExecutionRequest{
				Language:     "python",
				Code:         "print('test')",
				TimeoutMs:    1000,
				MaxMemMB:     64,
				MaxProcesses: 10,
			},
			expectError: false,
		},
		{
			name: "Invalid process limit exceeds maximum",
			req: executor.ExecutionRequest{
				Language:     "python",
				Code:         "print('test')",
				TimeoutMs:    1000,
				MaxMemMB:     64,
				MaxProcesses: 500,
			},
			expectError: true,
			errContains: "max_processes must be between",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("expected error message to contain '%s', got '%s'", tc.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestExecutionRequest_TimeoutDuration(t *testing.T) {
	req := executor.ExecutionRequest{
		TimeoutMs: 2500,
	}
	expected := 2500 * time.Millisecond
	if req.TimeoutDuration() != expected {
		t.Errorf("expected %v, got %v", expected, req.TimeoutDuration())
	}
}

func TestExecutionResult_JSONSerialization(t *testing.T) {
	result := executor.ExecutionResult{
		ExecutionID:      "exec-12345",
		Status:           executor.StatusCompleted,
		Stdout:           "hello\n",
		Stderr:           "",
		ExitCode:         0,
		ExecutionTimeMs:  24,
		MemoryUsageBytes: 15728640,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to serialize ExecutionResult: %v", err)
	}

	var deserialized executor.ExecutionResult
	if err := json.Unmarshal(data, &deserialized); err != nil {
		t.Fatalf("failed to deserialize ExecutionResult: %v", err)
	}

	if deserialized.ExecutionID != result.ExecutionID {
		t.Errorf("expected execution ID %s, got %s", result.ExecutionID, deserialized.ExecutionID)
	}
	if deserialized.Status != executor.StatusCompleted {
		t.Errorf("expected status %s, got %s", executor.StatusCompleted, deserialized.Status)
	}
	if deserialized.Stdout != "hello\n" {
		t.Errorf("expected stdout 'hello\\n', got '%s'", deserialized.Stdout)
	}
}
