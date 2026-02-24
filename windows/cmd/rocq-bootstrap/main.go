package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	rootfs "github.com/justme0606/rocq-bootstrap/windows"
	"github.com/justme0606/rocq-bootstrap/windows/internal/gui"
	"github.com/justme0606/rocq-bootstrap/windows/internal/manifest"
)

func main() {
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

	gui.Run(m, rootfs.EmbeddedTemplates, rootfs.EmbeddedIcon)
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
