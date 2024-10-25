package alg

import (
	"strings"
)

//Задана строка и массив числовых значений, числовые значения массива указывают на
//буквы строки которые нужно привести к прописным

func Capitalize(st string, arr []int) string {
	var buff strings.Builder
	if len(arr) <= 0 {
		return st
	}
	//buff.Grow()
	//r := []rune(s) //перевод строки в массив
	m := arrToMap(arr)

	for idx, ch := range st {
		if _, ok := m[idx]; ok {
			buff.WriteString(strings.ToUpper(string(ch)))
		} else {
			buff.WriteString(string(ch))
		}
	}

	return buff.String()

}

func arrToMap(arr []int) map[int]struct{} {

	m := make(map[int]struct{})
	for _, idx := range arr {
		if idx < 0 {
			continue
		}
		m[idx] = struct{}{}
	}

	return m
}
