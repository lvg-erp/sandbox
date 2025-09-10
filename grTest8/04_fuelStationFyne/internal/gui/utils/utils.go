package utils

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	_ "fyne.io/fyne/v2/container"
	_ "fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

// Цвета
var (
	Black = color.Black
	Gray  = color.RGBA{128, 128, 128, 255}
	Green = color.RGBA{0, 128, 0, 255}
	Red   = color.RGBA{255, 0, 0, 255}
)

// NewFixedHSeparator создаёт горизонтальную разделительную линию
func NewFixedHSeparator() *canvas.Rectangle {
	separator := canvas.NewRectangle(Black)
	separator.SetMinSize(fyne.NewSize(1280, 2))
	return separator
}

// NewFixedVSeparator создаёт вертикальную разделительную линию
func NewFixedVSeparator() *canvas.Rectangle {
	separator := canvas.NewRectangle(Black)
	separator.SetMinSize(fyne.NewSize(2, 720))
	return separator
}

// NewCustomSpacer создаёт пустое пространство заданного размера
func NewCustomSpacer(size fyne.Size) *widget.Label {
	return widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})
}

// NewFixedSpacer создаёт фиксированное пустое пространство
func NewFixedSpacer() *canvas.Rectangle {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(0, 15))
	return spacer
}

// NewZeroHSeparator создаёт горизонтальную линию нулевой ширины
func NewZeroHSeparator() *canvas.Rectangle {
	separator := canvas.NewRectangle(Black)
	separator.SetMinSize(fyne.NewSize(0, 2))
	return separator
}

// NewСustomSeparator создаёт разделитель произвольного размера
func NewСustomSeparator(size fyne.Size) *canvas.Rectangle {
	separator := canvas.NewRectangle(Black)
	separator.SetMinSize(size)
	return separator
}

// FormatRussianDate форматирует дату на русском языке
func FormatRussianDate(t time.Time) string {
	months := []string{
		"января", "февраля", "марта", "апреля", "мая", "июня",
		"июля", "августа", "сентября", "октября", "ноября", "декабря",
	}
	return t.Format("02 ") + months[t.Month()-1] + t.Format(" 2006")
}

// ConvertFloat32ToStringFull форматирует float32 с двумя знаками после запятой
func ConvertFloat32ToStringFull(f float32) string {
	return fmt.Sprintf("%.2f", f)
}

// ConvertFloat32ToStringShort форматирует float32 без дробной части
func ConvertFloat32ToStringShort(f float32) string {
	return fmt.Sprintf("%.0f", f)
}

// ConvertUnixToString конвертирует Unix-время в строку
func ConvertUnixToString(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return FormatRussianDate(t)
}
