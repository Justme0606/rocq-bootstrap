package vscode

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const ExtensionID = "rocq-prover.vsrocq"

// FindCode searches for the VSCode CLI executable on macOS.
func FindCode() (string, error) {
	// 1. Try PATH first
	path, err := exec.LookPath("code")
	if err == nil {
		return path, nil
	}

	// 2. Standard macOS app bundle location
	appBundlePath := "/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code"
	if info, err := os.Stat(appBundlePath); err == nil && !info.IsDir() {
		return appBundlePath, nil
	}

	// 3. User Applications folder
	home, _ := os.UserHomeDir()
	if home != "" {
		userAppPath := filepath.Join(home, "Applications/Visual Studio Code.app/Contents/Resources/app/bin/code")
		if info, err := os.Stat(userAppPath); err == nil && !info.IsDir() {
			return userAppPath, nil
		}
	}

	// 4. Homebrew cask location
	brewPaths := []string{
		"/opt/homebrew/bin/code",
		"/usr/local/bin/code",
	}
	for _, p := range brewPaths {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p, nil
		}
	}

	return "", fmt.Errorf("VSCode (code) not found in PATH or common locations")
}

// InstallExtension installs the vsrocq extension if not already present.
func InstallExtension(codeBin string) error {
	// Check if already installed
	out, err := exec.Command(codeBin, "--list-extensions").Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.EqualFold(strings.TrimSpace(line), ExtensionID) {
				return nil // already installed
			}
		}
	}

	cmd := exec.Command(codeBin, "--install-extension", ExtensionID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("install extension: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// OpenWorkspace opens VSCode with the given workspace directory.
func OpenWorkspace(codeBin, workspaceDir string) error {
	cmd := exec.Command(codeBin, workspaceDir)
	return cmd.Start()
}
