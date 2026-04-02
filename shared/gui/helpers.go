package gui

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const VSCodeDownloadURL = "https://code.visualstudio.com/Download"

// VersionDisplayName returns "Rocq X.Y.Z" or "Coq X.Y.Z" depending on the major version.
func VersionDisplayName(version string) string {
	parts := strings.SplitN(version, ".", 2)
	if len(parts) > 0 {
		if major, err := strconv.Atoi(parts[0]); err == nil && major < 9 {
			return fmt.Sprintf("Coq %s", version)
		}
	}
	return fmt.Sprintf("Rocq %s", version)
}

// FormatDuration formats a duration as "Xm Ys" or "Xs".
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// ShowVSCodeDialog shows a dialog informing the user that VSCode was not found
// and offering to open the download page.
func ShowVSCodeDialog(w fyne.Window) {
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
		u, _ := url.Parse(VSCodeDownloadURL)
		fyne.CurrentApp().OpenURL(u)
		d.Hide()
	}
	closeBtn.OnTapped = func() {
		d.Hide()
	}

	d.Show()
}

// ShowError re-enables the install button and shows an error dialog.
func ShowError(w fyne.Window, installBtn *widget.Button, msg string) {
	installBtn.Enable()
	dialog.ShowError(fmt.Errorf("%s", msg), w)
}

// ShowSuccess displays a success dialog with the given message.
func ShowSuccess(w fyne.Window, msg string) {
	successMsg := widget.NewLabel(msg)
	successMsg.Wrapping = fyne.TextWrapWord

	okBtn := widget.NewButton("OK", nil)
	okBtn.Importance = widget.HighImportance

	successContent := container.NewVBox(successMsg, container.NewHBox(layout.NewSpacer(), okBtn))
	successDialog := dialog.NewCustomWithoutButtons("Success", successContent, w)
	successDialog.Resize(fyne.NewSize(460, 250))
	okBtn.OnTapped = func() { successDialog.Hide() }
	successDialog.Show()
}
