package gui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"time"
	"wordcount/internal/cache"

	"github.com/sirupsen/logrus"
)

// AppGui управляет интерфейсом Fyne
type AppGui struct {
	FyneApp    fyne.App
	FyneWindow fyne.Window
	cache      *cache.WordCache
	logger     *logrus.Logger
	input      *widget.Entry
	result     *widget.Label
	history    *widget.Label
}

// NewAppGui создаёт новый GUI
func NewAppGui(wordCache *cache.WordCache, logger *logrus.Logger) (*AppGui, error) {
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("WordCount")

	appGui := &AppGui{
		FyneApp:    fyneApp,
		FyneWindow: fyneWindow,
		cache:      wordCache,
		logger:     logger,
		input:      widget.NewEntry(),
		result:     widget.NewLabel("Result: "),
		history:    widget.NewLabel("History: \n"),
	}

	appGui.setupUI()
	return appGui, nil
}

// setupUI настраивает интерфейс
func (appGui *AppGui) setupUI() {
	appGui.input.SetPlaceHolder("Enter a string...")

	submitButton := widget.NewButton("Count Words", func() {
		go appGui.handleInput(appGui.input.Text)
	})

	content := container.NewVBox(
		widget.NewLabel("Word Counter"),
		appGui.input,
		submitButton,
		appGui.result,
		widget.NewSeparator(),
		widget.NewLabel("Request History"),
		appGui.history,
	)

	appGui.FyneWindow.SetContent(container.New(layout.NewVBoxLayout(), content))
	appGui.FyneWindow.Resize(fyne.NewSize(400, 400))
}

// handleInput обрабатывает ввод строки
func (appGui *AppGui) handleInput(input string) {
	fyne.Do(func() {
		appGui.result.SetText("Processing...")
	})

	count, err := appGui.cache.CountWords(input)
	fyne.Do(func() {
		if err != nil {
			appGui.result.SetText(fmt.Sprintf("Error: %v", err))
			appGui.history.Text += fmt.Sprintf("[%v] Error: %v\n", time.Now().Format("15:04:05"), err)
		} else {
			appGui.result.SetText(fmt.Sprintf("Result: %d words", count))
			appGui.history.Text += fmt.Sprintf("[%v] Input: %s, Words: %d\n", time.Now().Format("15:04:05"), input, count)
		}
		appGui.history.Refresh()
	})
}
