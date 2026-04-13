package main

import "fmt"

func main() {
	fillData(nil, nil)
}

func fillData(s []int, m map[string]int) {
	// NIL APPEND - РАБОТАЕТ КОРРЕКТНО!!!!
	// МАПА ДОЛЖНА БЫТЬ ИНИЦИАЛИЗИРОВАННА!!!!!!!!!!!
	if m == nil {
		m = make(map[string]int)
	}

	m = make(map[string]int)
	s = append(s, 10, 20, 30)

	m["count"] = len(s)
	m["total"] = 60

	fmt.Println("Срез: ", s)
	fmt.Println("Мапа: ", m)
}
