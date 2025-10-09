package main

import "fmt"

func main() {

	input := make(chan int, 10)
	output := make(chan int, 10)

	for i := 1; i < 11; i++ {
		input <- i
	}

	close(input)

	go func() {
		DoubleNumbers(input, output)
	}()

	for v := range output {
		fmt.Println(v)
	}

}

func DoubleNumbers(input <-chan int, output chan<- int) {

	for i := range input {
		output <- i * 2
	}
	close(output)

}
