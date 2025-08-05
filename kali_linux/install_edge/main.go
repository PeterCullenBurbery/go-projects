package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
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
		installEdgeDebian()
	case "fedora", "rhel", "centos":
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
	fmt.Println("Installing Microsoft Edge on Debian/Ubuntu...")

	commands := []string{
		"sudo apt update",
		"sudo apt install -y wget gnupg2 software-properties-common",
		"sh -c \"wget -q https://packages.microsoft.com/keys/microsoft.asc -O- | gpg --dearmor > microsoft.gpg\"",
		"sudo install -o root -g root -m 644 microsoft.gpg /etc/apt/trusted.gpg.d/",
		"sudo sh -c \"echo 'deb [arch=amd64] https://packages.microsoft.com/repos/edge stable main' > /etc/apt/sources.list.d/microsoft-edge.list\"",
		"sudo apt update",
		"sudo apt install -y microsoft-edge-stable",
		"rm microsoft.gpg",
	}

	runCommands(commands)
}

func installEdgeRHEL() {
	fmt.Println("Installing Microsoft Edge on Fedora/RHEL/CentOS...")

	commands := []string{
		"sudo dnf install -y https://packages.microsoft.com/yumrepos/edge/microsoft-edge-stable.x86_64.rpm",
	}

	runCommands(commands)
}

func runCommands(commands []string) {
	for _, cmd := range commands {
		fmt.Printf("Running: %s\n", cmd)

		var execCmd *exec.Cmd
		if strings.ContainsAny(cmd, "|><") {
			execCmd = exec.Command("bash", "-c", cmd)
		} else {
			parts := strings.Fields(cmd)
			execCmd = exec.Command(parts[0], parts[1:]...)
		}

		execCmd.Stdin = nil
		execCmd.Stdout = nil
		execCmd.Stderr = nil

		err := execCmd.Run()
		if err != nil {
			log.Fatalf("Command failed: %s\nError: %v", cmd, err)
		}
	}
	fmt.Println("Microsoft Edge installation completed.")
}