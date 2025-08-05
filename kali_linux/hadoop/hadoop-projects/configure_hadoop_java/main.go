package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	home, _ := os.UserHomeDir()
	hadoop_env_path := filepath.Join(home, "hadoop", "etc", "hadoop", "hadoop-env.sh")
	backup_path := hadoop_env_path + ".bak"

	// Backup first
	if _, err := os.Stat(hadoop_env_path); err == nil {
		if err := os.Rename(hadoop_env_path, backup_path); err != nil {
			fmt.Printf("❌ Failed to create backup: %v\n", err)
			return
		}
		fmt.Printf("🔁 Backed up %s → %s\n", hadoop_env_path, backup_path)
	}

	inputFile, err := os.Open(backup_path)
	if err != nil {
		fmt.Printf("❌ Failed to open backup file: %v\n", err)
		return
	}
	defer inputFile.Close()

	outputFile, err := os.Create(hadoop_env_path)
	if err != nil {
		fmt.Printf("❌ Failed to create updated file: %v\n", err)
		return
	}
	defer outputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	writer := bufio.NewWriter(outputFile)

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
