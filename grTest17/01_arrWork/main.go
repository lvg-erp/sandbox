package main

import (
	"fmt"
)

type S struct {
	v int
}

func (s S) Value() int {
	return s.v
}

type I interface {
	Value() int
}

func mutate(i I) {
	// Мы думаем, что меняем реальный объект...
	//if s, ok := i.(S); // если так то значение останется которое было
	if s, ok := i.(*S); ok {
		s.v = 100
	}
}

func main() {
	s := S{v: 10}
	//var i I = s // если так то значение останется которое было
	var i I = &s

	mutate(i)

	fmt.Println(s.v)
	fmt.Println(i.Value())
}

//
//func main() {
//	//a := []int{10, 20, 30, 40}
//	a := []int{10, 20, 30}
//	b := a[:2]
//	c := append(b, 99)
//
//	a[1] = 777
//
//	c = append(c, 555)
//
//	fmt.Println("a: ", a)
//	fmt.Println("b: ", b)
//	fmt.Println("c: ", c)
//
//}
