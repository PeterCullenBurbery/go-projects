package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	dfs_unit = `[Unit]
Description=Hadoop Distributed File System (HDFS)
After=network.target

[Service]
Type=forking
User=%s
Group=%s
Environment=HADOOP_HOME=%s/hadoop
Environment=JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64
ExecStart=%s/hadoop/sbin/start-dfs.sh
ExecStop=%s/hadoop/sbin/stop-dfs.sh
Restart=on-failure
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
`

	yarn_unit = `[Unit]
Description=Hadoop Yet Another Resource Negotiator (YARN)
After=network.target hadoop-dfs.service
Requires=hadoop-dfs.service

[Service]
Type=forking
User=%s
Group=%s
Environment=HADOOP_HOME=%s/hadoop
Environment=JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64
ExecStart=%s/hadoop/sbin/start-yarn.sh
ExecStop=%s/hadoop/sbin/stop-yarn.sh
Restart=on-failure
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
`
)

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func write_unit(name, content string) error {
	unit_path := filepath.Join("/etc/systemd/system", name)
	return os.WriteFile(unit_path, []byte(content), 0644)
}

func main() {
	fmt.Println("🛠️ Installing Hadoop systemd services...")

	// Determine user and home
	user := os.Getenv("SUDO_USER")
	if user == "" {
		user = os.Getenv("USER")
	}
	if user == "" {
		fmt.Println("❌ Could not determine user. Run with sudo.")
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("❌ Could not determine home directory:", err)
		return
	}

	// Write DFS unit
	fmt.Println("📄 Writing hadoop-dfs.service...")
	dfs_content := fmt.Sprintf(dfs_unit, user, user, home, home, home)
	if err := write_unit("hadoop-dfs.service", dfs_content); err != nil {
		fmt.Println("❌ Failed to write hadoop-dfs.service:", err)
		return
	}

	// Write YARN unit
	fmt.Println("📄 Writing hadoop-yarn.service...")
	yarn_content := fmt.Sprintf(yarn_unit, user, user, home, home, home)
	if err := write_unit("hadoop-yarn.service", yarn_content); err != nil {
		fmt.Println("❌ Failed to write hadoop-yarn.service:", err)
		return
	}

	// Reload systemd
	fmt.Println("🔄 Reloading systemd daemon...")
	if err := run("systemctl", "daemon-reload"); err != nil {
		fmt.Println("❌ Failed to reload systemd:", err)
		return
	}

	// Enable and start services
	fmt.Println("✅ Enabling and starting hadoop-dfs.service...")
	_ = run("systemctl", "enable", "--now", "hadoop-dfs.service")

	fmt.Println("✅ Enabling and starting hadoop-yarn.service...")
	_ = run("systemctl", "enable", "--now", "hadoop-yarn.service")

	fmt.Println("🎉 Hadoop HDFS and YARN are now installed as systemd services, enabled on boot, and started.")
}