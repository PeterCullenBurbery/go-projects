package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		// Trim newline and skip empty input
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Execute the input line
		if err := execInput(input); err != nil {
			fmt.Fprintln(os.Stderr, "Execution error:", err)
		}
	}
}

// ErrNoPath is returned when 'cd' was called without a second argument.
var ErrNoPath = errors.New("path required")

func execInput(input string) error {
	args := strings.Fields(input)
	if len(args) == 0 {
		return nil
	}

	switch args[0] {
	case "cd":
		if len(args) < 2 {
			return ErrNoPath
		}
		return os.Chdir(args[1])
	case "exit":
		os.Exit(0)
	}

	// If command includes a slash or backslash, treat as path
	binary := args[0]
	if strings.ContainsAny(binary, `\/`) {
		abs, err := filepath.Abs(binary)
		if err == nil {
			if _, statErr := os.Stat(abs); statErr == nil {
				binary = `\\?\` + abs
			}
		}
	}

	cmd := exec.Command(binary, args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin // Allow interactive commands
	return cmd.Run()
}
