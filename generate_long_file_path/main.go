package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	base_path          = `C:\long-file-paths`
	alphanumeric_chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	random_name_length = 10
	max_path_length    = 260
	module_name        = "example.com/deep/project"
	exe_name           = "hello_world.exe"
)

func generate_random_string(length int) string {
	var builder strings.Builder
	for i := 0; i < length; i++ {
		random_index, _ := rand.Int(rand.Reader, big.NewInt(int64(len(alphanumeric_chars))))
		builder.WriteByte(alphanumeric_chars[random_index.Int64()])
	}
	return builder.String()
}

func copy_file(src, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = io.Copy(output, input)
	return err
}

func main() {
	current_path := base_path

	for len(current_path) <= max_path_length {
		random_folder_name := generate_random_string(random_name_length)
		current_path = filepath.Join(current_path, random_folder_name)
	}

	fmt.Printf("Creating directory path: %s\n", current_path)
	err := os.MkdirAll(current_path, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}

	// Step 1: Create Go project in a temp short path
	temp_path, err := os.MkdirTemp("", "short_goproject")
	if err != nil {
		log.Fatalf("Failed to create temporary directory: %v", err)
	}

	// go.mod
	cmd := exec.Command("go", "mod", "init", module_name)
	cmd.Dir = temp_path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to run 'go mod init': %v", err)
	}

	// main.go
	main_go_content := `package main

import "fmt"

func main() {
	fmt.Println("Hello from the deep nested project!")
}
`
	main_go_path := filepath.Join(temp_path, "main.go")
	err = os.WriteFile(main_go_path, []byte(main_go_content), 0644)
	if err != nil {
		log.Fatalf("Failed to write main.go: %v", err)
	}

	// go build
	cmd = exec.Command("go", "build", "-o", exe_name)
	cmd.Dir = temp_path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to run 'go build': %v", err)
	}

	// Copy files to long path
	files := []string{"go.mod", "main.go", exe_name}
	for _, file := range files {
		src := filepath.Join(temp_path, file)
		dst := filepath.Join(current_path, file)
		err = copy_file(src, dst)
		if err != nil {
			log.Fatalf("Failed to copy %s to long path: %v", file, err)
		}
	}

	fmt.Printf("Go project built and copied to:\n%s\n", current_path)
}
