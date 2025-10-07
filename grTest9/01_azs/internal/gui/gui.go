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

type QR struct {
	TYPE int    `json:"TYPE"`
	TID  string `json:"TID"`
}

type TopSection struct {
	Content            *fyne.Container
	TimeLable          *canvas.Text
	DateLable          *canvas.Text
	LogoLable          *canvas.Image
	SupportNumberLable *canvas.Text
	KazsNumberLable    *canvas.Text
	Logo               string
	SupportNumber      string
	KazsNumber         string
	Timezone           string
}

type LeftSection struct {
	Content            *fyne.Container
	ActiveDialogCancel context.CancelFunc
	ActiveDialog       *fyne.Container
	ActiveProcess      context.CancelFunc
	Timer              context.CancelFunc
}

type RightSection struct {
	Content            *fyne.Container
	ActiveDialogCancel context.CancelFunc
	ActiveDialog       *fyne.Container
	ActiveProcess      context.CancelFunc
	Timer              context.CancelFunc
}

var logo string = "iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg=="

func NewTopSection() *TopSection {
	topSection, timeLable, dateLable, logoImage, phoneNumber, kazsNumber := CreateHeader(logo, "", "")

	return &TopSection{
		Content:            topSection,
		TimeLable:          timeLable,
		DateLable:          dateLable,
		LogoLable:          logoImage,
		SupportNumberLable: phoneNumber,
		KazsNumberLable:    kazsNumber,
	}
}

func NewLeftSection() *LeftSection {

	defaultScreen := CreateDefaultScreen("1")
	return &LeftSection{
		Content:      defaultScreen,
		ActiveDialog: nil,
	}
}

func NewRightSection() *RightSection {

	defaultScreen := CreateDefaultScreen("2")
	return &RightSection{
		Content:      defaultScreen,
		ActiveDialog: nil,
	}
}

type Gui struct {
	TopSection   *TopSection
	LeftSection  *LeftSection
	RightSection *RightSection

	BottomSection *fyne.Container
	MainContent   *fyne.Container
}

func NewGui() *Gui {

	// Инициализация шапки
	topSection := NewTopSection()

	// Инициализация левой части экрана (1 пистолет)
	leftSection := NewLeftSection()

	// Инициализация правой части экрана (2 пистолет)
	rightSection := NewRightSection()

	bottomSectionSeparator := container.NewBorder(nil, nil, nil, NewFixedVSeparator(), leftSection.Content)
	bottomSection := container.NewGridWithColumns(2, bottomSectionSeparator, rightSection.Content)

	mainContent := container.NewVBox(topSection.Content, bottomSection)

	return &Gui{topSection, leftSection, rightSection, bottomSection, mainContent}
}

func (gui *Gui) SetUI(logo string, supportNumber string, kazsNumber string, timezone string) {
	//TODO:
	gui.TopSection.Logo = logo
	gui.TopSection.SupportNumber = supportNumber
	gui.TopSection.KazsNumber = kazsNumber
	gui.TopSection.Timezone = timezone
}

// CreateHeader Метод отрисовки заголовка
func CreateHeader(logo, supportNumber, kazsNumber string) (*fyne.Container, *canvas.Text, *canvas.Text, *canvas.Image, *canvas.Text, *canvas.Text) {
	// --- Левая часть заголовка ---
	decodedImageBytes, err := base64.StdEncoding.DecodeString(logo)
	if err != nil {
		log.Printf("decode image error: %v", err)
	}

	imageReader := bytes.NewReader(decodedImageBytes)

	_, _, _ = image.Decode(bytes.NewReader(decodedImageBytes))

	itecoImage := canvas.NewImageFromReader(imageReader, "iteco_logo.png")
	if itecoImage != nil {
		itecoImage.FillMode = canvas.ImageFillContain
		itecoImage.SetMinSize(fyne.NewSize(368, 150))
	}

	// --- Средняя часть заголовка ---
	phoneLabel := canvas.NewText(supportNumber, color.Black)
	supportLabel := canvas.NewText("Техническая поддержка", color.Black)

	phoneLabel.Alignment = fyne.TextAlignCenter
	supportLabel.Alignment = fyne.TextAlignCenter

	phoneLabel.TextStyle = fyne.TextStyle{Bold: true}

	phoneLabel.TextSize = 30
	supportLabel.TextSize = 26

	centerHeaderContent := container.NewVBox(NewCustomSpacer(fyne.NewSize(0, 15)), phoneLabel, supportLabel)

	// --- Правая часть заголовка ---
	now := time.Now()
	azsLabel := canvas.NewText(fmt.Sprintf("АЗС №%v", kazsNumber), color.Black)
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

	topSectionContent := container.New(
		layout.NewBorderLayout(nil, nil, itecoImage, rightHeader),
		itecoImage,
		rightHeader,
		centerHeaderContent,
	)

	topSectionContainer := container.NewVBox(topSectionContent, NewFixedHSeparator())

	return topSectionContainer, timeLabel, dateLabel, itecoImage, phoneLabel, azsLabel
}

