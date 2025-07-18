package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func main() {
	longPath := `\\?\C:\long-file-paths\mSKuGP9gFX\ZfioqxCduF\EewaJDhlYK\8r4kgHrLFS\8EV918wFvX\x64D0YHG3G\k77o7QCq0x\DwEedEczG5\4gY9gRh8J0\aR9t5gANa8\8MgHxqVUZp\fnVWu6f1Th\IqfNeTCn6Y\0Qw1kIulol\ahehDwqU20\OtvSoKCOjL\1IkjZhqGLf\GRfIjrZjiZ\W0CINpwjlT\x2u1v5DVn2\9d7TFoMCcs\cgpwIxGtd1\OHTmuXve1s\hello_world.exe`

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