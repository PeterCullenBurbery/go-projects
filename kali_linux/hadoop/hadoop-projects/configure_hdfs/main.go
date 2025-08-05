package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Property struct {
	Name  string `xml:"name"`
	Value string `xml:"value"`
}

type Configuration struct {
	XMLName    xml.Name   `xml:"configuration"`
	Properties []Property `xml:"property"`
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func update_hdfs_site(hdfsSitePath string, nameDir, dataDir string) error {
	f, err := os.Open(hdfsSitePath)
	if err != nil {
		return err
	}
	defer f.Close()

	var conf Configuration
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
		Property{"dfs.namenode.name.dir", "file:" + nameDir},
		Property{"dfs.datanode.data.dir", "file:" + dataDir},
	)

	// Write back
	f2, err := os.Create(hdfsSitePath)
	if err != nil {
		return err
	}
	defer f2.Close()

	f2.WriteString(xml.Header)
	encoder := xml.NewEncoder(f2)
	encoder.Indent("", "  ")
	return encoder.Encode(conf)
}

func main() {
	fmt.Println("🔧 Configuring HDFS persistent directories...")

	home, _ := os.UserHomeDir()
	hadoopConf := filepath.Join(home, "hadoop", "etc", "hadoop")
	hdfsSite := filepath.Join(hadoopConf, "hdfs-site.xml")
	nameDir := filepath.Join(home, "hdfs", "namenode")
	dataDir := filepath.Join(home, "hdfs", "datanode")

	// Step 1: Update hdfs-site.xml
	fmt.Println("📝 Updating hdfs-site.xml...")
	if err := update_hdfs_site(hdfsSite, nameDir, dataDir); err != nil {
		fmt.Println("❌ Failed to update hdfs-site.xml:", err)
		return
	}

	// Step 2: Create directories
	fmt.Println("📁 Creating data directories...")
	if err := os.MkdirAll(nameDir, 0755); err != nil {
		fmt.Println("❌ Failed to create namenode dir:", err)
		return
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		fmt.Println("❌ Failed to create datanode dir:", err)
		return
	}

	// Step 3: Format NameNode
	fmt.Println("🧹 Formatting HDFS...")
	if err := run("hdfs", "namenode", "-format"); err != nil {
		fmt.Println("❌ Failed to format HDFS.")
		return
	}

	// Step 4: Restart HDFS
	fmt.Println("♻️ Restarting Hadoop DFS services...")
	if err := run("stop-dfs.sh"); err != nil {
		fmt.Println("⚠️ stop-dfs.sh failed")
	}
	if err := run("start-dfs.sh"); err != nil {
		fmt.Println("⚠️ start-dfs.sh failed")
		return
	}

	fmt.Println("✅ HDFS reconfigured with persistent storage.")
}