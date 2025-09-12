package gui

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"image"
	"image/color"
	"log"
	"time"
)

// ProcessorFunc - функция обратного вызова для обработки JSON
type ProcessorFunc func(ctx context.Context, g *Gui, db *sql.DB, filePath string, action string, jarNumber string) error

// Gui структура для хранения состояния приложения
type Gui struct {
	App          fyne.App
	DB           *sql.DB
	Window       fyne.Window
	MainContent  *fyne.Container
	LeftSection  *Section
	RightSection *Section
	TopSection   *TopSection
	processJSON  ProcessorFunc
}

// Section структура для секций GUI
type Section struct {
	Content            *fyne.Container
	ActiveDialogCancel context.CancelFunc
	ActiveDialog       *fyne.Container
}

// TopSection структура для верхней секции с временем
type TopSection struct {
	Content            *fyne.Container
	TimeLabel          *canvas.Text
	DateLabel          *canvas.Text
	LogoLabel          *canvas.Image
	SupportNumberLabel *canvas.Text
	KazsNumberLabel    *canvas.Text
	Logo               string
	SupportNumber      string
	KazsNumber         string
	Timezone           string
	Timer              context.CancelFunc
}

// NewFyneApp создаёт новое приложение Fyne
func NewFyneApp() fyne.App {
	return app.New()
}

// NewGui создаёт новый объект Gui
func NewGui() *Gui {
	logo := "iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg=="
	return &Gui{
		MainContent: container.NewVBox(),
		LeftSection: &Section{
			Content:      container.NewVBox(),
			ActiveDialog: nil,
		},
		RightSection: &Section{
			Content:      container.NewVBox(),
			ActiveDialog: nil,
		},
		TopSection: &TopSection{
			Logo:          logo,
			SupportNumber: "8-800-555-35-35",
			KazsNumber:    "1",
			Timezone:      "Asia/Almaty",
		},
	}
}

// SetProcessorFunc устанавливает функцию обработки JSON
func (gui *Gui) SetProcessorFunc(processJSON ProcessorFunc) {
	gui.processJSON = processJSON
}

// CalculateOffsetSecondsFromString вычисляет смещение часового пояса в секундах
func CalculateOffsetSecondsFromString(timezone string) (int, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return 0, fmt.Errorf("не удалось загрузить часовой пояс %s: %s", timezone, err)
	}
	_, offset := time.Now().In(loc).Zone()
	return offset, nil
}

// FormatRussianDate форматирует дату на русском языке
func FormatRussianDate(t time.Time) string {
	months := []string{
		"января", "февраля", "марта", "апреля", "мая", "июня",
		"июля", "августа", "сентября", "октября", "ноября", "декабря",
	}
	day := t.Day()
	month := months[t.Month()-1]
	year := t.Year()
	return fmt.Sprintf("%d %s %d", day, month, year)
}

// CreateHeader создаёт заголовок
func (gui *Gui) CreateHeader() (*fyne.Container, *canvas.Text, *canvas.Text, *canvas.Image, *canvas.Text, *canvas.Text) {
	// Левая часть заголовка (логотип)
	decodedImageBytes, err := base64.StdEncoding.DecodeString(gui.TopSection.Logo)
	if err != nil {
		log.Printf("Ошибка декодирования base64 изображения: %v", err)
	}

	imageReader := bytes.NewReader(decodedImageBytes)
	_, _, _ = image.Decode(bytes.NewReader(decodedImageBytes))

	itecoImage := canvas.NewImageFromReader(imageReader, "iteco_logo.png")
	if itecoImage != nil {
		itecoImage.FillMode = canvas.ImageFillContain
		itecoImage.SetMinSize(fyne.NewSize(368, 150))
	}

	// Средняя часть заголовка
	phoneLabel := canvas.NewText(gui.TopSection.SupportNumber, color.Black)
	supportLabel := canvas.NewText("Техническая поддержка", color.Black)
	phoneLabel.Alignment = fyne.TextAlignCenter
	supportLabel.Alignment = fyne.TextAlignCenter
	phoneLabel.TextStyle = fyne.TextStyle{Bold: true}
	phoneLabel.TextSize = 30
	supportLabel.TextSize = 26
	centerHeaderContent := container.NewVBox(NewCustomSpacer(fyne.NewSize(0, 15)), phoneLabel, supportLabel)

	// Правая часть заголовка
	now := time.Now()
	azsLabel := canvas.NewText(fmt.Sprintf("АЗС №%v", gui.TopSection.KazsNumber), color.Black)
	dateLabel := canvas.NewText(FormatRussianDate(now), color.Black)
	_, offsetSeconds := now.Zone()
	offsetHours := offsetSeconds / 3600
	timeString := now.Format("15:04") + fmt.Sprintf(" (GMT%+d)", offsetHours)
	timeLabel := canvas.NewText(timeString, color.Black)
	azsLabel.Alignment = fyne.TextAlignTrailing
	dateLabel.Alignment = fyne.TextAlignTrailing
	timeLabel.Alignment = fyne.TextAlignTrailing
	azsLabel.TextStyle = fyne.TextStyle{Bold: true}
	azsLabel.TextSize = 30
	dateLabel.TextSize = 24
	timeLabel.TextSize = 34
	rightHeaderContent := container.NewVBox(azsLabel, dateLabel, timeLabel)
	rightHeaderContentCentered := container.NewCenter(rightHeaderContent)
	rightHeader := container.NewHBox(rightHeaderContentCentered, NewCustomSpacer(fyne.NewSize(15, 0)))

	// Собираем заголовок
	topSectionContent := container.New(
		layout.NewBorderLayout(nil, nil, itecoImage, rightHeader),
		itecoImage,
		rightHeader,
		centerHeaderContent,
	)
	topSectionContainer := container.NewVBox(topSectionContent, NewFixedHSeparator())

	fyne.Do(func() {
		gui.TopSection.DateLabel = dateLabel
		gui.TopSection.TimeLabel = timeLabel
		gui.TopSection.LogoLabel = itecoImage
		gui.TopSection.SupportNumberLabel = phoneLabel
		gui.TopSection.KazsNumberLabel = azsLabel
		gui.TopSection.Content = topSectionContainer
		gui.TopSection.Content.Refresh()
	})

	return topSectionContainer, timeLabel, dateLabel, itecoImage, phoneLabel, azsLabel
}

