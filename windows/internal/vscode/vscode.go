package vscode

import (
	"fmt"
	"os/exec"
	"strings"
)

const ExtensionID = "rocq-prover.vsrocq"

// FindCode searches for the VSCode CLI executable.
func FindCode() (string, error) {
	// Try PATH first
	path, err := exec.LookPath("code")
	if err == nil {
		return path, nil
	}

	// Try common Windows install locations
	candidates := []string{
		`C:\Program Files\Microsoft VS Code\bin\code.cmd`,
		`C:\Program Files (x86)\Microsoft VS Code\bin\code.cmd`,
	}
	for _, c := range candidates {
		if _, err := exec.LookPath(c); err == nil {
			return c, nil
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
