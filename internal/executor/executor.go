package executor

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

type Request struct {
	Language string
	Code     string
	Timeout  time.Duration
}

type Result struct {
	Stdout     string
	Stderr     string
	ExitCode   int
	Status     string
	ExecutionTime time.Duration
}

func Run(req Request) Result {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), req.Timeout)
	defer cancel()

	var cmd *exec.Cmd

	switch req.Language {
	case "python":
		cmd = exec.CommandContext(
			ctx,
			"python3",
			"-c",
			req.Code,
		)

	default:
		return Result{
			Status: "UNSUPPORTED_LANGUAGE",
		}
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := Result{
		Stdout:        stdout.String(),
		Stderr:        stderr.String(),
		ExitCode:      0,
		Status:        "COMPLETED",
		ExecutionTime: time.Since(start),
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.Status = "TIME_LIMIT_EXCEEDED"
		result.ExitCode = -1
		return result
	}

	if err != nil {
		result.Status = "RUNTIME_ERROR"

		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}

	return result
}
