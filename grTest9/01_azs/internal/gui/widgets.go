package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"image/color"
)

func NewFixedSpacer() fyne.CanvasObject {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(0, 50))
	return spacer
}

func NewCustomSpacer(size fyne.Size) fyne.CanvasObject {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(size)
	return spacer
}

func NewFixedHSeparator() fyne.CanvasObject {
	separator := canvas.NewRectangle(color.Black)

	separator.SetMinSize(fyne.NewSize(0, 1))

	return separator
}

func NewFixedVSeparator() fyne.CanvasObject {
	separator := canvas.NewRectangle(color.Black)

	separator.SetMinSize(fyne.NewSize(1, 860))

	return separator
}

func NewСustomSeparator(size fyne.Size) fyne.CanvasObject {
	separator := canvas.NewRectangle(color.Black)

	separator.SetMinSize(size)

	return separator
}

func NewZeroHSeparator() fyne.CanvasObject {
	separator := canvas.NewRectangle(color.Black)

	separator.SetMinSize(fyne.NewSize(0, 1))

	return separator
}
