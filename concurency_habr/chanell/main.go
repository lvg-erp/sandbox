package main

import (
	"fmt"
	"time"
)

func downloadFile(fileName string, done chan bool) {
	fmt.Printf("Starting download of file %s\n", fileName)
	time.Sleep(2 * time.Second)
	fmt.Printf("Finish download of file %s\n", fileName)

	done <- true
}

func sender(ch chan string, done chan bool) {
	for i := 0; i <= 2; i++ {
		ch <- fmt.Sprintf("message %d", i)
		time.Sleep(100 * time.Millisecond)
	}
	close(ch)
	done <- true

}

func receiver(ch chan string, done chan bool) {
	for msg := range ch {
		fmt.Println("Received", msg)
	}

	done <- true
}

func main() {
	fmt.Println("Starting downloading....")
	startTime := time.Now()
	//done := make(chan bool)
	//go downloadFile("./file1.txt", done)
	//go downloadFile("./file2.txt", done)
	//go downloadFile("./file3.txt", done)
	//
	//for i := 0; i < 3; i++ {
	//	<-done // Получаем сигнал от каждой завершенной горутины
	//}
	//ОБЩЕНИЕ МЕЖДУ ГОРУТИНАМИ
	//	маин инициализирует 3 канала
	//	ch - для сообщений
	//	senderDone - сигнал о завершении
	//	receiverDone - сигнал о завершении
	//	маин запускает 2 горутины
	//	sender
	//	receiver
	//	Маин блокируется и ждет сигналов о завершении
	//	Первое сообщение (t=1ms)
	//	sender отправляет "message 1" в канал ch.
	//		receiver просыпается и обрабатывает сообщение:
	//Выводит: "Received: message 1".
	//	Отправитель засыпает на 100 мс.
	//		Второе сообщение (t=101ms)
	//	sender просыпается и отправляет "message 2" в канал ch.
	//		receiver обрабатывает сообщение:
	//Выводит: "Received: message 2".
	//	Отправитель снова засыпает на 100 мс.
	//		Третье сообщение (t=201ms)
	//	sender просыпается и отправляет "message 3" в канал ch.
	//		receiver обрабатывает сообщение:
	//Выводит: "Received: message 3".
	//	Отправитель засыпает в последний раз.
	//		Закрытие канала (t=301ms)
	//	Отправитель завершает сон и закрывает канал ch.
	//		Отправитель отправляет сигнал true в канал senderDone, указывая на завершение работы.
	//		Получатель обнаруживает, что канал ch закрыт.
	//		Получатель выходит из цикла for-range.
	//	Завершение (t=302-303ms)
	//	маин получает сигнал от senderDone и прекращает ожидание.
	//		маин блокируется до сигнала от receiverDone.
	//		Получатель отправляет сигнал о завершении в канал receiverDone.
	//		маин получает сигнал и выводит:
	//	"All operations completed!".
	//		Программа завершается.
	//ch := make(chan string)
	//senderDone := make(chan bool)
	//receiverDone := make(chan bool)
	//
	//go sender(ch, senderDone)
	//go receiver(ch, receiverDone)
	//
	//<-senderDone
	//<-receiverDone
	// БУФЕРИЗИРОВАННЫЕ КАНАЛЫ
	ch := make(chan string, 2)
	ch <- "first"
	fmt.Println("Send message 1")
	ch <- "second"
	fmt.Println("Send message 2")
	// При попытке отправить третье сообщение будет блок
	//ch <- "third" // разкомментируй
	fmt.Println(<-ch)
	fmt.Printf(<-ch)
	elapsedTime := time.Since(startTime)
	fmt.Printf("All downloads completed! %s\n", elapsedTime)
}
