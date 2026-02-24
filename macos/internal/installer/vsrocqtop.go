package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FindVsrocqtop searches for the vsrocqtop binary.
// Search order:
// 1. Inside the installed .app bundle (Contents/ walk, max depth 6)
// 2. exec.LookPath
// 3. Known paths: /usr/local/bin, /opt/homebrew/bin
// 4. Scan /Applications and ~/Applications for *rocq*.app / *coq*.app
func FindVsrocqtop(installedAppPath string) (string, error) {
	debugLog("[vsrocqtop] searching for vsrocqtop")

	// 1. Search inside the installed .app bundle
	if installedAppPath != "" {
		contentsDir := filepath.Join(installedAppPath, "Contents")
		if info, err := os.Stat(contentsDir); err == nil && info.IsDir() {
			debugLog("[vsrocqtop] searching in %s (max depth 6)", contentsDir)
			found := walkForVsrocqtop(contentsDir, 6)
			if found != "" {
				debugLog("[vsrocqtop] FOUND in app bundle: %s", found)
				return found, nil
			}
		}
	}

	// 2. PATH lookup
	if path, err := exec.LookPath("vsrocqtop"); err == nil {
		debugLog("[vsrocqtop] FOUND in PATH: %s", path)
		return path, nil
	}

	// 3. Known paths
	knownPaths := []string{
		"/usr/local/bin/vsrocqtop",
		"/opt/homebrew/bin/vsrocqtop",
	}
	for _, p := range knownPaths {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			debugLog("[vsrocqtop] FOUND at known path: %s", p)
			return p, nil
		}
	}

	// 4. Scan /Applications and ~/Applications for rocq/coq .app bundles
	home, _ := os.UserHomeDir()
	searchDirs := []string{"/Applications"}
	if home != "" {
		searchDirs = append(searchDirs, filepath.Join(home, "Applications"))
	}

	for _, baseDir := range searchDirs {
		entries, err := os.ReadDir(baseDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() || !strings.HasSuffix(e.Name(), ".app") {
				continue
			}
			nameLower := strings.ToLower(e.Name())
			if !strings.Contains(nameLower, "rocq") && !strings.Contains(nameLower, "coq") {
				continue
			}
			appContents := filepath.Join(baseDir, e.Name(), "Contents")
			if info, err := os.Stat(appContents); err == nil && info.IsDir() {
				found := walkForVsrocqtop(appContents, 6)
				if found != "" {
					debugLog("[vsrocqtop] FOUND in app scan: %s", found)
					return found, nil
				}
			}
		}
	}

	debugLog("[vsrocqtop] NOT FOUND")
	return "", fmt.Errorf("vsrocqtop not found")
}

// walkForVsrocqtop walks a directory tree up to maxDepth levels looking for vsrocqtop.
func walkForVsrocqtop(root string, maxDepth int) string {
	var found string
	rootDepth := strings.Count(root, string(os.PathSeparator))

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Enforce max depth
		depth := strings.Count(path, string(os.PathSeparator)) - rootDepth
		if depth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() && info.Name() == "vsrocqtop" {
			// Verify it's executable
			if info.Mode()&0o111 != 0 {
				found = path
				return filepath.SkipAll
			}
		}
		return nil
	})

	return found
}
