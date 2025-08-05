package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func run_command(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func append_to_shell_rc(files []string, lines string) error {
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("failed to open %s: %v", file, err)
			}
			defer f.Close()
			if _, err := f.WriteString("\n" + lines + "\n"); err != nil {
				return fmt.Errorf("failed to write to %s: %v", file, err)
			}
			fmt.Printf("✅ Updated %s\n", file)
		} else {
			fmt.Printf("⚠️  Skipped missing shell config: %s\n", file)
		}
	}
	return nil
}

func main() {
	hadoop_version := "3.3.6"
	hadoop_tar := fmt.Sprintf("hadoop-%s.tar.gz", hadoop_version)
	hadoop_url := fmt.Sprintf("https://downloads.apache.org/hadoop/common/hadoop-%s/%s", hadoop_version, hadoop_tar)
	home, _ := os.UserHomeDir()
	hadoop_extract_path := filepath.Join(home, "hadoop")

	fmt.Println("📥 Downloading Hadoop...")
	if err := run_command("wget", hadoop_url); err != nil {
		fmt.Println("❌ Failed to download Hadoop.")
		return
	}

	fmt.Println("📦 Extracting Hadoop...")
	if err := run_command("tar", "-xvzf", hadoop_tar); err != nil {
		fmt.Println("❌ Failed to extract Hadoop.")
		return
	}

	fmt.Println("📂 Moving Hadoop to ~/hadoop...")
	if err := os.Rename("hadoop-"+hadoop_version, hadoop_extract_path); err != nil {
		fmt.Println("❌ Failed to move Hadoop directory.")
		return
	}

	fmt.Println("🛠  Updating shell config files with Hadoop environment variables...")
	env_vars := fmt.Sprintf(`# Hadoop environment variables
export HADOOP_HOME=%s
export HADOOP_INSTALL=$HADOOP_HOME
export HADOOP_MAPRED_HOME=$HADOOP_HOME
export HADOOP_COMMON_HOME=$HADOOP_HOME
export HADOOP_HDFS_HOME=$HADOOP_HOME
export YARN_HOME=$HADOOP_HOME
export HADOOP_COMMON_LIB_NATIVE_DIR=$HADOOP_HOME/lib/native
export JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64
export PATH=$PATH:$HADOOP_HOME/sbin:$HADOOP_HOME/bin`, hadoop_extract_path)

	shell_rc_files := []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".zshrc"),
	}

	if err := append_to_shell_rc(shell_rc_files, env_vars); err != nil {
		fmt.Println("❌ Failed to update shell config files.")
		return
	}

	fmt.Println("✅ Hadoop downloaded and environment configured.")
	fmt.Println("📢 Run 'source ~/.zshrc' or 'source ~/.bashrc' or restart your terminal to apply the changes.")
}