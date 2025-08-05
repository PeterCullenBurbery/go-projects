package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	if runtime.GOOS != "linux" {
		log.Fatal("This installer only supports Linux.")
	}

	linux_distro, err := detect_linux_distro()
	if err != nil {
		log.Fatalf("Failed to detect distro: %v", err)
	}

	fmt.Printf("Detected distro: %s\n", linux_distro)

	switch linux_distro {
	case "ubuntu", "debian", "kali":
		install_gitkraken_debian()
	case "fedora", "rhel", "centos":
		install_gitkraken_rhel()
	default:
		log.Fatalf("Unsupported distro: %s", linux_distro)
	}
}

func detect_linux_distro() (string, error) {
	output, err := exec.Command("sh", "-c", `grep -E '^ID=' /etc/os-release | cut -d= -f2`).Output()
	if err != nil {
		return "", err
	}
	return strings.Trim(string(output), "\"\n"), nil
}

func install_gitkraken_debian() {
	fmt.Println("Installing GitKraken on Debian/Ubuntu/Kali...")

	command_list := []string{
		"sudo apt update",
		"sudo apt install -y wget",
		"wget https://release.gitkraken.com/linux/gitkraken-amd64.deb -O gitkraken.deb",
		"sudo dpkg -i gitkraken.deb || sudo apt install -f -y",
		"rm gitkraken.deb",
	}

	run_command_list(command_list)
}

func install_gitkraken_rhel() {
	fmt.Println("Installing GitKraken on Fedora/RHEL/CentOS...")

	command_list := []string{
		"sudo dnf install -y wget",
		"wget https://release.gitkraken.com/linux/gitkraken-amd64.rpm -O gitkraken.rpm",
		"sudo dnf install -y gitkraken.rpm",
		"rm gitkraken.rpm",
	}

	run_command_list(command_list)
}

func run_command_list(command_list []string) {
	for _, command := range command_list {
		fmt.Printf("Running: %s\n", command)

		var exec_command *exec.Cmd
		if strings.ContainsAny(command, "|><") {
			exec_command = exec.Command("bash", "-c", command)
		} else {
			parts := strings.Fields(command)
			exec_command = exec.Command(parts[0], parts[1:]...)
		}

		exec_command.Stdin = os.Stdin
		exec_command.Stdout = os.Stdout
		exec_command.Stderr = os.Stderr

		err := exec_command.Run()
		if err != nil {
			log.Fatalf("Command failed: %s\nError: %v", command, err)
		}
	}
	fmt.Println("✅ GitKraken installation completed.")
}
