package trk

import (
	"encoding/json"
	"fmt"
	"fuelazs/internal/driver/topaz/trk/config"
	"os"
	"strconv"
	"strings"
)

func ConvertNetworkAddress(b []byte) (netAddr string, workStatus string) {
	if len(b) < 9 {
		return "", ""
	}

	netAddr = strconv.Itoa(parseDigit(b[2])) + strconv.Itoa(parseDigit(b[4])) + strconv.Itoa(parseDigit(b[6]))

	workStatus = strconv.Itoa(parseDigit(b[8]))

	return netAddr, workStatus
}

func RemoveLeadingZeros(s string) string {
	firstSigIdx := len(s)

	for i := 0; i < len(s); i++ {
		if s[i] != '0' {
			firstSigIdx = i
			break
		}
	}

	if firstSigIdx == len(s) {
		if len(s) > 0 {
			return "0"
		}
		return ""
	}

	return s[firstSigIdx:]
}

func StringToBytesSimple(s string) []byte {
	return []byte(s)
}

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

func ByteToHexString(b byte) string {
	return fmt.Sprintf("%02X", b)
}

func ByteToHexStringSimple(b byte) string {
	return fmt.Sprintf("0x%02X", b)
}

func SaveConfig(filename string, config interface{}) error {
	// Сериализуем с красивым форматированием
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Записываем с правами 0644 (-rw-r--r--)
	return os.WriteFile(filename, data, 0644)
}

func InvertByte(b byte) byte {
	return b ^ 0x7F
}

func InvertByte2(b byte) byte {
	return ^b
}

func TOPAZCalculateChecksum(data []byte) byte {
	if len(data) < 4 {
		return 0 // Ошибка: слишком короткий пакет
	}

	var checksum byte = 0

	// XOR только для байтов с нечётными индексами (после STX)
	for i := 2; i < len(data)-1; i += 2 {
		checksum ^= data[i]
	}

	// Добавляем 0x40
	checksum |= 0x40
	return checksum
}

func BytesToFloat(data []byte) (float32, error) {

	var value float32

	if len(data) == 17 {
		// Преобразуем каждый байт в цифру (ASCII или прямое значение)
		hundreds := parseDigit(data[4])    // сотни
		tens := parseDigit(data[6])        // десятки
		ones := parseDigit(data[8])        // единицы
		tenths := parseDigit(data[10])     // десятые
		hundredths := parseDigit(data[12]) // сотые

		// Собираем число
		value = float32(hundreds)*100 +
			float32(tens)*10 +
			float32(ones) +
			float32(tenths)/10 +
			float32(hundredths)/100

	} else if len(data) == 37 {
		hundreds := parseDigit(data[2])
		tens := parseDigit(data[4])
		ones := parseDigit(data[6])
		tenths := parseDigit(data[8])
		hundredths := parseDigit(data[10])

		// Собираем число
		value = float32(hundreds)*100 +
			float32(tens)*10 +
			float32(ones) +
			float32(tenths)/10 +
			float32(hundredths)/100

	} else {
		return 0, fmt.Errorf("data length is %d, expected 17 or 34", len(data))
	}

	return value, nil
}

func ConvertTRKID(b []byte) (string, error) {
	if len(b) < 6 {
		return "", nil
	}

	var builder strings.Builder

	for i := 2; i < len(b)-3; i += 2 {
		value := parseDigit(b[i])
		builder.WriteString(strconv.Itoa(value))
	}

	return builder.String(), nil
}

func parseDigit(b byte) int {
	if b >= '0' && b <= '9' { // Если это ASCII-цифра (0x30..0x39)
		return int(b - '0')
	}
	return int(b) // Если байт уже содержит числовое значение (0x00..0x09)
}

func FloatToBytes(num float32) ([]byte, error) {
	if num < 0 || num >= 1000 {
		return nil, fmt.Errorf("число должно быть в диапазоне 0 <= num < 1000")
	}

	// Умножаем на 100, чтобы перенести две десятичные цифры в целую часть.
	scaled := int(num * 100)

	// Проверяем, что число не превысило 99999 после умножения.
	if scaled > 99999 {
		return nil, fmt.Errorf("число слишком велико после масштабирования")
	}

	// Форматируем как 5-значное число с ведущими нулями.
	str := fmt.Sprintf("%05d", scaled)

	// Преобразуем строку в ASCII-байты.
	bytes := []byte(str)
	return bytes, nil
}