func (gui *Gui) CreateActivation() *fyne.Container {
	qrText1 := canvas.NewText("Активация", Black)
	qrText1.Alignment = fyne.TextAlignCenter
	qrText1.TextSize = 42

	sectionVBox := container.NewVBox(NewCustomSpacer(fyne.NewSize(0, 450)), qrText1)
	bottomSectionSeparator := container.NewBorder(nil, nil, nil, NewСustomSeparator(fyne.NewSize(0, 1000)), qrText1)
	bottomSectionSeparator2 := container.NewBorder(nil, nil, NewСustomSeparator(fyne.NewSize(0, 1000)), nil, qrText1)

	fyne.Do(func() {
		gui.TopSection.Content.RemoveAll()
		gui.TopSection.Content.Refresh()

		gui.LeftSection.Content.RemoveAll()
		gui.LeftSection.Content.Add(bottomSectionSeparator)
		gui.LeftSection.Content.Refresh()

		gui.RightSection.Content.RemoveAll()
		gui.RightSection.Content.Add(bottomSectionSeparator2)
		gui.RightSection.Content.Refresh()
	})

	return sectionVBox
}

// CreateHeader Метод отрисовки заголовка
func (gui *Gui) CreateHeader() (*fyne.Container, *canvas.Text, *canvas.Text, *canvas.Image, *canvas.Text, *canvas.Text) {
	// --- Левая часть заголовка ---
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

	// --- Средняя часть заголовка ---
	phoneLabel := canvas.NewText(gui.TopSection.SupportNumber, color.Black)
	supportLabel := canvas.NewText("Техническая поддержка", color.Black)

	phoneLabel.Alignment = fyne.TextAlignCenter
	supportLabel.Alignment = fyne.TextAlignCenter

	phoneLabel.TextStyle = fyne.TextStyle{Bold: true}

	phoneLabel.TextSize = 30
	supportLabel.TextSize = 26

	centerHeaderContent := container.NewVBox(NewCustomSpacer(fyne.NewSize(0, 15)), phoneLabel, supportLabel)

	// --- Правая часть заголовка ---
	now := time.Now()
	azsLabel := canvas.NewText(fmt.Sprintf("АЗС %v", gui.TopSection.KazsNumber), color.Black)
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

	topSectionContent := container.New(
		layout.NewBorderLayout(nil, nil, itecoImage, rightHeader),
		itecoImage,
		rightHeader,
		centerHeaderContent,
	)

	topSectionContainer := container.NewVBox(topSectionContent, NewFixedHSeparator())

	fyne.Do(func() {
		gui.TopSection.DateLable = dateLabel
		gui.TopSection.TimeLable = timeLabel
		gui.TopSection.Content.RemoveAll()
		gui.TopSection.Content.Add(topSectionContainer)
		gui.TopSection.Content.Refresh()
	})

	return topSectionContainer, timeLabel, dateLabel, itecoImage, phoneLabel, azsLabel
}

// CreateDefaultScreen Метод отрисовки стартового экрана
func CreateDefaultScreen(jarNumber string) *fyne.Container {
	qrText1 := canvas.NewText("Отсканируйте QR-код", Black)
	qrText2 := canvas.NewText(fmt.Sprintf("для пистолета №%v", jarNumber), Black)
	qrText1.Alignment = fyne.TextAlignCenter
	qrText1.TextSize = 42

	qrText2.Alignment = fyne.TextAlignCenter
	qrText2.TextSize = 42

	sectionVBox := container.NewVBox(qrText1, qrText2)
	sectionCenter := container.NewCenter(sectionVBox)

	return sectionCenter
}

func (gui *Gui) CreateDefaultScreen(jarNumber string) *fyne.Container {
	qrText1 := canvas.NewText("Отсканируйте QR-код", Black)
	qrText2 := canvas.NewText(fmt.Sprintf("для пистолета №%v", jarNumber), Black)
	qrText1.Alignment = fyne.TextAlignCenter
	qrText1.TextSize = 42

	qrText2.Alignment = fyne.TextAlignCenter
	qrText2.TextSize = 42

	sectionVBox := container.NewVBox(qrText1, qrText2)
	sectionCenter := container.NewCenter(sectionVBox)

	if jarNumber == "1" {
		fyne.Do(func() {
			gui.LeftSection.Content.RemoveAll()
			gui.LeftSection.Content.Add(sectionCenter)
			gui.LeftSection.Content.Refresh()
		})
	} else {
		fyne.Do(func() {
			gui.RightSection.Content.RemoveAll()
			gui.RightSection.Content.Add(sectionCenter)
			gui.RightSection.Content.Refresh()
		})
	}

	return sectionCenter
}

func (gui *Gui) CreateDownloadScreen(jarNumber string) *fyne.Container {
	qrText1 := canvas.NewText("Обработка", Green)
	qrText1.Alignment = fyne.TextAlignCenter
	qrText1.TextSize = 42

	sectionVBox := container.NewVBox(qrText1)
	sectionCenter := container.NewCenter(sectionVBox)

	if jarNumber == "1" {
		fyne.Do(func() {
			gui.LeftSection.Content.RemoveAll()
			gui.LeftSection.Content.Add(sectionCenter)
			gui.LeftSection.Content.Refresh()
		})
	} else {
		fyne.Do(func() {
			gui.RightSection.Content.RemoveAll()
			gui.RightSection.Content.Add(sectionCenter)
			gui.RightSection.Content.Refresh()
		})
	}

	return sectionCenter
}

func (gui *Gui) CreateStartScreen(jarNumber string) *fyne.Container {
	qrText1 := canvas.NewText("Загрузка приложения...", Green)
	qrText1.Alignment = fyne.TextAlignCenter
	qrText1.TextSize = 42

	sectionVBox := container.NewVBox(qrText1)
	sectionCenter := container.NewCenter(sectionVBox)

	if jarNumber == "1" {
		fyne.Do(func() {
			gui.LeftSection.Content.RemoveAll()
			gui.LeftSection.Content.Add(sectionCenter)
			gui.LeftSection.Content.Refresh()
		})
	} else {
		fyne.Do(func() {
			gui.RightSection.Content.RemoveAll()
			gui.RightSection.Content.Add(sectionCenter)
			gui.RightSection.Content.Refresh()
		})
	}

	return sectionCenter
}

