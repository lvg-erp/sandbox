package main

import (
	"fmt"
	"golang.org/x/net/html"
	"strings"
	"sync"
	"unicode"
)

func main() {

	urls := []string{
		"https://example.com/page1",
		"https://example.com/page2",
		"https://example.com/page3",
		"https://example.com/page4",
		"https://example.com/page5",
	}

	const stream = 3
	wg := sync.WaitGroup{}
	chResult := make(chan string, len(urls))
	cSize := (len(urls) + stream - 1) / stream
	for i := 0; i < stream; i++ {

		start := cSize * i
		end := start + cSize
		if end > len(urls) {
			end = len(urls)
		}
		s, e := start, end
		wg.Go(func() {
			chunk := urls[s:e]
			// получим текст страницы
			//fmt.Println(chunk)
			for _, url := range chunk {
				contentPage := getPageContent(url)
				textPage := extractText(contentPage)
				countLetter := countingLettersToString(textPage)
				chResult <- fmt.Sprintf("Страница %s, содержит символов %d", url, countLetter)
			}

		})
	}

	go func() {
		wg.Wait()
		close(chResult)
	}()

	for c := range chResult {
		fmt.Println(c)
	}

}

// генерируем страницы хтмл
func getPageContent(url string) string {
	switch url {
	case "https://example.com/page1":
		return `<html><body>Это <b>содержимое</b> страницы 1</body></html>`
	case "https://example.com/page2":
		return `<html><body>Другая страница с <a href="#">текстом</a></body></html>`
	case "https://example.com/page3":
		return `<html><body>Еще одна страница <div>для тестирования</div></body></html>`
	case "https://example.com/page4":
		return `<html><body>Еще одна горутины страница <div>для тестирования</div></body></html>`
	case "https://example.com/page5":
		return `<html><body>Еще одна горутины страница <div>для тестирования</div></body></html>`
	default:
		return `<html><body>Страница по умолчанию</body></html>`
	}
}

// парсим на наличие знаков хтмл используя библиотеку
// извлекаем только текст, без знаков разметки
// golang.org/x/net/html
func extractText(htmlStr string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return ""
	}

	var buf strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			// Добавляем текст, если он не пустой
			text := strings.TrimSpace(n.Data)
			if len(text) > 0 {
				buf.WriteString(text + " ")
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return strings.TrimSpace(buf.String())
}

func countingLettersToString(s string) int {
	count := 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			count++
		}
	}
	return count
}
