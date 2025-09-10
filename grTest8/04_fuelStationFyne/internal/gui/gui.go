package gui

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"image"
	"log"
	"time"

	_ "fuelstation/internal/db"
	_ "fuelstation/internal/gui/utils"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Корректная строка Base64 для минимального PNG
var Logo string = "iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAAYSURBVFhH7cEBAQAACAIwAAEAAABAAADg9wL8"

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
	Timer              context.CancelFunc
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

type Gui struct {
	TopSection    *TopSection
	LeftSection   *LeftSection
	RightSection  *RightSection
	BottomSection *fyne.Container
	MainContent   *fyne.Container
	Window        fyne.Window
	DB            *sql.DB
	updateChan    chan func() // Канал для обновления GUI в главном потоке
}

func NewFyneApp() fyne.App {
	log.Println("NewFyneApp: Создание нового приложения Fyne")
	return app.New()
}

func NewGui() *Gui {
	gui := &Gui{
		TopSection:    &TopSection{},
		LeftSection:   &LeftSection{},
		RightSection:  &RightSection{},
		BottomSection: container.NewVBox(),
		updateChan:    make(chan func(), 100), // Буферизированный канал для обновлений GUI
	}
	log.Println("NewGui: Создан новый объект Gui")
	return gui
}

func (g *Gui) RunGui(a fyne.App, ready chan<- struct{}, db *sql.DB) error {
	log.Println("RunGui: Начало инициализации GUI")
	g.DB = db
	g.Window = a.NewWindow("Fuel Station")
	log.Println("RunGui: Вызов setupTopSection")
	if err := g.setupTopSection(); err != nil {
		log.Printf("RunGui: Ошибка в setupTopSection: %v", err)
		// Продолжаем выполнение, несмотря на ошибку логотипа
	}
	log.Println("RunGui: Вызов setupMainContent")
	if err := g.setupMainContent(); err != nil {
		log.Printf("RunGui: Ошибка в setupMainContent: %v", err)
		return err
	}
	if g.MainContent == nil {
		log.Println("RunGui: g.MainContent равен nil после setupMainContent")
		return fmt.Errorf("MainContent не инициализирован")
	}
	log.Println("RunGui: Установка содержимого окна")
	g.Window.SetContent(g.MainContent)
	g.Window.Resize(fyne.NewSize(800, 600))

	// Запуск обработчика обновлений GUI в главном потоке
	go g.runUpdateHandler()

	g.updateTime()
	log.Println("RunGui: Отправка сигнала ready и запуск приложения")
	close(ready)
	g.Window.ShowAndRun()
	return nil
}

func (g *Gui) runUpdateHandler() {
	for update := range g.updateChan {
		update()
	}
}

func (g *Gui) ShowError(err error) {
	log.Printf("ShowError: Отображение ошибки: %v", err)
	g.updateChan <- func() {
		dialog.ShowError(err, g.Window)
	}
}

func (g *Gui) setupTopSection() error {
	log.Println("setupTopSection: Начало настройки верхней секции")
	g.TopSection.Logo = Logo
	g.TopSection.SupportNumber = "8-800-555-35-35"
	g.TopSection.KazsNumber = "KAZS-123"
	g.TopSection.Timezone = "Asia/Almaty"

	var img image.Image
	imgData, err := base64.StdEncoding.DecodeString(g.TopSection.Logo)
	if err != nil {
		log.Printf("setupTopSection: Ошибка декодирования логотипа: %v", err)
	} else {
		img, _, err = image.Decode(bytes.NewReader(imgData))
		if err != nil {
			log.Printf("setupTopSection: Ошибка создания изображения логотипа: %v", err)
		}
	}

	if img == nil {
		log.Println("setupTopSection: Логотип не загружен, используется пустое изображение")
		g.TopSection.LogoLable = canvas.NewImageFromImage(nil)
	} else {
		g.TopSection.LogoLable = canvas.NewImageFromImage(img)
	}
	g.TopSection.LogoLable.SetMinSize(fyne.NewSize(32, 32))
	g.TopSection.TimeLable = canvas.NewText("00:00:00", nil)
	g.TopSection.DateLable = canvas.NewText("01.01.1970", nil)
	g.TopSection.SupportNumberLable = canvas.NewText(g.TopSection.SupportNumber, nil)
	g.TopSection.KazsNumberLable = canvas.NewText(g.TopSection.KazsNumber, nil)
	g.TopSection.Content = container.NewHBox(
		g.TopSection.LogoLable,
		g.TopSection.TimeLable,
		g.TopSection.DateLable,
		g.TopSection.SupportNumberLable,
		g.TopSection.KazsNumberLable,
	)
	log.Println("setupTopSection: Верхняя секция настроена")
	return nil
}