func (gui *Gui) CreateTechnicalErrorScreen(jarNumber string) *fyne.Container {
	qrText1 := canvas.NewText("Технические неполадки", Green)
	qrText1.Alignment = fyne.TextAlignCenter
	qrText1.TextSize = 42

	sectionVBox := container.NewVBox(qrText1)
	sectionCenter := container.NewCenter(sectionVBox)
	if jarNumber == "1" {
		fyne.Do(func() {
			gui.LeftSection.Content.RemoveAll()
			gui.LeftSection.Content.Add(sectionCenter)
			gui.LeftSection.Content.Refresh()
		})
	} else {
		fyne.Do(func() {
			gui.RightSection.Content.RemoveAll()
			gui.RightSection.Content.Add(sectionCenter)
			gui.RightSection.Content.Refresh()
		})
	}

	return sectionCenter
}

// CreateFuelGiveStartScreen Метод отрисовки стартового экрана заправки
func (gui *Gui) CreateFuelGiveStartScreen(jarNumber string, liters float32, fuelType string, timer int) *fyne.Container {
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
	fuelTypeText.Alignment = fyne.TextAlignLeading
	fuelTypeText.Wrapping = fyne.TextWrapWord

	// Текст "Максимальный объем"
	maxVolumeText := canvas.NewText("Максимальный объем", Gray)
	maxVolumeText.Alignment = fyne.TextAlignLeading
	maxVolumeText.TextSize = 32

	// Текст "количество_литров"
	litersText := canvas.NewText(fmt.Sprintf("%v литров", liters), Black)
	litersText.TextSize = 40
	litersText.TextStyle = fyne.TextStyle{Bold: true}

	maxVolumeAndLitersArea := container.NewVBox(maxVolumeText, NewCustomSpacer(fyne.NewSize(0, 5)), litersText)
	maxVolumeAndLitersContainer := container.NewHBox(NewCustomSpacer(fyne.NewSize(5, 0)), maxVolumeAndLitersArea)

	// Нижняя рамка
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

	topCenterContent := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(0, 20)),
		insertText,
		gunText,
		NewCustomSpacer(fyne.NewSize(0, 60)),
		fuelTypeText,
		NewCustomSpacer(fyne.NewSize(0, 15)),
		maxVolumeAndLitersContainer,
	)

	columnContent := container.New(layout.NewBorderLayout(topCenterContent, buttonAreaContainer, nil, nil),
		topCenterContent,
		NewCustomSpacer(fyne.NewSize(0, 225)),
		buttonAreaContainer,
	)

	if jarNumber == "1" {
		fyne.Do(func() {
			gui.LeftSection.Content.RemoveAll()
			gui.LeftSection.Content.Add(columnContent)
			gui.LeftSection.Content.Refresh()
		})
	} else {
		fyne.Do(func() {
			gui.RightSection.Content.RemoveAll()
			gui.RightSection.Content.Add(columnContent)
			gui.RightSection.Content.Refresh()
		})
	}

	return columnContent
}

// CreateFuelGiveInProgressScreen Метод отрисовки процесса заправки
func (gui *Gui) CreateFuelGiveInProgressScreen(jarNumber string, fuelType string, liters float32, expectedLiters float32) *fyne.Container {
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
	fuelTypeText.Alignment = fyne.TextAlignLeading
	fuelTypeText.Wrapping = fyne.TextWrapWord

	// Текст "Максимальный объем"
	maxVolumeText := canvas.NewText("Максимальный объем", Gray)
	maxVolumeText.Alignment = fyne.TextAlignLeading
	maxVolumeText.TextSize = 32

	// Текст "количество_литров"
	volumeValueText := canvas.NewText(fmt.Sprintf("%v литров", expectedLiters), Black)
	volumeValueText.TextStyle = fyne.TextStyle{Bold: true}
	volumeValueText.TextSize = 40

	maxVolumeAndVolumeArea := container.NewVBox(maxVolumeText, volumeValueText)
	maxVolumeAndVolumeContainer := container.NewHBox(NewCustomSpacer(fyne.NewSize(5, 0)), maxVolumeAndVolumeArea)

	// Текст "количество_заправленных_литров"
	amountText := canvas.NewText(fmt.Sprintf("%.2f", liters), Black)
	amountText.Alignment = fyne.TextAlignCenter
	amountText.TextSize = 100

	// Текст "литров залито"
	litersText := canvas.NewText("литров залито", Gray)
	litersText.Alignment = fyne.TextAlignCenter
	litersText.TextSize = 40

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

	// Имитация рамки кнопки с помощью Rectangle
	borderRect := canvas.NewRectangle(color.Transparent)
	borderRect.StrokeColor = color.Black
	borderRect.CornerRadius = 10.0
	borderRect.StrokeWidth = 2

	paddedButtonText := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 2)), NewCustomSpacer(fyne.NewSize(0, 2)), NewCustomSpacer(fyne.NewSize(20, 0)), NewCustomSpacer(fyne.NewSize(20, 0)), buttonText)

	buttonArea := container.NewStack(borderRect, paddedButtonText)

	buttonAreaContainer := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(10, 0)), NewCustomSpacer(fyne.NewSize(10, 0)), buttonArea)

	topCenterContent := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(0, 20)),
		ifProgressText,
		gunText,
		NewCustomSpacer(fyne.NewSize(0, 60)),
		fuelTypeText,
		NewCustomSpacer(fyne.NewSize(0, 10)),
		maxVolumeAndVolumeContainer,
		NewCustomSpacer(fyne.NewSize(0, 85)),
		amountText,
		litersText,
	)

	columnContent := container.New(layout.NewBorderLayout(topCenterContent, buttonAreaContainer, nil, nil),
		topCenterContent,
		NewCustomSpacer(fyne.NewSize(0, 20)),
		buttonAreaContainer,
	)

	if jarNumber == "1" {
		fyne.Do(func() {
			gui.LeftSection.Content.RemoveAll()
			gui.LeftSection.Content.Add(columnContent)
			gui.LeftSection.Content.Refresh()
		})
	} else {
		fyne.Do(func() {
			gui.RightSection.Content.RemoveAll()
			gui.RightSection.Content.Add(columnContent)
			gui.RightSection.Content.Refresh()
		})
	}

	return columnContent
}

