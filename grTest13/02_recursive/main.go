package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Result struct {
	Count int64
	Size  int64
}

func walkDir(ctx context.Context, path string, threshold int64, results chan<- Result, wg *sync.WaitGroup) {
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
		// директория, запускаем в рекурсию
		if entry.IsDir() {
			wg.Add(1)
			go walkDir(ctx, fullPath, threshold, results, wg)
		} else {
			info, err := entry.Info()
			if err != nil {
				log.Printf("Ошибка получения информации о файле %s: %v", fullPath, err)
			}
			if info.Size() > threshold {
				results <- Result{
					Count: 1,
					Size:  info.Size(),
				}
			}
		}

	}
}

func main() {
	root := "/home/vladimir" // Замените на нужный путь, например, "/home/user"
	threshold := int64(1024) // Порог размера файла в байтах (например, 1 КБ)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results := make(chan Result)
	var wg sync.WaitGroup

	wg.Add(1)
	go walkDir(ctx, root, threshold, results, &wg)

	go func() {
		wg.Wait()
		close(results)
	}()

	var total Result

	for res := range results {
		total.Count += res.Count
		total.Size += res.Size
	}

	//fmt.Printf("Найдено файлов %d , Общий размер %d байт \n", total.Count, total.Size)
	fmt.Printf("Найдено файлов: %d, Общий размер: %.2f ГБ\n", total.Count, float64(total.Size)/1_073_741_824)
}
