package main

import (
	"fmt"
	"incube/alg"
)

func main() {
	//c := alg.Gen(2, 3, 6)
	//out := alg.Sq(c)
	//
	//fmt.Println(<-out)
	//fmt.Println(<-out)
	//fmt.Println(<-out)
	//выводим квадрат
	for n := range alg.Sq(alg.Sq(alg.Gen(4, 6, 9))) {
		fmt.Println(n)
	}

}