// CreateFuelGiveCompleteScreen Метод отрисовки завершения процесса заправки
func (gui *Gui) CreateFuelGiveCompleteScreen(jarNumber string, fuelType, documentNumber string, planAmount, factAmount float32, startDate, endDate int64, timer int) *fyne.Container {

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
	// Текст "№ документа"
	documentText := canvas.NewText("№ документа", Gray)
	documentValueText := canvas.NewText(documentNumber, Black)
	documentText.TextSize = 32
	documentValueText.TextSize = 32
	documentValueText.Alignment = fyne.TextAlignTrailing
	documentValueText.TextStyle = fyne.TextStyle{Bold: true}

	// Текст "Заправка план"
	planText := canvas.NewText("Заправка план", Gray)
	planValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringFull(planAmount)), Black)
	planText.TextSize = 32
	planValueText.TextSize = 32
	planValueText.Alignment = fyne.TextAlignTrailing
	planValueText.TextStyle = fyne.TextStyle{Bold: true}

	// Текст "Заправка факт"
	factText := canvas.NewText("Заправка факт", Gray)
	factValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringFull(factAmount)), Black)
	factText.TextSize = 32
	factValueText.TextSize = 32
	factValueText.Alignment = fyne.TextAlignTrailing
	factValueText.TextStyle = fyne.TextStyle{Bold: true}

	// Текст "Дата начала"
	startDateText := canvas.NewText("Дата начала", Gray)
	startDateValueText := canvas.NewText(ConvertUnixToString(startDate), Black)
	startDateText.TextSize = 32
	startDateValueText.TextSize = 32
	startDateValueText.Alignment = fyne.TextAlignTrailing
	startDateValueText.TextStyle = fyne.TextStyle{Bold: true}

	// Текст "Дата окончания"
	endDateText := canvas.NewText("Дата окончания", Gray)
	endDateValueText := canvas.NewText(ConvertUnixToString(endDate), Black)
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

	buttonText2 := canvas.NewText(fmt.Sprintf("возможна через %v секунд", timer), Black)
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

	topCenterContent := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(0, 15)),
		completeText,
		NewFixedSpacer(),
		fuelTypePistolText,
		NewFixedSpacer(),
		dataBorder,
	)

	columnContent := container.New(layout.NewBorderLayout(topCenterContent, buttonAreaContainer, nil, nil),
		topCenterContent,
		NewCustomSpacer(fyne.NewSize(0, 75)),
		buttonAreaContainer,
	)

	if jarNumber == "1" {
		fyne.Do(func() {
			gui.LeftSection.Content.RemoveAll()
			gui.LeftSection.Content.Add(columnContent)
			gui.LeftSection.Content.Refresh()
		})
	} else {
		fyne.Do(func() {
			gui.RightSection.Content.RemoveAll()
			gui.RightSection.Content.Add(columnContent)
			gui.RightSection.Content.Refresh()
		})
	}

	return columnContent

}