func ByteToSingleDigitFloat32(b byte) float32 {
	lowerNibble := b & 0x0F
	if lowerNibble <= 9 {
		return float32(lowerNibble)
	}
	return 0 // или другое значение по умолчанию
}

func StringToSlice(s string) []byte {
	result := make([]byte, 0, len(s))

	for _, v := range s {
		digit := byte(v)
		result = append(result, digit)
	}

	return result
}

func getAddress(address string) byte {
	switch address {
	case "1":
		return 0x21
	case "2":
		return 0x22
	case "3":
		return 0x23
	case "4":
		return 0x24
	case "5":
		return 0x25
	case "6":
		return 0x26
	case "7":
		return 0x27
	case "8":
		return 0x28
	case "9":
		return 0x29
	}
	return 0
}

func TRKСonvertByteToFloat(data map[string][]byte) map[string]float32 {
	var result = make(map[string]float32, len(data))
	for key, value := range data {
		if len(value) == 0 {
			continue
		}
		switch key {
		case "0x3C":
			result[key] = 10*ByteToSingleDigitFloat32(value[0]) + 1*ByteToSingleDigitFloat32(value[1])
		case "0x45":
			result[key] = 10*ByteToSingleDigitFloat32(value[0]) + 1*ByteToSingleDigitFloat32(value[1])
		case "0x3D":
			result[key] = 10*ByteToSingleDigitFloat32(value[0]) + 1*ByteToSingleDigitFloat32(value[1])
		case "0x3F":
			result[key] = 10*ByteToSingleDigitFloat32(value[0]) + 1*ByteToSingleDigitFloat32(value[1])
		case "0x37":
			result[key] = 1*ByteToSingleDigitFloat32(value[0]) + 0.1*ByteToSingleDigitFloat32(value[1])
		case "0x57":
			result[key] = 10*ByteToSingleDigitFloat32(value[0]) + 1*ByteToSingleDigitFloat32(value[1])
		case "0x51":
			result[key] = 10*ByteToSingleDigitFloat32(value[0]) + 1*ByteToSingleDigitFloat32(value[1])
		case "0x5A":
			result[key] = 0.1*ByteToSingleDigitFloat32(value[0]) + 0.01*ByteToSingleDigitFloat32(value[1])
		case "0x5B":
			result[key] = 0.1*ByteToSingleDigitFloat32(value[0]) + 0.01*ByteToSingleDigitFloat32(value[1])
		case "0x35":
			result[key] = 10*ByteToSingleDigitFloat32(value[0]) + 1*ByteToSingleDigitFloat32(value[1]) + 0.1*ByteToSingleDigitFloat32(value[2])
		case "0x39":
			result[key] = 1*ByteToSingleDigitFloat32(value[0]) + 0.1*ByteToSingleDigitFloat32(value[1]) + 0.01*ByteToSingleDigitFloat32(value[2])
		case "0x3A":
			result[key] = 1*ByteToSingleDigitFloat32(value[0]) + 0.1*ByteToSingleDigitFloat32(value[1]) + 0.01*ByteToSingleDigitFloat32(value[2])
		case "0x3B":
			result[key] = 100*ByteToSingleDigitFloat32(value[0]) + 10*ByteToSingleDigitFloat32(value[1]) + 1*ByteToSingleDigitFloat32(value[2])
		case "0x44":
			result[key] = 10*ByteToSingleDigitFloat32(value[0]) + 1*ByteToSingleDigitFloat32(value[1]) + 0.1*ByteToSingleDigitFloat32(value[2])
		case "0x47":
			result[key] = 1*ByteToSingleDigitFloat32(value[0]) + 0.1*ByteToSingleDigitFloat32(value[1]) + 0.01*ByteToSingleDigitFloat32(value[2])
		case "0x55":
			if value[0] == 0 {
				result[key] = 10*ByteToSingleDigitFloat32(value[1]) + 1*ByteToSingleDigitFloat32(value[2])
			} else {
				result[key] = -10*ByteToSingleDigitFloat32(value[1]) - 1*ByteToSingleDigitFloat32(value[2])
			}
		case "0x56":
			if value[0] == 0 {
				result[key] = 10*ByteToSingleDigitFloat32(value[1]) + 1*ByteToSingleDigitFloat32(value[2])
			} else {
				result[key] = -10*ByteToSingleDigitFloat32(value[1]) - 1*ByteToSingleDigitFloat32(value[2])
			}
		case "0x52":
			result[key] = 100*ByteToSingleDigitFloat32(value[0]) + 10*ByteToSingleDigitFloat32(value[1]) + 1*ByteToSingleDigitFloat32(value[2])
		case "0x48":
			result[key+"1"] = 100*ByteToSingleDigitFloat32(value[0]) + 10*ByteToSingleDigitFloat32(value[1]) + 1*ByteToSingleDigitFloat32(value[2])
			result[key+"2"] = 100*ByteToSingleDigitFloat32(value[3]) + 10*ByteToSingleDigitFloat32(value[4]) + 1*ByteToSingleDigitFloat32(value[5])
		default:
			result[key] = 1 * ByteToSingleDigitFloat32(value[0])
		}
	}

	return result
}

