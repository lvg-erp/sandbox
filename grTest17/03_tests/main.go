package main

import "fmt"

//func logArrayValues(label string, arr []int) {
//	fmt.Printf("%s (адрес: %p): %v\n", label, &arr, arr)
//}
//
//func modify(x []int) {
//	fmt.Printf("Внутри modify, начальный адрес массива x: %p\n", &x)
//	logArrayValues("Перед изменением", x)
//	x = append(x, 100)
//	fmt.Printf("После append, адрес массива x: %p\n", &x)
//	logArrayValues("После append", x)
//	x[0] = 999
//	logArrayValues("После изменения первого элемента", x)
//}
//
//func main() {
//	s := make([]int, 2, 3)
//	s[0], s[1] = 1, 2
//
//	logArrayValues("Изначальный s", s)
//
//	a := s[:2]
//	logArrayValues("a", a)
//
//	b := append(s, 3)
//	logArrayValues("b", b)
//
//	fmt.Println("До вызова modify(a):")
//	logArrayValues("s", s)
//	logArrayValues("a", a)
//	logArrayValues("b", b)
//
//	modify(a)
//
//	fmt.Println("После вызова modify(a):")
//	logArrayValues("s", s)
//	logArrayValues("a", a)
//	logArrayValues("b", b)
//
//	modify(b)
//	fmt.Println("После вызова modify(b):")
//	logArrayValues("s", s)
//	logArrayValues("a", a)
//	logArrayValues("b", b)
//}

func main() {
	s := make([]int, 2, 3)
	s[0], s[1] = 1, 2 // s{1,2,0}
	//a := s[:2]        // a{1, 2}
	//b := append(s, 3) // b{1,2,3}, пока хватает емкости s новый массив не создается
	// чтобы значения сохранялись необходимо сделать копии слайсов
	a := append([]int{}, s[:2]...)
	b := append([]int{}, s[:2]...)
	b = append(b, 3)
	//-------------------------------
	//fmt.Println("Тест: ", b)
	modify(&a) // a{999, 2, 100} // хватает емкости a и b
	modify(&b) // s заполнен полностью значит релоцируем новый массив b{999, 2, 100, 100}
	//c := modify(b)

	fmt.Println(s) // 999 2
	fmt.Println(a) // 999 2
	fmt.Println(b) // 999 2 100 100
}

// чтобы увидеть что изменился перелоцировался слайс необходимо передать указатель
// либо переписать метод который возращает значение //c := modify(b)
func modify(x *[]int) {
	*x = append(*x, 100)
	(*x)[0] = 999
	fmt.Println("Вызов в процедуре модификации(х):", *x)
}

//func modify(x []int) {
//	x = append(x, 100)
//	x[0] = 999
//	fmt.Println("Вызов в процедуре модификации(х):", x)
//}
