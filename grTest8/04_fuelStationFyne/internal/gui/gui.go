package gui

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"image"
	"image/color"
	"log"
	"time"
)

var (
	Black = color.RGBA{A: 255}
	Gray  = color.RGBA{R: 128, G: 128, B: 128, A: 255}
	Green = color.RGBA{G: 128, A: 255}
	White = color.RGBA{R: 255, G: 255, B: 255, A: 255}
)

var logo = "iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg=="

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
}

type Section struct {
	Content            *fyne.Container
	ActiveDialogCancel context.CancelFunc
	ActiveDialog       *fyne.Container
	ActiveProcess      context.CancelFunc
	Timer              context.CancelFunc
}

type Gui struct {
	app           fyne.App
	Window        fyne.Window
	TopSection    *TopSection
	LeftSection   *Section
	RightSection  *Section
	BottomSection *fyne.Container
	MainContent   *fyne.Container
}

type SectionInterface interface {
	CreateDefaultScreen(jarNumber string)
	CreateDownloadScreen(jarNumber string)
	CreateFuelGiveStartScreen(jarNumber string, liters float32, fuelType string, timer int)
	CreateFuelGiveInProgressScreen(jarNumber string, liters, expectedLiters float32)
	CreateFuelGetInProgressScreen(jarNumber string, expectedAmount, drainedAmount, fuelVolume, jarVolume float32, timer int)
	CreateFuelGiveCompleteScreen(jarNumber string, liters float32, fuelType string)
	CreateFuelGetStartScreen(jarNumber string, expectedAmount, drainedAmount, fuelVolume, jarVolume float32, timer int)
	ShowSectionDialog(sectionStack *fyne.Container, title, message string, timerSeconds int, onClose func())
	GetSection(jarNumber string) *Section
}

func NewGui(a fyne.App) *Gui {
	topSection := newTopSection()
	leftSection := newSection("1")
	rightSection := newSection("2")

	bottomSection := container.NewHBox(
		container.NewVBox(leftSection.Content),
		container.NewVBox(rightSection.Content),
	)
	mainContent := container.NewVBox(topSection.Content, bottomSection)
	return &Gui{
		app:           a,
		Window:        a.NewWindow("Fuel Station"),
		TopSection:    topSection,
		LeftSection:   leftSection,
		RightSection:  rightSection,
		BottomSection: bottomSection,
		MainContent:   mainContent,
	}
}

func (g *Gui) RunGui() {
	g.Window.SetContent(g.MainContent)
	g.Window.Resize(fyne.NewSize(800, 800))
	go g.updateTime()
	g.Window.ShowAndRun()
}

func (g *Gui) updateTime() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		fyne.Do(func() {
			now := time.Now()
			g.TopSection.DateLabel.Text = formatRussianDate(now)
			_, offsetSeconds := now.Zone()
			offsetHours := offsetSeconds / 3600
			g.TopSection.TimeLabel.Text = now.Format("15:04") + fmt.Sprintf(" (GMT%+d)", offsetHours)
			g.TopSection.DateLabel.Refresh()
			g.TopSection.TimeLabel.Refresh()
		})
	}
}

func newTopSection() *TopSection {
	topSection, timeLabel, dateLabel, logoImage, phoneLabel, kazsLabel := createHeader(logo, "8-111-555-11-11", "99699")
	return &TopSection{
		Content:            topSection,
		TimeLabel:          timeLabel,
		DateLabel:          dateLabel,
		LogoLabel:          logoImage,
		SupportNumberLabel: phoneLabel,
		KazsNumberLabel:    kazsLabel,
		Logo:               logo,
		SupportNumber:      "8-111-555-11-11",
		KazsNumber:         "99699",
		Timezone:           "UTC+3",
	}
}

func newSection(jarNumber string) *Section {
	defaultScreen := createDefaultScreen(jarNumber)
	return &Section{
		Content: defaultScreen,
	}
}