func TRKConvertFloatToByte(data map[string]float32) map[string][]byte {
	result := make(map[string][]byte, len(data))

	for key, value := range data {
		valueStr := float32ToString(value)
		switch key {
		case "0x33":
			result[key] = []byte{StringToBytesSimple(valueStr)[0]}
		case "0x36":
			result[key] = []byte{StringToBytesSimple(valueStr)[0]}
		case "0x38":
			result[key] = []byte{StringToBytesSimple(valueStr)[0]}
		case "0x3E":
			result[key] = []byte{StringToBytesSimple(valueStr)[0]}
		case "0x46":
			result[key] = []byte{StringToBytesSimple(valueStr)[0]}
		case "0x30":
			result[key] = []byte{StringToBytesSimple(valueStr)[0]}
		case "0x5D":
			result[key] = []byte{StringToBytesSimple(valueStr)[0]}
		case "0x3C":
			if value < 10 {
				result[key] = []byte{0x30, StringToBytesSimple(valueStr)[0]}
			} else {
				result[key] = []byte{StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[1]}
			}
		case "0x45":
			if value < 10 {
				result[key] = []byte{0x30, StringToBytesSimple(valueStr)[0]}
			} else {
				result[key] = []byte{StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[1]}
			}
		case "0x3D":
			if value < 10 {
				result[key] = []byte{0x30, StringToBytesSimple(valueStr)[0]}
			} else {
				result[key] = []byte{StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[1]}
			}
		case "0x3F":
			if value < 10 {
				result[key] = []byte{0x30, StringToBytesSimple(valueStr)[0]}
			} else {
				result[key] = []byte{StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[1]}
			}
		case "0x37":
			if value < 1 {
				result[key] = []byte{0x30, StringToBytesSimple(valueStr)[2]}
			} else {
				result[key] = []byte{StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[2]}
			}
		case "0x57":
			if value < 10 {
				result[key] = []byte{0x30, StringToBytesSimple(valueStr)[0]}
			} else {
				result[key] = []byte{StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[1]}
			}
		case "0x51":
			if value < 10 {
				result[key] = []byte{0x30, StringToBytesSimple(valueStr)[0]}
			} else {
				result[key] = []byte{StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[1]}
			}
		case "0x5A":
			switch value {
			case 0:
				result[key] = []byte{0x30, 0x30}
			case 0.99:
				result[key] = []byte{0x39, 0x39}
			case 0.01:
				result[key] = []byte{0x30, 0x31}
			case 0.98:
				result[key] = []byte{0x39, 0x38}
			}
		case "0x5B":
			result[key] = []byte{StringToBytesSimple(valueStr)[2], StringToBytesSimple(valueStr)[3]}
		case "0x35":
			if value < 10 {
				result[key] = []byte{0x30, StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[2]}
			} else {
				result[key] = []byte{StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[1], StringToBytesSimple(valueStr)[3]}
			}
		case "0x39":
			if value < 1 {
				result[key] = []byte{0x30, StringToBytesSimple(valueStr)[2], StringToBytesSimple(valueStr)[3]}
			} else {
				result[key] = []byte{StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[2], StringToBytesSimple(valueStr)[3]}
			}
		case "0x3A":
			if value < 1 {
				result[key] = []byte{0x30, StringToBytesSimple(valueStr)[2], StringToBytesSimple(valueStr)[3]}
			} else {
				result[key] = []byte{StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[2], StringToBytesSimple(valueStr)[3]}
			}
		case "0x3B":
			if value < 100 {
				if value < 10 {
					result[key] = []byte{0x30, 0x30, StringToBytesSimple(valueStr)[0]}
				} else {
					result[key] = []byte{0x30, StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[1]}
				}
			} else {
				result[key] = []byte{StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[1], StringToBytesSimple(valueStr)[2]}
			}
		case "0x44":
			if value < 10 {
				result[key] = []byte{0x30, StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[2]}
			} else {
				result[key] = []byte{StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[1], StringToBytesSimple(valueStr)[3]}
			}
		case "0x47":
			if value < 10 {
				result[key] = []byte{0x30, StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[2]}
			} else {
				result[key] = []byte{StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[1], StringToBytesSimple(valueStr)[3]}
			}
		case "0x55":
			switch value {
			case 0:
				result[key] = []byte{0x32, 0x30, 0x30}
			case 1:
				result[key] = []byte{0x30, 0x30, 0x31}
			}
		case "0x52":
			if value < 100 {
				if value < 10 {
					result[key] = []byte{0x30, 0x30, StringToBytesSimple(valueStr)[0]}
				} else {
					result[key] = []byte{0x30, StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[1]}
				}
			} else {
				result[key] = []byte{StringToBytesSimple(valueStr)[0], StringToBytesSimple(valueStr)[1], StringToBytesSimple(valueStr)[2]}
			}
		}
	}

	return result
}

