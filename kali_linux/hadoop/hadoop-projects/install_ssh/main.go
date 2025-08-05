package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func run(command_name string, command_args ...string) error {
	cmd := exec.Command(command_name, command_args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func file_exists(file_path string) bool {
	_, err := os.Stat(file_path)
	return err == nil
}

func main() {
	fmt.Println("🔧 Step 2: Setting up SSH and passwordless login...")

	// 1. Install openssh-server
	fmt.Println("📦 Installing openssh-server...")
	if err := run("sudo", "apt", "install", "openssh-server", "-y"); err != nil {
		fmt.Println("❌ Failed to install openssh-server.")
		return
	}

	home_dir, _ := os.UserHomeDir()
	ssh_dir := filepath.Join(home_dir, ".ssh")
	private_key := filepath.Join(ssh_dir, "id_rsa")
	public_key := filepath.Join(ssh_dir, "id_rsa.pub")
	authorized_keys := filepath.Join(ssh_dir, "authorized_keys")

	// 2. Generate SSH key if not present
	if !file_exists(private_key) || !file_exists(public_key) {
		fmt.Println("🔑 Generating new SSH key pair...")
		if err := run("ssh-keygen", "-t", "rsa", "-P", "", "-f", private_key); err != nil {
			fmt.Println("❌ Failed to generate SSH key.")
			return
		}
	} else {
		fmt.Println("✅ SSH key pair already exists.")
	}

	// 3. Ensure authorized_keys exists and append if needed
	public_key_content, err := os.ReadFile(public_key)
	if err != nil {
		fmt.Printf("❌ Failed to read public key: %v\n", err)
		return
	}

	os.MkdirAll(ssh_dir, 0700)

	auth_file, err := os.OpenFile(authorized_keys, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("❌ Failed to open authorized_keys: %v\n", err)
		return
	}
	defer auth_file.Close()

	_, err = auth_file.Write(public_key_content)
	if err != nil {
		fmt.Printf("❌ Failed to write to authorized_keys: %v\n", err)
		return
	}
	auth_file.WriteString("\n")

	// 4. Set correct permissions
	fmt.Println("🔐 Setting permissions on authorized_keys...")
	if err := os.Chmod(authorized_keys, 0600); err != nil {
		fmt.Printf("❌ Failed to chmod: %v\n", err)
		return
	}

	// 5. SSH to localhost to trigger host acceptance
	fmt.Println("🔌 Connecting to localhost to confirm setup...")
	if err := run("ssh", "-o", "StrictHostKeyChecking=no", "localhost", "echo", "✅ SSH to localhost successful!"); err != nil {
		fmt.Println("⚠️ SSH to localhost failed. Try manually running: ssh localhost")
		return
	}

	fmt.Println("✅ SSH passwordless setup complete.")
}