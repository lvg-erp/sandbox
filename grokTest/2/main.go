package main

import "fmt"

//Обратный порядок строк:
//Функция reverseStrings принимает слайс строк и возвращает новый слайс в обратном порядке, не изменяя исходный.
//Пример: ["hello", "world", "go"] → ["go", "world", "hello"].
//Подсказка: Выдели новый слайс с make([]string, len(input)) и заполни его в обратном порядке.

func reverseStrings(in []string) []string {
	//out := make([]string, 0, len(in))
	out := make([]string, len(in))

	//for i := len(in) - 1; i >= 0; i-- {
	//	out = append(out, in[i])
	//}

	for i := 0; i < len(in); i++ {
		out[i] = in[len(in)-1-i]
	}

	return out
}

func main() {

	in := []string{"hello", "world", "go"}
	out := reverseStrings(in)

	fmt.Println(out)
	fmt.Println(in)

}
