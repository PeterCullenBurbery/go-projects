package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

func run_command(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func resolve_hadoop_sbin() (string, error) {
	hadoop_home := os.Getenv("HADOOP_HOME")
	if hadoop_home == "" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		hadoop_home = filepath.Join(usr.HomeDir, "hadoop")
	}
	return filepath.Join(hadoop_home, "sbin"), nil
}

func main() {
	fmt.Println("🛑 Stopping Hadoop services...")

	hadoop_sbin, err := resolve_hadoop_sbin()
	if err != nil {
		fmt.Println("❌ Failed to resolve Hadoop sbin path:", err)
		return
	}

	stop_dfs := filepath.Join(hadoop_sbin, "stop-dfs.sh")
	stop_yarn := filepath.Join(hadoop_sbin, "stop-yarn.sh")

	fmt.Println("📴 Running stop-dfs.sh...")
	if err := run_command(stop_dfs); err != nil {
		fmt.Println("❌ Failed to stop HDFS:", err)
	}

	fmt.Println("📴 Running stop-yarn.sh...")
	if err := run_command(stop_yarn); err != nil {
		fmt.Println("❌ Failed to stop YARN:", err)
	}

	fmt.Println("✅ Hadoop services have been stopped.")
}