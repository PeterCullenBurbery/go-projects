package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
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

	home, _ := os.UserHomeDir()
	sshDir := filepath.Join(home, ".ssh")
	privateKey := filepath.Join(sshDir, "id_rsa")
	publicKey := filepath.Join(sshDir, "id_rsa.pub")
	authKeys := filepath.Join(sshDir, "authorized_keys")

	// 2. Generate SSH key if not present
	if !fileExists(privateKey) || !fileExists(publicKey) {
		fmt.Println("🔑 Generating new SSH key pair...")
		if err := run("ssh-keygen", "-t", "rsa", "-P", "", "-f", privateKey); err != nil {
			fmt.Println("❌ Failed to generate SSH key.")
			return
		}
	} else {
		fmt.Println("✅ SSH key pair already exists.")
	}

	// 3. Ensure authorized_keys exists and append if needed
	pubKeyContent, err := os.ReadFile(publicKey)
	if err != nil {
		fmt.Printf("❌ Failed to read public key: %v\n", err)
		return
	}

	os.MkdirAll(sshDir, 0700)

	authFile, err := os.OpenFile(authKeys, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("❌ Failed to open authorized_keys: %v\n", err)
		return
	}
	defer authFile.Close()

	_, err = authFile.Write(pubKeyContent)
	if err != nil {
		fmt.Printf("❌ Failed to write to authorized_keys: %v\n", err)
		return
	}
	authFile.WriteString("\n")

	// 4. Set correct permissions
	fmt.Println("🔐 Setting permissions on authorized_keys...")
	if err := os.Chmod(authKeys, 0600); err != nil {
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

//Use snake case. only make necessary changes.