// CreateFuelGetStartScreen Метод отрисовки стартового окна пополнения
func (gui *Gui) CreateFuelGetStartScreen(jarNumber string, fuelType string, filledAmount float32, availableAmount float32, expectedAmount float32, timer int) *fyne.Container {
	percentage := int(filledAmount / (filledAmount + availableAmount) * 100.0)
	_ = percentage

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
	fuelTypeText.Alignment = fyne.TextAlignLeading
	fuelTypeText.Wrapping = fyne.TextWrapWord

	// Данные о заправке
	// Текст "Заполнено"
	filledAmountText := canvas.NewText("Заполнено", Gray)
	filledAmountText.Alignment = fyne.TextAlignLeading
	filledAmountText.TextSize = 32
	filledAmountValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringShort(filledAmount)), Black)
	filledAmountValueText.Alignment = fyne.TextAlignLeading
	filledAmountValueText.TextSize = 40
	filledAmountValueText.TextStyle = fyne.TextStyle{Bold: true}

	// Текст "Доступный объем"
	availableAmountText := canvas.NewText("Доступный объем", Gray)
	availableAmountText.Alignment = fyne.TextAlignLeading
	availableAmountText.TextSize = 32
	availableAmountValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringShort(availableAmount)), Black)
	availableAmountValueText.Alignment = fyne.TextAlignLeading
	availableAmountValueText.TextSize = 40
	availableAmountValueText.TextStyle = fyne.TextStyle{Bold: true}

	// Текст "Ожидаемый слив"
	expectedAmountText := canvas.NewText("Ожидаемый слив", Gray)
	expectedAmountText.Alignment = fyne.TextAlignLeading
	expectedAmountText.TextSize = 32
	expectedAmountValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringShort(expectedAmount)), Black)
	expectedAmountValueText.Alignment = fyne.TextAlignLeading
	expectedAmountValueText.TextSize = 40
	expectedAmountValueText.TextStyle = fyne.TextStyle{Bold: true}

	fuelGetDataContainer := container.NewGridWithRows(6, filledAmountText, filledAmountValueText, availableAmountText, availableAmountValueText, expectedAmountText, expectedAmountValueText)

	// Прогресс бар

	progressBarHeight := float32(329)
	progressBarWidth := float32(85)
	borderThickness := float32(3)

	// Рамка

	progressBarBackground := canvas.NewRectangle(color.Transparent)
	progressBarBackground.StrokeColor = color.Black
	progressBarBackground.StrokeWidth = borderThickness
	progressBarBackground.CornerRadius = 10.0
	progressBarBackground.SetMinSize(fyne.NewSize(progressBarWidth, progressBarHeight))

	// Заполненная часть

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
	progressBarFilled.SetMinSize(fyne.NewSize(innerBarWidth, filledHeight))
	progressBarFilled.CornerRadius = 10.0

	spaceAboveFilled := innerBarHeight - filledHeight
	if spaceAboveFilled < 0 {
		spaceAboveFilled = 0
	}

	filledBarContent := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(innerBarWidth, spaceAboveFilled)),
		progressBarFilled,
	)

	progressBarArea := container.NewStack(
		progressBarBackground,
		filledBarContent,
	)

	// Текст процента
	percentageText := canvas.NewText(fmt.Sprintf("%v%%", percentage), color.Black)
	percentageText.Alignment = fyne.TextAlignLeading
	percentageText.TextSize = 40
	percentageText.TextStyle.Bold = true
	percentageContainer := container.NewHBox(NewCustomSpacer(fyne.NewSize(5, 0)), percentageText)

	barAndPercentage := container.NewVBox(
		progressBarArea,
		percentageContainer,
		layout.NewSpacer(),
	)

	barAndPercentageCentered := container.NewCenter(barAndPercentage)

	amountBarPercentRow := container.NewHBox(
		NewCustomSpacer(fyne.NewSize(10, 0)),
		fuelGetDataContainer,
		layout.NewSpacer(),
		barAndPercentageCentered,
		NewCustomSpacer(fyne.NewSize(10, 0)),
	)

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

	topCenterContent := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(0, 20)),
		drainText,
		tankText,
		NewCustomSpacer(fyne.NewSize(0, 20)),
		fuelTypeText,
		NewCustomSpacer(fyne.NewSize(0, 10)),
		amountBarPercentRow,
	)

	columnContent := container.New(layout.NewBorderLayout(topCenterContent, buttonAreaContainer, nil, nil),
		topCenterContent,
		NewCustomSpacer(fyne.NewSize(0, 35)),
		buttonAreaContainer,
	)

	if jarNumber == "1" {
		fyne.Do(func() {
			gui.LeftSection.Content.RemoveAll()
			gui.LeftSection.Content.Add(columnContent)
			gui.LeftSection.Content.Refresh()
		})
	} else {
		fyne.Do(func() {
			gui.RightSection.Content.RemoveAll()
			gui.RightSection.Content.Add(columnContent)
			gui.RightSection.Content.Refresh()
		})
	}
	return columnContent
}

