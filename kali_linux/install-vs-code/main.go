package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func must_run(command string, args ...string) {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive") // suppress interactive prompts
	err := cmd.Run()
	if err != nil {
		log.Fatalf("❌ Failed to run: %s %v\nError: %v", command, args, err)
	}
}

func check_root() {
	if os.Geteuid() != 0 {
		log.Fatalln("❌ This program must be run as root. Use sudo.")
	}
}

func main() {
	check_root()

	deb_url := "https://code.visualstudio.com/sha/download?build=stable&os=linux-deb-x64"
	deb_file := "vscode.deb"

	fmt.Println("📥 Downloading Visual Studio Code...")
	must_run("wget", "-O", deb_file, deb_url)

	fmt.Println("📦 Installing Visual Studio Code non-interactively...")
	must_run("apt", "install", "-y", "./"+deb_file)

	fmt.Println("🧹 Cleaning up...")
	err := os.Remove(deb_file)
	if err != nil {
		log.Printf("⚠️ Could not delete temporary file: %v\n", err)
	}

	fmt.Println("✅ VS Code installation completed successfully!")
	fmt.Println("🔁 Run it using: code")
}
