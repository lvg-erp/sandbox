package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

//const workers = 3

type WordCount struct {
	FileCount int32
	Words     map[string]int
}

func main() {

	dir := "/home/vladimir/testgo/"
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var (
		results        = make(chan WordCount)
		wg             sync.WaitGroup
		total          WordCount
		totalFileCount int32
		done           = make(chan struct{})
	)
	total.Words = make(map[string]int)

	// Тикер для промежуточной статистики
	go func() {
		fmt.Println("Горутина тикера запущена") // Отладка
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("Итог (таймаут): файлов %d, слов %d\n", atomic.LoadInt32(&totalFileCount), len(total.Words))
				return
			case <-ticker.C:
				fmt.Println("Тикер сработал") //TODO:
				fmt.Printf("Промежуточный итог: файлов %d, слов %d\n", atomic.LoadInt32(&totalFileCount), len(total.Words))
			case <-done:
				fmt.Println("Получен сигнал done") //TODO:
				fmt.Printf("Итог (завершение): файлов %d, слов %d\n", atomic.LoadInt32(&totalFileCount), len(total.Words))
				return
			}
		}
	}()

	// Запуск обхода директории
	wg.Add(1)
	go walkFiles(ctx, dir, results, &wg)

	// Сбор результатов
	go func() {
		wg.Wait()
		fmt.Println("Все задачи завершены, закрываем results") // Отладка
		close(results)
	}()

	// Суммирование результатов
	for res := range results {
		atomic.AddInt32(&totalFileCount, res.FileCount)
		for word, count := range res.Words {
			total.Words[word] += count
		}
	}
	close(done)
	fmt.Printf("Итог: файлов %d, слов %d\n", totalFileCount, len(total.Words))

}

func walkFiles(ctx context.Context, path string, results chan<- WordCount, wg *sync.WaitGroup) {
	defer wg.Done()
	select {
	case <-ctx.Done():
		return
	default:
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Printf("Ошибка чтения директории %s: %v", path, err)
		return
	}
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return
		default:
		}
		fullPath := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			wg.Add(1)
			go walkFiles(ctx, fullPath, results, wg)
		} else if strings.HasSuffix(strings.ToLower(entry.Name()), ".txt") {
			wg.Add(1)
			go processFile(ctx, fullPath, results, wg)
		}
	}

}

func processFile(ctx context.Context, path string, results chan<- WordCount, wg *sync.WaitGroup) {
	defer wg.Done()
	select {
	case <-ctx.Done():
		return
	default:
	}
	// Увеличенная задержка для тестирования
	time.Sleep(3000 * time.Millisecond)
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Ошибка чтения %s: %v", path, err)
		return
	}
	words := make(map[string]int)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			word := strings.ToLower(scanner.Text())
			words[word]++
		}
	}
	fmt.Printf("Обработан файл %s: %d слов\n", path, len(words)) // Отладка
	results <- WordCount{FileCount: 1, Words: words}

}

//// получим файлы по указанному расширению
//func filterFilesByType(entries []fs.DirEntry, fileType string) []string {
//	var filteredFiles []string
//	for _, entry := range entries {
//		// Проверяем, что это файл
//		if !entry.IsDir() {
//			fileName := entry.Name()
//			// Проверяем, что расширение совпадает с искомым
//			if strings.HasSuffix(fileName, fileType) {
//				filteredFiles = append(filteredFiles, fileName)
//			}
//		}
//	}
//	return filteredFiles
//}
