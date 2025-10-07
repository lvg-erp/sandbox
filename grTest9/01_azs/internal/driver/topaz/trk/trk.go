package trk

import (
	"encoding/json"
	"fmt"
	"fuelazs/internal/driver/topaz/topazAdapter"
	"fuelazs/internal/driver/topaz/trk/config"
	"path/filepath"
	"time"
)

const (
	DriverName = "trk"

	StartByte = 0x7F
	STX       = 0x02
	ETX       = 0x03
	ErrByte   = 0x15
	ErrByte2  = 0x18
	FirstByte = 0x31

	// Команды
	GetTRKStatus          = 0x31
	ApprovalTRK           = 0x32
	GetFuelGiveStatus     = 0x34
	GetFullFuelGiveStatus = 0x35
	FuelGiveSuccess       = 0x38
	SetFuelGive           = 0x54
	ReadGeneralParams     = 0x4E
	WriteGeneralParams    = 0x4F
	GetOtherTRKStatus     = 0x39
	GetTRKID              = 0x57

	StatusIdle         = "30" // Пистолет установлен
	StatusNozzleLifted = "31" // Пистолет снят
	StatusAuthorized   = "32" // Санкционирован отпуск топлива
	StatusBusy         = "33" // Отпуск топлива
	StatusIncomplete   = "34" // Отпуск топлива завершен
)

// OneByteParameters Мапа адресов с ответом 1 байт
var OneByteParameters map[byte]struct{} = map[byte]struct{}{
	0x32: struct{}{},
	0x33: struct{}{},
	0x36: struct{}{},
	0x38: struct{}{},
	0x3E: struct{}{},
	0x46: struct{}{},
	0x30: struct{}{},
	0x5D: struct{}{},
}

// TwoByteParameters Мапа адресов с ответом 2 байта
var TwoByteParameters map[byte]struct{} = map[byte]struct{}{
	0x3C: struct{}{},
	0x45: struct{}{},
	0x3D: struct{}{},
	0x3F: struct{}{},
	0x37: struct{}{},
	0x57: struct{}{},
	0x51: struct{}{},
	0x5A: struct{}{},
	0x5B: struct{}{},
}

// ThreeByteParameters Мапа адресов с ответом 3 байта
var ThreeByteParameters map[byte]struct{} = map[byte]struct{}{
	0x35: struct{}{},
	0x39: struct{}{},
	0x3A: struct{}{},
	0x3B: struct{}{},
	0x44: struct{}{},
	0x47: struct{}{},
	0x55: struct{}{},
	0x56: struct{}{},
	0x52: struct{}{},
}

// SixByteParameters Мапа адресов с ответом 6 байт
var SixByteParameters map[byte]struct{} = map[byte]struct{}{
	0x48: struct{}{},
}

type TRK struct {
	Adapter *map[string]*topazAdapter.TopazAdapter
}

func NewTRK(adapter *map[string]*topazAdapter.TopazAdapter) *TRK {
	return &TRK{
		Adapter: adapter,
	}
}

// GetTRKStatus Метод получения статуса ТРК
func (trk *TRK) GetTRKStatus(devicePNumber string) (string, error) {
	// 1. Загрузить конфиг из файла
	//filePath := "trk/" + DriverName + "_" + devicePNumber + ".json"
	filePath := filepath.Join("internal", "driver", "topaz", "trk", "config", fmt.Sprintf("trk_%s.json", devicePNumber))
	fileConfig, err := config.ReadConfig(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to load config file %s: %w", filePath, err)
	}

	// Проверка базовых полей конфига
	if fileConfig.DriverName != DriverName {
		return "", fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, fileConfig.DriverName)
	}

	if fileConfig.DevicePNumber != devicePNumber {
		return "", fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, fileConfig.DevicePNumber)
	}

	deviceAddressByte := getAddress(fileConfig.Address)

	// 2. Получение статуса
	cmd := []byte{StartByte, STX, deviceAddressByte, InvertByte(deviceAddressByte), GetTRKStatus, InvertByte(GetTRKStatus), ETX, ETX}
	crc := TOPAZCalculateChecksum(cmd)
	cmd = append(cmd, crc)

	res, err := (*trk.Adapter)[devicePNumber].SendCommand(cmd, "Long")
	if err != nil {
		return "", fmt.Errorf("SendCommand error for command 0x%X: %w", GetTRKStatus, err)
	}

	trkStatus := ByteToHexString(res[2])

	return trkStatus, nil
}

