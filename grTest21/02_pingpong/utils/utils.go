package utils

import (
	"fmt"
	"strings"
	"sync/atomic"
)

func UpperWorker(in <-chan string, out chan<- string, done <-chan struct{}, counter *int64) {
	for {
		select {
		case <-done:
			return
		case s := <-in:
			// защищаем И отправку — иначе дедлок при завершении
			// TODO: mutex
			//Что происходит в момент завершения:
			//Главная горутина отсчитала 20 секунд и делает close(done).
			//Воркер upper уже прочитал сообщение из ping (ветка case s := <-in сработала) и сейчас стоит на строке out <- strings.ToUpper(s).
			//Он ждёт, пока кто-нибудь прочитает из pong.
			//Но воркеры lower тоже получили done и вышли из цикла — читать pong больше некому.
			//Воркер upper висит вечно на этой строке. select он уже прошёл, до следующей проверки <-done не дойдёт.
			//wg.Wait() в main никогда не вернётся → программа не завершится.
			//Ключевая мысль: проверка done в select защищает только чтение. А запись — отдельная блокирующая операция, её тоже нужно защитить.
			n := atomic.AddInt64(counter, 1) // один инкремент, запоминаем результат
			result := strings.ToUpper(s)     // считаем один раз

			if n%1000 == 1 {
				fmt.Printf("[UPPER] %q -> %q\n", s, result)
			}
			select {
			case out <- strings.ToUpper(s):
			case <-done:
				return
			}
		}
	}
}

func LowerWorker(in <-chan string, out chan<- string, done <-chan struct{}, counter *int64) {
	for {
		select {
		case <-done:
			return
		case s := <-in:
			// защищаем И отправку — иначе дедлок при завершении
			// TODO: mutex
			//Что происходит в момент завершения:
			//Главная горутина отсчитала 20 секунд и делает close(done).
			//Воркер upper уже прочитал сообщение из ping (ветка case s := <-in сработала) и сейчас стоит на строке out <- strings.ToUpper(s).
			//Он ждёт, пока кто-нибудь прочитает из pong.
			//Но воркеры lower тоже получили done и вышли из цикла — читать pong больше некому.
			//Воркер upper висит вечно на этой строке. select он уже прошёл, до следующей проверки <-done не дойдёт.
			//wg.Wait() в main никогда не вернётся → программа не завершится.
			//Ключевая мысль: проверка done в select защищает только чтение. А запись — отдельная блокирующая операция, её тоже нужно защитить.
			n := atomic.AddInt64(counter, 1) // один инкремент, запоминаем результат
			result := strings.ToLower(s)     // считаем один раз

			if n%1000 == 0 {
				fmt.Printf("[LOWER] %q -> %q\n", s, result)
			}
			select {
			case out <- strings.ToLower(s):
			case <-done:
				return
			}
		}
	}
}
