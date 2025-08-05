package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

type Property struct {
	XMLName xml.Name `xml:"property"`
	Name    string   `xml:"name"`
	Value   string   `xml:"value"`
}

type Configuration struct {
	XMLName   xml.Name   `xml:"configuration"`
	Properties []Property `xml:"property"`
}

func backup_file(path string) error {
	backup_path := path + ".bak"
	if _, err := os.Stat(path); err == nil {
		err = os.Rename(path, backup_path)
		if err != nil {
			return fmt.Errorf("failed to create backup: %v", err)
		}
		fmt.Printf("🔁 Backed up %s → %s\n", path, backup_path)
	} else {
		fmt.Printf("⚠️ No original file found to backup: %s (creating new)\n", path)
	}
	return nil
}

func write_config_file(path string, config Configuration) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write XML with indentation and header
	file.WriteString(xml.Header)
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	return encoder.Encode(config)
}

func build_config(properties []Property) Configuration {
	return Configuration{
		Properties: properties,
	}
}

func main() {
	home, _ := os.UserHomeDir()
	config_dir := filepath.Join(home, "hadoop", "etc", "hadoop")

	// Define all the config files and their properties
	configs := map[string][]Property{
		"core-site.xml": {
			{Name: "fs.defaultFS", Value: "hdfs://localhost:9000"},
		},
		"hdfs-site.xml": {
			{Name: "dfs.replication", Value: "1"},
		},
		"mapred-site.xml": {
			{Name: "mapreduce.framework.name", Value: "yarn"},
		},
		"yarn-site.xml": {
			{Name: "yarn.nodemanager.aux-services", Value: "mapreduce_shuffle"},
		},
	}

	fmt.Println("🔧 Updating Hadoop XML configuration files using XML encoding...")

	for filename, properties := range configs {
		full_path := filepath.Join(config_dir, filename)
		fmt.Printf("📄 Processing %s...\n", full_path)

		if err := backup_file(full_path); err != nil {
			fmt.Printf("❌ Backup failed: %v\n", err)
			continue
		}

		config := build_config(properties)

		if err := write_config_file(full_path, config); err != nil {
			fmt.Printf("❌ Failed to write XML: %v\n", err)
			continue
		}
	}

	fmt.Println("✅ Hadoop config files updated with structured XML.")
}