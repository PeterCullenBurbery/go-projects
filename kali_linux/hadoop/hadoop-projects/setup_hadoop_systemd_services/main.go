package main

import (
	"fmt"
	"os"
	"os/exec"
)

const hdfsUnit = `[Unit]
Description=Hadoop HDFS (NameNode, DataNode, SecondaryNameNode)
After=network.target
Requires=network.target

[Service]
User=%s
Group=%s
Environment=HADOOP_HOME=${HOME}/hadoop
Environment=JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64
ExecStart=${HOME}/hadoop/sbin/start-dfs.sh
ExecStop=${HOME}/hadoop/sbin/stop-dfs.sh
Restart=on-failure
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
`

const yarnUnit = `[Unit]
Description=Hadoop YARN (ResourceManager, NodeManager)
After=network.target hadoop-hdfs.service
Requires=hadoop-hdfs.service

[Service]
User=%s
Group=%s
Environment=HADOOP_HOME=${HOME}/hadoop
Environment=JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64
ExecStart=${HOME}/hadoop/sbin/start-yarn.sh
ExecStop=${HOME}/hadoop/sbin/stop-yarn.sh
Restart=on-failure
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
`

func writeUnitFile(name, content string) error {
	path := "/etc/systemd/system/" + name
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	user := os.Getenv("SUDO_USER")
	if user == "" {
		user = os.Getenv("USER")
	}

	if user == "" {
		fmt.Println("❌ Could not determine username. Run with sudo or set USER environment variable.")
		return
	}

	fmt.Println("🔧 Writing systemd unit files...")
	hdfsContent := fmt.Sprintf(hdfsUnit, user, user)
	yarnContent := fmt.Sprintf(yarnUnit, user, user)

	if err := writeUnitFile("hadoop-hdfs.service", hdfsContent); err != nil {
		fmt.Println("❌ Failed to write hadoop-hdfs.service:", err)
		return
	}
	if err := writeUnitFile("hadoop-yarn.service", yarnContent); err != nil {
		fmt.Println("❌ Failed to write hadoop-yarn.service:", err)
		return
	}

	fmt.Println("🔄 Reloading systemd...")
	if err := runCommand("systemctl", "daemon-reload"); err != nil {
		fmt.Println("❌ Failed to reload systemd:", err)
		return
	}

	fmt.Println("📌 Enabling hadoop-hdfs.service...")
	_ = runCommand("systemctl", "enable", "hadoop-hdfs.service")

	fmt.Println("📌 Enabling hadoop-yarn.service...")
	_ = runCommand("systemctl", "enable", "hadoop-yarn.service")

	fmt.Println("✅ Done! Services will start automatically after reboot.")
}