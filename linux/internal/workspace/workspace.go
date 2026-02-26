package workspace

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Create creates the workspace directory with template files.
// Existing files are not overwritten.
func Create(workspaceDir string, templates fs.FS) error {
	log.Printf("[workspace] creating workspace at %s", workspaceDir)

	if err := os.MkdirAll(filepath.Join(workspaceDir, ".vscode"), 0o755); err != nil {
		return fmt.Errorf("create workspace dir: %w", err)
	}

	files := []struct {
		embeddedPath string
		destName     string
	}{
		{"embedded/templates/test.v", "test.v"},
		{"embedded/templates/main.v", "main.v"},
		{"embedded/templates/_RocqProject", "_RocqProject"},
	}

	for _, f := range files {
		dest := filepath.Join(workspaceDir, f.destName)
		if _, err := os.Stat(dest); err == nil {
			log.Printf("[workspace]   %s already exists, skipping", f.destName)
			continue // don't overwrite existing files
		}

		data, err := fs.ReadFile(templates, f.embeddedPath)
		if err != nil {
			log.Printf("[workspace]   ERROR reading template %s: %v", f.embeddedPath, err)
			return fmt.Errorf("read template %s: %w", f.embeddedPath, err)
		}

		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", dest, err)
		}
		log.Printf("[workspace]   wrote %s", dest)
	}

	log.Printf("[workspace] workspace created successfully")
	return nil
}

// WriteVSCodeSettings writes .vscode/settings.json with the vsrocqtop path.
func WriteVSCodeSettings(workspaceDir, vsrocqtopPath string, templates fs.FS) error {
	log.Printf("[workspace] writing VSCode settings with vsrocqtop=%s", vsrocqtopPath)

	tpl, err := fs.ReadFile(templates, "embedded/templates/vscode-settings.json")
	if err != nil {
		log.Printf("[workspace]   ERROR reading settings template: %v", err)
		return fmt.Errorf("read settings template: %w", err)
	}

	content := strings.ReplaceAll(string(tpl), "__VSROCQTOP__", vsrocqtopPath)

	dest := filepath.Join(workspaceDir, ".vscode", "settings.json")
	if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}

	log.Printf("[workspace]   wrote %s", dest)
	return nil
}

// WriteActivationScripts generates shell activation scripts for the opam switch.
func WriteActivationScripts(workspaceDir, switchName string) error {
	log.Printf("[workspace] writing activation scripts for switch %s", switchName)

	// activate.sh — sources opam env for the switch
	activateSh := fmt.Sprintf(`#!/usr/bin/env bash
# Activate the Rocq Platform opam switch.
# Usage: source activate.sh
eval "$(opam env --switch=%s --set-switch)"
echo "Rocq Platform activated (switch: %s)"
`, switchName, switchName)

	activatePath := filepath.Join(workspaceDir, "activate.sh")
	if err := os.WriteFile(activatePath, []byte(activateSh), 0o755); err != nil {
		return fmt.Errorf("write activate.sh: %w", err)
	}
	log.Printf("[workspace]   wrote %s", activatePath)

	// activate-shell.sh — spawns a new shell with the opam env
	activateShellSh := fmt.Sprintf(`#!/usr/bin/env bash
# Spawn a new shell with the Rocq Platform opam switch activated.
# Usage: ./activate-shell.sh
echo "Launching shell with Rocq Platform (switch: %s)..."
opam exec --switch=%s -- "${SHELL:-/bin/bash}"
`, switchName, switchName)

	activateShellPath := filepath.Join(workspaceDir, "activate-shell.sh")
	if err := os.WriteFile(activateShellPath, []byte(activateShellSh), 0o755); err != nil {
		return fmt.Errorf("write activate-shell.sh: %w", err)
	}
	log.Printf("[workspace]   wrote %s", activateShellPath)

	return nil
}
