package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type property struct {
	Name  string `xml:"name"`
	Value string `xml:"value"`
}

type configuration struct {
	XMLName    xml.Name   `xml:"configuration"`
	Properties []property `xml:"property"`
}

func run_command(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func update_hdfs_site(hdfs_site_path, name_dir, data_dir string) error {
	f, err := os.Open(hdfs_site_path)
	if err != nil {
		return err
	}
	defer f.Close()

	var conf configuration
	if err := xml.NewDecoder(f).Decode(&conf); err != nil {
		return err
	}

	// Filter out existing name.dir or data.dir
	filtered := conf.Properties[:0]
	for _, prop := range conf.Properties {
		if prop.Name != "dfs.namenode.name.dir" && prop.Name != "dfs.datanode.data.dir" {
			filtered = append(filtered, prop)
		}
	}
	conf.Properties = filtered

	// Add new properties
	conf.Properties = append(conf.Properties,
		property{"dfs.namenode.name.dir", "file:" + name_dir},
		property{"dfs.datanode.data.dir", "file:" + data_dir},
	)

	// Write back
	f2, err := os.Create(hdfs_site_path)
	if err != nil {
		return err
	}
	defer f2.Close()

	f2.WriteString(xml.Header)
	encoder := xml.NewEncoder(f2)
	encoder.Indent("", "  ")
	return encoder.Encode(conf)
}

func find_namenode_pid() (int, error) {
	cmd := exec.Command("jps")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "NameNode") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				return strconv.Atoi(fields[0])
			}
		}
	}
	return 0, nil
}

func main() {
	fmt.Println("🔧 Configuring HDFS persistent directories...")

	home, _ := os.UserHomeDir()
	hadoop_conf := filepath.Join(home, "hadoop", "etc", "hadoop")
	hdfs_site := filepath.Join(hadoop_conf, "hdfs-site.xml")
	name_dir := filepath.Join(home, "hdfs", "namenode")
	data_dir := filepath.Join(home, "hdfs", "datanode")

	// Step 1: Update hdfs-site.xml
	fmt.Println("📝 Updating hdfs-site.xml...")
	if err := update_hdfs_site(hdfs_site, name_dir, data_dir); err != nil {
		fmt.Println("❌ Failed to update hdfs-site.xml:", err)
		return
	}

	// Step 2: Create directories
	fmt.Println("📁 Creating data directories...")
	os.MkdirAll(name_dir, 0755)
	os.MkdirAll(data_dir, 0755)

	// Step 3: Check for running NameNode
	fmt.Println("🛑 Checking for running NameNode...")
	if pid, err := find_namenode_pid(); err == nil && pid > 0 {
		fmt.Printf("🔍 NameNode running with PID %d. Stopping services...\n", pid)
		_ = run_command("stop-dfs.sh")
		_ = run_command("stop-yarn.sh")
		_ = run_command("kill", "-9", fmt.Sprint(pid))
	}

	// Step 4: Remove stale pid file
	pid_file := "/tmp/hadoop-" + filepath.Base(home) + "-namenode.pid"
	fmt.Println("🧹 Removing stale pid file if it exists...")
	_ = os.Remove(pid_file)

	// Step 5: Format HDFS
	fmt.Println("🧹 Formatting HDFS...")
	if err := run_command("hdfs", "namenode", "-format"); err != nil {
		fmt.Println("❌ Failed to format HDFS.")
		return
	}

	// Step 6: Restart DFS
	fmt.Println("♻️ Restarting Hadoop DFS services...")
	_ = run_command("start-dfs.sh")

	fmt.Println("✅ HDFS reconfigured with persistent storage.")
}