package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// Exclude_from_Microsoft_Windows_Defender attempts to exclude the specified file or folder
// from Microsoft Defender's real-time protection.
//
// This function first checks whether the Microsoft Defender Antivirus service (WinDefend)
// is currently running. If it is not running (e.g., disabled by policy or replaced by another
// antivirus solution), the function skips the exclusion step without error.
//
// If the provided path refers to a file, its parent directory is excluded instead.
// If the path refers to a directory, it is used directly.
//
// This operation requires administrative privileges if Microsoft Defender is enabled.
//
// Parameters:
//   - path_to_exclude: The absolute or relative path to a file or folder to exclude.
//
// Returns:
//   - nil if the exclusion was successful, unnecessary (because Defender is not running),
//     or if the path was already excluded.
//   - An error if any part of the exclusion process fails (e.g., bad path, PowerShell failure).
//
// Example:
//
//      err := Exclude_from_Microsoft_Windows_Defender("C:\\downloads\\nirsoft")
//      if err != nil {
//          log.Fatalf("Failed to exclude: %v", err)
//      }
func Exclude_from_Microsoft_Windows_Defender(path_to_exclude string) error {
        // Step 0: Check if Microsoft Defender is running
        check_cmd := exec.Command("powershell", "-NoProfile", "-Command",
                `(Get-Service WinDefend).Status`)
        output_bytes, err := check_cmd.Output()
        if err != nil {
                log.Println("ℹ️ Unable to query WinDefend service; skipping exclusion step.")
                return nil
        }
        output := string(output_bytes)
        if output != "Running\r\n" && output != "Running\n" {
                log.Println("ℹ️ Microsoft Defender is not running; skipping exclusion step.")
                return nil
        }

        // Resolve absolute path
        absolute_path, err := filepath.Abs(path_to_exclude)
        if err != nil {
                return fmt.Errorf("❌ Failed to resolve absolute path: %w", err)
        }

        // Stat to determine if it's a file or folder
        file_info, err := os.Stat(absolute_path)
        if err != nil {
                return fmt.Errorf("❌ Failed to stat path: %w", err)
        }

        // If it's a file, get parent directory
        if !file_info.IsDir() {
                absolute_path = filepath.Dir(absolute_path)
        }

        // Normalize (trim trailing slash)
        normalized_path := filepath.Clean(absolute_path)

        // Build PowerShell command to exclude from Defender
        exclude_cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command",
                fmt.Sprintf(`Add-MpPreference -ExclusionPath "%s"`, normalized_path))

        exclude_output_bytes, err := exclude_cmd.CombinedOutput()
        if err != nil {
                return fmt.Errorf("❌ Failed to exclude from Defender: %w\nOutput: %s", err, string(exclude_output_bytes))
        }

        fmt.Printf("✅ Excluded from Microsoft Defender: %s\n", normalized_path)
        return nil
}

func main() {
	// Create a test folder inside the system temp directory
	temp_dir := filepath.Join(os.TempDir(), "test_defender_exclude")

	// Ensure the directory exists
	if err := os.MkdirAll(temp_dir, 0755); err != nil {
		log.Fatalf("❌ Failed to create temp test directory: %v", err)
	}

	log.Printf("🧪 Testing exclusion for: %s", temp_dir)

	// Call the exclusion function
	err := Exclude_from_Microsoft_Windows_Defender(temp_dir)
	if err != nil {
		log.Fatalf("❌ Exclusion test failed: %v", err)
	}

	log.Println("✅ Exclusion test completed successfully.")
}
