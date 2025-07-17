package main

import (
	"log"

	"github.com/PeterCullenBurbery/go_functions_002/v4/system_management_functions"
)

func main() {
	exePath := `C:\downloads\sql-developer\2025-007-011 017.040.040.768691900 America slash New_York 2025-W028-005 2025-192\sqldeveloper\sqldeveloper.exe`

	log.Println("🔗 Creating desktop shortcut using Create_desktop_shortcut from v4...")
	err := system_management_functions.Create_desktop_shortcut(
		exePath,
		"SQL Developer.lnk",       // Shortcut file name
		"Oracle SQL Developer",    // Description
		3,                         // Window style (3 = Maximized)
		true,                      // All users (requires admin)
	)
	if err != nil {
		log.Fatalf("❌ Failed to create shortcut: %v", err)
	}

	log.Println("✅ Shortcut successfully created.")
}