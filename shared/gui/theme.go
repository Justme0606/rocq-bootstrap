package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Rocq brand colors from https://rocq-prover.org/ logo.
// Logo uses #260085 (deep blue) and #ff540a (bright orange).
var (
	rocqLightBg   = color.NRGBA{R: 0xf0, G: 0xef, B: 0xf5, A: 0xff} // light background with blue tint
	rocqAccent    = color.NRGBA{R: 0xff, G: 0x54, B: 0x0a, A: 0xff} // brand orange for focus (#ff540a)
	rocqHover     = color.NRGBA{R: 0xe8, G: 0xe6, B: 0xf0, A: 0xff} // subtle blue-tinted hover
	rocqSelection = color.NRGBA{R: 0xd8, G: 0xd4, B: 0xe8, A: 0xff} // soft blue-tinted selection
	rocqSeparator = color.NRGBA{R: 0xd0, G: 0xce, B: 0xda, A: 0xff} // subtle separator
	rocqInputBg   = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff} // white input fields
	rocqDarkText  = color.NRGBA{R: 0x1a, G: 0x0a, B: 0x3a, A: 0xff} // dark text with blue undertone
	rocqMutedText = color.NRGBA{R: 0x6b, G: 0x6b, B: 0x6b, A: 0xff} // muted text
	rocqSuccess   = color.NRGBA{R: 0x2e, G: 0x7d, B: 0x32, A: 0xff} // green for success
	rocqError     = color.NRGBA{R: 0xc6, G: 0x28, B: 0x28, A: 0xff} // red for errors

	// RocqOrange is the brand orange (#ff540a), exported for use in GUI components.
	RocqOrange = color.NRGBA{R: 0xff, G: 0x54, B: 0x0a, A: 0xff}
	// RocqBlue is the brand deep blue (#260085), exported for use in GUI components.
	RocqBlue = color.NRGBA{R: 0x26, G: 0x00, B: 0x85, A: 0xff}
)

// rocqTheme implements fyne.Theme with Rocq brand colors.
type rocqTheme struct {
	base fyne.Theme
}

// NewRocqTheme creates a new Rocq-branded Fyne theme.
func NewRocqTheme() fyne.Theme {
	return &rocqTheme{base: theme.DefaultTheme()}
}

func (t *rocqTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return RocqOrange
	case theme.ColorNameButton:
		return RocqOrange
	case theme.ColorNameHover:
		return rocqHover
	case theme.ColorNameFocus:
		return rocqAccent
	case theme.ColorNameSelection:
		return rocqSelection
	case theme.ColorNameBackground:
		return rocqLightBg
	case theme.ColorNameForeground:
		return rocqDarkText
	case theme.ColorNameInputBackground:
		return rocqInputBg
	case theme.ColorNameSeparator:
		return rocqSeparator
	case theme.ColorNameDisabled:
		return rocqMutedText
	case theme.ColorNameSuccess:
		return rocqSuccess
	case theme.ColorNameError:
		return rocqError
	}
	return t.base.Color(name, variant)
}

func (t *rocqTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *rocqTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *rocqTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 12
	case theme.SizeNameText:
		return 13
	case theme.SizeNameSubHeadingText:
		return 14
	case theme.SizeNameHeadingText:
		return 18
	}
	return t.base.Size(name)
}
