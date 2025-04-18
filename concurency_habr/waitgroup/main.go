package main

import (
	"fmt"
	"sync"
	"time"
)

func downloadFile(fileName string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Downloading file: %s\n", fileName)
	// Эмулируем задержку
	time.Sleep(2 * time.Second)
	fmt.Printf("Finished downloading file: %s\n", fileName)
}

func main() {
	//Выполнение заняло 6 секунд, каждый вызов downloadFile должен завершиться прежде чем начнется следующий.
	fmt.Println("Start downloading...")
	startTime := time.Now()

	//downloadFile("./file1.txt")
	//downloadFile("./file2.txt")
	//downloadFile("./file3.txt")
	//// Мы можем снизить это время, давай внесем изменения в код, добавим горутины:
	// важно: перед вызовом функции используется `go`
	// Но ьайн завершилась до того как отработали горутины
	// майн - это тоже горутина

	//go downloadFile("./file1.txt")
	//go downloadFile("./file2.txt")
	//go downloadFile("./file3.txt")

	// будем использовать вайтгруппы и каналы
	// не будем использовать костыль в виде задержки
	//Как это работает
	//WaitGroup инициализирует внутренний счетчик
	//wg.Add(n) увеличивает счетчик на n
	//wg.Done() уменьшает счетчик на 1
	//wg.Wait() блокирует main до тех пор пока счетчик не станет 0
	var wg sync.WaitGroup

	//добавляем три потока
	wg.Add(3)

	go downloadFile("./file1.txt", &wg)
	go downloadFile("./file2.txt", &wg)
	go downloadFile("./file3.txt", &wg)
	//ждем выполнения горутин
	wg.Wait()

	elapsedTime := time.Since(startTime)
	fmt.Printf("All downloads completed! Time elapsed: %s\n", elapsedTime)

}
