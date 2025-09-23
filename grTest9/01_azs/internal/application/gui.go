package application

import (
	"fmt"
	"fuelazs/internal/gui"
	"fuelazs/internal/repository/postgres"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"

	"time"
)

type AppGui struct {
	FyneApp     fyne.App
	FyneWindow  fyne.Window
	MainContent *gui.Gui
	repo        *postgres.Activation
}

func NewAppGui(repo *postgres.Activation) *AppGui {
	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow("AZS.APP")

	mainContent := gui.NewGui()

	return &AppGui{
		FyneApp:     fyneApp,
		FyneWindow:  fyneWindow,
		MainContent: mainContent,
		repo:        repo,
	}
}

// Show Метод отрисовки интерфейса программы
func (appGui *AppGui) Show() error {
	appGui.FyneApp.Settings().SetTheme(&gui.CustomTheme{})
	appGui.FyneWindow.SetContent(appGui.MainContent.MainContent)
	appGui.FyneWindow.SetFullScreen(true)
	// Горутина для обновления шапки
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if appGui.MainContent.TopSection.TimeLable != nil && appGui.MainContent.TopSection.DateLable != nil {
				now := time.Now().UTC()
				var timeToDisplay time.Time
				var displayOffsetHours int
				var displayOffsetMinutes int
				var gmtFormat string

				offsetSeconds, err := gui.CalculateOffsetSecondsFromString(appGui.MainContent.TopSection.Timezone)
				if err != nil {
					continue
				}

				timeToDisplay = now.Add(time.Second * time.Duration(offsetSeconds))
				displayOffsetHours = offsetSeconds / 3600
				displayOffsetMinutes = (offsetSeconds % 3600) / 60
				if displayOffsetMinutes == 0 {
					gmtFormat = fmt.Sprintf("GMT%+d", displayOffsetHours)
				} else {
					gmtFormat = fmt.Sprintf("GMT%+d:%02d", displayOffsetHours, displayOffsetMinutes)
				}

				timeString := timeToDisplay.Format("15:04") + fmt.Sprintf(" (%s)", gmtFormat)

				dateString := gui.FormatRussianDate(timeToDisplay)

				fyne.Do(func() {
					if appGui != nil && appGui.MainContent.TopSection.TimeLable != nil {
						appGui.MainContent.TopSection.TimeLable.Text = timeString
						canvas.Refresh(appGui.MainContent.TopSection.TimeLable)
					}
					if appGui != nil && appGui.MainContent.TopSection.DateLable != nil {
						appGui.MainContent.TopSection.DateLable.Text = dateString
						canvas.Refresh(appGui.MainContent.TopSection.DateLable)
					}
					if appGui != nil && appGui.MainContent.TopSection.KazsNumberLable != nil {
						appGui.MainContent.TopSection.KazsNumberLable.Text = fmt.Sprintf("АЗС №%v", appGui.MainContent.TopSection.KazsNumber)
						appGui.MainContent.TopSection.KazsNumberLable.Refresh()
					}
					if appGui != nil && appGui.MainContent.TopSection.SupportNumberLable != nil {
						appGui.MainContent.TopSection.SupportNumberLable.Text = appGui.MainContent.TopSection.SupportNumber
						appGui.MainContent.TopSection.SupportNumberLable.Refresh()
					}
					if appGui.MainContent.TopSection != nil {
						appGui.MainContent.TopSection.Content.Refresh()
					}
				})
			}
		}
	}()

	appGui.FyneWindow.ShowAndRun()
	return nil
}
