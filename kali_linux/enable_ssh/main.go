package main

import (
	"fmt"
	"os"
	"os/exec"
)

func run_command(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func main() {
	// ✅ Check for root privileges
	if os.Geteuid() != 0 {
		fmt.Println("❌ This program must be run as root. Try using: sudo ./enable_ssh")
		return
	}

	fmt.Println("🔧 Enabling SSH on Kali Linux...")

	// 1. Update apt package index
	fmt.Println("📦 Updating package index...")
	if err := run_command("apt", "update"); err != nil {
		fmt.Println("❌ Failed to update package index:", err)
		return
	}

	// 2. Install openssh-server
	fmt.Println("📥 Installing openssh-server...")
	if err := run_command("apt", "install", "-y", "openssh-server"); err != nil {
		fmt.Println("❌ Failed to install openssh-server:", err)
		return
	}

	// 3. Start SSH service
	fmt.Println("▶️ Starting SSH service...")
	if err := run_command("systemctl", "start", "ssh"); err != nil {
		fmt.Println("❌ Failed to start ssh service:", err)
		return
	}

	// 4. Enable SSH service to start on boot
	fmt.Println("🔄 Enabling SSH service on boot...")
	if err := run_command("systemctl", "enable", "ssh"); err != nil {
		fmt.Println("❌ Failed to enable ssh service:", err)
		return
	}

	// 5. Show status
	fmt.Println("📈 SSH service status:")
	run_command("systemctl", "status", "ssh")

	fmt.Println("✅ SSH is now enabled.")
}