func (gui *Gui) RunGui(a fyne.App, ready chan<- struct{}, db *sql.DB, processJSON ProcessorFunc) error {
	log.Println("RunGui: Начало инициализации GUI")
	gui.App = a
	gui.DB = db
	gui.Window = a.NewWindow("Fuel Station")
	gui.processJSON = processJSON

	// Настройка верхней секции
	_, _, _, _, _, _ = gui.CreateHeader()

	// Основной контент с серым фоном
	background := canvas.NewRectangle(color.RGBA{128, 128, 128, 255})
	gui.MainContent = container.NewStack(
		background,
		container.NewHBox(
			gui.LeftSection.Content,
			NewFixedVSeparator(),
			gui.RightSection.Content,
		),
	)
	gui.Window.SetContent(container.NewVBox(
		gui.TopSection.Content,
		NewFixedHSeparator(),
		gui.MainContent,
	))
	gui.Window.Resize(fyne.NewSize(800, 600))

	// Горутина для обновления времени
	ctx, cancel := context.WithCancel(context.Background())
	gui.TopSection.Timer = cancel
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if gui.TopSection != nil && gui.TopSection.TimeLabel != nil && gui.TopSection.DateLabel != nil {
					now := time.Now()
					offsetSeconds, err := CalculateOffsetSecondsFromString(gui.TopSection.Timezone)
					if err != nil {
						log.Printf("RunGui: Ошибка вычисления смещения часового пояса: %v", err)
						continue
					}
					timeToDisplay := now.Add(time.Second * time.Duration(offsetSeconds))
					displayOffsetHours := offsetSeconds / 3600
					displayOffsetMinutes := (offsetSeconds % 3600) / 60
					gmtFormat := fmt.Sprintf("GMT%+d", displayOffsetHours)
					if displayOffsetMinutes != 0 {
						gmtFormat = fmt.Sprintf("GMT%+d:%02d", displayOffsetHours, displayOffsetMinutes)
					}
					timeString := timeToDisplay.Format("15:04") + fmt.Sprintf(" (%s)", gmtFormat)
					dateString := FormatRussianDate(timeToDisplay)

					fyne.Do(func() {
						gui.TopSection.TimeLabel.Text = timeString
						gui.TopSection.DateLabel.Text = dateString
						gui.TopSection.KazsNumberLabel.Text = fmt.Sprintf("АЗС №%v", gui.TopSection.KazsNumber)
						gui.TopSection.SupportNumberLabel.Text = gui.TopSection.SupportNumber
						gui.TopSection.TimeLabel.Refresh()
						gui.TopSection.DateLabel.Refresh()
						gui.TopSection.KazsNumberLabel.Refresh()
						gui.TopSection.SupportNumberLabel.Refresh()
						gui.TopSection.Content.Refresh()
					})
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Инициализация начальных экранов для колонок
	gui.CreateFuelGiveStartScreen("1", 0, "", 0)
	gui.CreateFuelGiveStartScreen("2", 0, "", 0)

	log.Println("RunGui: Отправка сигнала ready")
	close(ready)

	log.Println("RunGui: Запуск приложения")
	gui.Window.ShowAndRun()
	return nil
}

func (gui *Gui) CreateFuelGiveStartScreen(jarNumber string, liters float32, fuelType string, timer int) *fyne.Container {
	if jarNumber == "" {
		log.Printf("CreateFuelGiveStartScreen: Пустой jarNumber, пропуск обновления")
		return nil
	}

	// Текст "Вставьте в бензобак"
	insertText := canvas.NewText("Вставьте в бензобак", Black)
	insertText.Alignment = fyne.TextAlignCenter
	insertText.TextSize = 32

	// Текст "ПИСТОЛЕТ №X"
	gunText := canvas.NewText(fmt.Sprintf("ПИСТОЛЕТ №%v", jarNumber), Black)
	gunText.Alignment = fyne.TextAlignCenter
	gunText.TextSize = 40
	gunText.TextStyle = fyne.TextStyle{Bold: true}

	// Текст "Тип_топлива"
	fuelTypeText := widget.NewLabel(fuelType)
	fuelTypeText.Alignment = fyne.TextAlignCenter
	fuelTypeText.Wrapping = fyne.TextWrapWord

	// Текст "Максимальный объем"
	maxVolumeText := canvas.NewText("Максимальный объем", Gray)
	maxVolumeText.Alignment = fyne.TextAlignCenter
	maxVolumeText.TextSize = 32

	// Текст "количество_литров"
	litersText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringFull(liters)), Black)
	litersText.TextSize = 40
	litersText.TextStyle = fyne.TextStyle{Bold: true}
	litersText.Alignment = fyne.TextAlignCenter

	maxVolumeAndLitersArea := container.NewVBox(maxVolumeText, NewCustomSpacer(fyne.NewSize(0, 5)), litersText)
	maxVolumeAndLitersContainer := container.NewCenter(maxVolumeAndLitersArea)

	// Кнопка "Залить"
	fillButtonText := canvas.NewText("Залить", color.RGBA{0, 128, 0, 255})
	fillButtonText.Alignment = fyne.TextAlignCenter
	fillButtonText.TextSize = 32
	fillButtonBorder := canvas.NewRectangle(color.Transparent)
	fillButtonBorder.StrokeColor = color.Black
	fillButtonBorder.CornerRadius = 10.0
	fillButtonBorder.StrokeWidth = 2
	fillButtonBorder.SetMinSize(fyne.NewSize(150, 60))
	fillButtonContent := container.NewPadded(container.NewCenter(fillButtonText))
	fillButtonPadded := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 5)), NewCustomSpacer(fyne.NewSize(0, 5)), NewCustomSpacer(fyne.NewSize(20, 0)), NewCustomSpacer(fyne.NewSize(20, 0)), fillButtonContent)
	fillButtonVisual := container.NewStack(fillButtonBorder, fillButtonPadded)
	fillButton := widget.NewButton("", func() {
		log.Printf("CreateFuelGiveStartScreen: Нажата кнопка Залить для колонки %s", jarNumber)
		if err := gui.processJSON(context.Background(), gui, gui.DB, "operations.json", "fill", jarNumber); err != nil {
			log.Printf("CreateFuelGiveStartScreen: Ошибка обработки fill для колонки %s: %v", jarNumber, err)
			gui.ShowSectionDialog(gui.getSectionContent(jarNumber), "Ошибка", fmt.Sprintf("Не удалось выполнить заправку: %v", err), 10, nil)
		}
	})
	fillButton.SetText("") // Прозрачный текст для кнопки
	fillButtonArea := container.NewStack(fillButtonVisual, fillButton)

	// Кнопка "Слить"
	drainButtonText := canvas.NewText("Слить", color.RGBA{0, 128, 0, 255})
	drainButtonText.Alignment = fyne.TextAlignCenter
	drainButtonText.TextSize = 32
	drainButtonBorder := canvas.NewRectangle(color.Transparent)
	drainButtonBorder.StrokeColor = color.Black
	drainButtonBorder.CornerRadius = 10.0
	drainButtonBorder.StrokeWidth = 2
	drainButtonBorder.SetMinSize(fyne.NewSize(150, 60))
	drainButtonContent := container.NewPadded(container.NewCenter(drainButtonText))
	drainButtonPadded := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 5)), NewCustomSpacer(fyne.NewSize(0, 5)), NewCustomSpacer(fyne.NewSize(20, 0)), NewCustomSpacer(fyne.NewSize(20, 0)), drainButtonContent)
	drainButtonVisual := container.NewStack(drainButtonBorder, drainButtonPadded)
	drainButton := widget.NewButton("", func() {
		log.Printf("CreateFuelGiveStartScreen: Нажата кнопка Слить для колонки %s", jarNumber)
		if err := gui.processJSON(context.Background(), gui, gui.DB, "operations.json", "drain", jarNumber); err != nil {
			log.Printf("CreateFuelGiveStartScreen: Ошибка обработки drain для колонки %s: %v", jarNumber, err)
			gui.ShowSectionDialog(gui.getSectionContent(jarNumber), "Ошибка", fmt.Sprintf("Не удалось выполнить слив: %v", err), 10, nil)
		}
	})
	drainButton.SetText("") // Прозрачный текст для кнопки
	drainButtonArea := container.NewStack(drainButtonVisual, drainButton)

	buttons := container.NewCenter(container.NewHBox(fillButtonArea, NewCustomSpacer(fyne.NewSize(10, 0)), drainButtonArea))

	// Нижняя рамка с текстом
	buttonText1 := canvas.NewText("Для начала заправки вставьте", Black)
	buttonText1.Alignment = fyne.TextAlignCenter
	buttonText1.TextSize = 32
	buttonText1.TextStyle = fyne.TextStyle{Bold: true}

	buttonText2 := canvas.NewText(fmt.Sprintf("ПИСТОЛЕТ №%v в бензобак в", jarNumber), Black)
	buttonText2.Alignment = fyne.TextAlignCenter
	buttonText2.TextSize = 32
	buttonText2.TextStyle = fyne.TextStyle{Bold: true}

	buttonText3 := canvas.NewText(fmt.Sprintf("в течение %v секунд", timer), Black)
	buttonText3.Alignment = fyne.TextAlignCenter
	buttonText3.TextSize = 32
	buttonText3.TextStyle = fyne.TextStyle{Bold: true}

	buttonText := container.NewVBox(buttonText1, buttonText2, buttonText3)

	borderRect := canvas.NewRectangle(color.Transparent)
	borderRect.CornerRadius = 10.0
	borderRect.StrokeColor = color.Black
	borderRect.StrokeWidth = 2

	paddedButtonText := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 2)), NewCustomSpacer(fyne.NewSize(0, 2)), NewCustomSpacer(fyne.NewSize(20, 0)), NewCustomSpacer(fyne.NewSize(20, 0)), buttonText)
	buttonArea := container.NewStack(borderRect, paddedButtonText)
	buttonAreaContainer := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(10, 0)), NewCustomSpacer(fyne.NewSize(10, 0)), buttonArea)

	// Собираем контент
	topCenterContent := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(0, 20)),
		insertText,
		gunText,
		NewCustomSpacer(fyne.NewSize(0, 60)),
		fuelTypeText,
		NewCustomSpacer(fyne.NewSize(0, 15)),
		maxVolumeAndLitersContainer,
		NewCustomSpacer(fyne.NewSize(0, 20)),
		buttons,
		NewCustomSpacer(fyne.NewSize(0, 10)),
	)

	columnContent := container.New(layout.NewBorderLayout(topCenterContent, buttonAreaContainer, nil, nil),
		topCenterContent,
		NewCustomSpacer(fyne.NewSize(0, 180)),
		buttonAreaContainer,
	)

	// Обновляем соответствующую секцию
	fyne.Do(func() {
		section := gui.getSectionContent(jarNumber)
		section.RemoveAll()
		section.Add(columnContent)
		section.Refresh()
	})

	return columnContent
}

