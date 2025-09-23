package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"

	_ "embed"
)

//go:embed Montserrat-Regular.ttf
var montserratRegular []byte

//go:embed Montserrat-Bold.ttf
var montserratBold []byte

type CustomTheme struct{}

func (t *CustomTheme) Color(name fyne.ThemeColorName, state fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameForeground:
		return color.Black
	default:
		return theme.DefaultTheme().Color(name, state)
	}
}

func (t *CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Bold {
		return fyne.NewStaticResource("Montserrat-Bold.ttf", montserratBold)
	}
	return fyne.NewStaticResource("Montserrat-Regular.ttf", montserratRegular)
}

func (t *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 30
	default:
		return theme.DefaultTheme().Size(name)
	}
}

func (t *CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
