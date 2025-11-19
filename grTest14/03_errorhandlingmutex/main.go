package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type WordCount struct {
	FileCount int32
	Words     map[string]int
	Errors    map[string]int
}

func main() {

}

// метод чтения файла и подсчета слов
func wordCountFromFile(ctx context.Context, path string, results chan<- WordCount, wg *sync.WaitGroup) {
	defer wg.Done()
	select {
	case <-ctx.Done():
		return
	default:
	}
	// возможно нужна будет задержка
	data, err := os.ReadFile(path)
	if err != nil {
		we := WordCount{
			FileCount: 1,
			Words:     make(map[string]int),
			Errors:    map[string]int{err.Error(): 1},
		}
		results <- we
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

	results <- WordCount{
		FileCount: 1,
		Words:     words,
		Errors:    map[string]int{},
	}

}

// метода обхода файлов в указанной директории
func walkFiles(ctx context.Context, path string, results chan<- WordCount, wg *sync.WaitGroup) {
	defer wg.Done()
	select {
	case <-ctx.Done():
		return
	default:
	}

	entities, err := os.ReadDir(path)
	if err != nil {
		log.Printf("Ошибка чтения директории: %s: %v\n", path, err)
		return
	}

	for _, entity := range entities {
		select {
		case <-ctx.Done():
			return
		default:
		}
		fullPath := filepath.Join(path, entity.Name())
		if entity.IsDir() {
			wg.Add(1)
			go walkFiles(ctx, fullPath, results, wg)
		} else {
			wg.Add(1)
			go wordCountFromFile(ctx, fullPath, results, wg)
		}
	}

}