func (gui *Gui) CreateFuelGiveInProgressScreen(jarNumber string, fuelType string, liters float32, maxLiters float32) *fyne.Container {
	// Устанавливаем текущее время начала заправки
	fillTimestamp := time.Now().Unix()

	// Текст "Заправка в процессе"
	ifProgressText := canvas.NewText("Заправка в процессе", Black)
	ifProgressText.Alignment = fyne.TextAlignCenter
	ifProgressText.TextSize = 32

	// Текст "ПИСТОЛЕТ №X"
	gunText := canvas.NewText(fmt.Sprintf("ПИСТОЛЕТ №%v", jarNumber), Black)
	gunText.Alignment = fyne.TextAlignCenter
	gunText.TextStyle = fyne.TextStyle{Bold: true}
	gunText.TextSize = 40

	// Текст "Тип_топлива"
	fuelTypeText := widget.NewLabel(fuelType)
	fuelTypeText.Alignment = fyne.TextAlignCenter
	fuelTypeText.Wrapping = fyne.TextWrapWord

	// Текст "Максимальный объем"
	maxVolumeText := canvas.NewText("Максимальный объем", Gray)
	maxVolumeText.Alignment = fyne.TextAlignCenter
	maxVolumeText.TextSize = 32

	// Текст "количество_литров"
	volumeValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringFull(maxLiters)), Black)
	volumeValueText.TextStyle = fyne.TextStyle{Bold: true}
	volumeValueText.TextSize = 40
	volumeValueText.Alignment = fyne.TextAlignCenter

	maxVolumeAndVolumeArea := container.NewVBox(maxVolumeText, volumeValueText)
	maxVolumeAndVolumeContainer := container.NewCenter(maxVolumeAndVolumeArea)

	// Текст "количество_заправленных_литров"
	amountText := canvas.NewText(fmt.Sprintf("%.2f", liters), Black)
	amountText.Alignment = fyne.TextAlignCenter
	amountText.TextSize = 100

	// Текст "литров залито"
	litersText := canvas.NewText("литров залито", Gray)
	litersText.Alignment = fyne.TextAlignCenter
	litersText.TextSize = 40

	// Прогресс-бар с анимацией (вертикально слева, чёрный)
	progressBarHeight := float32(329)
	progressBarWidth := float32(85)
	borderThickness := float32(2)
	progressBarBackground := canvas.NewRectangle(color.Transparent)
	progressBarBackground.StrokeColor = color.Black
	progressBarBackground.StrokeWidth = borderThickness
	progressBarBackground.CornerRadius = 10.0
	progressBarBackground.SetMinSize(fyne.NewSize(progressBarWidth, progressBarHeight))

	innerBarHeight := progressBarHeight - 2*borderThickness
	innerBarWidth := progressBarWidth - 2*borderThickness
	progressBarFilled := canvas.NewRectangle(color.Black)
	progressBarFilled.SetMinSize(fyne.NewSize(innerBarWidth, 0)) // Начальная высота 0
	progressBarFilled.CornerRadius = 10.0

	filledBarContent := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(innerBarWidth, innerBarHeight)),
		progressBarFilled,
	)
	progressBarArea := container.NewStack(progressBarBackground, filledBarContent)

	// Текст процента
	percentageText := canvas.NewText("0%", Black)
	percentageText.Alignment = fyne.TextAlignCenter
	percentageText.TextSize = 40
	percentageText.TextStyle.Bold = true
	percentageContainer := container.NewCenter(percentageText)

	amountAndLiters := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(0, 10)),
		amountText,
		litersText,
	)
	amountAndLitersAligned := container.NewCenter(amountAndLiters)

	// Прогресс-бар слева
	progressBarContainer := container.NewVBox(progressBarArea, percentageContainer)
	progressBarAligned := container.NewCenter(progressBarContainer)

	// Основной контент справа
	mainContent := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(0, 20)),
		ifProgressText,
		gunText,
		NewCustomSpacer(fyne.NewSize(0, 60)),
		fuelTypeText,
		NewCustomSpacer(fyne.NewSize(0, 10)),
		maxVolumeAndVolumeContainer,
		NewCustomSpacer(fyne.NewSize(0, 10)),
		amountAndLitersAligned,
	)

	// Собираем контент с прогресс-баром слева
	contentRow := container.NewHBox(
		NewCustomSpacer(fyne.NewSize(10, 0)),
		progressBarAligned,
		NewCustomSpacer(fyne.NewSize(20, 0)),
		mainContent,
		NewCustomSpacer(fyne.NewSize(10, 0)),
	)

	// Анимация прогресс-бара (20 секунд)
	duration := 20 * time.Second
	animation := fyne.NewAnimation(duration, func(t float32) {
		percentage := t * 100
		filledHeight := percentage * innerBarHeight / 100.0
		if filledHeight > innerBarHeight {
			filledHeight = innerBarHeight
		}
		fyne.Do(func() {
			progressBarFilled.SetMinSize(fyne.NewSize(innerBarWidth, filledHeight))
			filledBarContent.Objects[0].(*fyne.Container).Objects[0].(*canvas.Rectangle).SetMinSize(fyne.NewSize(innerBarWidth, innerBarHeight-filledHeight))
			percentageText.Text = fmt.Sprintf("%.0f%%", percentage)
			filledBarContent.Refresh()
			progressBarArea.Refresh()
			percentageText.Refresh()
		})
	})
	animation.AutoReverse = false
	animation.Start()

	// Нижняя рамка
	buttonText1 := canvas.NewText("Для завершения заправки", Black)
	buttonText1.Alignment = fyne.TextAlignCenter
	buttonText1.TextSize = 32
	buttonText1.TextStyle = fyne.TextStyle{Bold: true}

	buttonText2 := canvas.NewText(fmt.Sprintf("повесьте ПИСТОЛЕТ №%v", jarNumber), Black)
	buttonText2.Alignment = fyne.TextAlignCenter
	buttonText2.TextSize = 32
	buttonText2.TextStyle = fyne.TextStyle{Bold: true}

	buttonText := container.NewVBox(buttonText1, buttonText2)
	borderRect := canvas.NewRectangle(color.Transparent)
	borderRect.StrokeColor = color.Black
	borderRect.CornerRadius = 10.0
	borderRect.StrokeWidth = 2
	paddedButtonText := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 2)), NewCustomSpacer(fyne.NewSize(0, 2)), NewCustomSpacer(fyne.NewSize(20, 0)), NewCustomSpacer(fyne.NewSize(20, 0)), buttonText)
	buttonArea := container.NewStack(borderRect, paddedButtonText)
	buttonAreaContainer := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(10, 0)), NewCustomSpacer(fyne.NewSize(10, 0)), buttonArea)

	// Собираем контент
	columnContent := container.New(layout.NewBorderLayout(contentRow, buttonAreaContainer, nil, nil),
		contentRow,
		NewCustomSpacer(fyne.NewSize(0, 20)),
		buttonAreaContainer,
	)

	// Обновляем соответствующую секцию и переходим к экрану завершения
	fyne.Do(func() {
		section := gui.getSectionContent(jarNumber)
		section.RemoveAll()
		section.Add(columnContent)
		section.Refresh()
	})

	// Переход к экрану завершения через 20 секунд
	go func() {
		time.Sleep(20 * time.Second)
		fyne.Do(func() {
			gui.CreateFuelGiveCompleteScreen(jarNumber, fuelType, "DOC123", liters, maxLiters, fillTimestamp, fillTimestamp+20)
		})
	}()

	return columnContent
}