// ApprovalTRK Санкционирование ТРК
func (trk *TRK) ApprovalTRK(devicePNumber string) error {
	// 1. Загрузить конфиг из файла
	filePath := "trk/" + DriverName + "_" + devicePNumber + ".json"
	fileConfig, err := config.ReadConfig(filePath)
	if err != nil {
		return fmt.Errorf("failed to load config file %s: %w", filePath, err)
	}

	// Проверка базовых полей конфига
	if fileConfig.DriverName != DriverName {
		return fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, fileConfig.DriverName)
	}

	if fileConfig.DevicePNumber != devicePNumber {
		return fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, fileConfig.DevicePNumber)
	}

	deviceAddressByte := getAddress(fileConfig.Address)

	// Санкционирование
	cmd := []byte{StartByte, STX, deviceAddressByte, InvertByte(deviceAddressByte), ApprovalTRK, InvertByte(ApprovalTRK), ETX, ETX}
	crc := TOPAZCalculateChecksum(cmd)
	cmd = append(cmd, crc)

	res, err := (*trk.Adapter)[devicePNumber].SendCommand(cmd, "Short")
	if err != nil {
		return fmt.Errorf("SendCommand error for command 0x%X: %w", GetTRKStatus, err)
	}
	trkRes := ByteToHexString(res[1])
	_ = trkRes
	if res[1] != 0x06 {
		return fmt.Errorf("trk response 0x%X", res[1])
	}

	return nil
}

// GetFuelGiveStatus Запрос текущих данных отпуска топлива
func (trk *TRK) GetFuelGiveStatus(devicePNumber string) (float32, error) {
	// 1. Загрузить конфиг из файла
	filePath := "trk/" + DriverName + "_" + devicePNumber + ".json"
	fileConfig, err := config.ReadConfig(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to load config file %s: %w", filePath, err)
	}

	// Проверка базовых полей конфига
	if fileConfig.DriverName != DriverName {
		return 0, fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, fileConfig.DriverName)
	}

	if fileConfig.DevicePNumber != devicePNumber {
		return 0, fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, fileConfig.DevicePNumber)
	}

	deviceAddressByte := getAddress(fileConfig.Address)

	// Узнаем текущий отпуск топлива
	cmd := []byte{StartByte, STX, deviceAddressByte, InvertByte(deviceAddressByte), GetFuelGiveStatus, InvertByte(GetFuelGiveStatus), ETX, ETX}
	crc := TOPAZCalculateChecksum(cmd)
	cmd = append(cmd, crc)

	res, err := (*trk.Adapter)[devicePNumber].SendCommand(cmd, "Long")
	if err != nil {
		return 0, fmt.Errorf("SendCommand error for command 0x%X: %w", GetFuelGiveStatus, err)
	}

	liter, err := BytesToFloat(res)
	if err != nil {
		return 0, fmt.Errorf("failed to convert liter: %w", err)
	}

	return liter, nil
}

