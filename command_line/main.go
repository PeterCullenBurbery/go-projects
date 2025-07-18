package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func main() {
	longPath := `\\?\C:\long-file-paths\SLALHxvTze\RRcnGFMT4b\D405pydQKV\5Pgy3yE4W7\o0s8CF942d\aO16Hmnkox\IV5rfDe3C9\Qw0MvV4WDb\W9tFG7ADJC\SrKGaCwRNQ\jxKIMaMkNS\AjR6EMdoS0\MuMgEU92Df\FZwd2QGuW2\dr06QAboY0\aw14j7DjJQ\bkUZqOXRl2\xhxwKiGiDO\bMiOKX9BCD\nnblvnjad2\Lykd8bASVA\It0zFMbCF9\tM0FrJsgfz\hello_world.exe`

	fmt.Println("== Path Information ==")
	fmt.Printf("Path   : %s\n", longPath)
	fmt.Printf("Length : %d characters\n\n", len(longPath))

	// Convert path and command line to UTF-16
	appNameUTF16, _ := syscall.UTF16PtrFromString(longPath)
	cmdLineUTF16, _ := syscall.UTF16PtrFromString(`"` + longPath + `"`)

	// Setup process startup info
	var si syscall.StartupInfo
	var pi syscall.ProcessInformation
	si.Cb = uint32(unsafe.Sizeof(si))

	// Call CreateProcessW
	err := syscall.CreateProcess(
		appNameUTF16,     // lpApplicationName
		cmdLineUTF16,     // lpCommandLine (quoted)
		nil,              // lpProcessAttributes
		nil,              // lpThreadAttributes
		false,            // bInheritHandles
		0,                // dwCreationFlags
		nil,              // lpEnvironment
		nil,              // lpCurrentDirectory
		&si,              // lpStartupInfo
		&pi,              // lpProcessInformation
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "CreateProcessW failed: %v\n", err)
		os.Exit(1)
	}

	// Wait for the process to finish
	_, _ = syscall.WaitForSingleObject(pi.Process, syscall.INFINITE)

	// Close handles
	_ = syscall.CloseHandle(pi.Process)
	_ = syscall.CloseHandle(pi.Thread)

	fmt.Println("Execution completed.")
}