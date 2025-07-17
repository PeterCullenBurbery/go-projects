package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	// File path
	path := `C:\Users\Administrator\AppData\Roaming\SQL Developer\system24.3.1.347.1826\o.ide.14.1.2.0.42.240731.1054\dtcache.xml`

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("❌ File does not exist.")
			return
		}
		fmt.Printf("⚠️ Error accessing file: %v\n", err)
		return
	}

	// Print file properties
	fmt.Println("📄 File Properties:")
	fmt.Printf("📍 Path: %s\n", path)
	fmt.Printf("📏 Size: %d bytes\n", info.Size())
	fmt.Printf("🕒 Last Modified: %s\n", info.ModTime().Format(time.RFC3339))
	fmt.Printf("📁 Is Directory: %v\n", info.IsDir())
	fmt.Printf("🔒 Permissions: %v\n", info.Mode())
}