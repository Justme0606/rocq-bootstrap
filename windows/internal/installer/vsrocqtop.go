package installer

import (
	"fmt"
	"os"
	"path/filepath"
)

// vsrocqtopNames lists the binary names to search for.
var vsrocqtopNames = []string{"vsrocqtop", "vsrocqtop.exe"}

// FindVsrocqtop searches for vsrocqtop in the installation directory.
// It first checks <installDir>/bin/, then searches recursively.
func FindVsrocqtop(installDir string) (string, error) {
	debugLog("[vsrocqtop] searching in %s", installDir)

	// Check the expected locations first
	for _, name := range vsrocqtopNames {
		direct := filepath.Join(installDir, "bin", name)
		debugLog("[vsrocqtop] checking %s", direct)
		if info, err := os.Stat(direct); err == nil && !info.IsDir() {
			debugLog("[vsrocqtop] FOUND at %s", direct)
			return direct, nil
		}
	}

	// Recursive search
	debugLog("[vsrocqtop] not in bin/, starting recursive search...")
	var found string
	err := filepath.Walk(installDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}
		if !info.IsDir() {
			name := info.Name()
			if name == "vsrocqtop" || name == "vsrocqtop.exe" {
				found = path
				return filepath.SkipAll
			}
		}
		return nil
	})
	if err != nil {
		debugLog("[vsrocqtop] walk error: %v", err)
		return "", fmt.Errorf("search vsrocqtop: %w", err)
	}

	if found == "" {
		debugLog("[vsrocqtop] NOT FOUND anywhere in %s", installDir)
		return "", fmt.Errorf("vsrocqtop not found in %s", installDir)
	}

	debugLog("[vsrocqtop] FOUND at %s", found)
	return found, nil
}
