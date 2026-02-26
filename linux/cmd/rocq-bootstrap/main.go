package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	rootfs "github.com/justme0606/rocq-bootstrap/linux"
	"github.com/justme0606/rocq-bootstrap/linux/internal/gui"
	"github.com/justme0606/rocq-bootstrap/linux/internal/manifest"
)

var Version = "dev"

const (
	binaryName  = "rocq-bootstrap"
	desktopFile = `[Desktop Entry]
Name=Rocq Bootstrap
Comment=Rocq Platform Installer
Exec=rocq-bootstrap
Icon=rocq-bootstrap
Terminal=false
Type=Application
Categories=Development;Education;Science;
Keywords=Rocq;Coq;proof;assistant;opam;
`
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--install":
			if err := installDesktop(); err != nil {
				fmt.Fprintf(os.Stderr, "Install failed: %v\n", err)
				os.Exit(1)
			}
			return
		case "--uninstall":
			if err := uninstallDesktop(); err != nil {
				fmt.Fprintf(os.Stderr, "Uninstall failed: %v\n", err)
				os.Exit(1)
			}
			return
		case "--help", "-h":
			fmt.Println("Usage: rocq-bootstrap [--install | --uninstall | --help]")
			fmt.Println()
			fmt.Println("  (no args)     Launch the GUI installer")
			fmt.Println("  --install     Install as desktop application (~/.local)")
			fmt.Println("  --uninstall   Remove desktop application")
			fmt.Println("  --help        Show this help")
			return
		}
	}

	// Early log file to capture errors before GUI starts
	earlyLog := setupEarlyLog()
	if earlyLog != nil {
		defer earlyLog.Close()
		fmt.Fprintf(earlyLog, "[%s] rocq-bootstrap starting\n", time.Now().Format("15:04:05"))
	}

	m, err := manifest.Load(rootfs.EmbeddedManifest, "embedded/manifest/latest.json")
	if err != nil {
		msg := fmt.Sprintf("Fatal: %v", err)
		if earlyLog != nil {
			fmt.Fprintln(earlyLog, msg)
		}
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	if earlyLog != nil {
		fmt.Fprintf(earlyLog, "[%s] manifest loaded: Rocq %s (platform %s)\n",
			time.Now().Format("15:04:05"), m.RocqVersion, m.PlatformRelease)
		fmt.Fprintf(earlyLog, "[%s] launching GUI\n", time.Now().Format("15:04:05"))
	}

	gui.Run(m, rootfs.EmbeddedTemplates, rootfs.EmbeddedIcon, Version)
}

func setupEarlyLog() *os.File {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	logDir := filepath.Join(home, ".rocq-setup", "logs")
	os.MkdirAll(logDir, 0o755)

	name := fmt.Sprintf("rocq-bootstrap-%s.log", time.Now().Format("20060102-150405"))
	f, err := os.Create(filepath.Join(logDir, name))
	if err != nil {
		return nil
	}
	return f
}

func installDesktop() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	binDir := filepath.Join(home, ".local", "bin")
	iconDir := filepath.Join(home, ".local", "share", "icons", "hicolor", "256x256", "apps")
	desktopDir := filepath.Join(home, ".local", "share", "applications")

	// Create directories
	for _, dir := range []string{binDir, iconDir, desktopDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}

	// Copy self to ~/.local/bin/
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}
	destBin := filepath.Join(binDir, binaryName)
	if err := copyFile(self, destBin, 0o755); err != nil {
		return fmt.Errorf("copy binary: %w", err)
	}
	fmt.Printf("Installed binary:  %s\n", destBin)

	// Write embedded icon
	destIcon := filepath.Join(iconDir, "rocq-bootstrap.png")
	if err := os.WriteFile(destIcon, rootfs.EmbeddedIcon, 0o644); err != nil {
		return fmt.Errorf("write icon: %w", err)
	}
	fmt.Printf("Installed icon:    %s\n", destIcon)

	// Write .desktop file
	destDesktop := filepath.Join(desktopDir, "rocq-bootstrap.desktop")
	if err := os.WriteFile(destDesktop, []byte(desktopFile), 0o644); err != nil {
		return fmt.Errorf("write desktop file: %w", err)
	}
	fmt.Printf("Installed desktop: %s\n", destDesktop)

	fmt.Println()
	fmt.Println("Rocq Bootstrap installed as desktop application.")
	fmt.Printf("Make sure %s is in your PATH.\n", binDir)
	return nil
}

func uninstallDesktop() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	files := []string{
		filepath.Join(home, ".local", "bin", binaryName),
		filepath.Join(home, ".local", "share", "icons", "hicolor", "256x256", "apps", "rocq-bootstrap.png"),
		filepath.Join(home, ".local", "share", "applications", "rocq-bootstrap.desktop"),
	}

	for _, f := range files {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: could not remove %s: %v\n", f, err)
		} else if err == nil {
			fmt.Printf("Removed: %s\n", f)
		}
	}

	fmt.Println("Rocq Bootstrap uninstalled.")
	return nil
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
