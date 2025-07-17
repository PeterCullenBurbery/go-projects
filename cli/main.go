package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
	input = strings.TrimSpace(input)
	args := strings.Fields(input)

	if len(args) == 0 {
		return nil
	}

	// Handle built-in commands
	switch args[0] {
	case "cd":
		if len(args) < 2 {
			return errors.New("path required")
		}
		return os.Chdir(args[1])
	case "exit":
		os.Exit(0)
	case "pwd":
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		fmt.Println(dir)
		return nil
	}

	// Try exec.Command directly
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err == nil {
		return nil
	}

	// If it failed, try through PowerShell
	powershellArgs := append([]string{"-Command"}, input)
	psCmd := exec.Command("powershell", powershellArgs...)
	psCmd.Stdout = os.Stdout
	psCmd.Stderr = os.Stderr
	return psCmd.Run()
}