// CreateFuelGetInProgressScreen Метод отрисовки процесса пополнения
func (gui *Gui) CreateFuelGetInProgressScreen(jarNumber string, expectedAmount, drainedAmount, fuelVolume, jarVolume float32, timer int) *fyne.Container {
	var percentage int
	percentage = int(fuelVolume / jarVolume * 100.0)
	if percentage >= 100 {
		percentage = 99
	}

	// Текст "Слив бензовоза"
	drainText := canvas.NewText("Слив бензовоза", Black)
	drainText.Alignment = fyne.TextAlignCenter
	drainText.TextSize = 32

	// Текст "Емкость №X"
	jarText := canvas.NewText(fmt.Sprintf("ЁМКОСТЬ №%v", jarNumber), Black)
	jarText.Alignment = fyne.TextAlignCenter
	jarText.TextStyle = fyne.TextStyle{Bold: true}
	jarText.TextSize = 40

	// Текст "Ожидаемый слив"
	expectedText := canvas.NewText("Ожидаемый слив", Gray)
	expectedText.Alignment = fyne.TextAlignLeading
	expectedText.TextSize = 32

	expectedContainer := container.NewHBox(NewCustomSpacer(fyne.NewSize(25, 0)), expectedText)

	// Текс "количество_литров"
	expectedValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringShort(expectedAmount)), Black)
	expectedValueText.Alignment = fyne.TextAlignLeading
	expectedValueText.TextStyle = fyne.TextStyle{Bold: true}
	expectedValueText.TextSize = 40

	expectedValueContainer := container.NewHBox(NewCustomSpacer(fyne.NewSize(25, 0)), expectedValueText)

	// Текст "количество_слитых_литров"

	drainedValueText := canvas.NewText(fmt.Sprintf("%v", ConvertFloat32ToStringShort(drainedAmount)), Black)
	drainedValueText.Alignment = fyne.TextAlignCenter
	drainedValueText.TextStyle = fyne.TextStyle{Bold: true}
	drainedValueText.TextSize = 100

	// Текст "Литров слито"
	drainedText := canvas.NewText("литров слито", Gray)
	drainedText.Alignment = fyne.TextAlignCenter
	drainedText.TextSize = 40

	amountAndLiters := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(0, 100)),
		drainedValueText,
		drainedText,
	)

	amountAndLitersAligned := container.NewHBox(amountAndLiters)

	// Прогресс бар

	progressBarHeight := float32(329)
	progressBarWidth := float32(85)
	borderThickness := float32(2)

	// Рамка

	progressBarBackground := canvas.NewRectangle(color.Transparent)
	progressBarBackground.StrokeColor = color.Black
	progressBarBackground.StrokeWidth = borderThickness
	progressBarBackground.CornerRadius = 10.0
	progressBarBackground.SetMinSize(fyne.NewSize(progressBarWidth, progressBarHeight))

	// Заполненная часть

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
	progressBarFilled.SetMinSize(fyne.NewSize(innerBarWidth, filledHeight))
	progressBarFilled.CornerRadius = 10.0

	spaceAboveFilled := innerBarHeight - filledHeight
	if spaceAboveFilled < 0 {
		spaceAboveFilled = 0
	}

	filledBarContent := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(innerBarWidth, spaceAboveFilled)),
		progressBarFilled,
	)

	progressBarArea := container.NewStack(
		progressBarBackground,
		filledBarContent,
	)

	// Текст процента
	percentageText := canvas.NewText(fmt.Sprintf("%v%%", percentage), color.Black)
	percentageText.Alignment = fyne.TextAlignLeading
	percentageText.TextSize = 40
	percentageText.TextStyle.Bold = true
	percentageContainer := container.NewHBox(NewCustomSpacer(fyne.NewSize(5, 0)), percentageText)
	_ = percentageContainer

	amountBarPercentRow := container.NewHBox(
		NewCustomSpacer(fyne.NewSize(25, 0)),
		amountAndLitersAligned,
		layout.NewSpacer(),
		progressBarArea,
		NewCustomSpacer(fyne.NewSize(25, 0)),
	)

	// Рамка снизу

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

	// Имитация рамки кнопки с помощью Rectangle
	borderRect := canvas.NewRectangle(color.Transparent)
	borderRect.StrokeColor = color.Black
	borderRect.CornerRadius = 10.0
	borderRect.StrokeWidth = 2

	// Контейнер для текста кнопки с отступами и центрированием
	paddedButtonText := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 2)), NewCustomSpacer(fyne.NewSize(0, 2)), NewCustomSpacer(fyne.NewSize(20, 0)), NewCustomSpacer(fyne.NewSize(20, 0)), buttonText)

	// StackLayout для наложения рамки под текстом кнопки
	buttonArea := container.NewStack(borderRect, paddedButtonText)

	buttonAreaContainer := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(10, 0)), NewCustomSpacer(fyne.NewSize(10, 0)), buttonArea)

	topCenterContent := container.NewVBox(
		NewCustomSpacer(fyne.NewSize(0, 25)),
		drainText,
		jarText,
		NewCustomSpacer(fyne.NewSize(0, 30)),
		expectedContainer,
		expectedValueContainer,
		layout.NewSpacer(),
		amountBarPercentRow,
	)

	percentageContainerPadded := container.NewHBox(NewCustomSpacer(fyne.NewSize(460, 0)), percentageContainer)

	columnContent := container.New(layout.NewBorderLayout(topCenterContent, buttonAreaContainer, nil, nil),
		topCenterContent,
		NewCustomSpacer(fyne.NewSize(515, 0)),
		percentageContainerPadded,
		NewCustomSpacer(fyne.NewSize(0, 70)),
		buttonAreaContainer,
	)

	if jarNumber == "1" {
		fyne.Do(func() {
			gui.LeftSection.Content.RemoveAll()
			gui.LeftSection.Content.Add(columnContent)
			gui.LeftSection.Content.Refresh()
		})
	} else {
		fyne.Do(func() {
			gui.RightSection.Content.RemoveAll()
			gui.RightSection.Content.Add(columnContent)
			gui.RightSection.Content.Refresh()
		})
	}

	return columnContent
}

