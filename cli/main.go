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

// ErrNoPath is returned when 'cd' has no target.
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
		return os.Chdir(normalizePath(args[1]))
	case "pwd":
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		fmt.Println(dir)
		return nil
	case "exit":
		os.Exit(0)
	}

	// Decide whether the first token is a path or just a command name.
	prog := args[0]
	if strings.ContainsAny(prog, `\/`) {
		prog = normalizePath(prog)
	}

	// Native execution first.
	cmd := exec.Command(prog, args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err == nil {
		return nil
	}

	// Fallback: treat whole line as a PowerShell command.
	ps := exec.Command("powershell",
		"-NoProfile", "-NonInteractive", "-Command", input)
	ps.Stdin = os.Stdin
	ps.Stdout = os.Stdout
	ps.Stderr = os.Stderr
	return ps.Run()
}

// normalizePath converts any supplied path to an absolute Windows path
// and unconditionally prefixes it with \\?\  (unless already present).
func normalizePath(p string) string {
	p = strings.Trim(p, `"'`) // strip quotes
	abs, err := filepath.Abs(p)
	if err != nil {
		abs = p // fall back to original if Abs fails
	}
	if strings.HasPrefix(abs, `\\?\`) {
		return abs
	}
	return `\\?\` + abs
}