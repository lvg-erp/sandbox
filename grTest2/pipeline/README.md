Пайплайн обработки данных
Описание: Реализуйте конвейер (pipeline) с использованием каналов для обработки чисел.

Условия:
Создайте три этапа:
Генерация: горутина генерирует числа от 1 до 10.
Умножение: горутина умножает каждое число на 2.
Фильтрация: горутина пропускает только четные числа.
Результаты выводятся в консоль.
Подсказка:
Используйте отдельные каналы для связи между этапами.
Как закрыть каналы, чтобы сигнализировать о завершении этапов?
Как избежать блокировки (deadlock)?
`````Вариант решения (возможно не верный)`````
go func() {
wg.Wait()
close(ch)
close(ch1)
close(ch2)
}()

	for {
		select {
		case num, ok := <-ch:
			if !ok {
				fmt.Println("Channel closed")
				return
			}
			if num > 0 {
				fmt.Printf("Multiplie: %d\n", num)
			}
		case num1, ok := <-ch1:
			if !ok {
				fmt.Println("Channel closed")
				return
			}
			fmt.Printf("Addition: %d\n", num1)
		case num2, ok := <-ch2:
			if !ok {
				fmt.Println("Channel closed")
				return
			}
			if num2 > 0 {
				fmt.Printf("evenOnly: %d\n", num2)
			}
		case <-ctx.Done():
			fmt.Println("Context done")
			return
		}
	}



``````````````````````````````````````````