package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type step struct {
	cmd        string
	maxTries   int           // how many attempts before we give up
	timeout    time.Duration // per-attempt timeout
	skipIfFile string        // if this file exists, skip step (idempotence helper)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	if runtime.GOOS != "linux" {
		log.Fatal("This program only supports Linux.")
	}

	distro, err := detectDistro()
	if err != nil {
		log.Fatalf("Could not detect distro: %v", err)
	}
	fmt.Printf("Detected distro: %s\n", distro)

	switch distro {
	case "ubuntu", "debian", "kali":
		if isDebPackageInstalled("microsoft-edge-stable") {
			fmt.Println("Microsoft Edge already installed. Nothing to do.")
			return
		}
		installEdgeDebian()
	case "fedora", "rhel", "centos":
		if isRpmPackageInstalled("microsoft-edge-stable") {
			fmt.Println("Microsoft Edge already installed. Nothing to do.")
			return
		}
		installEdgeRHEL()
	default:
		log.Fatalf("Unsupported Linux distro: %s", distro)
	}
}

func detectDistro() (string, error) {
	output, err := exec.Command("sh", "-c", `grep -E '^ID=' /etc/os-release | cut -d= -f2`).Output()
	if err != nil {
		return "", err
	}
	distro := strings.Trim(string(output), "\"\n")
	return distro, nil
}

func installEdgeDebian() {
	fmt.Println("Installing Microsoft Edge on Debian/Ubuntu/Kali...")

	// Newer Debian/Ubuntu prefer keyrings; the below works broadly. Retries cover transient network issues.
	steps := []step{
		{cmd: "sudo apt update", maxTries: 5, timeout: 5 * time.Minute},
		{cmd: "sudo apt install -y wget gnupg2 software-properties-common", maxTries: 5, timeout: 10 * time.Minute},
		// Fetch MS key and write to file. This is the one that failed for you; give it more tries.
		{cmd: `sh -c "wget -q https://packages.microsoft.com/keys/microsoft.asc -O- | gpg --dearmor > microsoft.gpg"`, maxTries: 8, timeout: 2 * time.Minute, skipIfFile: "microsoft.gpg"},
		{cmd: "sudo install -o root -g root -m 644 microsoft.gpg /etc/apt/trusted.gpg.d/", maxTries: 5, timeout: 2 * time.Minute},
		{cmd: `sudo sh -c "echo 'deb [arch=amd64] https://packages.microsoft.com/repos/edge stable main' > /etc/apt/sources.list.d/microsoft-edge.list"`, maxTries: 5, timeout: 2 * time.Minute, skipIfFile: "/etc/apt/sources.list.d/microsoft-edge.list"},
		{cmd: "sudo apt update", maxTries: 5, timeout: 5 * time.Minute},
		{cmd: "sudo apt install -y microsoft-edge-stable", maxTries: 8, timeout: 20 * time.Minute},
		{cmd: "rm -f microsoft.gpg", maxTries: 3, timeout: 30 * time.Second},
	}

	runSteps(steps)
	fmt.Println("Microsoft Edge installation completed.")
}

func installEdgeRHEL() {
	fmt.Println("Installing Microsoft Edge on Fedora/RHEL/CentOS...")

	steps := []step{
		{cmd: "sudo dnf install -y https://packages.microsoft.com/yumrepos/edge/microsoft-edge-stable.x86_64.rpm", maxTries: 8, timeout: 20 * time.Minute},
	}

	runSteps(steps)
	fmt.Println("Microsoft Edge installation completed.")
}

func runSteps(steps []step) {
	for _, s := range steps {
		if s.skipIfFile != "" {
			if _, err := os.Stat(s.skipIfFile); err == nil {
				fmt.Printf("Skipping (exists): %s\n", s.skipIfFile)
				continue
			}
		}

		attempt := 0
		for {
			attempt++
			fmt.Printf("Running (attempt %d/%d): %s\n", attempt, s.maxTries, s.cmd)

			// Use bash -lc to support pipes, redirects, quotes, etc.
			ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
			defer cancel()

			cmd := exec.CommandContext(ctx, "bash", "-lc", s.cmd)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Success
			if err == nil {
				// Optional: show minimal output when it's informative
				out := strings.TrimSpace(stdout.String())
				if out != "" {
					fmt.Println(out)
				}
				break
			}

			// Timeout is also an error; show stderr/stdout to help debugging.
			fmt.Printf("Command failed: %s\n", s.cmd)
			se := strings.TrimSpace(stderr.String())
			so := strings.TrimSpace(stdout.String())
			if so != "" {
				fmt.Printf("STDOUT:\n%s\n", so)
			}
			if se != "" {
				fmt.Printf("STDERR:\n%s\n", se)
			}
			if attempt >= s.maxTries {
				log.Fatalf("Giving up after %d attempts: %v", attempt, err)
			}

			// Exponential backoff with jitter.
			sleep := backoffWithJitter(attempt, 2*time.Second, 15*time.Second)
			fmt.Printf("Retrying in %s...\n", sleep.Round(time.Second))
			time.Sleep(sleep)
		}
	}
}

func backoffWithJitter(attempt int, base, max time.Duration) time.Duration {
	// Exponential: base * 2^(attempt-1), capped at max, then add +/- 20% jitter
	back := base << (attempt - 1)
	if back > max {
		back = max
	}
	jitterFrac := 0.2 * (rand.Float64()*2 - 1) // -20% .. +20%
	jitter := time.Duration(float64(back) * jitterFrac)
	return back + jitter
}

func isDebPackageInstalled(name string) bool {
	cmd := exec.Command("bash", "-lc", fmt.Sprintf("dpkg -s %s >/dev/null 2>&1", shellEscape(name)))
	return cmd.Run() == nil
}

func isRpmPackageInstalled(name string) bool {
	cmd := exec.Command("bash", "-lc", fmt.Sprintf("rpm -q %s >/dev/null 2>&1", shellEscape(name)))
	return cmd.Run() == nil
}

func shellEscape(s string) string {
	// Minimal escape for simple package names
	return strings.ReplaceAll(s, `'`, `'\''`)
}