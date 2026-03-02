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

func (lp *logPanel) clear() {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	lp.lines = nil
	lp.display.ParseMarkdown("")
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
	titleRocq.TextSize = 20
	titleRocq.TextStyle = fyne.TextStyle{Bold: true}

	titleRest := canvas.NewText(" Platform Installer", rocqDarkText)
	titleRest.TextSize = 20
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

	// --- Log panel (defined early so release callbacks can update it) ---
	logP := newLogPanel()

	resetLogForManifest := func(mf *manifest.Manifest) {
		logP.clear()
		logP.append(fmt.Sprintf("Rocq version: %s", mf.RocqVersion))
		logP.append(fmt.Sprintf("Platform release: %s", mf.PlatformRelease))
		logP.append("Click 'Install' to begin.")
	}
	resetLogForManifest(m)

	// --- Release selector ---
	currentManifest := m
	labelToTag := map[string]string{}

	initialLabel := m.PlatformRelease + " \u2014 " + versionDisplayName(m.RocqVersion)
	labelToTag[initialLabel] = m.PlatformRelease
	labelToTag[m.PlatformRelease] = m.PlatformRelease

	releaseSelect := widget.NewSelect([]string{initialLabel}, func(selected string) {})
	releaseSelect.Selected = initialLabel

	releaseLabel := widget.NewLabelWithStyle("Release:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	releaseRow := container.NewBorder(nil, nil, releaseLabel, nil, releaseSelect)

	resolveTag := func(label string) string {
		if tag, ok := labelToTag[label]; ok {
			return tag
		}
		return label
	}

	// Fetch available releases in background with versions
	go func() {
		tags, err := releases.FetchReleases()
		if err != nil {
			return
		}

		// Fetch Rocq version for each tag in parallel
		type tagVersion struct {
			tag     string
			version string
		}
		results := make(chan tagVersion, len(tags))
		for _, tag := range tags {
			go func(t string) {
				ver, err := releases.FetchRocqVersion(t)
				if err != nil {
					results <- tagVersion{tag: t}
					return
				}
				results <- tagVersion{tag: t, version: ver}
			}(tag)
		}

		versionMap := map[string]string{}
		for range tags {
			r := <-results
			if r.version != "" {
				versionMap[r.tag] = r.version
			}
		}

		var options []string
		for _, tag := range tags {
			label := tag
			if ver, ok := versionMap[tag]; ok {
				label = tag + " \u2014 " + versionDisplayName(ver)
			}
			labelToTag[label] = tag
			options = append(options, label)
		}

		// Update current selection label
		currentTag := currentManifest.PlatformRelease
		var currentLabel string
		for _, opt := range options {
			if resolveTag(opt) == currentTag {
				currentLabel = opt
				break
			}
		}

		releaseSelect.Options = options
		if currentLabel != "" {
			releaseSelect.Selected = currentLabel
		}
		releaseSelect.Refresh()

		// Fetch full manifest for current release
		newManifest, err := releases.FetchManifestForTag(currentTag)
		if err != nil {
			return
		}
		currentManifest = newManifest
		resetLogForManifest(currentManifest)
	}()

	releaseSelect.OnChanged = func(selected string) {
		tag := resolveTag(selected)
		if tag == currentManifest.PlatformRelease {
			return
		}
		releaseSelect.Disable()
		go func() {
			newManifest, err := releases.FetchManifestForTag(tag)
			if err != nil {
				releaseSelect.Enable()
				return
			}
			currentManifest = newManifest
			resetLogForManifest(currentManifest)
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

	infiniteBar := widget.NewProgressBarInfinite()
	infiniteBar.Hide()

	progressStack := container.NewStack(progressBar, infiniteBar)

	statusRow := container.NewBorder(nil, nil, nil, stepLabel, statusLabel)

	progressSection := container.NewVBox(
		statusRow,
		progressStack,
	)

	// --- Log panel layout ---
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
					go runInstallWithOptions(w, currentManifest, templates, statusLabel, progressBar, infiniteBar, stepLabel, installBtn, logP, false)
				} else {
					logP.append(fmt.Sprintf("Reusing switch %s...", selected))
					go runInstallWithOptions(w, currentManifest, templates, statusLabel, progressBar, infiniteBar, stepLabel, installBtn, logP, true)
				}
			}

			d.Show()
		} else {
			logP.append("Starting installation...")
			go runInstallWithOptions(w, currentManifest, templates, statusLabel, progressBar, infiniteBar, stepLabel, installBtn, logP, false)
		}
	})
	installBtn.Importance = widget.HighImportance

	// --- Doctor button ---
	var doctorBtn *widget.Button
	doctorBtn = widget.NewButtonWithIcon("Doctor", theme.InfoIcon(), func() {
		installBtn.Disable()
		doctorBtn.Disable()
		statusLabel.SetText("Running diagnostics...")
		progressBar.Hide()
		infiniteBar.Show()

		go func() {
			var lines []string
			doctor.Run(func(msg string) {
				lines = append(lines, msg)
			})

			infiniteBar.Hide()
			progressBar.Show()
			statusLabel.SetText("Ready to install")

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
	infiniteBar *widget.ProgressBarInfinite,
	stepLabel *widget.Label, installBtn *widget.Button, logP *logPanel,
	skipInstall bool) {

	startTime := time.Now()

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

			// Show infinite progress bar during long opam operations (steps 2-5)
			if step >= 2 && step <= 5 && fraction < 1.0 {
				progressBar.Hide()
				infiniteBar.Show()
			} else {
				infiniteBar.Hide()
				progressBar.Show()
				progressBar.SetValue(overall)
			}

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

	elapsed := formatDuration(time.Since(startTime))

	if !result.VSCodeFound {
		statusLabel.SetText(fmt.Sprintf("Rocq Platform installed in %s — VSCode not found", elapsed))
		logP.append(fmt.Sprintf("Rocq Platform installed successfully in %s.", elapsed))
		logP.append("VSCode was not found. Install VSCode then re-run this installer to configure the workspace.")
		logP.append(fmt.Sprintf("Opam switch: %s", switchName))
		logP.append("Activate with: source ~/rocq-workspace/activate.sh")

		showVSCodeDialog(w)
		return
	}

	statusLabel.SetText(fmt.Sprintf("Installation complete! (%s)", elapsed))
	logP.append(fmt.Sprintf("Installation complete! (%s)", elapsed))
	logP.append(fmt.Sprintf("Opam switch: %s", switchName))
	logP.append(fmt.Sprintf("Workspace: ~/%s", installer.WorkspaceName))
	logP.append("Activate with: source ~/rocq-workspace/activate.sh")

	successMsg := widget.NewLabel(
		fmt.Sprintf("Rocq Platform has been installed successfully in %s.\n\n", elapsed) +
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

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
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
