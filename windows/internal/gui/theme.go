package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Rocq brand colors extracted from the logo.
var (
	rocqLightBg   = color.NRGBA{R: 0xf7, G: 0xf0, B: 0xeb, A: 0xff} // warm light background
	rocqAccent    = color.NRGBA{R: 0xf0, G: 0xa0, B: 0x30, A: 0xff} // warm golden for focus
	rocqHover     = color.NRGBA{R: 0xed, G: 0xe5, B: 0xdf, A: 0xff} // subtle warm hover tint
	rocqSelection = color.NRGBA{R: 0xe0, G: 0xd6, B: 0xd0, A: 0xff} // soft selection highlight
	rocqSeparator = color.NRGBA{R: 0xe0, G: 0xd6, B: 0xcf, A: 0xff} // subtle separator
	rocqInputBg   = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff} // white input fields
	rocqDarkText  = color.NRGBA{R: 0x2a, G: 0x2a, B: 0x2a, A: 0xff} // near-black text
	rocqMutedText = color.NRGBA{R: 0x6b, G: 0x6b, B: 0x6b, A: 0xff} // muted text
	rocqSuccess   = color.NRGBA{R: 0x2e, G: 0x7d, B: 0x32, A: 0xff} // green for success
	rocqError     = color.NRGBA{R: 0xc6, G: 0x28, B: 0x28, A: 0xff} // red for errors
	rocqOrange    = color.NRGBA{R: 0xe8, G: 0x8a, B: 0x1a, A: 0xff} // warm orange for "Rocq" branding
)

// rocqTheme implements fyne.Theme with Rocq brand colors.
type rocqTheme struct {
	base fyne.Theme
}

func newRocqTheme() fyne.Theme {
	return &rocqTheme{base: theme.DefaultTheme()}
}

func (t *rocqTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return rocqOrange
	case theme.ColorNameButton:
		return rocqOrange
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
