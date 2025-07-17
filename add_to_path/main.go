package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/PeterCullenBurbery/go_functions_002/v4/system_management_functions"
)

func main() {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ Failed to get current directory: %v\n", err)
		return
	}

	// For the test, we explicitly use "." to simulate relative path input
	fmt.Printf("📂 Current directory: %s\n", cwd)
	err = system_management_functions.Add_to_path(".")
	if err != nil {
		fmt.Printf("❌ Add_to_path failed: %v\n", err)
	} else {
		fmt.Printf("✅ Successfully added %s to system PATH (via .)\n", filepath.Clean(cwd))
	}
}