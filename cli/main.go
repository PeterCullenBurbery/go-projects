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
	args := strings.Fields(input)
	if len(args) == 0 {
		return nil
	}

	switch args[0] {
	case "cd":
		if len(args) < 2 {
			return ErrNoPath
		}
		newPath := addLongPathPrefix(args[1])
		return os.Chdir(newPath)
	case "exit":
		os.Exit(0)
	case "pwd":
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		fmt.Println(addLongPathPrefix(dir))
		return nil
	}

	// Try native execution
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = addLongPathPrefixSafe()
	if err := cmd.Run(); err == nil {
		return nil
	}

	// Fallback to PowerShell Core (pwsh)
	psCmd := exec.Command("pwsh", "-NoProfile", "-NonInteractive", "-Command", input)
	psCmd.Stdout = os.Stdout
	psCmd.Stderr = os.Stderr
	psCmd.Dir = addLongPathPrefixSafe()
	return psCmd.Run()
}

func addLongPathPrefix(path string) string {
	if strings.HasPrefix(path, `\\?\`) {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return `\\?\` + abs
}

func addLongPathPrefixSafe() string {
	dir, err := os.Getwd()
	if err != nil {
		return "C:\\"
	}
	return addLongPathPrefix(dir)
}