// CreateFuelGetCompleteScreen Метод отрисовки завершения процесса пополнения
func (gui *Gui) CreateFuelGetCompleteScreen(jarNumber string, fuelType string, documentNumber string, beforeFuelGet float32, afterFuelGet float32, fuelGetPlan float32, fuelGetFact float32, startTime, endTime int64, timer int) *fyne.Container {
	// Текст "СЛИВ ЗАВЕРШЁН"
	completeText := canvas.NewText("СЛИВ ЗАВЕРШЁН", Black)
	completeText.Alignment = fyne.TextAlignCenter
	completeText.TextSize = 40
	completeText.TextStyle = fyne.TextStyle{Bold: true}

	// Текст "тип_топлива + Пистолет №X"
	fuelTypePistolText := widget.NewLabel(fmt.Sprintf("%s Пистолет №%v", fuelType, jarNumber))
	fuelTypePistolText.Alignment = fyne.TextAlignCenter
	fuelTypePistolText.Wrapping = fyne.TextWrapWord

	// Данные о заправке
	// Текст "№ документа"
	documentText := canvas.NewText("№ документа", Gray)
	documentValueText := canvas.NewText(documentNumber, Black)
	documentText.TextSize = 32
	documentValueText.TextSize = 32
	documentValueText.Alignment = fyne.TextAlignTrailing
	documentValueText.TextStyle = fyne.TextStyle{Bold: true}

	// Текст "До слива"
	beforeFuelGetText := canvas.NewText("До слива", Gray)
	beforeFuelGetValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringShort(beforeFuelGet)), Black)
	beforeFuelGetText.TextSize = 32
	beforeFuelGetValueText.TextSize = 32
	beforeFuelGetValueText.Alignment = fyne.TextAlignTrailing
	beforeFuelGetValueText.TextStyle = fyne.TextStyle{Bold: true}

	// Текст "После слива"
	afterFuelGetText := canvas.NewText("После слива", Gray)
	afterFuelGetValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringShort(afterFuelGet)), Black)
	afterFuelGetText.TextSize = 32
	afterFuelGetValueText.TextSize = 32
	afterFuelGetValueText.Alignment = fyne.TextAlignTrailing
	afterFuelGetValueText.TextStyle = fyne.TextStyle{Bold: true}

	// Текст "Слив план"
	fuelGetPlanText := canvas.NewText("Слив план", Gray)
	fuelGetPlanValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringShort(fuelGetPlan)), Black)
	fuelGetPlanText.TextSize = 32
	fuelGetPlanValueText.TextSize = 32
	fuelGetPlanValueText.Alignment = fyne.TextAlignTrailing
	fuelGetPlanValueText.TextStyle = fyne.TextStyle{Bold: true}

	// Текст "Слив факт"
	fuelGetFactText := canvas.NewText("Слив факт", Gray)
	fuelGetFactValueText := canvas.NewText(fmt.Sprintf("%v литров", ConvertFloat32ToStringShort(fuelGetFact)), Black)
	fuelGetFactText.TextSize = 32
	fuelGetFactValueText.TextSize = 32
	fuelGetFactValueText.Alignment = fyne.TextAlignTrailing
	fuelGetFactValueText.TextStyle = fyne.TextStyle{Bold: true}

	// Текст "Дата начала"
	startTimeText := canvas.NewText("Дата начала", Gray)
	startTimeValueText := canvas.NewText(ConvertUnixToString(startTime), Black)
	startTimeText.TextSize = 32
	startTimeValueText.TextSize = 32
	startTimeValueText.Alignment = fyne.TextAlignTrailing
	startTimeValueText.TextStyle = fyne.TextStyle{Bold: true}

	// Текст "Дата окончания"
	endTimeText := canvas.NewText("Дата окончания", Gray)
	endTimeValueText := canvas.NewText(ConvertUnixToString(endTime), Black)
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

	buttonText2 := canvas.NewText(fmt.Sprintf("возможна через %v секунд", timer), Black)
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

	if jarNumber == "1" {
		fyne.Do(func() {
			gui.LeftSection.Content.RemoveAll()
			gui.LeftSection.Content.Add(columnContent)
			gui.LeftSection.Content.Refresh()
		})
	} else {
		fyne.Do(func() {
			gui.RightSection.Content.RemoveAll()
			gui.RightSection.Content.Add(columnContent)
			gui.RightSection.Content.Refresh()
		})
	}

	return columnContent
}

// ShowSectionDialog Метод отрисовки диалогового окна
func (g *Gui) ShowSectionDialog(sectionStack *fyne.Container, title string, message string, timerSeconds int, onClose func()) {

	if sectionStack == nil {
		log.Println("Ошибка: Контейнер секции равен nil.")
		// Возвращаем nil для всех возвращаемых значений, чтобы избежать паники.
		return
	}

	// Контекст для управления отменой горутины таймера.
	ctx, cancel := context.WithCancel(context.Background())

	// Канал для сигнализации о закрытии диалога.
	dialogClosed := make(chan struct{})

	// Функция, которая будет вызвана при завершении таймера или принудительном закрытии.
	// Она гарантирует, что канал dialogClosed будет закрыт.
	onDialogClose := func() {
		// Вызываем пользовательскую функцию обратного вызова, если она предоставлена.
		if onClose != nil {
			onClose()
		}
		// Отменяем контекст, чтобы остановить горутину таймера.
		cancel()
		// Закрываем канал, сигнализируя о завершении.
		select {
		case <-dialogClosed:
			// Канал уже закрыт, ничего не делаем.
		default:
			close(dialogClosed)
		}
	}

	// Создаем полупрозрачный фон для затемнения содержимого секции
	overlayBackground := canvas.NewRectangle(color.RGBA{128, 128, 128, 200})

	// Размер будет установлен StackLayout'ом
	overlayBackground.SetMinSize(sectionStack.Size())
	overlayBackground.Show()

	// --- Создаем содержимое самого диалога ---
	// Заголовок (красный)
	titleText := canvas.NewText(title, color.RGBA{191, 7, 7, 255})
	titleText.Alignment = fyne.TextAlignLeading
	titleText.TextSize = 40
	titleText.TextStyle = fyne.TextStyle{Bold: true}

	// Разделитель
	separator := NewZeroHSeparator()

	// Текст сообщения (серый)
	messageText := widget.NewLabel(message)
	messageText.Wrapping = fyne.TextWrapWord
	messageText.Alignment = fyne.TextAlignLeading

	// Кнопка с таймером
	timerButton := createTimerButtonContent(ctx, timerSeconds, onDialogClose)

	// Объединяем элементы диалога вертикально
	dialogContentVBox := container.NewVBox(
		titleText,
		separator,
		messageText,
		NewCustomSpacer(fyne.NewSize(0, 5)),
		timerButton,
	)

	// Добавляем белый фон и отступы к содержимому диалога
	dialogBackground := canvas.NewRectangle(color.RGBA{255, 255, 255, 255})
	dialogBackground.SetMinSize(fyne.NewSize(550, 150))
	dialogContentPadded := container.NewPadded(dialogContentVBox) // Отступы вокруг содержимого
	dialogContent := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(0, 10)), NewCustomSpacer(fyne.NewSize(10, 0)), NewCustomSpacer(fyne.NewSize(10, 0)), dialogContentPadded)

	// StackLayout для наложения белого фона под содержимым диалога
	dialogArea := container.NewStack(dialogBackground, dialogContent)

	// Центрируем область диалога внутри полупрозрачного фона (и, соответственно, внутри Stack секции)
	centeredDialog := container.NewCenter(dialogArea)

	// Объединяем полупрозрачный фон и центрированный диалог в один контейнер-оверлей
	overlayContainer := container.NewStack(overlayBackground, centeredDialog)

	// Размер будет установлен StackLayout'ом родителя
	overlayContainer.Show()

	// Добавляем контейнер-оверлей в Stack секции.
	fyne.Do(func() {
		sectionStack.Add(overlayContainer)
		sectionStack.Refresh()
	})

}