func createHeader(logo, supportNumber, kazsNumber string) (*fyne.Container, *canvas.Text, *canvas.Text, *canvas.Image, *canvas.Text, *canvas.Text) {
	decodedImageBytes, err := base64.StdEncoding.DecodeString(logo)
	if err != nil {
		log.Printf("decode image error: %v", err)
	}
	imageReader := bytes.NewReader(decodedImageBytes)
	_, _, _ = image.Decode(bytes.NewReader(decodedImageBytes))
	itecoImage := canvas.NewImageFromReader(imageReader, "logo.png") //logo iteco_logo.png
	itecoImage.FillMode = canvas.ImageFillContain
	itecoImage.SetMinSize(fyne.NewSize(368, 150))

	phoneLabel := canvas.NewText(supportNumber, White) //or Black
	supportLabel := canvas.NewText("Техническая поддержка", White)
	phoneLabel.Alignment = fyne.TextAlignCenter
	supportLabel.Alignment = fyne.TextAlignCenter
	phoneLabel.TextStyle = fyne.TextStyle{Bold: true}
	phoneLabel.TextSize = 30
	supportLabel.TextSize = 26
	centerHeaderContent := container.NewVBox(newCustomSpacer(fyne.NewSize(0, 15)), phoneLabel, supportLabel)

	now := time.Now()
	azsLabel := canvas.NewText(fmt.Sprintf("АЗС №%v", kazsNumber), Black)
	dateLabel := canvas.NewText(formatRussianDate(now), Black)
	_, offsetSeconds := now.Zone()
	offsetHours := offsetSeconds / 3600
	timeLabel := canvas.NewText(now.Format("15:04")+fmt.Sprintf(" (GMT%+d)", offsetHours), Black)
	azsLabel.Alignment = fyne.TextAlignTrailing
	dateLabel.Alignment = fyne.TextAlignTrailing
	timeLabel.Alignment = fyne.TextAlignTrailing
	azsLabel.TextStyle = fyne.TextStyle{Bold: true}
	azsLabel.TextSize = 30
	dateLabel.TextSize = 24
	timeLabel.TextSize = 34
	rightHeaderContent := container.NewVBox(azsLabel, dateLabel, timeLabel)
	rightHeaderContentCentered := container.NewCenter(rightHeaderContent)
	rightHeader := container.NewHBox(rightHeaderContentCentered, newCustomSpacer(fyne.NewSize(15, 0)))

	topSectionContent := container.New(
		layout.NewBorderLayout(nil, nil, itecoImage, rightHeader),
		itecoImage, rightHeader, centerHeaderContent,
	)
	topSectionContainer := container.NewVBox(topSectionContent, newFixedHSeparator())
	return topSectionContainer, timeLabel, dateLabel, itecoImage, phoneLabel, azsLabel
}

func createDefaultScreen(jarNumber string) *fyne.Container {
	qrText1 := canvas.NewText("Отсканируйте QR-код", White)                       //or Black
	qrText2 := canvas.NewText(fmt.Sprintf("для пистолета №%v", jarNumber), White) //or Black
	qrText1.Alignment = fyne.TextAlignCenter
	qrText1.TextSize = 42
	qrText2.Alignment = fyne.TextAlignCenter
	qrText2.TextSize = 42
	sectionVBox := container.NewVBox(
		newCustomSpacer(fyne.NewSize(0, 20)), // Верхний отступ
		qrText1,
		newCustomSpacer(fyne.NewSize(0, 10)), // Между надписями
		qrText2,
	)
	sectionWithLeftPadding := container.NewHBox(
		newCustomSpacer(fyne.NewSize(0, 0)), // Отступ слева 20px
		sectionVBox,
	)
	sectionCenter := container.NewCenter(sectionWithLeftPadding)                                   // Центрирование содержимого
	sectionCenter = container.New(layout.NewGridWrapLayout(fyne.NewSize(800, 800)), sectionCenter) // Фиксированный размер
	return sectionCenter
}

func (g *Gui) CreateDefaultScreen(jarNumber string) {
	section := g.getSection(jarNumber)
	content := createDefaultScreen(jarNumber)
	fyne.Do(func() {
		section.Content.RemoveAll()
		section.Content.Add(content)
		section.Content.Refresh()
		log.Printf("CreateDefaultScreen: jarNumber=%s, size=%v", jarNumber, section.Content.Size())
	})
}

func (g *Gui) CreateDownloadScreen(jarNumber string) {
	section := g.getSection(jarNumber)
	qrText1 := canvas.NewText("Обработка", Green)
	qrText1.Alignment = fyne.TextAlignCenter
	qrText1.TextSize = 42
	sectionVBox := container.NewVBox(qrText1)
	sectionCenter := container.NewCenter(sectionVBox)
	sectionCenter = container.New(layout.NewGridWrapLayout(fyne.NewSize(400, 600)), sectionCenter)
	fyne.Do(func() {
		section.Content.RemoveAll()
		section.Content.Add(sectionCenter)
		section.Content.Refresh()
		log.Printf("CreateDownloadScreen: jarNumber=%s, size=%v", jarNumber, section.Content.Size())
	})
}

func (g *Gui) CreateTechnicalErrorScreen(jarNumber string) *fyne.Container {
	section := g.getSection(jarNumber)
	qrText1 := canvas.NewText("Технические неполадки", Green)
	qrText1.Alignment = fyne.TextAlignCenter
	qrText1.TextSize = 42
	sectionVBox := container.NewVBox(qrText1)
	sectionCenter := container.NewCenter(sectionVBox)
	sectionCenter = container.New(layout.NewGridWrapLayout(fyne.NewSize(400, 600)), sectionCenter)
	fyne.Do(func() {
		section.Content.RemoveAll()
		section.Content.Add(sectionCenter)
		section.Content.Refresh()
		log.Printf("CreateTechnicalErrorScreen: jarNumber=%s, size=%v", jarNumber, section.Content.Size())
	})
	return sectionCenter
}