// GetFullFuelGiveStatus Запрос полных данных отпуска топлива
func (trk *TRK) GetFullFuelGiveStatus(devicePNumber string) (float32, error) {
	// 1. Загрузить конфиг из файла
	filePath := "trk/" + DriverName + "_" + devicePNumber + ".json"
	fileConfig, err := config.ReadConfig(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to load config file %s: %w", filePath, err)
	}

	// Проверка базовых полей конфига
	if fileConfig.DriverName != DriverName {
		return 0, fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, fileConfig.DriverName)
	}

	if fileConfig.DevicePNumber != devicePNumber {
		return 0, fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, fileConfig.DevicePNumber)
	}

	deviceAddressByte := getAddress(fileConfig.Address)

	// Узнаем текущий отпуск топлива
	cmd := []byte{StartByte, STX, deviceAddressByte, InvertByte(deviceAddressByte), GetFullFuelGiveStatus, InvertByte(GetFullFuelGiveStatus), ETX, ETX}
	crc := TOPAZCalculateChecksum(cmd)
	cmd = append(cmd, crc)

	res, err := (*trk.Adapter)[devicePNumber].SendCommand(cmd, "Long")
	if err != nil {
		return 0, fmt.Errorf("SendCommand error for command 0x%X: %w", GetFullFuelGiveStatus, err)
	}

	liter, err := BytesToFloat(res)
	if err != nil {
		return 0, fmt.Errorf("failed to convert liter: %w", err)
	}

	return liter, nil
}

// FuelGiveSuccess Подтверждение записи итогов отпуска
func (trk *TRK) FuelGiveSuccess(devicePNumber string) error {
	// 1. Загрузить конфиг из файла
	filePath := "trk/" + DriverName + "_" + devicePNumber + ".json"
	fileConfig, err := config.ReadConfig(filePath)
	if err != nil {
		return fmt.Errorf("failed to load config file %s: %w", filePath, err)
	}

	// Проверка базовых полей конфига
	if fileConfig.DriverName != DriverName {
		return fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, fileConfig.DriverName)
	}

	if fileConfig.DevicePNumber != devicePNumber {
		return fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, fileConfig.DevicePNumber)
	}

	deviceAddressByte := getAddress(fileConfig.Address)

	// Подтверждаем запись итогов отпуска
	cmd := []byte{StartByte, STX, deviceAddressByte, InvertByte(deviceAddressByte), FuelGiveSuccess, InvertByte(FuelGiveSuccess), ETX, ETX}
	crc := TOPAZCalculateChecksum(cmd)
	cmd = append(cmd, crc)

	res, err := (*trk.Adapter)[devicePNumber].SendCommand(cmd, "Short")
	if err != nil {
		return fmt.Errorf("SendCommand error for command 0x%X: %w", FuelGiveSuccess, err)
	}

	trkRes := ByteToHexString(res[1])

	_ = trkRes
	if res[1] != 0x06 {
		return fmt.Errorf("trk response 0x%X", res[1])
	}

	return nil
}

// SetFuelGive Установка дозы отпуска топлива в литрах
func (trk *TRK) SetFuelGive(devicePNumber string, literCount float32) error {
	// 1. Загрузить конфиг из файла
	filePath := "trk/" + DriverName + "_" + devicePNumber + ".json"
	fileConfig, err := config.ReadConfig(filePath)
	if err != nil {
		return fmt.Errorf("failed to load config file %s: %w", filePath, err)
	}

	// Проверка базовых полей конфига
	if fileConfig.DriverName != DriverName {
		return fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, fileConfig.DriverName)
	}

	if fileConfig.DevicePNumber != devicePNumber {
		return fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, fileConfig.DevicePNumber)
	}

	deviceAddressByte := getAddress(fileConfig.Address)

	var payload []byte
	literCountByte, err := FloatToBytes(literCount)
	for _, v := range literCountByte {
		buf := InvertByte(v)
		payload = append(payload, v, buf)
	}

	if err != nil {
		return fmt.Errorf("failed to convert literCount: %w", err)
	}

	// Установка дозы отпуска топлива в литрах - T (0x54)
	cmd := []byte{StartByte, STX, deviceAddressByte, InvertByte(deviceAddressByte), SetFuelGive, InvertByte(SetFuelGive)}
	cmd = append(cmd, payload...)
	cmd = append(cmd, 0x30, InvertByte(0x30), 0x30, InvertByte(0x30))
	cmd = append(cmd, ETX, ETX)
	crc := TOPAZCalculateChecksum(cmd)
	cmd = append(cmd, crc)

	res, err := (*trk.Adapter)[devicePNumber].SendCommand(cmd, "Short")
	if err != nil {
		return fmt.Errorf("SendCommand error for command 0x%X: %w", SetFuelGive, err)
	}

	trkRes := ByteToHexString(res[1])

	_ = trkRes
	if res[1] != 0x06 {
		return fmt.Errorf("trk response 0x%X", res[1])
	}

	return nil
}

