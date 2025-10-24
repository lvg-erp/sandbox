package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

type Product struct {
	Name  string
	Price string
}

func main() {
	url := "https://www.tmktools.ru/catalog/sadovaya-tekhnika/snegouborshchiki/akkumulyatornye/"
	products, err := parseTMKTools(url)
	if err != nil {
		log.Fatalf("Ошибка: %v", err)
	}

	fmt.Printf("Найдено товаров: %d\n", len(products))
	if len(products) > 0 {
		for i, p := range products[:5] {
			fmt.Printf("%d. %s — %s\n", i+1, p.Name, p.Price)
		}
	} else {
		fmt.Println("Нет данных по указанной странице")
	}
}

func parseTMKTools(url string) ([]Product, error) {
	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Создаем контекст для chromedp с User-Agent
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()
	allocCtx, allocCancel = chromedp.NewContext(allocCtx)
	defer allocCancel()

	var nodes []struct {
		Name  string `json:"name"`
		Price string `json:"price"`
	}
	var htmlContent string
	var screenshot []byte

	err := chromedp.Run(allocCtx,
		chromedp.Navigate(url),
		// Ждем появления контейнера товаров
		chromedp.WaitVisible(".product-card", chromedp.ByQuery),
		// Дополнительная задержка
		chromedp.Sleep(10*time.Second),
		// Точные селекторы
		chromedp.Evaluate(`Array.from(document.querySelectorAll('.product-card')).map(item => ({
			name: item.querySelector('.product-card__title')?.textContent.trim() || '',
			price: item.querySelector('.product-price__current')?.textContent.trim() || ''
		})).filter(item => item.name && item.price)`, &nodes),
		// Получаем HTML и скриншот для отладки
		chromedp.OuterHTML("html", &htmlContent),
		chromedp.Screenshot("body", &screenshot, chromedp.ByQuery),
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка chromedp: %v", err)
	}

	// Сохраняем скриншот
	if err := os.WriteFile("screenshot.png", screenshot, 0644); err != nil {
		log.Printf("Ошибка сохранения скриншота: %v", err)
	}

	// Отладка: выводим первые 2000 символов HTML, если товары не найдены
	if len(nodes) == 0 {
		fmt.Printf("Отладка: первые 2000 символов HTML:\n%s\n", htmlContent[:min(2000, len(htmlContent))])
	}

	var products []Product
	for _, node := range nodes {
		products = append(products, Product{Name: node.Name, Price: node.Price})
	}

	return products, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