func (gui *Gui) CreateFuelGiveCompleteScreen(jarNumber string, fuelType string, doc string, liters float32, maxLiters float32, fillTimestamp int64, completeTimestamp int64) *fyne.Container {
	// Текст "ЗАПРАВКА ЗАВЕРШЕНА"
	completeText := canvas.NewText("ЗАПРАВКА ЗАВЕРШЕНА", Black)
	completeText.Alignment = fyne.TextAlignCenter
	completeText.TextSize = 40
	completeText.TextStyle = fyne.TextStyle{Bold: true}

	// Текст "тип_топлива + Пистолет №X"
	fuelTypePistolText := widget.NewLabel(fmt.Sprintf("%s Пистолет №%v", fuelType, jarNumber))
	fuelTypePistolText.Alignment = fyne.TextAlignCenter
	fuelTypePistolText.Wrapping = fyne.TextWrapWord

	// Данные о заправке
	documentText := canvas.NewText("№ документа", Gray)
	documentValueText := canvas.NewText(doc, Black)
	documentText.TextSize = 32
	documentValueText.TextSize = 32
	documentValueText.Alignment = fyne.TextAlignTrailing
	documentValueText.TextStyle = fyne.TextStyle{Bold: true}

	planText := canvas.NewText("Заправка план", Gray)
	planValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringFull(maxLiters)), Black)
	planText.TextSize = 32
	planValueText.TextSize = 32
	planValueText.Alignment = fyne.TextAlignTrailing
	planValueText.TextStyle = fyne.TextStyle{Bold: true}

	factText := canvas.NewText("Заправка факт", Gray)
	factValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringFull(liters)), Black)
	factText.TextSize = 32
	factValueText.TextSize = 32
	factValueText.Alignment = fyne.TextAlignTrailing
	factValueText.TextStyle = fyne.TextStyle{Bold: true}

	startDateText := canvas.NewText("Дата начала", Gray)
	startDateValueText := canvas.NewText(ConvertUnixToString(fillTimestamp), Black)
	startDateText.TextSize = 32
	startDateValueText.TextSize = 32
	startDateValueText.Alignment = fyne.TextAlignTrailing
	startDateValueText.TextStyle = fyne.TextStyle{Bold: true}

	endDateText := canvas.NewText("Дата окончания", Gray)
	endDateValueText := canvas.NewText(ConvertUnixToString(completeTimestamp), Black)
	endDateText.TextSize = 32
	endDateValueText.TextSize = 32
	endDateValueText.Alignment = fyne.TextAlignTrailing
	endDateValueText.TextStyle = fyne.TextStyle{Bold: true}

	data1 := container.NewGridWithColumns(2, documentText, documentValueText)
	data1Container := container.NewVBox(data1, NewCustomSpacer(fyne.NewSize(0, 15)))
	data2 := container.NewGridWithColumns(2, planText, planValueText)
	data2Container := container.NewVBox(data2, NewCustomSpacer(fyne.NewSize(0, 15)))
	data3 := container.NewGridWithColumns(2, factText, factValueText)
	data3Container := container.NewVBox(data3, NewCustomSpacer(fyne.NewSize(0, 15)))
	data4 := container.NewGridWithColumns(2, startDateText, startDateValueText)
	data4Container := container.NewVBox(data4, NewCustomSpacer(fyne.NewSize(0, 15)))
	data5 := container.NewGridWithColumns(2, endDateText, endDateValueText)
	data5Container := container.NewVBox(data5, NewCustomSpacer(fyne.NewSize(0, 15)))

	data := container.NewGridWithRows(5, data1Container, data2Container, data3Container, data4Container, data5Container)
	dataPadding := container.NewPadded(data)
	dataBorder := container.NewBorder(nil, nil, NewCustomSpacer(fyne.NewSize(5, 0)), NewCustomSpacer(fyne.NewSize(5, 0)), dataPadding)

	// Рамка снизу
	buttonText1 := canvas.NewText(fmt.Sprintf("Заправка с ПИСТОЛЕТА №%v", jarNumber), Black)
	buttonText1.Alignment = fyne.TextAlignCenter
	buttonText1.TextSize = 32
	buttonText1.TextStyle = fyne.TextStyle{Bold: true}

	buttonText2 := canvas.NewText("возможна через 10 секунд", Black)
	buttonText2.Alignment = fyne.TextAlignCenter
	buttonText2.TextSize = 32
	buttonText2.TextStyle = fyne.TextStyle{Bold: true}

	buttonText := container.NewVBox(buttonText1, buttonText2)
	borderRect := canvas.NewRectangle(color.Transparent)
	borderRect.StrokeColor = color.Black
	borderRect.CornerRadius = 10.0
	borderRect.StrokeWidth = 2
	paddedButtonText := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 2)), NewCustomSpacer(fyne.NewSize(0, 2)), NewCustomSpacer(fyne.NewSize(20, 0)), NewCustomSpacer(fyne.NewSize(20, 0)), buttonText)
	buttonArea := container.NewStack(borderRect, paddedButtonText)
	buttonAreaContainer := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(10, 0)), NewCustomSpacer(fyne.NewSize(10, 0)), buttonArea)

	// Собираем контент
	topCenterContent := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(0, 15)),
		completeText,
		NewCustomSpacer(fyne.NewSize(0, 15)),
		fuelTypePistolText,
		NewCustomSpacer(fyne.NewSize(0, 15)),
		dataBorder,
	)

	columnContent := container.New(layout.NewBorderLayout(topCenterContent, buttonAreaContainer, nil, nil),
		topCenterContent,
		NewCustomSpacer(fyne.NewSize(0, 75)),
		buttonAreaContainer,
	)

	// Обновляем соответствующую секцию
	fyne.Do(func() {
		section := gui.getSectionContent(jarNumber)
		section.RemoveAll()
		section.Add(columnContent)
		section.Refresh()
	})

	// Таймер для возврата к начальному экрану через 10 секунд
	go func() {
		time.Sleep(10 * time.Second)
		fyne.Do(func() {
			gui.CreateFuelGiveStartScreen(jarNumber, 0, "", 0)
		})
	}()

	return columnContent
}