// GetSettings Чтение параметров ТРК
func (trk *TRK) GetSettings(devicePNumber string) ([]byte, error) {
	//filePath := "trk/" + DriverName + "_" + devicePNumber + ".json"
	filePath := filepath.Join("internal", "driver", "topaz", "trk", "config", fmt.Sprintf("trk_%s.json", devicePNumber))
	fileConfig, err := config.ReadConfig(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file %s: %w", filePath, err)
	}
	// Проверка базовых полей конфига
	if fileConfig.DriverName != DriverName {
		return nil, fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, fileConfig.DriverName)
	}

	if fileConfig.DevicePNumber != devicePNumber {
		return nil, fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, fileConfig.DevicePNumber)
	}

	deviceAddressByte := getAddress(fileConfig.Address)

	// Ответ 1 байт

	unReadParameters := make([]byte, 0, 32)

	generalParams := make(map[string][]byte)
	for k, _ := range OneByteParameters {
		cmd := []byte{StartByte, STX, deviceAddressByte, InvertByte(deviceAddressByte), ReadGeneralParams, InvertByte(ReadGeneralParams),
			k, InvertByte(k), ETX, ETX}
		crc := TOPAZCalculateChecksum(cmd)
		cmd = append(cmd, crc)

		res, err := (*trk.Adapter)[devicePNumber].SendCommand(cmd, "Long")
		if err != nil {
			return nil, fmt.Errorf("SendCommand error for command 0x%X: %w", cmd, err)
		}

		if res[1] == ErrByte || res[1] == ErrByte2 {
			unReadParameters = append(unReadParameters, k)
			continue
		}

		generalParams[ByteToHexStringSimple(res[2])] = []byte{res[4]}
		time.Sleep(10 * time.Millisecond)
	}

	// Ответ 2 байта

	for k, _ := range TwoByteParameters {
		cmd := []byte{StartByte, STX, deviceAddressByte, InvertByte(deviceAddressByte), ReadGeneralParams, InvertByte(ReadGeneralParams),
			k, InvertByte(k), ETX, ETX}
		crc := TOPAZCalculateChecksum(cmd)
		cmd = append(cmd, crc)

		res, err := (*trk.Adapter)[devicePNumber].SendCommand(cmd, "Long")
		if err != nil {
			return nil, fmt.Errorf("SendCommand error for command 0x%X: %w", cmd, err)
		}

		if res[1] == ErrByte || res[1] == ErrByte2 {
			unReadParameters = append(unReadParameters, k)
			continue
		}

		generalParams[ByteToHexStringSimple(res[2])] = []byte{res[4], res[6]}
		time.Sleep(10 * time.Millisecond)
	}

	// Ответ 3 байта

	for k, _ := range ThreeByteParameters {
		cmd := []byte{StartByte, STX, deviceAddressByte, InvertByte(deviceAddressByte), ReadGeneralParams, InvertByte(ReadGeneralParams),
			k, InvertByte(k), ETX, ETX}
		crc := TOPAZCalculateChecksum(cmd)
		cmd = append(cmd, crc)
		res, err := (*trk.Adapter)[devicePNumber].SendCommand(cmd, "Long")
		if err != nil {
			return nil, fmt.Errorf("SendCommand error for command 0x%X: %w", cmd, err)
		}

		if res[1] == ErrByte || res[1] == ErrByte2 {
			unReadParameters = append(unReadParameters, k)
			continue
		}

		generalParams[ByteToHexStringSimple(res[2])] = []byte{res[4], res[6], res[8]}
		time.Sleep(10 * time.Millisecond)
	}

	//Ответ 6 байта

	for k, _ := range SixByteParameters {
		cmd := []byte{StartByte, STX, deviceAddressByte, InvertByte(deviceAddressByte), ReadGeneralParams, InvertByte(ReadGeneralParams),
			k, InvertByte(k), ETX, ETX}
		crc := TOPAZCalculateChecksum(cmd)
		cmd = append(cmd, crc)

		res, err := (*trk.Adapter)[devicePNumber].SendCommand(cmd, "Long")
		if err != nil {
			return nil, fmt.Errorf("SendCommand error for command 0x%X: %w", cmd, err)
		}

		if res[1] == ErrByte || res[1] == ErrByte2 {
			unReadParameters = append(unReadParameters, k)
			continue
		}

		generalParams[ByteToHexStringSimple(res[2])] = []byte{res[4], res[6], res[8], res[10], res[12], res[14], res[16]}
		time.Sleep(10 * time.Millisecond)
	}

	result := TRKСonvertByteToFloat(generalParams)

	response := TRKResponse{
		Address:           fileConfig.Address,
		GeneralParameters: result,
		DriverName:        DriverName,
		DevicePNumber:     devicePNumber,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("json marshal error: %w", err)
	}

	// TODO: Сделать отработку параметров, которые не удалось прочитать
	if len(unReadParameters) != 0 {
		fmt.Printf("%X\n", unReadParameters)
	}

	return responseJSON, nil
}