func (g *Gui) CreateFuelGiveCompleteScreen(jarNumber string, liters float32, fuelType string) {
	section := g.getSection(jarNumber)
	completeText := canvas.NewText("Заправка завершена", Black)
	completeText.Alignment = fyne.TextAlignCenter
	completeText.TextSize = 32
	gunText := canvas.NewText(fmt.Sprintf("ПИСТОЛЕТ №%v", jarNumber), Black)
	gunText.Alignment = fyne.TextAlignCenter
	gunText.TextSize = 40
	gunText.TextStyle = fyne.TextStyle{Bold: true}
	fuelTypeText := widget.NewLabel(fuelType)
	fuelTypeText.Alignment = fyne.TextAlignCenter
	fuelTypeText.Wrapping = fyne.TextWrapWord
	volumeText := canvas.NewText("Залито", Gray)
	volumeText.Alignment = fyne.TextAlignCenter
	volumeText.TextSize = 32
	litersText := canvas.NewText(fmt.Sprintf("%.2f литров", liters), Black)
	litersText.Alignment = fyne.TextAlignCenter
	litersText.TextSize = 40
	litersText.TextStyle = fyne.TextStyle{Bold: true}
	volumeAndLitersArea := container.NewVBox(volumeText, newCustomSpacer(fyne.NewSize(0, 5)), litersText)
	volumeAndLitersContainer := container.NewCenter(volumeAndLitersArea)
	instructionText := canvas.NewText("Спасибо за заправку!", Black)
	instructionText.Alignment = fyne.TextAlignCenter
	instructionText.TextSize = 32
	instructionText.TextStyle = fyne.TextStyle{Bold: true}
	topCenterContent := container.NewVBox(
		newCustomSpacer(fyne.NewSize(0, 20)),
		completeText,
		gunText,
		newCustomSpacer(fyne.NewSize(0, 60)),
		fuelTypeText,
		newCustomSpacer(fyne.NewSize(0, 15)),
		volumeAndLitersContainer,
		newCustomSpacer(fyne.NewSize(0, 20)),
		instructionText,
	)
	columnContent := container.NewCenter(topCenterContent)
	columnContent = container.New(layout.NewGridWrapLayout(fyne.NewSize(400, 600)), columnContent)
	fyne.Do(func() {
		section.Content.RemoveAll()
		section.Content.Add(columnContent)
		section.Content.Refresh()
		log.Printf("CreateFuelGiveCompleteScreen: jarNumber=%s, liters=%.2f, fuelType=%s, size=%v", jarNumber, liters, fuelType, section.Content.Size())
	})
}

func (g *Gui) CreateFuelGiveStartScreen(jarNumber string, liters float32, fuelType string, timer int) {
	section := g.getSection(jarNumber)
	insertText := canvas.NewText("Вставьте в бензобак", Black)
	insertText.Alignment = fyne.TextAlignCenter
	insertText.TextSize = 32
	gunText := canvas.NewText(fmt.Sprintf("ПИСТОЛЕТ №%v", jarNumber), Black)
	gunText.Alignment = fyne.TextAlignCenter
	gunText.TextSize = 40
	gunText.TextStyle = fyne.TextStyle{Bold: true}
	fuelTypeText := widget.NewLabel(fuelType)
	fuelTypeText.Alignment = fyne.TextAlignCenter
	fuelTypeText.Wrapping = fyne.TextWrapWord
	maxVolumeText := canvas.NewText("Максимальный объем", Gray)
	maxVolumeText.Alignment = fyne.TextAlignCenter
	maxVolumeText.TextSize = 32
	litersText := canvas.NewText(fmt.Sprintf("%v литров", liters), Black)
	litersText.Alignment = fyne.TextAlignCenter
	litersText.TextSize = 40
	litersText.TextStyle = fyne.TextStyle{Bold: true}
	maxVolumeAndLitersArea := container.NewVBox(maxVolumeText, newCustomSpacer(fyne.NewSize(0, 5)), litersText)
	maxVolumeAndLitersContainer := container.NewCenter(maxVolumeAndLitersArea)
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
	borderRect.StrokeColor = Black
	borderRect.StrokeWidth = 2
	paddedButtonText := container.NewBorder(newCustomSpacer(fyne.NewSize(0, 2)), newCustomSpacer(fyne.NewSize(0, 2)), newCustomSpacer(fyne.NewSize(20, 0)), newCustomSpacer(fyne.NewSize(20, 0)), buttonText)
	buttonArea := container.NewStack(borderRect, paddedButtonText)
	buttonAreaContainer := container.NewCenter(buttonArea)
	paddedButtonArea := container.NewHBox(
		newCustomSpacer(fyne.NewSize(20, 0)), // Отступ 20px слева
		buttonAreaContainer,
	)
	topCenterContent := container.NewVBox(
		newCustomSpacer(fyne.NewSize(0, 20)),
		insertText,
		gunText,
		newCustomSpacer(fyne.NewSize(0, 60)),
		fuelTypeText,
		newCustomSpacer(fyne.NewSize(0, 15)),
		maxVolumeAndLitersContainer,
	)
	columnContent := container.New(layout.NewBorderLayout(topCenterContent, paddedButtonArea, nil, nil),
		topCenterContent,
		newCustomSpacer(fyne.NewSize(0, 225)),
		paddedButtonArea,
	)
	columnContent = container.New(layout.NewGridWrapLayout(fyne.NewSize(400, 600)), columnContent)
	fyne.Do(func() {
		section.Content.RemoveAll()
		section.Content.Add(columnContent)
		section.Content.Refresh()
		log.Printf("CreateFuelGiveStartScreen: jarNumber=%s, size=%v", jarNumber, section.Content.Size())
	})
}

