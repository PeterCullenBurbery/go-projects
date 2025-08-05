package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	home_dir, _ := os.UserHomeDir()
	hadoop_env_path := filepath.Join(home_dir, "hadoop", "etc", "hadoop", "hadoop-env.sh")
	backup_path := hadoop_env_path + ".bak"

	// Backup first
	if _, err := os.Stat(hadoop_env_path); err == nil {
		if err := os.Rename(hadoop_env_path, backup_path); err != nil {
			fmt.Printf("❌ Failed to create backup: %v\n", err)
			return
		}
		fmt.Printf("🔁 Backed up %s → %s\n", hadoop_env_path, backup_path)
	}

	input_file, err := os.Open(backup_path)
	if err != nil {
		fmt.Printf("❌ Failed to open backup file: %v\n", err)
		return
	}
	defer input_file.Close()

	output_file, err := os.Create(hadoop_env_path)
	if err != nil {
		fmt.Printf("❌ Failed to create updated file: %v\n", err)
		return
	}
	defer output_file.Close()

	scanner := bufio.NewScanner(input_file)
	writer := bufio.NewWriter(output_file)

	java_home_set := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "# export JAVA_HOME=") && !java_home_set {
			writer.WriteString("export JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64\n")
			java_home_set = true
		} else {
			writer.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("❌ Error reading backup: %v\n", err)
		return
	}

	writer.Flush()
	fmt.Println("✅ JAVA_HOME has been set in hadoop-env.sh.")
}