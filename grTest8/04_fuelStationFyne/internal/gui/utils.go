package gui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"image/color"
	"time"
)

var (
	Black = color.RGBA{0, 0, 0, 255}
	Gray  = color.RGBA{128, 128, 128, 255}
	Green = color.RGBA{0, 128, 0, 255}
)

func NewCustomSpacer(size fyne.Size) *fyne.Container {
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(size)
	return container.New(layout.NewVBoxLayout(), rect)
}

func NewFixedVSeparator() *canvas.Rectangle {
	separator := canvas.NewRectangle(color.Black)
	separator.SetMinSize(fyne.NewSize(2, 0))
	return separator
}

func NewCustomSeparator(size fyne.Size) *canvas.Rectangle {
	separator := canvas.NewRectangle(color.Black)
	separator.SetMinSize(size)
	return separator
}

func NewFixedHSeparator() *canvas.Rectangle {
	separator := canvas.NewRectangle(color.Black)
	separator.SetMinSize(fyne.NewSize(0, 2))
	return separator
}

func NewZeroHSeparator() *canvas.Rectangle {
	separator := canvas.NewRectangle(color.Transparent)
	separator.SetMinSize(fyne.NewSize(0, 2))
	return separator
}

func NewFixedSpacer() *fyne.Container {
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(0, 15))
	return container.New(layout.NewVBoxLayout(), rect)
}

func ConvertFloat32ToStringFull(f float32) string {
	return fmt.Sprintf("%.2f", f)
}

func ConvertFloat32ToStringShort(f float32) string {
	return fmt.Sprintf("%.0f", f)
}

func ConvertUnixToString(t int64) string {
	return time.Unix(t, 0).Format("02.01.2006 15:04:05")
}