func (g *Gui) CreateFuelGiveInProgressScreen(jarNumber string, liters, expectedLiters float32) {
	section := g.getSection(jarNumber)
	ifProgressText := canvas.NewText("Заправка в процессе", Black)
	ifProgressText.Alignment = fyne.TextAlignCenter
	ifProgressText.TextSize = 32
	gunText := canvas.NewText(fmt.Sprintf("ПИСТОЛЕТ №%v", jarNumber), Black)
	gunText.Alignment = fyne.TextAlignCenter
	gunText.TextStyle = fyne.TextStyle{Bold: true}
	gunText.TextSize = 40
	fuelTypeText := widget.NewLabel("Petrol") // Пример значения, замените на динамическое
	fuelTypeText.Alignment = fyne.TextAlignCenter
	fuelTypeText.Wrapping = fyne.TextWrapWord
	maxVolumeText := canvas.NewText("Максимальный объем", Gray)
	maxVolumeText.Alignment = fyne.TextAlignCenter
	maxVolumeText.TextSize = 32
	volumeValueText := canvas.NewText(fmt.Sprintf("%v литров", expectedLiters), Black)
	volumeValueText.Alignment = fyne.TextAlignCenter
	volumeValueText.TextStyle = fyne.TextStyle{Bold: true}
	volumeValueText.TextSize = 40
	maxVolumeAndVolumeArea := container.NewVBox(maxVolumeText, volumeValueText)
	maxVolumeAndVolumeContainer := container.NewCenter(maxVolumeAndVolumeArea)
	amountText := canvas.NewText(fmt.Sprintf("%.2f", liters), Black)
	amountText.Alignment = fyne.TextAlignCenter
	amountText.TextSize = 100
	litersText := canvas.NewText("литров залито", Gray)
	litersText.Alignment = fyne.TextAlignCenter
	litersText.TextSize = 40
	buttonText1 := canvas.NewText("Для завершения заправки", Black)
	buttonText1.Alignment = fyne.TextAlignCenter
	buttonText1.TextSize = 32
	buttonText1.TextStyle = fyne.TextStyle{Bold: true}
	buttonText2 := canvas.NewText(fmt.Sprintf("повесьте ПИСТОЛЕТ №%v", jarNumber), Black)
	buttonText2.Alignment = fyne.TextAlignCenter
	buttonText2.TextSize = 32
	buttonText2.TextStyle = fyne.TextStyle{Bold: true}
	buttonText := container.NewVBox(
		newCustomSpacer(fyne.NewSize(0, 0)), // Отступ внутри контейнера
		buttonText1,
		buttonText2,
	)

	borderRect := canvas.NewRectangle(color.Transparent)
	borderRect.StrokeColor = Black
	borderRect.CornerRadius = 10.0
	borderRect.StrokeWidth = 2
	paddedButtonText := container.NewBorder(newCustomSpacer(fyne.NewSize(0, 2)), newCustomSpacer(fyne.NewSize(0, 2)), newCustomSpacer(fyne.NewSize(20, 0)), newCustomSpacer(fyne.NewSize(20, 0)), buttonText)
	buttonArea := container.NewStack(borderRect, paddedButtonText)
	buttonAreaContainer := container.NewCenter(buttonArea)
	progressBarHeight := float32(329)
	progressBarWidth := float32(85)
	borderThickness := float32(2)
	progressBarBackground := canvas.NewRectangle(color.Transparent)
	progressBarBackground.StrokeColor = Black
	progressBarBackground.StrokeWidth = borderThickness
	progressBarBackground.CornerRadius = 10.0
	progressBarBackground.SetMinSize(fyne.NewSize(progressBarWidth, progressBarHeight))
	innerBarHeight := progressBarHeight - 2*borderThickness
	innerBarWidth := progressBarWidth - 2*borderThickness
	percentage := int((liters / expectedLiters) * 100.0)
	if percentage > 99 {
		percentage = 99
	}
	filledHeight := float32(percentage) * innerBarHeight / 100.0
	if filledHeight < 0 {
		filledHeight = 0
	}
	if filledHeight > innerBarHeight {
		filledHeight = innerBarHeight
	}
	progressBarFilled := canvas.NewRectangle(Black)
	progressBarFilled.SetMinSize(fyne.NewSize(innerBarWidth, filledHeight))
	progressBarFilled.CornerRadius = 10.0
	spaceAboveFilled := innerBarHeight - filledHeight
	if spaceAboveFilled < 0 {
		spaceAboveFilled = 0
	}
	filledBarContent := container.NewVBox(
		newCustomSpacer(fyne.NewSize(innerBarWidth, spaceAboveFilled)),
		progressBarFilled,
	)
	progressBarArea := container.NewStack(
		progressBarBackground,
		filledBarContent,
	)
	percentageText := canvas.NewText(fmt.Sprintf("%v%%", percentage), Black)
	percentageText.Alignment = fyne.TextAlignLeading
	percentageText.TextSize = 40
	percentageText.TextStyle.Bold = true
	percentageContainer := container.NewHBox(
		newCustomSpacer(fyne.NewSize(20, 0)), // Отступ слева 20px
		percentageText,
		newCustomSpacer(fyne.NewSize(400, 0)), // Отступ справа
	)
	amountBarPercentRow := container.NewHBox(
		newCustomSpacer(fyne.NewSize(400, 0)),
		container.NewVBox(amountText, litersText),
		progressBarArea,
		newCustomSpacer(fyne.NewSize(25, 0)),
	)
	topContent := container.NewVBox(
		newCustomSpacer(fyne.NewSize(0, 20)),
		ifProgressText,
		gunText,
		newCustomSpacer(fyne.NewSize(0, 40)),
		fuelTypeText,
		newCustomSpacer(fyne.NewSize(0, 10)),
		maxVolumeAndVolumeContainer,
		newCustomSpacer(fyne.NewSize(0, 20)),
		percentageContainer,
		newCustomSpacer(fyne.NewSize(0, -60)),
		amountBarPercentRow,
	)
	columnContent := container.New(layout.NewBorderLayout(topContent, buttonAreaContainer, nil, nil),
		topContent,
		buttonAreaContainer,
	)
	columnContent = container.New(layout.NewGridWrapLayout(fyne.NewSize(800, 800)), columnContent)
	fyne.Do(func() {
		section.Content.RemoveAll()
		section.Content.Add(columnContent)
		section.Content.Refresh()
		log.Printf("CreateFuelGiveInProgressScreen: jarNumber=%s, liters=%.2f, expectedLiters=%.2f, size=%v", jarNumber, liters, expectedLiters, section.Content.Size())
	})
}