// SetSettings Настройка ТРК
func (trk *TRK) SetSettings(devicePNumber string, data []byte) error {
	filePath := "trk/" + DriverName + "_" + devicePNumber + ".json"

	if data != nil {
		var dataJSON config.TRKConfig
		err := json.Unmarshal(data, &dataJSON)
		if err != nil {
			return fmt.Errorf("data.json parse error: %w", err)
		}

		// Базовые проверки входных данных
		if dataJSON.DriverName != DriverName {
			return fmt.Errorf("data.json driver name mismatch. Expect: %s, Have: %s", DriverName, dataJSON.DriverName)
		}
		if dataJSON.DevicePNumber != devicePNumber {
			return fmt.Errorf("data.json device number mismatch. Expect: %s, Have: %s", devicePNumber, dataJSON.DevicePNumber)
		}

		// Загрузка существующего конфига для сравнения и валидации
		var baseConfig *config.TRKConfig
		fileExists := config.FileExists(filePath)

		if fileExists {
			baseConfig, err = config.ReadConfig(filePath)
			if err != nil {
				return fmt.Errorf("failed to load existing config %s: %w", filePath, err)
			}
		} else {
			return fmt.Errorf("initialization failed, config file %s not found or invalid: %w", filePath, err)
		}

		// Строгая валидация структуры dataJSON относительно baseConfig
		if err = validateSettingsStructure(baseConfig, &dataJSON); err != nil {
			return fmt.Errorf("data.json structure validation failed: %w", err)
		}

		// Вычисляем отличающиеся параметры для записи
		writingParamsMap := make(map[string]float32)
		for dataKey, dataParam := range dataJSON.GeneralParameters {
			baseValue, _ := baseConfig.GeneralParameters[dataKey]

			if baseValue.Value != dataParam.Value {
				writingParamsMap[dataKey] = dataParam.Value
			}
		}

		if len(writingParamsMap) == 0 {
			fmt.Println("No setting changes detected, skipping write.")
			return nil
		}

		deviceAddressByte := getAddress(baseConfig.Address)

		// Верификация сетевого адреса
		err = trk.validateNetworkAddress(devicePNumber, baseConfig.Address)
		if err != nil {
			return fmt.Errorf("network address validation error: %w ", err)
		}

		adminPass := baseConfig.AdminPass
		adminPassBytes := StringToBytesSimple(adminPass)
		adminCMD := []byte{StartByte, STX, deviceAddressByte, InvertByte(deviceAddressByte), 0x4D, InvertByte(0x4D),
			0x51, InvertByte(0x51), 0x35, InvertByte(0x35), 0x30, InvertByte(0x30),
			0x31, InvertByte(0x31), 0x41, InvertByte(0x41),
			adminPassBytes[0], InvertByte(adminPassBytes[0]), adminPassBytes[1], InvertByte(adminPassBytes[1]),
			adminPassBytes[2], InvertByte(adminPassBytes[2]), adminPassBytes[3], InvertByte(adminPassBytes[3]),
			ETX, ETX}
		crc := TOPAZCalculateChecksum(adminCMD)
		adminCMD = append(adminCMD, crc)

		res, err := (*trk.Adapter)[devicePNumber].SendCommand(adminCMD, "Short")
		if err != nil {
			return fmt.Errorf("SendCommand error for command 0x%X: %w", adminCMD, err)
		}

		if res[1] == ErrByte || res[1] == ErrByte2 {
			return fmt.Errorf("admin error", res[1])
		}
		fmt.Println("Admin success")

		writingParams := TRKConvertFloatToByte(writingParamsMap)
		unWritingParams := make([]string, 0, len(writingParams))
		// Запись отличающихся параметров
		for k, v := range writingParams {
			cmd := []byte{StartByte, STX, deviceAddressByte, InvertByte(deviceAddressByte), WriteGeneralParams, InvertByte(WriteGeneralParams),
				StringToBytes(k), InvertByte(StringToBytes(k))}
			for _, vv := range v {
				cmd = append(cmd, vv)
				cmd = append(cmd, InvertByte(vv))
			}
			cmd = append(cmd, ETX, ETX)
			crc = TOPAZCalculateChecksum(cmd)
			cmd = append(cmd, crc)

			res, err = (*trk.Adapter)[devicePNumber].SendCommand(cmd, "Short")
			if err != nil {
				return fmt.Errorf("SendCommand error for command 0x%X: %w", cmd, err)
			}

			time.Sleep(50 * time.Millisecond)
			if res[1] == ErrByte || res[1] == ErrByte2 {
				unWritingParams = append(unWritingParams, k)
			}
		}

		if len(unWritingParams) != 0 {
			return fmt.Errorf("unwriting parameters: %X", unWritingParams)
		}

		dataJSON.Address = baseConfig.Address

		// Сохраняем успешную конфигурацию в файл
		err = SaveConfig(filePath, dataJSON)
		if err != nil {
			return fmt.Errorf("SaveConfig error after writing parameters: %w", err)
		}

		fmt.Printf("Settings successfully written and saved to %s\n", filePath)
		return nil

	} else {
		// --- Сценарий 2: Синхронизация при запуске (data == nil) ---
		fmt.Printf("Starting settings synchronization for %s\n", devicePNumber)

		// 1. Загрузить конфиг из файла
		fileConfig, err := config.ReadConfig(filePath)
		if err != nil {
			return fmt.Errorf("initialization failed, config file %s not found or invalid: %w", filePath, err)
		}

		deviceAddressByte := getAddress(fileConfig.Address)

		// Верификация сетевого адреса
		err = trk.validateNetworkAddress(devicePNumber, fileConfig.Address)
		if err != nil {
			return fmt.Errorf("network address validation error: %w ", err)
		}

		readParams, err := trk.GetSettings(devicePNumber)
		if err != nil {
			return fmt.Errorf("GetSettings failed: %w", err)
		}

		var readParamsStruct *TRKResponse

		err = json.Unmarshal(readParams, &readParamsStruct)

		writingParamsMap := make(map[string]float32)
		for dataKey, dataParam := range fileConfig.GeneralParameters {
			baseValue, _ := readParamsStruct.GeneralParameters[dataKey]

			if baseValue != dataParam.Value {
				writingParamsMap[dataKey] = dataParam.Value
			}
		}

		if len(writingParamsMap) == 0 {
			fmt.Println("No setting changes detected, skipping write.")
			return nil
		}

		adminPass := fileConfig.AdminPass
		adminPassBytes := StringToBytesSimple(adminPass)
		adminCMD := []byte{StartByte, STX, deviceAddressByte, InvertByte(deviceAddressByte), 0x4D, InvertByte(0x4D),
			0x51, InvertByte(0x51), 0x35, InvertByte(0x35), 0x30, InvertByte(0x30),
			0x31, InvertByte(0x31), 0x41, InvertByte(0x41),
			adminPassBytes[0], InvertByte(adminPassBytes[0]), adminPassBytes[1], InvertByte(adminPassBytes[1]),
			adminPassBytes[2], InvertByte(adminPassBytes[2]), adminPassBytes[3], InvertByte(adminPassBytes[3]),
			ETX, ETX}
		crc := TOPAZCalculateChecksum(adminCMD)
		adminCMD = append(adminCMD, crc)

		res, err := (*trk.Adapter)[devicePNumber].SendCommand(adminCMD, "Short")
		if err != nil {
			return fmt.Errorf("SendCommand error for command 0x%X: %w", adminCMD, err)
		}

		if res[1] == ErrByte || res[1] == ErrByte2 {
			return fmt.Errorf("admin error", res[1])
		}
		fmt.Println("Admin success")

		writingParams := TRKConvertFloatToByte(writingParamsMap)
		unWritingParams := make([]string, 0, len(writingParams))
		// Запись отличающихся параметров
		for k, v := range writingParams {
			cmd := []byte{StartByte, STX, deviceAddressByte, InvertByte(deviceAddressByte), WriteGeneralParams, InvertByte(WriteGeneralParams),
				StringToBytes(k), InvertByte(StringToBytes(k))}
			for _, vv := range v {
				cmd = append(cmd, vv)
				cmd = append(cmd, InvertByte(vv))
			}
			cmd = append(cmd, ETX, ETX)
			crc = TOPAZCalculateChecksum(cmd)
			cmd = append(cmd, crc)

			res, err = (*trk.Adapter)[devicePNumber].SendCommand(cmd, "Short")
			if err != nil {
				return fmt.Errorf("SendCommand error for command 0x%X: %w", cmd, err)
			}

			time.Sleep(50 * time.Millisecond)
			if res[1] == 0x18 || res[1] == 0x15 {
				unWritingParams = append(unWritingParams, k)
			}
		}

		if len(unWritingParams) != 0 {
			return fmt.Errorf("unwriting parameters: %X", unWritingParams)
		}

		fmt.Printf("Settings successfully written to %s\n", DriverName)
		return nil

	}

}

