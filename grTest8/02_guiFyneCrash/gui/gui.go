package gui

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gui/cacheApp"
	"strings"
)

type AppGui struct {
	FyneApp     fyne.App
	FyneWindow  fyne.Window
	MainContent *fyne.Container
	cache       *cacheApp.Cache
}

func NewGUIApp(cache *cacheApp.Cache) (*AppGui, error) {
	if cache == nil {
		return nil, fmt.Errorf("cache is nil")
	}
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("String Processor")

	//интерфейс
	input := widget.NewEntry()
	input.SetPlaceHolder("Enter strings (spare-separated)")

	output := widget.NewMultiLineEntry()
	output.Disabled()
	output.Wrapping = fyne.TextWrapWord
	//обработчик кнопки
	submitButton := widget.NewButton("Process", func() {
		ctx := context.Background()
		inputs := strings.Fields(input.Text)
		var outputText strings.Builder
		results, errors, err := cacheApp.ProcessStringsWithCache(ctx, inputs, 2)
		if err != nil {
			outputText.WriteString(fmt.Sprintf("Error: %v\n", err))
		}
		if len(results) > 0 {
			outputText.WriteString("Results: \n")
			for _, r := range results {
				outputText.WriteString(fmt.Sprintf("Input: %s, Lenght: %v\n", r.Input, r.Output))
			}
		}
		if len(errors) > 0 {
			outputText.WriteString("Errors: \n")
			for _, e := range errors {
				outputText.WriteString(fmt.Sprintf("Input: %s, Error: %v\n", e.Input, e.Error))
			}
		}
		output.SetText(outputText.String())
	})

	mainContent := container.NewVBox(
		widget.NewLabel("Enter strings to calculate their lengths: "),
		input,
		submitButton,
		widget.NewLabel("Output:"),
		output,
	)

	fyneWindow.SetContent(mainContent)
	return &AppGui{
		FyneApp:     fyneApp,
		FyneWindow:  fyneWindow,
		MainContent: mainContent,
		cache:       cache,
	}, nil
}

func (a *AppGui) Run() {
	a.FyneWindow.ShowAndRun()
}