func (g *Gui) CreateFuelGetInProgressScreen(jarNumber string, expectedAmount, drainedAmount, fuelVolume, jarVolume float32, timer int) {
	section := g.getSection(jarNumber)
	drainText := canvas.NewText("Слив бензовоза", Black)
	drainText.Alignment = fyne.TextAlignCenter
	drainText.TextSize = 32
	jarText := canvas.NewText(fmt.Sprintf("ЁМКОСТЬ №%v", jarNumber), Black)
	jarText.Alignment = fyne.TextAlignCenter
	jarText.TextStyle = fyne.TextStyle{Bold: true}
	jarText.TextSize = 40
	fuelTypeText := widget.NewLabel("Diesel") // Пример значения
	fuelTypeText.Alignment = fyne.TextAlignCenter
	fuelTypeText.Wrapping = fyne.TextWrapWord
	expectedText := canvas.NewText("Ожидаемый слив", Gray)
	expectedText.Alignment = fyne.TextAlignCenter
	expectedText.TextSize = 32
	expectedContainer := container.NewCenter(expectedText)
	expectedValueText := canvas.NewText(fmt.Sprintf("%v литров", expectedAmount), Black)
	expectedValueText.Alignment = fyne.TextAlignCenter
	expectedValueText.TextStyle = fyne.TextStyle{Bold: true}
	expectedValueText.TextSize = 40
	expectedValueContainer := container.NewCenter(expectedValueText)
	drainedValueText := canvas.NewText(fmt.Sprintf("%v", drainedAmount), Black)
	drainedValueText.Alignment = fyne.TextAlignCenter
	drainedValueText.TextStyle = fyne.TextStyle{Bold: true}
	drainedValueText.TextSize = 100
	drainedText := canvas.NewText("литров слито", Gray)
	drainedText.Alignment = fyne.TextAlignCenter
	drainedText.TextSize = 40
	amountAndLiters := container.NewVBox(newCustomSpacer(fyne.NewSize(0, -150)), drainedValueText, drainedText)
	amountAndLitersAligned := container.NewCenter(amountAndLiters)
	progressBarHeight := float32(329)
	progressBarWidth := float32(85)
	borderThickness := float32(2)
	progressBarBackground := canvas.NewRectangle(color.Transparent)
	progressBarBackground.StrokeColor = Black
	progressBarBackground.StrokeWidth = borderThickness
	progressBarBackground.CornerRadius = 10.0
	progressBarBackground.SetMinSize(fyne.NewSize(progressBarWidth, progressBarHeight))
	innerBarHeight := progressBarHeight - 2*borderThickness
	innerBarWidth := progressBarWidth - 2*borderThickness
	percentage := int(fuelVolume / jarVolume * 100.0)
	if percentage >= 100 {
		percentage = 99
	}
	filledHeight := float32(percentage) * innerBarHeight / 100.0
	if filledHeight < 0 {
		filledHeight = 0
	}
	if filledHeight > innerBarHeight {
		filledHeight = innerBarHeight
	}
	progressBarFilled := canvas.NewRectangle(Black)
	progressBarFilled.SetMinSize(fyne.NewSize(innerBarWidth, filledHeight))
	progressBarFilled.CornerRadius = 10.0
	spaceAboveFilled := innerBarHeight - filledHeight
	if spaceAboveFilled < 0 {
		spaceAboveFilled = 0
	}
	filledBarContent := container.NewVBox(
		newCustomSpacer(fyne.NewSize(innerBarWidth, spaceAboveFilled)),
		progressBarFilled,
	)
	progressBarArea := container.NewStack(
		progressBarBackground,
		filledBarContent,
	)
	percentageText := canvas.NewText(fmt.Sprintf("%v%%", percentage), Black)
	percentageText.Alignment = fyne.TextAlignCenter
	percentageText.TextSize = 40
	percentageText.TextStyle.Bold = true
	percentageContainer := container.NewHBox(
		newCustomSpacer(fyne.NewSize(20, 0)), // Отступ слева
		percentageText,
		newCustomSpacer(fyne.NewSize(400, 0)), // Отступ справа
	)

	buttonText1 := canvas.NewText("Для завершения слива", Black)
	buttonText1.Alignment = fyne.TextAlignCenter
	buttonText1.TextSize = 24
	buttonText1.TextStyle = fyne.TextStyle{Bold: true}
	buttonText2 := canvas.NewText("закройте люк. Слив должен", Black)
	buttonText2.Alignment = fyne.TextAlignCenter
	buttonText2.TextSize = 24
	buttonText2.TextStyle = fyne.TextStyle{Bold: true}
	buttonText3 := canvas.NewText(fmt.Sprintf("быть завершен через %v минут", timer), Black)
	buttonText3.Alignment = fyne.TextAlignCenter
	buttonText3.TextSize = 24
	buttonText3.TextStyle = fyne.TextStyle{Bold: true}
	buttonText := container.NewVBox(
		newCustomSpacer(fyne.NewSize(0, 0)), // Отступ внутри контейнера
		buttonText1,
		buttonText2,
		buttonText3,
	)
	borderRect := canvas.NewRectangle(color.Transparent)
	borderRect.StrokeColor = Black
	borderRect.CornerRadius = 10.0
	borderRect.StrokeWidth = 2
	paddedButtonText := container.NewBorder(newCustomSpacer(fyne.NewSize(0, 2)), newCustomSpacer(fyne.NewSize(0, 2)), newCustomSpacer(fyne.NewSize(20, 0)), newCustomSpacer(fyne.NewSize(20, 0)), buttonText)
	buttonArea := container.NewStack(borderRect, paddedButtonText)
	buttonAreaContainer := container.NewHBox(
		newCustomSpacer(fyne.NewSize(200, 0)), // Отступ слева надписи для завершения слива....
		buttonArea,
	)
	amountBarPercentRow := container.NewHBox(
		newCustomSpacer(fyne.NewSize(400, 0)),
		amountAndLitersAligned,
		layout.NewSpacer(),
		progressBarArea,
		newCustomSpacer(fyne.NewSize(25, 0)),
	)

	topContent := container.NewVBox(
		newCustomSpacer(fyne.NewSize(0, 20)),
		drainText,
		jarText,
		newCustomSpacer(fyne.NewSize(0, 45)),
		fuelTypeText,
		newCustomSpacer(fyne.NewSize(0, 10)),
		expectedContainer,
		expectedValueContainer,
		newCustomSpacer(fyne.NewSize(0, 20)),
		percentageContainer,
		newCustomSpacer(fyne.NewSize(0, -60)),
		amountBarPercentRow,
	)
	columnContent := container.New(layout.NewBorderLayout(topContent, buttonAreaContainer, nil, nil),
		topContent,
		buttonAreaContainer,
	)
	columnContent = container.New(layout.NewGridWrapLayout(fyne.NewSize(800, 800)), columnContent)
	fyne.Do(func() {
		section.Content.RemoveAll()
		section.Content.Add(columnContent)
		section.Content.Refresh()
		log.Printf("CreateFuelGetInProgressScreen: jarNumber=%s, size=%v", jarNumber, section.Content.Size())
	})
}

