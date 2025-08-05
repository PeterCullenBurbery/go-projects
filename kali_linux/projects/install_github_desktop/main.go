package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const (
	download_path = "/tmp/github-desktop.deb"
	api_url       = "https://api.github.com/repos/shiftkey/desktop/releases/latest"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func main() {
	if _, err := exec.LookPath("dpkg"); err != nil {
		fmt.Println("This script only works on Debian-based systems with dpkg.")
		return
	}

	fmt.Println("Fetching latest GitHub Desktop release info...")
	release, err := get_latest_release()
	if err != nil {
		fmt.Printf("Failed to get release: %v\n", err)
		return
	}

	var deb_url string
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, ".deb") {
			deb_url = asset.BrowserDownloadURL
			break
		}
	}

	if deb_url == "" {
		fmt.Println("No .deb package found in the latest release.")
		return
	}

	fmt.Printf("Downloading: %s\n", deb_url)
	err = run_command("curl", "-L", "-o", download_path, deb_url)
	if err != nil {
		fmt.Printf("Download failed: %v\n", err)
		return
	}

	fmt.Println("Installing GitHub Desktop...")
	err = run_command("sudo", "dpkg", "-i", download_path)
	if err != nil {
		fmt.Printf("dpkg install failed: %v\n", err)
		fmt.Println("Attempting to fix dependencies...")
		_ = run_command("sudo", "apt", "-f", "install", "-y")
	}

	fmt.Println("GitHub Desktop installation complete.")
}

func get_latest_release() (*GitHubRelease, error) {
	resp, err := http.Get(api_url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var release GitHubRelease
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return nil, err
	}
	return &release, nil
}

func run_command(name string, args ...string) error {
	fmt.Printf("Running: %s %s\n", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
