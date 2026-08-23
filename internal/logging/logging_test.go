package logging_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aegisbox/aegisbox/internal/logging"
)

func TestLogger_LifecycleFormatting(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.NewLogger(&buf, logging.LevelInfo)

	logger.Lifecycle("execution.created", "exec-001", "python", "created execution environment", map[string]interface{}{
		"timeout_ms": 1000,
	})

	output := buf.String()
	if !strings.Contains(output, "exec-001") {
		t.Fatalf("expected log to contain execution ID 'exec-001', got: %s", output)
	}

	var entry logging.LogEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("expected valid JSON log entry, got error: %v", err)
	}

	if entry.Event != "execution.created" {
		t.Errorf("expected event 'execution.created', got '%s'", entry.Event)
	}
	if entry.ExecutionID != "exec-001" {
		t.Errorf("expected execution ID 'exec-001', got '%s'", entry.ExecutionID)
	}
	if entry.Language != "python" {
		t.Errorf("expected language 'python', got '%s'", entry.Language)
	}
}