// createTimerButtonContent создает содержимое кнопки с таймером.
func createTimerButtonContent(ctx context.Context, initialSeconds int, onTimerComplete func()) *fyne.Container {

	// Текст кнопки с таймером
	timerText := canvas.NewText(fmt.Sprintf("Закроется через (%d с)", initialSeconds), color.Black)
	timerText.Alignment = fyne.TextAlignCenter
	timerText.TextSize = 32 // Размер текста по скриншоту

	// Имитация рамки кнопки
	borderRect := canvas.NewRectangle(color.Transparent)
	borderRect.StrokeColor = color.Black
	borderRect.CornerRadius = 10.0 // Закругленные углы
	borderRect.StrokeWidth = 2
	borderRect.SetMinSize(fyne.NewSize(530, 60))

	// Контейнер для текста кнопки с отступами и центрированием
	paddedText := container.NewPadded(container.NewCenter(timerText))
	paddedContainer := container.NewBorder(NewCustomSpacer(fyne.NewSize(0, 5)), NewCustomSpacer(fyne.NewSize(0, 5)), NewCustomSpacer(fyne.NewSize(20, 0)), NewCustomSpacer(fyne.NewSize(20, 0)), paddedText)

	// StackLayout для наложения рамки под текстом кнопки
	buttonArea := container.NewStack(borderRect, paddedContainer)

	// *** Запускаем таймер в горутине ***
	go func() {

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		secondsRemaining := initialSeconds

		for {
			select {
			case <-ctx.Done(): // Ожидаем сигнал отмены из контекста
				log.Println("Горутина таймера отменена через контекст.")
				return // Выходим из горутины при отмене
			case <-ticker.C: // Ожидаем тик таймера

				secondsRemaining--
				if secondsRemaining < 0 {
					secondsRemaining = 0
				}

				// Обновляем текст таймера в главном потоке Fyne
				fyne.Do(func() {
					timerText.Text = fmt.Sprintf("Закроется через (%d с)", secondsRemaining)
					timerText.Refresh()
				})
				if secondsRemaining <= 0 {
					// Таймер завершен, вызываем функцию обратного вызова
					if onTimerComplete != nil {
						fyne.Do(onTimerComplete) // Вызываем в главном потоке
					}
					return // Выходим из горутины
				}
			}
		}
	}()

	return container.NewCenter(buttonArea)
}

// HideSectionDialog скрывает диалоговое окно из sectionStack.
func (g *Gui) HideSectionDialog(sectionStack *fyne.Container, overlayContainer fyne.CanvasObject) {

	if sectionStack == nil {
		log.Println("Ошибка HideSectionDialog: Контейнер секции равен nil.")
		return
	}

	fyne.Do(func() {
		var overlayToRemove fyne.CanvasObject
		switch {
		case overlayContainer != nil:
			overlayToRemove = overlayContainer
		case len(sectionStack.Objects) > 0:
			overlayToRemove = sectionStack.Objects[len(sectionStack.Objects)-1]
		default:
			log.Println("Ошибка HideSectionDialog: Нет элементов для удаления из Stack.")
			return
		}
		sectionStack.Remove(overlayToRemove)
		sectionStack.Refresh()
	})
}

func (g *Gui) IsSectionDialogActive(pistolNumber string) bool {
	switch pistolNumber {
	case "left":
		return g.LeftSection.ActiveDialog != nil
	case "right":
		return g.RightSection.ActiveDialog != nil
	default:
		log.Printf("Ошибка IsSectionDialogActive: Неизвестная колонка '%s'", pistolNumber)
		return false
	}
}

func (g *Gui) CloseDialog(pistolNumber string) {
	switch pistolNumber {
	case "1":
		if g.LeftSection.ActiveDialog != nil {
			g.LeftSection.ActiveDialog = nil
			g.LeftSection.ActiveDialogCancel()
		}
	case "2":
		if g.RightSection.ActiveDialog != nil {
			g.RightSection.ActiveDialog = nil
			g.RightSection.ActiveDialogCancel()
		}
	}
}
