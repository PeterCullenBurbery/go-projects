package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

// Expand_windows_env expands environment variables using the Windows API.
// For example, %SystemRoot% becomes C:\Windows.
func Expand_windows_env(input string) string {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procExpand := kernel32.NewProc("ExpandEnvironmentStringsW")

	inputPtr, _ := syscall.UTF16PtrFromString(input)
	buf := make([]uint16, 32767) // MAX_PATH

	ret, _, _ := procExpand.Call(
		uintptr(unsafe.Pointer(inputPtr)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)

	if ret == 0 {
		return input // fallback if expansion fails
	}

	return syscall.UTF16ToString(buf[:ret])
}

// Add_to_path adds the given path to the top of the system PATH (HKLM) if not already present.
// It expands environment variables, removes redundant entries (like %SystemRoot%), avoids duplicates,
// and broadcasts the environment change to Explorer. It also prints PowerShell instructions to refresh the session.
func Add_to_path(path_to_add string) error {
	fmt.Printf("🔧 Input path: %s\n", path_to_add)

	// Step 1: Resolve absolute path
	abs_path, err := filepath.Abs(path_to_add)
	if err != nil {
		return fmt.Errorf("❌ Failed to resolve absolute path: %w", err)
	}
	fmt.Printf("📁 Absolute path: %s\n", abs_path)

	info, err := os.Stat(abs_path)
	if err != nil {
		return fmt.Errorf("❌ Failed to stat path: %w", err)
	}
	if !info.IsDir() {
		abs_path = filepath.Dir(abs_path)
	}
	normalized := strings.TrimRight(abs_path, `\`)
	fmt.Printf("🧹 Normalized path: %s\n", normalized)

	// Step 2: Open system PATH from registry
	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("❌ Failed to open registry key: %w", err)
	}
	defer key.Close()
	fmt.Println("🔑 Opened HKLM system environment registry key.")

	raw_path, _, err := key.GetStringValue("Path")
	if err != nil {
		return fmt.Errorf("❌ Failed to read PATH: %w", err)
	}
	fmt.Println("📍 Current PATH (raw):")
	fmt.Println(raw_path)

	// Step 3: Process PATH entries
	entries := strings.Split(raw_path, ";")
	fmt.Println("🔍 Checking each existing PATH entry against target:")

	normalized_lower := strings.ToLower(normalized)
	already_exists := false
	seen := make(map[string]bool)
	rebuilt := []string{normalized} // New path goes first
	seen[normalized_lower] = true

	for _, entry := range entries {
		entry_trimmed := strings.TrimSpace(strings.TrimRight(entry, `\`))
		if entry_trimmed == "" {
			continue
		}

		expanded := strings.TrimRight(Expand_windows_env(entry_trimmed), `\`)
		lower_expanded := strings.ToLower(expanded)

		if !strings.EqualFold(entry_trimmed, expanded) {
			fmt.Printf("   - Original: %-70s →  Expanded: %s\n", entry_trimmed, expanded)
		}

		if lower_expanded == normalized_lower {
			already_exists = true
		}

		if !seen[lower_expanded] {
			rebuilt = append(rebuilt, expanded)
			seen[lower_expanded] = true
		}
	}

	if already_exists {
		fmt.Println("✅ Path already present in system PATH (via expanded match).")
		return nil
	}

	new_path := strings.Join(rebuilt, ";")
	fmt.Println("🧩 New PATH to set in registry:")
	fmt.Println(new_path)

	// Step 4: Write back to registry
	if err := key.SetStringValue("Path", new_path); err != nil {
		return fmt.Errorf("❌ Failed to update PATH in registry: %w", err)
	}
	fmt.Println("✅ Path added to the top of system PATH.")

	// Step 5: Broadcast change
	const (
		HWND_BROADCAST   = 0xffff
		WM_SETTINGCHANGE = 0x001A
		SMTO_ABORTIFHUNG = 0x0002
	)
	user32 := syscall.NewLazyDLL("user32.dll")
	procSendMessageTimeout := user32.NewProc("SendMessageTimeoutW")

	ret, _, _ := procSendMessageTimeout.Call(
		uintptr(HWND_BROADCAST),
		uintptr(WM_SETTINGCHANGE),
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Environment"))),
		uintptr(SMTO_ABORTIFHUNG),
		5000,
		uintptr(0),
	)
	if ret == 0 {
		fmt.Println("⚠️ Environment change broadcast may have failed.")
	} else {
		fmt.Println("📢 Environment update broadcast sent.")
	}

	// Step 6: Check for refreshenv and print accordingly
	if _, err := exec.LookPath("refreshenv"); err == nil {
		fmt.Println("♻️  'refreshenv' is available. To update this session, run:")
		fmt.Println("    refreshenv")
	} else {
		fmt.Println("ℹ️  'refreshenv' not available in this session.")
	}

	return nil
}

func main() {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ Failed to get current directory: %v\n", err)
		return
	}

	// For the test, we explicitly use "." to simulate relative path input
	fmt.Printf("📂 Current directory: %s\n", cwd)
	err = Add_to_path(".")
	if err != nil {
		fmt.Printf("❌ Add_to_path failed: %v\n", err)
	} else {
		fmt.Printf("✅ Successfully added %s to system PATH (via .)\n", filepath.Clean(cwd))
	}
}
