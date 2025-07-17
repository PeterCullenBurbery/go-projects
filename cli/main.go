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
			fmt.Fprintln(os.Stderr, "Input error:", err)
			continue
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if err := execInput(input); err != nil {
			fmt.Fprintln(os.Stderr, "Execution error:", err)
		}
	}
}

var ErrNoPath = errors.New("path required")

func execInput(input string) error {
	args, err := splitArgs(input)
	if err != nil || len(args) == 0 {
		return err
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

	// Resolve absolute path and prepend \\?\ for long path support
	binary := args[0]
	absBinary, err := filepath.Abs(binary)
	if err == nil {
		binary = `\\?\` + absBinary
	}

	cmd := exec.Command(binary, args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// splitArgs splits input into arguments respecting quoted substrings
func splitArgs(input string) ([]string, error) {
	var args []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(input); i++ {
		switch input[i] {
		case '"':
			inQuotes = !inQuotes
		case ' ':
			if inQuotes {
				current.WriteByte(input[i])
			} else if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(input[i])
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	if inQuotes {
		return nil, errors.New("unmatched quote")
	}
	return args, nil
}