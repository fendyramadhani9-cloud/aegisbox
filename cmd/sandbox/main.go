package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("AegisBox PID:", os.Getpid())

	cmd := exec.Command(
		"python3",
		"-c",
		"import os; print('Python PID:', os.getpid())",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("execution failed:", err)
		return
	}

	fmt.Println("execution completed")
}
