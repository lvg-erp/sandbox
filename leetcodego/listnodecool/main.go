package main

import (
	"fmt"
	"listnodecool/alg"
	"log"
)

func main() {

	data := "12345"

	dataToList, err := alg.NewList(data)
	if err != nil {
		log.Fatal(err)
	}
	dataToTest := alg.NewSolution(dataToList)
	fmt.Println(*dataToTest)
	result := alg.SwapPairs(dataToList)
	solution := alg.NewSolution(result)
	fmt.Println(*solution)

}
