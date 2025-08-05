package main

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

// get_zshrc_path determines the path to the .zshrc file using $ZDOTDIR or $HOME
func get_zshrc_path() (string, error) {
	z_dot_dir := os.Getenv("ZDOTDIR")
	if z_dot_dir != "" {
		return filepath.Join(z_dot_dir, ".zshrc"), nil
	}

	current_user, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(current_user.HomeDir, ".zshrc"), nil
}

// parse_aliases reads the file at zshrc_path and prints all alias definitions
func parse_aliases(zshrc_path string) error {
	file, err := os.Open(zshrc_path)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Match alias name='value', alias name="value", or alias name=value
	alias_regex := regexp.MustCompile(`^alias\s+(\S+)=['"]?(.*?)['"]?$`)

	fmt.Printf("Reading aliases from: %s\n\n", zshrc_path)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if alias_regex.MatchString(line) {
			matches := alias_regex.FindStringSubmatch(line)
			if len(matches) == 3 {
				alias_name := matches[1]
				alias_command := matches[2]
				fmt.Printf("%-25s → %s\n", alias_name, alias_command)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}

func main() {
	zshrc_path, err := get_zshrc_path()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to determine .zshrc path: %v\n", err)
		os.Exit(1)
	}

	if err := parse_aliases(zshrc_path); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse aliases: %v\n", err)
		os.Exit(1)
	}
}