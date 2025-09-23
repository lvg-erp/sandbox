package gui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"
)

var (
	Green color.Color = color.RGBA{1, 111, 84, 255}
	Gray  color.Color = color.RGBA{84, 84, 84, 255}
	Black color.Color = color.RGBA{47, 47, 47, 255}
)

func ConvertUnixToString(unitTime int64) string {
	t := time.Unix(unitTime, 0)

	layout := "02.01.2006 15:04"

	formattedTime := t.Format(layout)

	return formattedTime
}

func ConvertFloat32ToStringShort(number float32) string {
	// 1. Преобразуем число в строку с двумя знаками после запятой.
	s := fmt.Sprintf("%.2f", number)

	// 2. Находим позицию десятичной точки.
	decimalIndex := strings.Index(s, ".")

	// 3. Разделяем строку на целую и десятичную части.
	integerPart := s
	decimalPart := ""
	if decimalIndex != -1 {
		integerPart = s[:decimalIndex]
		decimalPart = s[decimalIndex:]
	}

	// 4. Обрабатываем знак минуса, если число отрицательное.
	isNegative := false
	if strings.HasPrefix(integerPart, "-") {
		isNegative = true
		integerPart = integerPart[1:]
	}

	// 5. Форматируем целую часть, вставляя пробелы каждые три цифры справа налево.
	formattedInteger := ""
	n := len(integerPart)
	for i := n - 1; i >= 0; i-- {
		formattedInteger = string(integerPart[i]) + formattedInteger
		if i > 0 && (n-i)%3 == 0 {
			formattedInteger = " " + formattedInteger
		}
	}

	// 6. Добавляем обратно знак минуса, если число было отрицательным.
	if isNegative {
		formattedInteger = "-" + formattedInteger
	}

	_ = decimalPart
	// 7. Объединяем отформатированную целую часть и десятичную часть.
	return formattedInteger
}

func ConvertFloat32ToStringFull(number float32) string {
	// 1. Преобразуем число в строку с двумя знаками после запятой.
	s := fmt.Sprintf("%.2f", number)

	// 2. Находим позицию десятичной точки.
	decimalIndex := strings.Index(s, ".")

	// 3. Разделяем строку на целую и десятичную части.
	integerPart := s
	decimalPart := ""
	if decimalIndex != -1 {
		integerPart = s[:decimalIndex]
		decimalPart = s[decimalIndex:]
	}

	// 4. Обрабатываем знак минуса, если число отрицательное.
	isNegative := false
	if strings.HasPrefix(integerPart, "-") {
		isNegative = true
		integerPart = integerPart[1:]
	}

	// 5. Форматируем целую часть, вставляя пробелы каждые три цифры справа налево.
	formattedInteger := ""
	n := len(integerPart)
	for i := n - 1; i >= 0; i-- {
		formattedInteger = string(integerPart[i]) + formattedInteger
		if i > 0 && (n-i)%3 == 0 {
			formattedInteger = " " + formattedInteger
		}
	}

	// 6. Добавляем обратно знак минуса, если число было отрицательным.
	if isNegative {
		formattedInteger = "-" + formattedInteger
	}

	// 7. Объединяем отформатированную целую часть и десятичную часть.
	return formattedInteger + decimalPart
}

func FormatRussianDate(t time.Time) string {
	weekdays := []string{"Воскресенье", "Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота"}
	months := []string{"января", "февраля", "марта", "апреля", "мая", "июня", "июля", "августа", "сентября", "октября", "ноября", "декабря"}

	weekday := weekdays[t.Weekday()]
	day := t.Format("2")
	month := months[t.Month()-1] // Month() returns 1-12, slice is 0-indexed
	year := t.Format("2006")

	return fmt.Sprintf("%s, %s %s, %s", weekday, day, month, year)
}

func CalculateOffsetSecondsFromString(offsetString string) (int, error) {
	parts := strings.Split(offsetString, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("format error. Expected \"+/-ЧЧ:ММ\"")
	}

	sign := 1
	hoursStr := parts[0]
	if strings.HasPrefix(hoursStr, "-") {
		sign = -1
		hoursStr = strings.TrimPrefix(hoursStr, "-")
	} else if strings.HasPrefix(hoursStr, "+") {
		hoursStr = strings.TrimPrefix(hoursStr, "+")
	}

	hours, err := strconv.Atoi(hoursStr)
	if err != nil {
		return 0, fmt.Errorf("parse hours err: %w", err)
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("parse minute err: %w", err)
	}

	offsetSeconds := (hours*3600 + minutes*60) * sign
	return offsetSeconds, nil
}
