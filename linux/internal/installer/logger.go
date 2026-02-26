package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Logger writes to a log file.
type Logger struct {
	file *os.File
}

// NewLogger creates a log file under ~/.rocq-setup/logs/.
func NewLogger() (*Logger, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	logDir := filepath.Join(home, ".rocq-setup", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}

	name := fmt.Sprintf("rocq-setup-%s.log", time.Now().Format("20060102-150405"))
	f, err := os.Create(filepath.Join(logDir, name))
	if err != nil {
		return nil, err
	}

	return &Logger{file: f}, nil
}

func (l *Logger) Log(format string, args ...interface{}) {
	if l == nil || l.file == nil {
		return
	}
	ts := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(l.file, "[%s] %s\n", ts, fmt.Sprintf(format, args...))
}

func (l *Logger) Close() {
	if l != nil && l.file != nil {
		l.file.Close()
	}
}
