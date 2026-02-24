package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MountDMG mounts a DMG file and returns the mount point path.
func MountDMG(dmgPath string) (string, error) {
	debugLog("[dmg] mounting %s", dmgPath)

	out, err := exec.Command("hdiutil", "attach", dmgPath, "-nobrowse").Output()
	if err != nil {
		return "", fmt.Errorf("hdiutil attach: %w", err)
	}

	// Parse output to find mount point (last column of last line containing /Volumes/)
	for _, line := range strings.Split(string(out), "\n") {
		if idx := strings.Index(line, "/Volumes/"); idx >= 0 {
			mountPoint := strings.TrimSpace(line[idx:])
			debugLog("[dmg] mounted at %s", mountPoint)
			return mountPoint, nil
		}
	}

	return "", fmt.Errorf("hdiutil attach: no mount point found in output")
}

// UnmountDMG detaches a mounted DMG volume.
func UnmountDMG(mountPoint string) error {
	debugLog("[dmg] detaching %s", mountPoint)
	err := exec.Command("hdiutil", "detach", mountPoint).Run()
	if err != nil {
		// Try force detach
		debugLog("[dmg] normal detach failed, trying force")
		return exec.Command("hdiutil", "detach", mountPoint, "-force").Run()
	}
	return nil
}

// FindAppInDMG searches for a .app bundle in the mounted DMG volume.
func FindAppInDMG(mountPoint string) (string, error) {
	debugLog("[dmg] searching for .app in %s", mountPoint)

	entries, err := os.ReadDir(mountPoint)
	if err != nil {
		return "", fmt.Errorf("read mount point: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() && strings.HasSuffix(e.Name(), ".app") {
			appPath := filepath.Join(mountPoint, e.Name())
			debugLog("[dmg] found app: %s", appPath)
			return appPath, nil
		}
	}

	// Try one level deeper
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		subEntries, err := os.ReadDir(filepath.Join(mountPoint, e.Name()))
		if err != nil {
			continue
		}
		for _, se := range subEntries {
			if se.IsDir() && strings.HasSuffix(se.Name(), ".app") {
				appPath := filepath.Join(mountPoint, e.Name(), se.Name())
				debugLog("[dmg] found app: %s", appPath)
				return appPath, nil
			}
		}
	}

	return "", fmt.Errorf("no .app found in DMG at %s", mountPoint)
}

// InstallApp copies the .app bundle to /Applications (or ~/Applications as fallback).
// If force is true, an existing installation will be replaced.
// Returns the destination path of the installed app.
func InstallApp(appSrc string, force bool) (string, error) {
	appName := filepath.Base(appSrc)

	// Determine destination: /Applications if writable, otherwise ~/Applications
	destDir := "/Applications"
	testFile := filepath.Join(destDir, ".rocq-write-test")
	if err := os.WriteFile(testFile, []byte("test"), 0o644); err != nil {
		// /Applications not writable, use ~/Applications
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get home dir: %w", err)
		}
		destDir = filepath.Join(home, "Applications")
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return "", fmt.Errorf("create ~/Applications: %w", err)
		}
	} else {
		os.Remove(testFile)
	}

	appDst := filepath.Join(destDir, appName)
	debugLog("[dmg] installing %s -> %s", appName, appDst)

	// Check if already installed
	if _, err := os.Stat(appDst); err == nil && !force {
		debugLog("[dmg] app already installed at %s (use force to replace)", appDst)
		return appDst, nil
	}

	// Remove existing installation if present
	if err := os.RemoveAll(appDst); err != nil {
		debugLog("[dmg] WARNING: failed to remove existing app: %v", err)
	}

	// Copy using rsync for reliable .app bundle copy
	cmd := exec.Command("rsync", "-a", "--delete", appSrc+"/", appDst+"/")
	if out, err := cmd.CombinedOutput(); err != nil {
		// Fallback to cp -R if rsync is not available
		debugLog("[dmg] rsync failed (%v), trying cp -R", err)
		cmd = exec.Command("cp", "-R", appSrc, appDst)
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("copy app: %w\nOutput: %s", err, string(out))
		}
	} else {
		_ = out
	}

	debugLog("[dmg] app installed at %s", appDst)
	return appDst, nil
}
