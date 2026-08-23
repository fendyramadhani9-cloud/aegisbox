package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/fendyramadhani9-cloud/aegisbox/internal/executor"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: aegisbox run [options]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		runCommand()

	default:
		fmt.Println("unknown command:", os.Args[1])
		os.Exit(1)
	}
}

func runCommand() {
	fs := flag.NewFlagSet("run", flag.ExitOnError)

	language := fs.String(
		"language",
		"python",
		"programming language",
	)

	code := fs.String(
		"code",
		"",
		"code to execute",
	)

	timeout := fs.Int(
		"timeout",
		1000,
		"timeout in milliseconds",
	)

	fs.Parse(os.Args[2:])

	if *code == "" {
		fmt.Println("code cannot be empty")
		os.Exit(1)
	}

	result := executor.Run(executor.Request{
		Language: *language,
		Code:     *code,
		Timeout:  time.Duration(*timeout) * time.Millisecond,
	})

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println("failed to encode result:", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}