func (gui *Gui) CreateFuelGetStartScreen(jarNumber string, fuelType string, tankLiters int, maxTankLiters int, liters float32, timer int) *fyne.Container {
	percentage := int(float32(tankLiters) / float32(maxTankLiters) * 100.0)

	// Текст "Слив бензовоза"
	drainText := canvas.NewText("Слив бензовоза", Black)
	drainText.Alignment = fyne.TextAlignCenter
	drainText.TextSize = 32

	// Текст "ЕМКОСТЬ №X"
	tankText := canvas.NewText(fmt.Sprintf("ЁМКОСТЬ №%v", jarNumber), Black)
	tankText.Alignment = fyne.TextAlignCenter
	tankText.TextSize = 40
	tankText.TextStyle.Bold = true

	// Текст с типом топлива
	fuelTypeText := widget.NewLabel(fuelType)
	fuelTypeText.Alignment = fyne.TextAlignCenter
	fuelTypeText.Wrapping = fyne.TextWrapWord

	// Данные о заправке
	filledAmountText := canvas.NewText("Заполнено", Gray)
	filledAmountText.Alignment = fyne.TextAlignCenter
	filledAmountText.TextSize = 32
	filledAmountValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringShort(float32(tankLiters))), Black)
	filledAmountValueText.Alignment = fyne.TextAlignCenter
	filledAmountValueText.TextSize = 40
	filledAmountValueText.TextStyle = fyne.TextStyle{Bold: true}

	availableAmountText := canvas.NewText("Доступный объем", Gray)
	availableAmountText.Alignment = fyne.TextAlignCenter
	availableAmountText.TextSize = 32
	availableAmountValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringShort(float32(maxTankLiters-tankLiters))), Black)
	availableAmountValueText.Alignment = fyne.TextAlignCenter
	availableAmountValueText.TextSize = 40
	availableAmountValueText.TextStyle = fyne.TextStyle{Bold: true}

	expectedAmountText := canvas.NewText("Ожидаемый слив", Gray)
	expectedAmountText.Alignment = fyne.TextAlignCenter
	expectedAmountText.TextSize = 32
	expectedAmountValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringShort(liters)), Black)
	expectedAmountValueText.Alignment = fyne.TextAlignCenter
	expectedAmountValueText.TextSize = 40
	expectedAmountValueText.TextStyle = fyne.TextStyle{Bold: true}

	fuelGetDataContainer := container.NewVBox(filledAmountText, filledAmountValueText, availableAmountText, availableAmountValueText, expectedAmountText, expectedAmountValueText)
	fuelGetDataAligned := container.NewCenter(fuelGetDataContainer)

	// Прогресс-бар с анимацией (вертикально слева)
	progressBarHeight := float32(329)
	progressBarWidth := float32(85)
	borderThickness := float32(3)
	progressBarBackground := canvas.NewRectangle(color.Transparent)
	progressBarBackground.StrokeColor = color.Black
	progressBarBackground.StrokeWidth = borderThickness
	progressBarBackground.CornerRadius = 10.0
	progressBarBackground.SetMinSize(fyne.NewSize(progressBarWidth, progressBarHeight))

	innerBarHeight := progressBarHeight - 2*borderThickness
	innerBarWidth := progressBarWidth - 2*borderThickness
	filledHeight := float32(percentage) * innerBarHeight / 100.0
	if filledHeight < 0 {
		filledHeight = 0
	}
	if filledHeight > innerBarHeight {
		filledHeight = innerBarHeight
	}
	progressBarFilled := canvas.NewRectangle(color.Black)
	progressBarFilled.SetMinSize(fyne.NewSize(innerBarWidth, 0)) // Начальная высота 0
	progressBarFilled.CornerRadius = 10.0

	filledBarContent := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(innerBarWidth, innerBarHeight-filledHeight)),
		progressBarFilled,
	)
	progressBarArea := container.NewStack(progressBarBackground, filledBarContent)

	// Текст процента
	percentageText := canvas.NewText(fmt.Sprintf("%v%%", percentage), Black)
	percentageText.Alignment = fyne.TextAlignCenter
	percentageText.TextSize = 40
	percentageText.TextStyle.Bold = true
	percentageContainer := container.NewCenter(percentageText)

	progressBarContainer := container.NewVBox(progressBarArea, percentageContainer)
	progressBarAligned := container.NewCenter(progressBarContainer)

	// Основной контент справа
	mainContent := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(0, 20)),
		drainText,
		tankText,
		NewCustomSpacer(fyne.NewSize(0, 20)),
		fuelTypeText,
		NewCustomSpacer(fyne.NewSize(0, 10)),
		fuelGetDataAligned,
	)

	// Собираем контент с прогресс-баром слева
	contentRow := container.NewHBox(
		NewCustomSpacer(fyne.NewSize(10, 0)),
		progressBarAligned,
		NewCustomSpacer(fyne.NewSize(20, 0)),
		mainContent,
		NewCustomSpacer(fyne.NewSize(10, 0)),
	)

	// Анимация прогресс-бара
	startHeight := float32(0)
	endHeight := filledHeight
	animation := fyne.NewAnimation(time.Second, func(t float32) {
		currentHeight := startHeight + (endHeight-startHeight)*t
		fyne.Do(func() {
			progressBarFilled.SetMinSize(fyne.NewSize(innerBarWidth, currentHeight))
			filledBarContent.Objects[0].(*fyne.Container).Objects[0].(*canvas.Rectangle).SetMinSize(fyne.NewSize(innerBarWidth, innerBarHeight-currentHeight))
			filledBarContent.Refresh()
			progressBarArea.Refresh()
		})
	})
	animation.AutoReverse = false
	animation.Start()

	// Рамка снизу
	buttonText1 := canvas.NewText("Для начала слива откройте", Black)
	buttonText1.Alignment = fyne.TextAlignCenter
	buttonText1.TextSize = 32
	buttonText1.TextStyle = fyne.TextStyle{Bold: true}

	buttonText2 := canvas.NewText(fmt.Sprintf("люк в течение %v секунд", timer), Black)
	buttonText2.Alignment = fyne.TextAlignCenter
	buttonText2.TextSize = 32
	buttonText2.TextStyle = fyne.TextStyle{Bold: true}

	buttonText := container.NewVBox(buttonText1, buttonText2)
	borderRect := canvas.NewRectangle(color.Transparent)
	borderRect.CornerRadius = 10.0
	borderRect.StrokeColor = color.Black
	borderRect.StrokeWidth = 2
	paddedButtonText := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 2)), NewCustomSpacer(fyne.NewSize(0, 2)), NewCustomSpacer(fyne.NewSize(20, 0)), NewCustomSpacer(fyne.NewSize(20, 0)), buttonText)
	buttonArea := container.NewStack(borderRect, paddedButtonText)
	buttonAreaContainer := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(10, 0)), NewCustomSpacer(fyne.NewSize(10, 0)), buttonArea)

	// Собираем контент
	columnContent := container.New(layout.NewBorderLayout(contentRow, buttonAreaContainer, nil, nil),
		contentRow,
		NewCustomSpacer(fyne.NewSize(0, 35)),
		buttonAreaContainer,
	)

	// Обновляем соответствующую секцию
	fyne.Do(func() {
		section := gui.getSectionContent(jarNumber)
		section.RemoveAll()
		section.Add(columnContent)
		section.Refresh()
	})

	return columnContent
}

