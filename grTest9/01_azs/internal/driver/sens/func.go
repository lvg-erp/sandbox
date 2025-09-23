package sens

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// SENSCalculateCRC Вычисление контрольной суммы
func SENSCalculateCRC(data []byte) byte {
	var crc byte // Инициализируется нулем
	for i, b := range data {
		if i == 0 {
			continue
		}
		crc += b
	}
	return crc
}

// AreSlicesEqual Сравнение слайсов
func AreSlicesEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// StringToBytes Конвертация строки в байт
func StringToBytes(str string) byte {
	// Удаляем префикс 0x если он есть
	if len(str) > 2 && strings.ToLower(str[:2]) == "0x" {
		str = str[2:]
		// Парсим как шестнадцатеричное число
		num, err := strconv.ParseUint(str, 16, 8)
		if err != nil {
			return 0
		}
		return byte(num)
	}

	// Пробуем парсить как десятичное число
	num, err := strconv.ParseUint(str, 10, 8)
	if err != nil {
		return 0
	}
	return byte(num)
}

// FileExists Проверка есть ли файл
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true // файл существует
	}
	if os.IsNotExist(err) {
		return false // файл не существует
	}
	return false // произошла другая ошибка (например, нет прав доступа)
}

// ByteToHexString Конвертация байта в строку (0x01 -> "01")
func ByteToHexString(b byte) string {
	return fmt.Sprintf("%02X", b)
}

// ByteToHexStringSimple Конвертация байта в строку (0x01 -> "0x01")
func ByteToHexStringSimple(b byte) string {
	return fmt.Sprintf("0x%02X", b)
}

// Convert24BytesToFloat32 Конвертация 3 байт в float32
func Convert24BytesToFloat32(data []byte, byteOrder binary.ByteOrder) (float32, error) {
	if len(data) != 3 {
		return 0, fmt.Errorf("data length is %d, expected 3", len(data))
	}

	// Проверка специального значения (-1)
	if data[0] == 0xFF && data[1] == 0xFF && data[2] == 0xFF {
		return -1.0, nil
	}

	var bits uint32
	if byteOrder == binary.LittleEndian {
		bits = uint32(data[0])<<8 | uint32(data[1])<<16 | uint32(data[2])<<24
	} else {
		bits = uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8
	}

	// Конвертируем биты в float32
	floatValue := math.Float32frombits(bits)

	return floatValue, nil
}

// SaveConfig Сохранение конфига
func SaveConfig(filename string, config interface{}) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// ConvertFloat32To24Bytes Конвертация float32 в 3 байта
func ConvertFloat32To24Bytes(value float32, byteOrder binary.ByteOrder) ([]byte, error) {
	bits := math.Float32bits(value)

	buf := make([]byte, 3)
	if byteOrder == binary.LittleEndian {
		buf[0] = byte(bits >> 8)
		buf[1] = byte(bits >> 16)
		buf[2] = byte(bits >> 24)
	} else {
		buf[0] = byte(bits >> 24)
		buf[1] = byte(bits >> 16)
		buf[2] = byte(bits >> 8)
	}

	return buf, nil
}

// BytesSliceToString Метод преобразования среда байтов в строку формата "01 02 03 04"
func BytesSliceToString(b []byte) string {
	var buf bytes.Buffer
	for i, b := range b {
		if i > 0 {
			buf.WriteByte(' ')
		}
		fmt.Fprintf(&buf, "%02x", b)
	}

	result := buf.String()
	return result
}
