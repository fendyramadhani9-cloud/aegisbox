package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level represents log severity.
type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// LogEntry defines the structured JSON log format for AegisBox events.
type LogEntry struct {
	Timestamp   string                 `json:"timestamp"`
	Level       Level                  `json:"level"`
	Message     string                 `json:"message"`
	Event       string                 `json:"event,omitempty"`
	ExecutionID string                 `json:"execution_id,omitempty"`
	Language    string                 `json:"language,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
}

// Logger provides structured JSON logging with thread-safety.
type Logger struct {
	mu     sync.Mutex
	out    io.Writer
	minLvl Level
}

// NewLogger creates a Logger writing to the specified writer.
func NewLogger(out io.Writer, minLevel Level) *Logger {
	if out == nil {
		out = os.Stdout
	}
	return &Logger{
		out:    out,
		minLvl: minLevel,
	}
}

// Default returns a standard stdout logger with INFO level.
func Default() *Logger {
	return NewLogger(os.Stdout, LevelInfo)
}

func (l *Logger) log(level Level, event string, execID string, lang string, msg string, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339Nano),
		Level:       level,
		Message:     msg,
		Event:       event,
		ExecutionID: execID,
		Language:    lang,
		Fields:      fields,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal log entry: %v\n", err)
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = fmt.Fprintln(l.out, string(data))
}

// Info logs an informational message.
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.log(LevelInfo, "", "", "", msg, fields)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	l.log(LevelWarn, "", "", "", msg, fields)
}

// Error logs an error message.
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.log(LevelError, "", "", "", msg, fields)
}

// Lifecycle logs an execution lifecycle event without leaking raw user code.
func (l *Logger) Lifecycle(event string, execID string, lang string, msg string, fields map[string]interface{}) {
	l.log(LevelInfo, event, execID, lang, msg, fields)
}