func (gui *Gui) CreateFuelGetInProgressScreen(jarNumber string, fuelType string, liters float32, maxLiters float32, tankLiters int, maxTankLiters int, timer int) *fyne.Container {
	log.Printf("CreateFuelGetInProgressScreen: jarNumber=%s, fuelType=%s, liters=%v, maxLiters=%v, tankLiters=%v, maxTankLiters=%v, timer=%v", jarNumber, fuelType, liters, maxLiters, tankLiters, maxTankLiters, timer)

	drainTimestamp := time.Now().Unix()

	drainText := canvas.NewText("Слив бензовоза", Black)
	drainText.Alignment = fyne.TextAlignCenter
	drainText.TextSize = 32

	jarText := canvas.NewText(fmt.Sprintf("ЁМКОСТЬ №%v", jarNumber), Black)
	jarText.Alignment = fyne.TextAlignCenter
	jarText.TextStyle = fyne.TextStyle{Bold: true}
	jarText.TextSize = 40

	fuelTypeText := widget.NewLabel(fuelType)
	if fuelType == "" {
		fuelTypeText.Text = "Неизвестно"
	}
	fuelTypeText.Alignment = fyne.TextAlignCenter
	fuelTypeText.Wrapping = fyne.TextWrapWord

	expectedText := canvas.NewText("Ожидаемый слив", Gray)
	expectedText.Alignment = fyne.TextAlignCenter
	expectedText.TextSize = 32
	expectedContainer := container.NewCenter(expectedText)

	expectedValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringShort(maxLiters)), Black)
	expectedValueText.Alignment = fyne.TextAlignCenter
	expectedValueText.TextStyle = fyne.TextStyle{Bold: true}
	expectedValueText.TextSize = 40
	expectedValueContainer := container.NewCenter(expectedValueText)

	drainedValueText := canvas.NewText(fmt.Sprintf("%v", ConvertFloat32ToStringShort(liters)), Black)
	drainedValueText.Alignment = fyne.TextAlignCenter
	drainedValueText.TextStyle = fyne.TextStyle{Bold: true}
	drainedValueText.TextSize = 100

	drainedText := canvas.NewText("литров слито", Gray)
	drainedText.Alignment = fyne.TextAlignCenter
	drainedText.TextSize = 40

	// Прогресс-бар (увеличиваем ширину для видимости)
	progressBarHeight := float32(329)
	progressBarWidth := float32(150) // Увеличено с 85 до 150
	borderThickness := float32(2)
	progressBarBackground := canvas.NewRectangle(color.Transparent)
	progressBarBackground.StrokeColor = color.Black
	progressBarBackground.StrokeWidth = borderThickness
	progressBarBackground.CornerRadius = 10.0
	progressBarBackground.SetMinSize(fyne.NewSize(progressBarWidth, progressBarHeight))
	progressBarBackground.Refresh() // Явное обновление

	innerBarHeight := progressBarHeight - 2*borderThickness
	innerBarWidth := progressBarWidth - 2*borderThickness
	progressBarFilled := canvas.NewRectangle(color.Black)
	progressBarFilled.SetMinSize(fyne.NewSize(innerBarWidth, 0))
	progressBarFilled.CornerRadius = 10.0
	progressBarFilled.Refresh() // Явное обновление

	filledBarContent := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(innerBarWidth, innerBarHeight)),
		progressBarFilled,
	)
	progressBarArea := container.NewStack(progressBarBackground, filledBarContent)
	progressBarArea.Refresh() // Явное обновление

	percentageText := canvas.NewText("0%", Black)
	percentageText.Alignment = fyne.TextAlignCenter
	percentageText.TextSize = 40
	percentageText.TextStyle.Bold = true
	percentageContainer := container.NewCenter(percentageText)

	progressBarContainer := container.NewVBox(progressBarArea, percentageContainer)
	progressBarAligned := container.NewCenter(progressBarContainer)

	mainContent := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(0, 25)),
		drainText,
		jarText,
		NewCustomSpacer(fyne.NewSize(0, 30)),
		fuelTypeText,
		NewCustomSpacer(fyne.NewSize(0, 10)),
		expectedContainer,
		expectedValueContainer,
		NewCustomSpacer(fyne.NewSize(0, 10)),
		container.NewCenter(container.NewVBox(drainedValueText, drainedText)),
	)

	buttonText1 := canvas.NewText("Для завершения слива", Black)
	buttonText1.Alignment = fyne.TextAlignCenter
	buttonText1.TextSize = 32
	buttonText1.TextStyle = fyne.TextStyle{Bold: true}

	buttonText2 := canvas.NewText("закройте люк. Слив должен", Black)
	buttonText2.Alignment = fyne.TextAlignCenter
	buttonText2.TextSize = 32
	buttonText2.TextStyle = fyne.TextStyle{Bold: true}

	buttonText3 := canvas.NewText(fmt.Sprintf("быть завершен через %v минут", timer), Black)
	buttonText3.Alignment = fyne.TextAlignCenter
	buttonText3.TextSize = 32
	buttonText3.TextStyle = fyne.TextStyle{Bold: true}

	buttonText := container.NewVBox(buttonText1, buttonText2, buttonText3)
	borderRect := canvas.NewRectangle(color.Transparent)
	borderRect.StrokeColor = color.Black
	borderRect.CornerRadius = 10.0
	borderRect.StrokeWidth = 2
	paddedButtonText := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 2)), NewCustomSpacer(fyne.NewSize(0, 2)), NewCustomSpacer(fyne.NewSize(20, 0)), NewCustomSpacer(fyne.NewSize(20, 0)), buttonText)
	buttonArea := container.NewStack(borderRect, paddedButtonText)
	buttonAreaContainer := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(10, 0)), NewCustomSpacer(fyne.NewSize(10, 0)), buttonArea)

	// Компоновка: прогресс-бар под кнопками
	columnContent := container.NewVBox(
		mainContent,
		buttonAreaContainer,
		progressBarAligned,
		NewCustomSpacer(fyne.NewSize(0, 20)),
	)
	columnContent.Refresh() // Явное обновление

	// Анимация прогресс-бара
	duration := 30 * time.Second
	animation := fyne.NewAnimation(duration, func(t float32) {
		percentage := t * 100
		filledHeight := percentage * innerBarHeight / 100.0
		if filledHeight > innerBarHeight {
			filledHeight = innerBarHeight
		}
		fyne.Do(func() {
			progressBarFilled.SetMinSize(fyne.NewSize(innerBarWidth, filledHeight))
			filledBarContent.Objects[0].(*fyne.Container).Objects[0].(*canvas.Rectangle).SetMinSize(fyne.NewSize(innerBarWidth, innerBarHeight-filledHeight))
			percentageText.Text = fmt.Sprintf("%.0f%%", percentage)
			filledBarContent.Refresh()
			progressBarArea.Refresh()
			percentageText.Refresh()
			columnContent.Refresh() // Обновление всей формы
			log.Printf("Анимация прогресс-бара для %s: процент %.0f%%", jarNumber, percentage)
		})
	})
	animation.AutoReverse = false
	animation.Start()

	fyne.Do(func() {
		section := gui.getSectionContent(jarNumber)
		section.RemoveAll()
		section.Add(columnContent)
		section.Refresh()
	})

	go func() {
		time.Sleep(30 * time.Second)
		fyne.Do(func() {
			localFuelType := fuelType
			if localFuelType == "" {
				localFuelType = "Неизвестно"
				log.Printf("CreateFuelGetInProgressScreen: fuelType пустой, установлено значение по умолчанию: %s", localFuelType)
			}
			gui.CreateFuelGetCompleteScreen(jarNumber, localFuelType, "DOC456", tankLiters, maxTankLiters, liters, maxLiters, drainTimestamp, drainTimestamp+30, timer)
		})
	}()

	return columnContent
}

