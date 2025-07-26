package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/PeterCullenBurbery/go_functions_002/v4/system_management_functions"
)

func main() {
	// Define download targets
	files := []struct {
		name string
		url  string
	}{
		{
			name: "instantclient-basic-windows.x64-23.8.0.25.04.zip",
			url:  "https://download.oracle.com/otn_software/nt/instantclient/2380000/instantclient-basic-windows.x64-23.8.0.25.04.zip",
		},
		{
			name: "instantclient-sdk-windows.x64-23.8.0.25.04.zip",
			url:  "https://download.oracle.com/otn_software/nt/instantclient/2380000/instantclient-sdk-windows.x64-23.8.0.25.04.zip",
		},
	}

	// Set working directories
	zipDir := `C:\downloads\oracle_instant_client\zips`
	extractDir := `C:\downloads\oracle_instant_client\instantclient_23_8`

	// Ensure zip and extract directories exist
	if err := os.MkdirAll(zipDir, 0755); err != nil {
		log.Fatalf("❌ Could not create zipDir: %v", err)
	}
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		log.Fatalf("❌ Could not create extractDir: %v", err)
	}

	// Download and extract each file
	for _, file := range files {
		zipPath := filepath.Join(zipDir, file.name)
		fmt.Printf("📥 Downloading %s...\n", file.url)

		// Download
		if err := system_management_functions.Download_file(zipPath, file.url); err != nil {
			log.Fatalf("❌ Failed to download %s: %v", file.name, err)
		}
		fmt.Printf("✅ Downloaded: %s\n", zipPath)

		// Extract
		fmt.Printf("📦 Extracting %s...\n", zipPath)
		if err := system_management_functions.Extract_zip(zipPath, extractDir); err != nil {
			log.Fatalf("❌ Failed to extract %s: %v", file.name, err)
		}
		fmt.Printf("✅ Extracted to: %s\n", extractDir)
	}

	fmt.Println("🎉 Oracle Instant Client Basic + SDK downloaded and extracted successfully.")
}