func validateSettingsStructure(template *config.TRKConfig, data *config.TRKConfig) error {
	if template == nil || data == nil {
		return fmt.Errorf("cannot validate nil config")
	}
	// Проверка наличия всех ключей из шаблона в данных
	for key := range template.GeneralParameters {
		if _, exists := data.GeneralParameters[key]; !exists {
			return fmt.Errorf("missing setting parameter key: %s", key)
		}
		switch key {
		case "0x32":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 2 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x33":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 7 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x35":
			if data.GeneralParameters[key].Value < 0.4 && data.GeneralParameters[key].Value > 50 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x36":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 8 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x38":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 4 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x39":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 2 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x3A":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 2 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x3B":
			if data.GeneralParameters[key].Value < 3 && data.GeneralParameters[key].Value > 180 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x3C":
			if data.GeneralParameters[key].Value < 3 && data.GeneralParameters[key].Value > 75 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x3E":
			if data.GeneralParameters[key].Value < 2 && data.GeneralParameters[key].Value > 4 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x45":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 20 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x3D":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 20 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x44":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 10 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x46":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 9 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x47":
			if data.GeneralParameters[key].Value < 0.01 && data.GeneralParameters[key].Value > 5 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x3F":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 50 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x30":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 1 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x37":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 9 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x55":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 1 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x56":
			if data.GeneralParameters[key].Value < -20 && data.GeneralParameters[key].Value > 20 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x57":
			if data.GeneralParameters[key].Value < 3 && data.GeneralParameters[key].Value > 15 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x52":
			if data.GeneralParameters[key].Value < 3 && data.GeneralParameters[key].Value > 180 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x51":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 30 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x5A":
			if data.GeneralParameters[key].Value != 0 && data.GeneralParameters[key].Value != 0.99 && data.GeneralParameters[key].Value != 0.01 && data.GeneralParameters[key].Value != 0.98 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x5B":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 0.5 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x5D":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 1 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x5E":
			if data.GeneralParameters[key].Value != 0 && data.GeneralParameters[key].Value < 3 && data.GeneralParameters[key].Value > 60 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x60":
			if data.GeneralParameters[key].Value != 0 && data.GeneralParameters[key].Value < 3 && data.GeneralParameters[key].Value > 60 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		case "0x61":
			if data.GeneralParameters[key].Value < 0 && data.GeneralParameters[key].Value > 10 {
				return fmt.Errorf("unvalide setting parameter key: %s", key)
			}
		}
	}
	// Проверка отсутствия лишних ключей в данных (опционально)
	for key := range data.GeneralParameters {
		if _, exists := template.GeneralParameters[key]; !exists {
			return fmt.Errorf("extra setting parameter key found: %s", key)
		}
	}

	return nil
}

func float32ToString(f float32) string {
	return fmt.Sprintf("%f", f)
}