// GetOtherStatus Получение дополнительного статуса ТРК
func (trk *TRK) GetOtherStatus(devicePNumber string) ([]byte, error) {
	filePath := "trk/" + DriverName + "_" + devicePNumber + ".json"

	fileConfig, err := config.ReadConfig(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file %s: %w", filePath, err)
	}
	// Проверка базовых полей конфига
	if fileConfig.DriverName != DriverName {
		return nil, fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, fileConfig.DriverName)
	}

	if fileConfig.DevicePNumber != devicePNumber {
		return nil, fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, fileConfig.DevicePNumber)
	}

	deviceAddressByte := getAddress(fileConfig.Address)

	// Верификация сетевого адреса
	err = trk.validateNetworkAddress(devicePNumber, fileConfig.Address)
	if err != nil {
		return nil, fmt.Errorf("network address validation error: %w ", err)
	}

	// Ответ 1 байт

	cmd := []byte{StartByte, STX, deviceAddressByte, InvertByte(deviceAddressByte), GetOtherTRKStatus, InvertByte(GetOtherTRKStatus), ETX, ETX}
	crc := TOPAZCalculateChecksum(cmd)
	cmd = append(cmd, crc)

	res, err := (*trk.Adapter)[devicePNumber].SendCommand(cmd, "Long")
	if err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	responseJSON, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("json marshal error: %w", err)
	}

	return responseJSON, nil
}

