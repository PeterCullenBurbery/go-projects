package main

import (
	"log"

	"github.com/PeterCullenBurbery/go_functions_002/v4/system_management_functions"
)

func main() {
	path := `C:\Users\Administrator\Downloads`

	err := system_management_functions.Exclude_from_Microsoft_Windows_Defender(path)
	if err != nil {
		log.Fatalf("❌ Failed to add exclusion: %v", err)
	}
}