func (gui *Gui) CreateFuelGetCompleteScreen(jarNumber string, fuelType string, doc string, tankLiters int, maxTankLiters int, liters float32, maxLiters float32, drainTimestamp int64, completeTimestamp int64, timer int) *fyne.Container {
	// Текст "СЛИВ ЗАВЕРШЁН"
	completeText := canvas.NewText("СЛИВ ЗАВЕРШЁН", Black)
	completeText.Alignment = fyne.TextAlignCenter
	completeText.TextSize = 40
	completeText.TextStyle = fyne.TextStyle{Bold: true}

	// Текст "тип_топлива + Емкость №X"
	fuelTypePistolText := widget.NewLabel(fmt.Sprintf("%s Ёмкость №%v", fuelType, jarNumber))
	fuelTypePistolText.Alignment = fyne.TextAlignCenter
	fuelTypePistolText.Wrapping = fyne.TextWrapWord

	// Данные о сливе
	documentText := canvas.NewText("№ документа", Gray)
	documentValueText := canvas.NewText(doc, Black)
	documentText.TextSize = 32
	documentValueText.TextSize = 32
	documentValueText.Alignment = fyne.TextAlignTrailing
	documentValueText.TextStyle = fyne.TextStyle{Bold: true}

	beforeFuelGetText := canvas.NewText("До слива", Gray)
	beforeFuelGetValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringShort(float32(tankLiters))), Black)
	beforeFuelGetText.TextSize = 32
	beforeFuelGetValueText.TextSize = 32
	beforeFuelGetValueText.Alignment = fyne.TextAlignTrailing
	beforeFuelGetValueText.TextStyle = fyne.TextStyle{Bold: true}

	afterFuelGetText := canvas.NewText("После слива", Gray)
	afterFuelGetValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringShort(float32(tankLiters)+liters)), Black)
	afterFuelGetText.TextSize = 32
	afterFuelGetValueText.TextSize = 32
	afterFuelGetValueText.Alignment = fyne.TextAlignTrailing
	afterFuelGetValueText.TextStyle = fyne.TextStyle{Bold: true}

	fuelGetPlanText := canvas.NewText("Слив план", Gray)
	fuelGetPlanValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringShort(maxLiters)), Black)
	fuelGetPlanText.TextSize = 32
	fuelGetPlanValueText.TextSize = 32
	fuelGetPlanValueText.Alignment = fyne.TextAlignTrailing
	fuelGetPlanValueText.TextStyle = fyne.TextStyle{Bold: true}

	fuelGetFactText := canvas.NewText("Слив факт", Gray)
	fuelGetFactValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringShort(liters)), Black)
	fuelGetFactText.TextSize = 32
	fuelGetFactValueText.TextSize = 32
	fuelGetFactValueText.Alignment = fyne.TextAlignTrailing
	fuelGetFactValueText.TextStyle = fyne.TextStyle{Bold: true}

	startTimeText := canvas.NewText("Дата начала", Gray)
	startTimeValueText := canvas.NewText(ConvertUnixToString(drainTimestamp), Black)
	startTimeText.TextSize = 32
	startTimeValueText.TextSize = 32
	startTimeValueText.Alignment = fyne.TextAlignTrailing
	startTimeValueText.TextStyle = fyne.TextStyle{Bold: true}

	endTimeText := canvas.NewText("Дата окончания", Gray)
	endTimeValueText := canvas.NewText(ConvertUnixToString(completeTimestamp), Black)
	endTimeText.TextSize = 32
	endTimeValueText.TextSize = 32
	endTimeValueText.Alignment = fyne.TextAlignTrailing
	endTimeValueText.TextStyle = fyne.TextStyle{Bold: true}

	data1 := container.NewGridWithColumns(2, documentText, documentValueText)
	data1Container := container.NewVBox(data1, NewCustomSpacer(fyne.NewSize(0, 15)))
	data2 := container.NewGridWithColumns(2, beforeFuelGetText, beforeFuelGetValueText)
	data2Container := container.NewVBox(data2, NewCustomSpacer(fyne.NewSize(0, 15)))
	data3 := container.NewGridWithColumns(2, afterFuelGetText, afterFuelGetValueText)
	data3Container := container.NewVBox(data3, NewCustomSpacer(fyne.NewSize(0, 15)))
	data4 := container.NewGridWithColumns(2, fuelGetPlanText, fuelGetPlanValueText)
	data4Container := container.NewVBox(data4, NewCustomSpacer(fyne.NewSize(0, 15)))
	data5 := container.NewGridWithColumns(2, fuelGetFactText, fuelGetFactValueText)
	data5Container := container.NewVBox(data5, NewCustomSpacer(fyne.NewSize(0, 15)))
	data6 := container.NewGridWithColumns(2, startTimeText, startTimeValueText)
	data6Container := container.NewVBox(data6, NewCustomSpacer(fyne.NewSize(0, 15)))
	data7 := container.NewGridWithColumns(2, endTimeText, endTimeValueText)
	data7Container := container.NewVBox(data7, NewCustomSpacer(fyne.NewSize(0, 15)))

	data := container.NewGridWithRows(7, data1Container, data2Container, data3Container, data4Container, data5Container, data6Container, data7Container)
	dataPadding := container.NewPadded(data)
	dataBorder := container.NewBorder(nil, nil, NewCustomSpacer(fyne.NewSize(5, 0)), NewCustomSpacer(fyne.NewSize(5, 0)), dataPadding)

	// Рамка снизу
	buttonText1 := canvas.NewText(fmt.Sprintf("Заправка с ПИСТОЛЕТА №%v", jarNumber), Black)
	buttonText1.Alignment = fyne.TextAlignCenter
	buttonText1.TextSize = 32
	buttonText1.TextStyle = fyne.TextStyle{Bold: true}

	buttonText2 := canvas.NewText("возможна через 10 секунд", Black)
	buttonText2.Alignment = fyne.TextAlignCenter
	buttonText2.TextSize = 32
	buttonText2.TextStyle = fyne.TextStyle{Bold: true}

	buttonText := container.NewVBox(buttonText1, buttonText2)
	borderRect := canvas.NewRectangle(color.Transparent)
	borderRect.StrokeColor = color.Black
	borderRect.CornerRadius = 10.0
	borderRect.StrokeWidth = 2
	paddedButtonText := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 2)), NewCustomSpacer(fyne.NewSize(0, 2)), NewCustomSpacer(fyne.NewSize(20, 0)), NewCustomSpacer(fyne.NewSize(20, 0)), buttonText)
	buttonArea := container.NewStack(borderRect, paddedButtonText)
	buttonAreaContainer := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(10, 0)), NewCustomSpacer(fyne.NewSize(10, 0)), buttonArea)

	// Собираем контент
	topCenterContent := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(0, 15)),
		completeText,
		NewCustomSpacer(fyne.NewSize(0, 15)),
		fuelTypePistolText,
		NewCustomSpacer(fyne.NewSize(0, 15)),
		dataBorder,
	)

	columnContent := container.New(layout.NewBorderLayout(topCenterContent, buttonAreaContainer, nil, nil),
		topCenterContent,
		NewCustomSpacer(fyne.NewSize(0, 25)),
		buttonAreaContainer,
	)

	// Обновляем соответствующую секцию
	fyne.Do(func() {
		section := gui.getSectionContent(jarNumber)
		section.RemoveAll()
		section.Add(columnContent)
		section.Refresh()
	})

	// Таймер для возврата к начальному экрану через 10 секунд
	go func() {
		time.Sleep(10 * time.Second)
		fyne.Do(func() {
			gui.CreateFuelGiveStartScreen(jarNumber, 0, "", 0)
		})
	}()

	return columnContent
}

