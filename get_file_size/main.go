// Package main provides a utility to calculate the size of a file or directory.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Get_file_size returns the size in bytes of the specified path.
// If the path is a regular file, its size is returned directly.
// If the path is a directory, the function walks through all files
// and returns the cumulative size of all non-directory files within it.
//
// Parameters:
//   - path: The path to the file or directory.
//
// Returns:
//   - int64: Total size in bytes.
//   - error: Any error encountered while accessing the file system.
//
// Example:
//   size, err := Get_file_size("C:\\Users\\Administrator\\Desktop")
//   if err != nil {
//       log.Fatal(err)
//   }
//   fmt.Println("Total size:", size)
func Get_file_size(path string) (int64, error) {
	var totalSize int64

	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	if !info.IsDir() {
		return info.Size(), nil
	}

	err = filepath.Walk(path, func(_ string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fileInfo.IsDir() {
			totalSize += fileInfo.Size()
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return totalSize, nil
}

// main is the entry point of the program.
// It demonstrates the usage of Get_file_size by printing
// the size of a predefined path.
func main() {
	path := `C:\Users\Administrator\Desktop\GitHub-repositories\configuration-003\go_projects\configuration\apps\configure_keyboard_shortcuts_for_vs_code\main.go`
	size, err := Get_file_size(path)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	fmt.Printf("📦 Total size of '%s': %d bytes\n", path, size)
}