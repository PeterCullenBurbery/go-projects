package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	longPath := `C:\long-file-paths\mSKuGP9gFX\ZfioqxCduF\EewaJDhlYK\8r4kgHrLFS\8EV918wFvX\x64D0YHG3G\k77o7QCq0x\DwEedEczG5\4gY9gRh8J0\aR9t5gANa8\8MgHxqVUZp\fnVWu6f1Th\IqfNeTCn6Y\0Qw1kIulol\ahehDwqU20\OtvSoKCOjL\1IkjZhqGLf\GRfIjrZjiZ\W0CINpwjlT\x2u1v5DVn2\9d7TFoMCcs\cgpwIxGtd1\OHTmuXve1s\hello_world.exe`

	fmt.Println("== Path Information ==")
	fmt.Printf("Long path : %s\n", longPath)
	fmt.Printf("Length    : %d characters\n", len(longPath))
	fmt.Println()

	cmd := exec.Command("cmd.exe", "/C", longPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Execution via cmd.exe failed: %v\n", err)
		os.Exit(1)
	}
}