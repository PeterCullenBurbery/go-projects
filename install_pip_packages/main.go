package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("❌ Usage: install_python_packages.exe <path-to-configuration-003>")
		os.Exit(1)
	}

	configuration_path := os.Args[1]
	yaml_file_path := filepath.Join(configuration_path, "python-packages.yaml")

	yaml_file, err := os.Open(yaml_file_path)
	if err != nil {
		log.Fatalf("❌ Could not open %s: %v", yaml_file_path, err)
	}
	defer yaml_file.Close()

	scanner := bufio.NewScanner(yaml_file)
	var package_list []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "-") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				package_list = append(package_list, fields[1])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("❌ Failed reading file: %v", err)
	}

	if len(package_list) == 0 {
		log.Println("ℹ️ No Python packages found in YAML.")
		return
	}

	pip_args := append([]string{"install"}, package_list...)
	pip_cmd := exec.Command("pip", pip_args...)
	pip_cmd.Stdout = os.Stdout
	pip_cmd.Stderr = os.Stderr

	log.Printf("🐍 Installing Python packages: %v\n", package_list)
	if err := pip_cmd.Run(); err != nil {
		log.Fatalf("❌ pip install failed: %v", err)
	}

	log.Println("✅ Python packages installation completed.")
}