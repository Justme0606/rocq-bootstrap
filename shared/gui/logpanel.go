package gui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// LogPanel is a thread-safe log buffer displayed in the GUI.
type LogPanel struct {
	mu      sync.Mutex
	lines   []string
	Display *widget.RichText
}

// NewLogPanel creates a new log panel widget.
func NewLogPanel() *LogPanel {
	lp := &LogPanel{
		Display: widget.NewRichText(),
	}
	lp.Display.Wrapping = fyne.TextWrapWord
	return lp
}

// Append adds a timestamped message to the log panel.
func (lp *LogPanel) Append(msg string) {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	ts := time.Now().Format("15:04:05")
	line := fmt.Sprintf("[%s]  %s", ts, msg)
	lp.lines = append(lp.lines, line)
	lp.Display.ParseMarkdown("```\n" + strings.Join(lp.lines, "\n") + "\n```")
}

// UpdateLast replaces the last log line with a new timestamped message.
// If the log is empty, it behaves like Append.
func (lp *LogPanel) UpdateLast(msg string) {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	ts := time.Now().Format("15:04:05")
	line := fmt.Sprintf("[%s]  %s", ts, msg)
	if len(lp.lines) == 0 {
		lp.lines = append(lp.lines, line)
	} else {
		lp.lines[len(lp.lines)-1] = line
	}
	lp.Display.ParseMarkdown("```\n" + strings.Join(lp.lines, "\n") + "\n```")
}

// Clear removes all log lines.
func (lp *LogPanel) Clear() {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	lp.lines = nil
	lp.Display.ParseMarkdown("")
}
