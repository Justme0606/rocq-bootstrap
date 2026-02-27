package gui

import (
	"fmt"
	"io/fs"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/justme0606/rocq-bootstrap/linux/internal/doctor"
	"github.com/justme0606/rocq-bootstrap/linux/internal/installer"
	"github.com/justme0606/rocq-bootstrap/linux/internal/manifest"
	"github.com/justme0606/rocq-bootstrap/linux/internal/releases"
)

const vscodeDownloadURL = "https://code.visualstudio.com/Download"

const (
	windowWidth  = 620
	windowHeight = 520
	totalSteps   = 7
)

// logPanel is a thread-safe log buffer displayed in the GUI.
type logPanel struct {
	mu      sync.Mutex
	lines   []string
	display *widget.RichText
}

func newLogPanel() *logPanel {
	lp := &logPanel{
		display: widget.NewRichText(),
	}
	lp.display.Wrapping = fyne.TextWrapWord
	return lp
}

func (lp *logPanel) append(msg string) {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	ts := time.Now().Format("15:04:05")
	line := fmt.Sprintf("[%s]  %s", ts, msg)
	lp.lines = append(lp.lines, line)
	lp.display.ParseMarkdown("```\n" + strings.Join(lp.lines, "\n") + "\n```")
}

// Run creates and runs the GUI application.
func Run(m *manifest.Manifest, templates fs.FS, icon []byte, version string) {
	a := app.New()
	a.Settings().SetTheme(newRocqTheme())

	var iconRes fyne.Resource
	if len(icon) > 0 {
		iconRes = fyne.NewStaticResource("rocq-icon.png", icon)
		a.SetIcon(iconRes)
	}

	windowTitle := "Rocq Platform Installer"
	if version != "" && version != "dev" {
		windowTitle += " - " + version
	}
	w := a.NewWindow(windowTitle)
	w.Resize(fyne.NewSize(windowWidth, windowHeight))
	w.SetFixedSize(false)

	// --- Header: icon + title + version info ---
	var headerIcon *canvas.Image
	if iconRes != nil {
		headerIcon = canvas.NewImageFromResource(iconRes)
		headerIcon.FillMode = canvas.ImageFillContain
		headerIcon.SetMinSize(fyne.NewSize(64, 64))
	}

	titleRocq := canvas.NewText("Rocq", rocqOrange)
	titleRocq.TextSize = 22
	titleRocq.TextStyle = fyne.TextStyle{Bold: true}

	titleRest := canvas.NewText(" Platform Installer", rocqNavy)
	titleRest.TextSize = 22
	titleRest.TextStyle = fyne.TextStyle{Bold: true}

	titleRow := container.NewHBox(titleRocq, titleRest)

	titleBlock := container.NewVBox(titleRow)

	var header *fyne.Container
	if headerIcon != nil {
		header = container.NewHBox(headerIcon, container.NewCenter(titleBlock))
	} else {
		header = container.NewHBox(container.NewCenter(titleBlock))
	}

	headerSep := widget.NewSeparator()

	// --- Release selector ---
	currentManifest := m

	releaseSelect := widget.NewSelect([]string{m.PlatformRelease}, func(selected string) {})
	releaseSelect.Selected = m.PlatformRelease

	releaseLabel := widget.NewLabelWithStyle("Release:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	rocqVersionLabel := widget.NewLabel(versionDisplayName(m.RocqVersion))
	releaseRow := container.NewBorder(nil, nil, releaseLabel, rocqVersionLabel, releaseSelect)

	// Fetch available releases in background and update Rocq version for current tag
	go func() {
		tags, err := releases.FetchReleases()
		if err != nil {
			return
		}
		if len(tags) > 0 {
			releaseSelect.Options = tags
			releaseSelect.Refresh()
		}
		// Fetch actual Rocq version for the initially selected release
		newManifest, err := releases.FetchManifestForTag(currentManifest.PlatformRelease)
		if err != nil {
			return
		}
		currentManifest = newManifest
		rocqVersionLabel.SetText(versionDisplayName(currentManifest.RocqVersion))
	}()

	releaseSelect.OnChanged = func(selected string) {
		if selected == currentManifest.PlatformRelease {
			return
		}
		releaseSelect.Disable()
		go func() {
			newManifest, err := releases.FetchManifestForTag(selected)
			if err != nil {
				releaseSelect.Enable()
				return
			}
			currentManifest = newManifest
			rocqVersionLabel.SetText(versionDisplayName(currentManifest.RocqVersion))
			releaseSelect.Enable()
		}()
	}

	// --- Progress section ---
	statusLabel := widget.NewLabel("Ready to install")
	statusLabel.Wrapping = fyne.TextWrapWord
	statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	stepLabel := widget.NewLabel(fmt.Sprintf("Step 0/%d", totalSteps))
	stepLabel.Alignment = fyne.TextAlignTrailing

	progressBar := widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 1.0

	statusRow := container.NewBorder(nil, nil, nil, stepLabel, statusLabel)

	progressSection := container.NewVBox(
		statusRow,
		progressBar,
	)

	// --- Log panel ---
	logP := newLogPanel()
	logP.append(fmt.Sprintf("Rocq version: %s", m.RocqVersion))
	logP.append(fmt.Sprintf("Platform release: %s", m.PlatformRelease))
	logP.append("Click 'Install' to begin.")

	logScroll := container.NewScroll(logP.display)
	logScroll.SetMinSize(fyne.NewSize(580, 220))

	logHeader := widget.NewLabelWithStyle("Log", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	logSection := container.NewBorder(
		logHeader,
		nil, nil, nil,
		logScroll,
	)

	// --- Install button ---
	var installBtn *widget.Button
	installBtn = widget.NewButtonWithIcon("Install", theme.DownloadIcon(), func() {
		installBtn.Disable()
		releaseSelect.Disable()

		existingSwitches := installer.FindExistingInstallations()
		if len(existingSwitches) > 0 {
			for _, sw := range existingSwitches {
				logP.append(fmt.Sprintf("Existing opam switch detected: %s", sw))
			}

			msg := widget.NewLabel("Existing opam switches were found.\nSelect one to reuse, or install a new switch:")
			msg.Wrapping = fyne.TextWrapWord

			newSwitchLabel := fmt.Sprintf("Install new (%s)", installer.SwitchName(currentManifest.RocqVersion, currentManifest.PlatformRelease))
			options := append(existingSwitches, newSwitchLabel)
			radio := widget.NewRadioGroup(options, nil)
			radio.SetSelected(existingSwitches[0])

			radioScroll := container.NewScroll(radio)
			radioScroll.SetMinSize(fyne.NewSize(400, 200))

			closeBtn := widget.NewButton("Close", nil)
			closeBtn.Importance = widget.HighImportance
			confirmBtn := widget.NewButton("Continue", nil)
			confirmBtn.Importance = widget.HighImportance

			buttons := container.NewHBox(layout.NewSpacer(), closeBtn, confirmBtn)
			content := container.NewVBox(msg, radioScroll, buttons)
			d := dialog.NewCustomWithoutButtons("Existing Installation Detected", content, w)

			closeBtn.OnTapped = func() {
				d.Hide()
				installBtn.Enable()
				releaseSelect.Enable()
			}
			confirmBtn.OnTapped = func() {
				d.Hide()
				selected := radio.Selected
				if selected == newSwitchLabel {
					logP.append("Starting fresh installation...")
					go runInstallWithOptions(w, currentManifest, templates, statusLabel, progressBar, stepLabel, installBtn, logP, false)
				} else {
					logP.append(fmt.Sprintf("Reusing switch %s...", selected))
					go runInstallWithOptions(w, currentManifest, templates, statusLabel, progressBar, stepLabel, installBtn, logP, true)
				}
			}

			d.Show()
		} else {
			logP.append("Starting installation...")
			go runInstallWithOptions(w, currentManifest, templates, statusLabel, progressBar, stepLabel, installBtn, logP, false)
		}
	})
	installBtn.Importance = widget.HighImportance

	// --- Doctor button ---
	var doctorBtn *widget.Button
	doctorBtn = widget.NewButtonWithIcon("Doctor", theme.InfoIcon(), func() {
		installBtn.Disable()
		doctorBtn.Disable()

		go func() {
			var lines []string
			doctor.Run(func(msg string) {
				lines = append(lines, msg)
			})

			richText := widget.NewRichText()
			richText.Wrapping = fyne.TextWrapWord
			richText.ParseMarkdown("```\n" + strings.Join(lines, "\n") + "\n```")

			scroll := container.NewScroll(richText)
			scroll.SetMinSize(fyne.NewSize(560, 350))

			closeBtn := widget.NewButton("Close", nil)
			closeBtn.Importance = widget.HighImportance

			content := container.NewBorder(nil, container.NewCenter(closeBtn), nil, nil, scroll)
			d := dialog.NewCustomWithoutButtons("Doctor \u2014 System Diagnostic", content, w)

			closeBtn.OnTapped = func() {
				d.Hide()
			}

			d.Show()

			installBtn.Enable()
			doctorBtn.Enable()
		}()
	})
	doctorBtn.Importance = widget.HighImportance

	bottomBar := container.NewPadded(container.NewCenter(container.NewHBox(doctorBtn, installBtn)))

	// --- Main layout ---
	content := container.NewPadded(
		container.NewBorder(
			container.NewVBox(
				header,
				headerSep,
				releaseRow,
				progressSection,
			),
			bottomBar,
			nil, nil,
			container.NewVBox(
				layout.NewSpacer(),
				logSection,
				layout.NewSpacer(),
			),
		),
	)

	w.SetContent(content)
	w.ShowAndRun()
}

func runInstallWithOptions(w fyne.Window, m *manifest.Manifest, templates fs.FS,
	statusLabel *widget.Label, progressBar *widget.ProgressBar,
	stepLabel *widget.Label, installBtn *widget.Button, logP *logPanel,
	skipInstall bool) {

	logger, err := installer.NewLogger()
	if err != nil {
		logP.append(fmt.Sprintf("WARNING: could not create log file: %v", err))
	}
	if logger != nil {
		defer logger.Close()
	}

	switchName := installer.SwitchName(m.RocqVersion, m.PlatformRelease)

	var lastLoggedStep int
	cfg := &installer.Config{
		Manifest:    m,
		Templates:   templates,
		SkipInstall: skipInstall,
		Logger:      logger,
		OnStep: func(step int, label string, fraction float64) {
			overall := (float64(step-1) + fraction) / float64(totalSteps)
			statusLabel.SetText(label)
			stepLabel.SetText(fmt.Sprintf("Step %d/%d", step, totalSteps))
			progressBar.SetValue(overall)
			if step != lastLoggedStep || fraction >= 1.0 {
				logP.append(label)
				lastLoggedStep = step
			}
		},
	}

	result, err := installer.Run(cfg)
	if err != nil {
		if logger != nil {
			logger.Log("ERROR: %v", err)
		}
		logP.append(fmt.Sprintf("ERROR: %v", err))
		showError(w, installBtn, err.Error())
		return
	}

	progressBar.SetValue(1.0)

	if !result.VSCodeFound {
		statusLabel.SetText("Rocq Platform installed — VSCode not found")
		logP.append("Rocq Platform installed successfully.")
		logP.append("VSCode was not found. Install VSCode then re-run this installer to configure the workspace.")
		logP.append(fmt.Sprintf("Opam switch: %s", switchName))
		logP.append("Activate with: source ~/rocq-workspace/activate.sh")

		showVSCodeDialog(w)
		return
	}

	statusLabel.SetText("Installation complete!")
	logP.append("Installation complete!")
	logP.append(fmt.Sprintf("Opam switch: %s", switchName))
	logP.append(fmt.Sprintf("Workspace: ~/%s", installer.WorkspaceName))
	logP.append("Activate with: source ~/rocq-workspace/activate.sh")

	successMsg := widget.NewLabel(
		"Rocq Platform has been installed successfully.\n\n" +
			fmt.Sprintf("Opam switch: %s\n", switchName) +
			fmt.Sprintf("Workspace: ~/%s\n\n", installer.WorkspaceName) +
			"Activate with:\n  source ~/rocq-workspace/activate.sh")
	successMsg.Wrapping = fyne.TextWrapWord

	okBtn := widget.NewButton("OK", nil)
	okBtn.Importance = widget.HighImportance

	successContent := container.NewVBox(successMsg, container.NewHBox(layout.NewSpacer(), okBtn))
	successDialog := dialog.NewCustomWithoutButtons("Success", successContent, w)
	successDialog.Resize(fyne.NewSize(460, 250))
	okBtn.OnTapped = func() { successDialog.Hide() }
	successDialog.Show()
}

func versionDisplayName(version string) string {
	parts := strings.SplitN(version, ".", 2)
	if len(parts) > 0 {
		if major, err := strconv.Atoi(parts[0]); err == nil && major < 9 {
			return fmt.Sprintf("Coq %s", version)
		}
	}
	return fmt.Sprintf("Rocq %s", version)
}

func showVSCodeDialog(w fyne.Window) {
	msg := widget.NewLabel(
		"Rocq Platform has been installed successfully.\n\n" +
			"However, VSCode was not found on your system.\n" +
			"VSCode is required to use the Rocq extension.\n\n" +
			"Would you like to download VSCode?")
	msg.Wrapping = fyne.TextWrapWord

	downloadBtn := widget.NewButton("Download VSCode", nil)
	downloadBtn.Importance = widget.HighImportance

	closeBtn := widget.NewButton("Close", nil)
	closeBtn.Importance = widget.LowImportance

	buttons := container.NewHBox(layout.NewSpacer(), closeBtn, downloadBtn)
	content := container.NewVBox(msg, buttons)

	d := dialog.NewCustomWithoutButtons("VSCode Not Found", content, w)
	d.Resize(fyne.NewSize(460, 250))

	downloadBtn.OnTapped = func() {
		u, _ := url.Parse(vscodeDownloadURL)
		fyne.CurrentApp().OpenURL(u)
		d.Hide()
	}
	closeBtn.OnTapped = func() {
		d.Hide()
	}

	d.Show()
}

func showError(w fyne.Window, installBtn *widget.Button, msg string) {
	installBtn.Enable()
	dialog.ShowError(fmt.Errorf("%s", msg), w)
}