func (g *Gui) ShowSectionDialog(sectionStack *fyne.Container, title, message string, timerSeconds int, onClose func()) {
	if sectionStack == nil {
		log.Println("ShowSectionDialog: sectionStack is nil")
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
	overlayBackground.SetMinSize(fyne.NewSize(400, 600))
	overlayBackground.Show()
	titleText := canvas.NewText(title, color.RGBA{191, 7, 7, 255})
	titleText.Alignment = fyne.TextAlignCenter
	titleText.TextSize = 40
	titleText.TextStyle = fyne.TextStyle{Bold: true}
	separator := newZeroHSeparator()
	messageText := widget.NewLabel(message)
	messageText.Wrapping = fyne.TextWrapWord
	messageText.Alignment = fyne.TextAlignCenter
	timerButton := createTimerButtonContent(ctx, timerSeconds, onDialogClose)
	dialogContentVBox := container.NewVBox(titleText, separator, messageText, newCustomSpacer(fyne.NewSize(0, 5)), timerButton)
	dialogBackground := canvas.NewRectangle(color.RGBA{255, 255, 255, 255})
	dialogBackground.SetMinSize(fyne.NewSize(350, 150))
	dialogContentPadded := container.NewPadded(dialogContentVBox)
	dialogContent := container.NewBorder(newCustomSpacer(fyne.NewSize(0, 10)), newCustomSpacer(fyne.NewSize(0, 10)), newCustomSpacer(fyne.NewSize(10, 0)), newCustomSpacer(fyne.NewSize(10, 0)), dialogContentPadded)
	dialogArea := container.NewStack(dialogBackground, dialogContent)
	centeredDialog := container.NewCenter(dialogArea)
	overlayContainer := container.NewStack(overlayBackground, centeredDialog)
	overlayContainer.Show()
	fyne.Do(func() {
		sectionStack.Add(overlayContainer)
		sectionStack.Refresh()
		log.Printf("ShowSectionDialog: title=%s, size=%v", title, sectionStack.Size())
	})
	section := g.getSectionByStack(sectionStack)
	if section != nil {
		section.ActiveDialog = overlayContainer
		section.ActiveDialogCancel = cancel
	}
}

func createTimerButtonContent(ctx context.Context, initialSeconds int, onTimerComplete func()) *fyne.Container {
	timerText := canvas.NewText(fmt.Sprintf("Закроется через (%d с)", initialSeconds), Black)
	timerText.Alignment = fyne.TextAlignCenter
	timerText.TextSize = 32
	borderRect := canvas.NewRectangle(color.Transparent)
	borderRect.StrokeColor = Black
	borderRect.CornerRadius = 10.0
	borderRect.StrokeWidth = 2
	borderRect.SetMinSize(fyne.NewSize(330, 60))
	paddedText := container.NewPadded(container.NewCenter(timerText))
	paddedContainer := container.NewBorder(newCustomSpacer(fyne.NewSize(0, 5)), newCustomSpacer(fyne.NewSize(0, 5)), newCustomSpacer(fyne.NewSize(20, 0)), newCustomSpacer(fyne.NewSize(20, 0)), paddedText)
	buttonArea := container.NewStack(borderRect, paddedContainer)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		secondsRemaining := initialSeconds
		for {
			select {
			case <-ctx.Done():
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

func (g *Gui) getSection(jarNumber string) *Section {
	if jarNumber == "1" {
		return g.LeftSection
	}
	return g.RightSection
}

func (g *Gui) GetSection(jarNumber string) *Section {
	return g.getSection(jarNumber)
}

func (g *Gui) getSectionByStack(sectionStack *fyne.Container) *Section {
	if g.LeftSection.Content == sectionStack {
		return g.LeftSection
	}
	if g.RightSection.Content == sectionStack {
		return g.RightSection
	}
	return nil
}

func (g *Gui) CreateFuelGetStartScreen(jarNumber string, expectedAmount, drainedAmount, fuelVolume, jarVolume float32, timer int) {
	section := g.getSection(jarNumber)
	insertText := canvas.NewText("Начните слив", Black)
	insertText.Alignment = fyne.TextAlignCenter
	insertText.TextSize = 32
	jarText := canvas.NewText(fmt.Sprintf("ЁМКОСТЬ №%v", jarNumber), Black)
	jarText.Alignment = fyne.TextAlignCenter
	jarText.TextSize = 40
	jarText.TextStyle = fyne.TextStyle{Bold: true}
	fuelTypeText := widget.NewLabel("Diesel") // Пример значения
	fuelTypeText.Alignment = fyne.TextAlignCenter
	fuelTypeText.Wrapping = fyne.TextWrapWord
	expectedText := canvas.NewText("Ожидаемый объем", Gray)
	expectedText.Alignment = fyne.TextAlignCenter
	expectedText.TextSize = 32
	expectedValueText := canvas.NewText(fmt.Sprintf("%v литров", expectedAmount), Black)
	expectedValueText.Alignment = fyne.TextAlignCenter
	expectedValueText.TextStyle = fyne.TextStyle{Bold: true}
	expectedValueText.TextSize = 40
	expectedArea := container.NewVBox(expectedText, newCustomSpacer(fyne.NewSize(0, 5)), expectedValueText)
	expectedContainer := container.NewCenter(expectedArea)
	buttonText1 := canvas.NewText("Для начала слива снимите", Black)
	buttonText1.Alignment = fyne.TextAlignCenter
	buttonText1.TextSize = 32
	buttonText1.TextStyle = fyne.TextStyle{Bold: true}
	buttonText2 := canvas.NewText(fmt.Sprintf("пистолет №%v в течение", jarNumber), Black)
	buttonText2.Alignment = fyne.TextAlignCenter
	buttonText2.TextSize = 32
	buttonText2.TextStyle = fyne.TextStyle{Bold: true}
	buttonText3 := canvas.NewText(fmt.Sprintf("%v секунд", timer), Black)
	buttonText3.Alignment = fyne.TextAlignCenter
	buttonText3.TextSize = 32
	buttonText3.TextStyle = fyne.TextStyle{Bold: true}
	buttonText := container.NewVBox(buttonText1, buttonText2, buttonText3)
	borderRect := canvas.NewRectangle(color.Transparent)
	borderRect.CornerRadius = 10.0
	borderRect.StrokeColor = Black
	borderRect.StrokeWidth = 2
	paddedButtonText := container.NewBorder(newCustomSpacer(fyne.NewSize(0, 2)), newCustomSpacer(fyne.NewSize(0, 2)), newCustomSpacer(fyne.NewSize(20, 0)), newCustomSpacer(fyne.NewSize(20, 0)), buttonText)
	buttonArea := container.NewStack(borderRect, paddedButtonText)
	buttonAreaContainer := container.NewCenter(buttonArea)
	paddedButtonArea := container.NewHBox(
		newCustomSpacer(fyne.NewSize(20, 0)), // Отступ 20px слева
		buttonAreaContainer,
	)
	topCenterContent := container.NewVBox(
		newCustomSpacer(fyne.NewSize(0, 20)),
		insertText,
		jarText,
		newCustomSpacer(fyne.NewSize(0, 60)),
		fuelTypeText,
		newCustomSpacer(fyne.NewSize(0, 15)),
		expectedContainer,
	)
	columnContent := container.New(layout.NewBorderLayout(topCenterContent, paddedButtonArea, nil, nil),
		topCenterContent,
		newCustomSpacer(fyne.NewSize(0, 225)),
		paddedButtonArea,
	)
	columnContent = container.New(layout.NewGridWrapLayout(fyne.NewSize(400, 600)), columnContent)
	fyne.Do(func() {
		section.Content.RemoveAll()
		section.Content.Add(columnContent)
		section.Content.Refresh()
		log.Printf("CreateFuelGetStartScreen: jarNumber=%s, size=%v", jarNumber, section.Content.Size())
	})
}

func newCustomSpacer(size fyne.Size) *canvas.Rectangle {
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(size)
	return rect
}

func newFixedHSeparator() *canvas.Rectangle {
	rect := canvas.NewRectangle(Gray)
	rect.SetMinSize(fyne.NewSize(800, 2)) // Обновлено для ширины 800
	return rect
}

func newZeroHSeparator() *canvas.Rectangle {
	rect := canvas.NewRectangle(Gray)
	rect.SetMinSize(fyne.NewSize(350, 2))
	return rect
}

func formatRussianDate(t time.Time) string {
	months := []string{"января", "февраля", "марта", "апреля", "мая", "июня", "июля", "августа", "сентября", "октября", "ноября", "декабря"}
	return fmt.Sprintf("%d %s %d", t.Day(), months[t.Month()-1], t.Year())
}
