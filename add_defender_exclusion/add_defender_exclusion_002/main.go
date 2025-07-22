package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/PeterCullenBurbery/go_functions_002/v4/system_management_functions"
)

func main() {
	// Create a test folder inside the system temp directory
	temp_dir := filepath.Join(os.TempDir(), "test_defender_exclude")

	// Ensure the directory exists
	if err := os.MkdirAll(temp_dir, 0755); err != nil {
		log.Fatalf("❌ Failed to create temp test directory: %v", err)
	}

	log.Printf("🧪 Testing exclusion for: %s", temp_dir)

	// Call the exclusion function
	err := system_management_functions.Exclude_from_Microsoft_Windows_Defender(temp_dir)
	if err != nil {
		log.Fatalf("❌ Exclusion test failed: %v", err)
	}

	log.Println("✅ Exclusion test completed successfully.")
}