// GetTRKID Получение ID номера устройства
func (trk *TRK) GetTRKID(devicePNumber string) (string, error) {
	filePath := "trk/" + DriverName + "_" + devicePNumber + ".json"

	fileConfig, err := config.ReadConfig(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to load config file %s: %w", filePath, err)
	}
	// Проверка базовых полей конфига
	if fileConfig.DriverName != DriverName {
		return "", fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, fileConfig.DriverName)
	}

	if fileConfig.DevicePNumber != devicePNumber {
		return "", fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, fileConfig.DevicePNumber)
	}

	cmd := []byte{StartByte, STX, GetTRKID, InvertByte(GetTRKID), FirstByte, InvertByte(FirstByte), 0x3F, InvertByte(0x3F), ETX, ETX}
	crc := TOPAZCalculateChecksum(cmd)
	cmd = append(cmd, crc)

	res, err := (*trk.Adapter)[devicePNumber].SendCommand(cmd, "Long")
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	id, err := ConvertTRKID(res)
	if err != nil {
		return "", fmt.Errorf("failed to convert trk id: %w", err)
	}

	return id, nil
}

// GetNetworkAddress Получение сетевого адреса устройства
func (trk *TRK) GetNetworkAddress(devicePNumber string) (string, string, error) {
	filePath := "trk/" + DriverName + "_" + devicePNumber + ".json"

	fileConfig, err := config.ReadConfig(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to load config file %s: %w", filePath, err)
	}
	// Проверка базовых полей конфига
	if fileConfig.DriverName != DriverName {
		return "", "", fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, fileConfig.DriverName)
	}

	if fileConfig.DevicePNumber != devicePNumber {
		return "", "", fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, fileConfig.DevicePNumber)
	}

	id, err := trk.GetTRKID(devicePNumber)
	if err != nil {
		return "", "", fmt.Errorf("failed to get trk id: %w", err)
	}
	idBytes := StringToSlice(id)

	cmd := []byte{StartByte, STX, 0x5D, InvertByte(0x5D)}
	for i, idByte := range idBytes {
		cmd = append(cmd, idByte, InvertByte2(idByte))
		if i == 4 {
			cmd = append(cmd, 0x30, InvertByte2(0x30), 0x30, InvertByte2(0x30), 0x30, InvertByte2(0x30))
		}
	}
	cmd = append(cmd, ETX, ETX)
	crc := TOPAZCalculateChecksum(cmd)
	cmd = append(cmd, crc)

	res, err := (*trk.Adapter)[devicePNumber].SendCommand(cmd, "Long")
	if err != nil {
		return "", "", fmt.Errorf("failed to send command: %w", err)
	}

	netAddr, workStatus := ConvertNetworkAddress(res)

	return netAddr, workStatus, nil
}

// validateNetworkAddress Валидация сетевого адреса устройства
func (trk *TRK) validateNetworkAddress(devicePNumber string, networkAddress string) error {
	netAddress, _, err := trk.GetNetworkAddress(devicePNumber)
	if err != nil {
		return fmt.Errorf("get network address error: %s", devicePNumber)
	}

	address := RemoveLeadingZeros(netAddress)

	if address != networkAddress {
		return fmt.Errorf("invalid network address: %s", networkAddress)
	}

	return nil
}
