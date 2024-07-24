package alg

// преобразуем список в канал
func Gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()

	return out

}

// забирает числа из канала и возвращает новый канал, который отдает квадрат каждого полученного числа
func Sq(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()

	return out

}
