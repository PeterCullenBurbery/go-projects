package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Prefer PowerShell 7+ (pwsh) with -NoProfile
	cmd := exec.Command("pwsh", "-NoLogo", "-NoExit", "-NoProfile", "-Command", "-")

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	// Start PowerShell
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to start pwsh:", err)
		return
	}

	// Stream pwsh output to terminal
	go streamOutput(stdout)
	go streamOutput(stderr)

	// Input loop
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\n> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)

		if line == "exit" {
			stdin.Write([]byte("exit\n"))
			break
		}

		stdin.Write([]byte(line + "\n"))
	}

	cmd.Wait()
}

func streamOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