func (g *Gui) setupMainContent() error {
	log.Println("setupMainContent: Начало настройки главного контента")
	g.LeftSection.Content = container.NewVBox(
		widget.NewButton("Заправить", func() {
			g.FuelGiveScreen()
		}),
		widget.NewButton("Слить", func() {
			g.FuelGetScreen()
		}),
	)
	g.RightSection.Content = container.NewVBox(
		widget.NewLabel("История операций"),
		widget.NewLabel("Здесь будет список операций"),
	)
	if g.TopSection.Content == nil {
		log.Println("setupMainContent: TopSection.Content равен nil, создание пустого контейнера")
		g.TopSection.Content = container.NewHBox()
	}
	g.MainContent = container.NewBorder(
		g.TopSection.Content,
		g.BottomSection,
		g.LeftSection.Content,
		g.RightSection.Content,
	)
	if g.LeftSection.Content == nil || g.RightSection.Content == nil || g.TopSection.Content == nil || g.BottomSection == nil {
		log.Printf("setupMainContent: Один из контейнеров равен nil: LeftSection.Content=%v, RightSection.Content=%v, TopSection.Content=%v, BottomSection=%v",
			g.LeftSection.Content, g.RightSection.Content, g.TopSection.Content, g.BottomSection)
		return fmt.Errorf("один из контейнеров не инициализирован")
	}
	log.Println("setupMainContent: Главный контент настроен")
	return nil
}

func (g *Gui) updateTime() {
	log.Println("updateTime: Запуск обновления времени")
	ticker := time.NewTicker(time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	g.TopSection.Timer = cancel

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("updateTime: Остановка обновления времени")
				ticker.Stop()
				return
			case t := <-ticker.C:
				g.updateChan <- func() {
					g.TopSection.TimeLable.Text = t.In(time.FixedZone("Asia/Almaty", 6*3600)).Format("15:04:05")
					g.TopSection.DateLable.Text = t.In(time.FixedZone("Asia/Almaty", 6*3600)).Format("02.01.2006")
					g.TopSection.TimeLable.Refresh()
					g.TopSection.DateLable.Refresh()
				}
			}
		}
	}()
}

func (g *Gui) FuelGiveScreen() {
	log.Println("FuelGiveScreen: Отображение экрана заправки")
	progress := widget.NewProgressBar()
	ctx, cancel := context.WithCancel(context.Background())
	g.LeftSection.ActiveProcess = cancel
	cancelButton := widget.NewButton("Отмена", func() {
		cancel()
		log.Println("FuelGiveScreen: Операция заправки отменена")
		g.updateChan <- func() {
			dialog.ShowInformation("Отмена", "Заправка отменена", g.Window)
		}
	})
	content := container.NewVBox(
		widget.NewLabel("Заправка топлива"),
		progress,
		cancelButton,
	)
	g.LeftSection.ActiveDialog = content
	dialog.ShowCustom("Заправка", "Закрыть", content, g.Window)
	go g.simulateFuelOperation(progress, ctx, cancel, "fill")
}

func (g *Gui) FuelGetScreen() {
	log.Println("FuelGetScreen: Отображение экрана слива")
	progress := widget.NewProgressBar()
	ctx, cancel := context.WithCancel(context.Background())
	g.RightSection.ActiveProcess = cancel
	cancelButton := widget.NewButton("Отмена", func() {
		cancel()
		log.Println("FuelGetScreen: Операция слива отменена")
		g.updateChan <- func() {
			dialog.ShowInformation("Отмена", "Слив отменён", g.Window)
		}
	})
	content := container.NewVBox(
		widget.NewLabel("Слив топлива"),
		progress,
		cancelButton,
	)
	g.RightSection.ActiveDialog = content
	dialog.ShowCustom("Слив", "Закрыть", content, g.Window)
	go g.simulateFuelOperation(progress, ctx, cancel, "drain")
}

func (g *Gui) simulateFuelOperation(progress *widget.ProgressBar, ctx context.Context, cancel context.CancelFunc, action string) {
	log.Printf("simulateFuelOperation: Симуляция операции %s", action)
	for i := 0; i <= 100; i++ {
		select {
		case <-time.After(100 * time.Millisecond):
			g.updateChan <- func() {
				progress.SetValue(float64(i) / 100)
			}
		case <-ctx.Done():
			log.Println("simulateFuelOperation: Операция отменена")
			g.updateChan <- func() {
				dialog.ShowInformation("Отмена", fmt.Sprintf("Операция %s отменена", action), g.Window)
			}
			return
		}
	}
	log.Printf("simulateFuelOperation: Операция %s завершена", action)
	g.updateChan <- func() {
		dialog.ShowInformation("Успех", fmt.Sprintf("Операция %s завершена", action), g.Window)
	}
}