// getSectionContent возвращает контейнер секции по номеру колонки
func (gui *Gui) getSectionContent(jarNumber string) *fyne.Container {
	if jarNumber == "1" && gui.LeftSection != nil && gui.LeftSection.Content != nil {
		return gui.LeftSection.Content
	} else if jarNumber == "2" && gui.RightSection != nil && gui.RightSection.Content != nil {
		return gui.RightSection.Content
	}
	log.Printf("getSectionContent: Не удалось найти секцию для jarNumber %s", jarNumber)
	return nil
}

// ShowSectionDialog отображает диалоговое окно
func (g *Gui) ShowSectionDialog(sectionStack *fyne.Container, title string, message string, timerSeconds int, onClose func()) {
	if sectionStack == nil {
		log.Println("Ошибка: Контейнер секции равен nil.")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	dialogClosed := make(chan struct{})

	onDialogClose := func() {
		if onClose != nil {
			onClose()
		}
		cancel()
		select {
		case <-dialogClosed:
		default:
			close(dialogClosed)
		}
	}

	overlayBackground := canvas.NewRectangle(color.RGBA{128, 128, 128, 200})
	overlayBackground.SetMinSize(sectionStack.Size())
	overlayBackground.Show()

	titleText := canvas.NewText(title, color.RGBA{191, 7, 7, 255})
	titleText.Alignment = fyne.TextAlignLeading
	titleText.TextSize = 40
	titleText.TextStyle = fyne.TextStyle{Bold: true}

	separator := NewZeroHSeparator()
	messageText := widget.NewLabel(message)
	messageText.Wrapping = fyne.TextWrapWord
	messageText.Alignment = fyne.TextAlignLeading

	timerButton := createTimerButtonContent(ctx, timerSeconds, onDialogClose)

	dialogContentVBox := container.NewVBox(
		titleText,
		separator,
		messageText,
		NewCustomSpacer(fyne.NewSize(0, 5)),
		timerButton,
	)

	dialogBackground := canvas.NewRectangle(color.RGBA{255, 255, 255, 255})
	dialogBackground.SetMinSize(fyne.NewSize(550, 150))
	dialogContentPadded := container.NewPadded(dialogContentVBox)
	dialogContent := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(10, 0)), NewCustomSpacer(fyne.NewSize(10, 0)), dialogContentPadded)
	dialogArea := container.NewStack(dialogBackground, dialogContent)
	centeredDialog := container.NewCenter(dialogArea)
	overlayContainer := container.NewStack(overlayBackground, centeredDialog)
	overlayContainer.Show()

	fyne.Do(func() {
		sectionStack.Add(overlayContainer)
		sectionStack.Refresh()
	})
}

// createTimerButtonContent создаёт кнопку с таймером
func createTimerButtonContent(ctx context.Context, initialSeconds int, onTimerComplete func()) *fyne.Container {
	timerText := canvas.NewText(fmt.Sprintf("Закроется через (%d с)", initialSeconds), color.Black)
	timerText.Alignment = fyne.TextAlignCenter
	timerText.TextSize = 32

	borderRect := canvas.NewRectangle(color.Transparent)
	borderRect.StrokeColor = color.Black
	borderRect.CornerRadius = 10.0
	borderRect.StrokeWidth = 2
	borderRect.SetMinSize(fyne.NewSize(530, 60))

	paddedText := container.NewPadded(container.NewCenter(timerText))
	paddedContainer := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 5)), NewCustomSpacer(fyne.NewSize(0, 5)), NewCustomSpacer(fyne.NewSize(20, 0)), NewCustomSpacer(fyne.NewSize(20, 0)), paddedText)
	buttonArea := container.NewStack(borderRect, paddedContainer)

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		secondsRemaining := initialSeconds
		for {
			select {
			case <-ctx.Done():
				log.Println("Горутина таймера отменена через контекст.")
				return
			case <-ticker.C:
				secondsRemaining--
				if secondsRemaining < 0 {
					secondsRemaining = 0
				}
				fyne.Do(func() {
					timerText.Text = fmt.Sprintf("Закроется через (%d с)", secondsRemaining)
					timerText.Refresh()
				})
				if secondsRemaining <= 0 {
					if onTimerComplete != nil {
						fyne.Do(onTimerComplete)
					}
					return
				}
			}
		}
	}()

	return container.NewCenter(buttonArea)
}
