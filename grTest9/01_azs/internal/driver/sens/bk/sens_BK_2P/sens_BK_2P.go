package sens_BK_2P

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"fuelazs/internal/driver/sens"
	"fuelazs/internal/driver/sens/bk/configBK"
	"fuelazs/internal/driver/sens/sensAdapter"
)

const (
	// DriverName Название устройства
	DriverName = "sens_BK_2P"
	// SyncByte Стартовый байт
	SyncByte byte = 0xB5
	// CmdReadInfo Команда для чтения параметров
	CmdReadInfo byte = 0x0F
	// CmdWrite Команда для записи параметров
	CmdWrite byte = 0x11
	// DeviceNumber Адрес номера устройства
	DeviceNumber = 0xF2
)

// DeviceList Список устройств, поддерживаемых драйвером
var DeviceList = map[string]string{
	"C720": DriverName,
	"C721": DriverName,
	"C722": DriverName,
	"C723": DriverName,
	"C724": DriverName,
	"C725": DriverName,
	"C726": DriverName,
	"C7A0": DriverName,
}

type BK struct {
	Adapter *map[string]*sensAdapter.SensAdapter
}

func NewBK(adapter *map[string]*sensAdapter.SensAdapter) *BK {
	return &BK{
		Adapter: adapter,
	}
}

// SetSettings Метод для настройки параметров блока коммутации
func (bk *BK) SetSettings(devicePNumber string, data []byte) error {

	filePath := "bk/" + DriverName + "_" + devicePNumber + ".json"

	// --- Сценарий 1: Применение новых настроек (data != nil) ---
	if data != nil {
		var dataJSON configBK.SensBk2P
		err := json.Unmarshal(data, &dataJSON)
		if err != nil {
			return fmt.Errorf("data.json parse error: %w", err)
		}

		// Базовая проверка входных данных
		if dataJSON.DriverName != DriverName {
			return fmt.Errorf("data.json driver name mismatch. Expect: %s, Have: %s", DriverName, dataJSON.DriverName)
		}
		if dataJSON.DevicePNumber != devicePNumber {
			return fmt.Errorf("data.json device number mismatch. Expect: %s, Have: %s", devicePNumber, dataJSON.DevicePNumber)
		}

		// Загрузка базового конфига
		var baseConfig *configBK.SensBk2P
		fileExists := sens.FileExists(filePath)

		if fileExists {
			baseConfig, err = configBK.ReadConfig(filePath)
			if err != nil {
				return fmt.Errorf("failed to load existing config %s: %w", filePath, err)
			}
		} else {
			return fmt.Errorf("initialization failed, config file %s not found or invalid: %w", filePath, err)
		}

		// Строгая валидация структуры dataJSON относительно baseConfig
		if err := validateSettingsStructure(baseConfig, &dataJSON); err != nil {
			return fmt.Errorf("data.json structure validation failed: %w", err)
		}

		// --- Вычисление изменений ---
		// Параметры
		writingParamsMap := make(map[string]float32) // Map[ParamHex]Value
		for dataKey, dataParam := range dataJSON.SettingsParameters {
			baseValue, ok := baseConfig.SettingsParameters[dataKey]
			if !ok || baseValue.Value != dataParam.Value {
				writingParamsMap[dataKey] = dataParam.Value
			}
		}

		// Таблицы
		writingTablesMap := make(map[string][]float32)
		for dataTableKey, dataTable := range dataJSON.Tables {
			baseTable, tableExistedInBase := baseConfig.Tables[dataTableKey]

			needsWrite := false
			if !tableExistedInBase {
				needsWrite = true
			} else {
				needsWrite = !sens.AreSlicesEqual(baseTable.Value, dataTable.Value)
			}

			if needsWrite {
				valuesToWrite := dataTable.Value

				// Проверка длины и дополнение нулями для ЗАПИСИ В УСТРОЙСТВО
				if tableExistedInBase && len(dataTable.Value) < len(baseTable.Value) {
					originalLength := len(baseTable.Value)
					fmt.Printf("Info: Table %s data is shorter (%d) than original (%d). Padding with zeros for device write.\n",
						dataTableKey, len(dataTable.Value), originalLength)
					paddedValues := make([]float32, originalLength)
					copy(paddedValues, dataTable.Value)
					valuesToWrite = paddedValues
				}

				writingTablesMap[dataTableKey] = valuesToWrite
			}

		}

		// Если нет изменений ни в параметрах, ни в таблицах
		if len(writingParamsMap) == 0 && len(writingTablesMap) == 0 {
			fmt.Println("No setting or table changes detected, skipping write.")
			return nil
		}

		// --- Запись в устройство ---
		deviceAddressByte := sens.StringToBytes(baseConfig.Address)

		// 1. Верификация устройства
		verifyPayload := []byte{DeviceNumber}
		res, err := bk.sendLinCommand(devicePNumber, deviceAddressByte, CmdReadInfo, verifyPayload, DriverName)
		if err != nil {
			return fmt.Errorf("device verification command failed for address %s: %w", baseConfig.Address, err)
		}

		// Проверка ответа верификации
		if len(res) < 9 {
			return fmt.Errorf("device verification response too short (expected >= 9, got %d) from address %s", len(res), baseConfig.Address)
		}
		deviceResponseStr := sens.ByteToHexString(res[6]) + sens.ByteToHexString(res[5])
		_, ok := DeviceList[deviceResponseStr]
		if !ok {
			return fmt.Errorf("unknown device response ID %s from address %s", deviceResponseStr, baseConfig.Address)
		}

		// 2. Запись отличающихся параметров (если есть)
		if len(writingParamsMap) > 0 {
			const bytesPerParam = 4
			writeParamsPayload := make([]byte, 0, len(writingParamsMap)*bytesPerParam)
			paramAddressesToWrite := []string{}

			for paramHex, paramValue := range writingParamsMap {
				paramAddrByte := sens.StringToBytes(paramHex)
				buf, err := sens.ConvertFloat32To24Bytes(paramValue, binary.LittleEndian)
				if err != nil {
					return fmt.Errorf("ConvertFloat32To24Bytes error for param %s: %w", paramHex, err)
				}
				if len(buf) != 3 {
					return fmt.Errorf("unexpected byte count from ConvertFloat32To24Bytes for param %s: expected 3, got %d", paramHex, len(buf))
				}
				writeParamsPayload = append(writeParamsPayload, paramAddrByte)
				writeParamsPayload = append(writeParamsPayload, buf...)
				paramAddressesToWrite = append(paramAddressesToWrite, paramHex)
			}

			res, err = bk.sendLinCommand(devicePNumber, deviceAddressByte, CmdWrite, writeParamsPayload, DriverName)
			if err != nil {
				return fmt.Errorf("write parameters command failed for address %s: %w", baseConfig.Address, err)
			}

			// Проверка ответа записи параметров
			successParamsMap := make(map[byte]bool)

			if len(res) >= 5 && len(res) == int(res[2])+5 {
				for i := 4; i < len(res)-1; i++ {
					successParamsMap[res[i]] = true
				}
			} else {
				return fmt.Errorf("invalid write parameters response format or length from address %s (got %d bytes, len field %d)", baseConfig.Address, len(res), res[2])
			}

			for _, paramHex := range paramAddressesToWrite {
				paramAddrByte := sens.StringToBytes(paramHex)
				if _, ok := successParamsMap[paramAddrByte]; !ok {
					return fmt.Errorf("parameter %s (0x%X) write failed (not confirmed in response)", paramHex, paramAddrByte)
				}
			}
			fmt.Println("Parameters written successfully.")
		} else {
			fmt.Println("No parameters to write.")
		}

		// 3. Запись отличающихся таблиц (если есть)
		if len(writingTablesMap) > 0 {
			fmt.Printf("Writing %d tables...\n", len(writingTablesMap))
			for tableKeyHex, tableValuesToWrite := range writingTablesMap {
				fmt.Printf("Writing table %s (length %d for device)...\n", tableKeyHex, len(tableValuesToWrite))
				tableKeyByte := sens.StringToBytes(tableKeyHex)

				// Кодируем данные таблицы в байты
				encodedData, err := encodeTableData(tableValuesToWrite) // Используем valuesToWrite
				if err != nil {
					// Если кодирование не удалось - критическая ошибка
					return fmt.Errorf("failed to encode data for table %s: %w", tableKeyHex, err)
				}
				if len(encodedData) == 0 && len(tableValuesToWrite) > 0 {
					return fmt.Errorf("internal error: encoding non-empty table %s resulted in zero bytes (check encodeTableData)", tableKeyHex)
				}

				// Формируем полезную нагрузку команды записи таблицы
				writeTablePayload := []byte{tableKeyByte, 0x00, 0x00}
				writeTablePayload = append(writeTablePayload, encodedData...)

				res, err = bk.sendLinCommand(devicePNumber, deviceAddressByte, 0x1A, writeTablePayload, DriverName)
				if err != nil {
					return fmt.Errorf("write table %s command failed for address %s: %w", tableKeyHex, baseConfig.Address, err)
				}

				// Проверка ответа записи таблицы
				if len(res) < 5 {
					return fmt.Errorf("invalid write table %s response: too short (got %d bytes)", tableKeyHex, len(res))
				}
				fmt.Printf("Table %s write command sent, response: %X\n", tableKeyHex, res)
			}
			fmt.Println("Tables write commands sent.")
		} else {
			fmt.Println("No tables to write.")
		}

		dataJSON.Address = baseConfig.Address

		// 4. Сохраняем успешную конфигурацию в файл
		// Сохраняем ПОСЛЕ успешной записи и параметров, и таблиц
		fmt.Printf("Saving configuration with potentially shorter tables back to %s...\n", filePath)
		err = sens.SaveConfig(filePath, dataJSON)
		if err != nil {
			return fmt.Errorf("settings/tables written to device, but SaveConfig failed for %s: %w", filePath, err)
		}

		fmt.Printf("Settings successfully written and saved to %s\n", filePath)
		return nil

	} else {
		// --- Сценарий 2: Синхронизация при запуске (data == nil) ---
		fmt.Printf("Starting settings synchronization for %s\n", devicePNumber)

		// 1. Загрузить конфиг из файла
		fileConfig, err := configBK.ReadConfig(filePath)
		if err != nil {
			// Если файла нет при запуске - это критическая ошибка инициализации
			return fmt.Errorf("initialization failed, config file %s not found or invalid: %w", filePath, err)
		}

		deviceAddressByte := sens.StringToBytes(fileConfig.Address)

		// 2. Верификация устройства
		verifyPayload := []byte{DeviceNumber}
		res, err := bk.sendLinCommand(devicePNumber, deviceAddressByte, CmdReadInfo, verifyPayload, DriverName)
		if err != nil {
			return fmt.Errorf("initialization failed: device verification command failed for address %s: %w", fileConfig.Address, err)
		}
		// Проверка ответа
		if len(res) < 9 {
			return fmt.Errorf("initialization failed: device verification response too short from address %s", fileConfig.Address)
		}
		deviceResponseStr := sens.ByteToHexString(res[6]) + sens.ByteToHexString(res[5])
		_, ok := DeviceList[deviceResponseStr]
		if !ok {
			return fmt.Errorf("initialization failed: unknown device response ID %s from address %s", deviceResponseStr, fileConfig.Address)
		}

		// --- 3. Чтение с устройства ---
		// Чтение параметров
		readValuesMap := make(map[byte]float32) // Map[ParamAddrByte]Value
		paramsToRead := make([]byte, 0, len(fileConfig.SettingsParameters))
		paramKeysFromFile := []string{}
		if len(fileConfig.SettingsParameters) > 0 {
			fmt.Printf("Reading %d parameters from device...\n", len(fileConfig.SettingsParameters))
			for key := range fileConfig.SettingsParameters {
				paramsToRead = append(paramsToRead, sens.StringToBytes(key))
				paramKeysFromFile = append(paramKeysFromFile, key)
			}

			res, err = bk.sendLinCommand(devicePNumber, deviceAddressByte, CmdReadInfo, paramsToRead, DriverName)
			if err != nil {
				return fmt.Errorf("initialization failed: read parameters command failed for address %s: %w", fileConfig.Address, err)
			}

			// Парсинг ответа чтения параметров
			expectedDataLength := len(paramsToRead) * 4
			if len(res) < 5+expectedDataLength || int(res[2]) != expectedDataLength {
				fmt.Printf("Warning: read parameters response length mismatch (Addr: %s). Expected data len %d, got %d bytes in response, response len field %d. Attempting partial parse.\n",
					fileConfig.Address, expectedDataLength, len(res)-5, int(res[2]))
			}

			actualDataLength := len(res) - 1 // Индекс до CRC
			for i := 4; i < actualDataLength; i += 4 {
				if i+3 >= actualDataLength {
					fmt.Printf("Warning: incomplete parameter data at end of read response (Addr: %s)\n", fileConfig.Address)
					break
				}
				paramAddrByte := res[i]
				value, err := sens.Convert24BytesToFloat32(res[i+1:i+4], binary.LittleEndian)
				if err != nil {
					fmt.Printf("Warning: ConvertFloat32ToFloat32 error for param Addr 0x%X from device %s: %v\n", paramAddrByte, fileConfig.Address, err)
					continue
				}
				readValuesMap[paramAddrByte] = value
			}
			fmt.Println("Parameters read.")
		} else {
			fmt.Println("No parameters defined in file, skipping parameter read.")
		}

		// Чтение таблиц
		readTablesMap := make(map[string][]float32, len(fileConfig.Tables))
		if len(fileConfig.Tables) > 0 {
			fmt.Printf("Reading %d tables from device...\n", len(fileConfig.Tables))
			for keyTableHex := range fileConfig.Tables {
				fmt.Printf("Reading table %s...\n", keyTableHex)
				tableKeyByte := sens.StringToBytes(keyTableHex)

				readTablePayload := []byte{tableKeyByte, 0x00, 0x00, 0x36, 0x00}

				res, err = bk.sendLinCommand(devicePNumber, deviceAddressByte, 0x0A, readTablePayload, DriverName)
				if err != nil {
					// Логируем ошибку чтения конкретной таблицы, но продолжаем с другими
					fmt.Printf("Warning: read table %s command failed for address %s: %v. Skipping table.\n", keyTableHex, fileConfig.Address, err)
					continue
					// return fmt.Errorf("initialization failed: read table %s command failed for address %s: %w", keyTableHex, fileConfig.Address, err)
				}

				if len(res) < 5 {
					fmt.Printf("Warning: read table %s response is too short (%d bytes). Skipping table.\n", keyTableHex, len(res))
					continue
				}
				tableDataBytes := res[7 : len(res)-1]
				effectiveTableDataBytes := tableDataBytes

				terminatorIndex := -1
				for i := 0; i < len(tableDataBytes)-1; i++ {
					if tableDataBytes[i] == 0x00 && tableDataBytes[i+1] == 0x00 {
						terminatorIndex = i // Нашли первый байт терминатора
						break
					}
				}

				if terminatorIndex != -1 {
					fmt.Printf("Info: Found 0x00 0x00 terminator at index %d in table %s data. Using data before terminator.\n", terminatorIndex, keyTableHex)
					effectiveTableDataBytes = tableDataBytes[:terminatorIndex]
				}

				decodedValues, err := decodeTableData(effectiveTableDataBytes)
				if err != nil {
					fmt.Printf("Warning: failed to decode data for table %s (Data: %X): %v. Skipping table.\n", keyTableHex, tableDataBytes, err)
					continue
				}
				readTablesMap[keyTableHex] = decodedValues
				fmt.Printf("Successfully read and processed table %s, stored %d values.\n", keyTableHex, len(decodedValues))
			}
			fmt.Println("Tables read.")
		} else {
			fmt.Println("No tables defined in file, skipping table read.")
		}

		// --- 4. Сравнение и подготовка к записи ---
		// Параметры
		writeParamsPayloadMap := make(map[byte]float32)
		writeParamsPayloadBytes := make([]byte, 0)
		paramsNotRead := []string{}

		for _, paramKey := range paramKeysFromFile {
			paramAddrByte := sens.StringToBytes(paramKey)
			fileValue := fileConfig.SettingsParameters[paramKey].Value

			deviceValue, ok := readValuesMap[paramAddrByte]
			if !ok {
				paramsNotRead = append(paramsNotRead, paramKey)
				fmt.Printf("Warning: Parameter %s (0x%X) from file was not found in read response from device %s. Scheduling write from file.\n", paramKey, paramAddrByte, fileConfig.Address)
				writeParamsPayloadMap[paramAddrByte] = fileValue
				buf, err := sens.ConvertFloat32To24Bytes(fileValue, binary.LittleEndian)
				if err != nil {
					return fmt.Errorf("init failed: ConvertFloat32To24Bytes error for param %s: %w", paramKey, err)
				}
				writeParamsPayloadBytes = append(writeParamsPayloadBytes, paramAddrByte)
				writeParamsPayloadBytes = append(writeParamsPayloadBytes, buf...)
				continue
			}

			if fileValue != deviceValue {
				fmt.Printf("Param mismatch %s (0x%X): File=%.4f, Device=%.4f. Scheduling write.\n", paramKey, paramAddrByte, fileValue, deviceValue)
				writeParamsPayloadMap[paramAddrByte] = fileValue
				buf, err := sens.ConvertFloat32To24Bytes(fileValue, binary.LittleEndian)
				if err != nil {
					return fmt.Errorf("init failed: ConvertFloat32To24Bytes error for param %s: %w", paramKey, err)
				}
				writeParamsPayloadBytes = append(writeParamsPayloadBytes, paramAddrByte)
				writeParamsPayloadBytes = append(writeParamsPayloadBytes, buf...)
			}
		}

		if len(paramsNotRead) > 0 {
			fmt.Printf("Warning: Could not read %d parameters from device %s: %v\n", len(paramsNotRead), fileConfig.Address, paramsNotRead)
		}

		writingTablesMap := make(map[string][]float32)
		for fileTableKey, fileTable := range fileConfig.Tables {
			readTable, tableExistedInBase := readTablesMap[fileTableKey]

			needsWrite := false
			if !tableExistedInBase {
				needsWrite = true
			} else {
				needsWrite = !sens.AreSlicesEqual(readTable, fileTable.Value)
			}

			if needsWrite {
				valuesToWrite := fileTable.Value

				if tableExistedInBase && len(fileTable.Value) < len(readTable) {
					originalLength := len(readTable)
					fmt.Printf("Info: Table %s data is shorter (%d) than original (%d). Padding with zeros for device write.\n",
						fileTableKey, len(fileTable.Value), originalLength)
					paddedValues := make([]float32, originalLength)
					copy(paddedValues, fileTable.Value)
					valuesToWrite = paddedValues
				}

				writingTablesMap[fileTableKey] = valuesToWrite
			}
		}

		// --- 5. Запись изменений в устройство ---
		// Запись таблиц
		if len(writingTablesMap) > 0 {
			fmt.Printf("Writing %d differing tables to device %s...\n", len(writingTablesMap), fileConfig.Address)
			for tableKeyHex, tableValues := range writingTablesMap {
				fmt.Printf("Writing table %s for sync...\n", tableKeyHex)
				tableKeyByte := sens.StringToBytes(tableKeyHex)
				encodedData, err := encodeTableData(tableValues)
				if err != nil {
					return fmt.Errorf("sync failed: failed to encode data for table %s: %w", tableKeyHex, err)
				}

				writeTablePayload := []byte{tableKeyByte, 0x00, 0x00}
				writeTablePayload = append(writeTablePayload, encodedData...)

				res, err = bk.sendLinCommand(devicePNumber, deviceAddressByte, 0x1A, writeTablePayload, DriverName)
				if err != nil {
					// Если запись таблицы не удалась - возможно, стоит прервать синхронизацию?
					return fmt.Errorf("sync failed: write table %s command failed for address %s: %w", tableKeyHex, fileConfig.Address, err)
				}

				if len(res) < 5 {
					return fmt.Errorf("sync failed: invalid write table %s response: too short (got %d bytes)", tableKeyHex, len(res))
				}
				fmt.Printf("Table %s write command sent for sync, response: %X\n", tableKeyHex, res)
			}
			fmt.Println("Tables written for sync.")
		} else {
			fmt.Println("No tables to write for sync.")
		}

		// Запись параметров
		if len(writeParamsPayloadMap) > 0 {
			fmt.Printf("Writing %d differing parameters to device %s...\n", len(writeParamsPayloadMap), fileConfig.Address)
			res, err = bk.sendLinCommand(devicePNumber, deviceAddressByte, CmdWrite, writeParamsPayloadBytes, DriverName)
			if err != nil {
				return fmt.Errorf("sync failed: write parameters command failed for address %s: %w", fileConfig.Address, err)
			}

			// Проверка ответа записи параметров
			successParamsMap := make(map[byte]bool)
			if len(res) >= 5 && len(res) == int(res[2])+5 {
				for i := 4; i < len(res)-1; i++ {
					successParamsMap[res[i]] = true
				}
			} else {
				return fmt.Errorf("sync failed: invalid write parameters response format or length from address %s (got %d bytes, len field %d)", fileConfig.Address, len(res), res[2])
			}

			for paramAddrByte := range writeParamsPayloadMap {
				if _, ok := successParamsMap[paramAddrByte]; !ok {
					return fmt.Errorf("sync failed: parameter 0x%X write failed (not confirmed in response)", paramAddrByte)
				}
			}
			fmt.Println("Parameters written for sync.")
		} else {
			fmt.Println("No parameters to write for sync.")
		}

		fmt.Printf("Device %s successfully synchronized with file %s.\n", fileConfig.Address, filePath)
	}

	return nil
}

// GetSettings Метод для получения параметров блока коммутации
func (bk *BK) GetSettings(devicePNumber string) ([]byte, error) {

	// 1. Загрузить конфиг из файла
	filePath := "bk/" + DriverName + "_" + devicePNumber + ".json"
	fileConfig, err := configBK.ReadConfig(filePath)
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

	deviceAddressByte := sens.StringToBytes(fileConfig.Address)

	// 2. Верификация устройства
	verifyPayload := []byte{DeviceNumber}
	res, err := bk.sendLinCommand(devicePNumber, deviceAddressByte, CmdReadInfo, verifyPayload, DriverName)
	if err != nil {
		return nil, fmt.Errorf("device verification command failed for address %s: %w", fileConfig.Address, err)
	}

	// Проверка ответа верификации
	if len(res) < 9 {
		return nil, fmt.Errorf("device verification response too short (expected >= 9, got %d) from address %s", len(res), fileConfig.Address)
	}
	deviceResponseStr := sens.ByteToHexString(res[6]) + sens.ByteToHexString(res[5])
	expectedDeviceName, ok := DeviceList[deviceResponseStr]
	if !ok {
		return nil, fmt.Errorf("unknown device response ID %s from address %s", deviceResponseStr, fileConfig.Address)
	}
	fmt.Printf("Device %s verified at address %s\n", expectedDeviceName, fileConfig.Address)

	// --- 3. Чтение с устройства ---
	// Параметры
	readValuesMap := make(map[byte]float32)
	paramsToRead := make([]byte, 0, len(fileConfig.SettingsParameters))
	paramKeysFromFile := []string{}
	if len(fileConfig.SettingsParameters) > 0 {
		fmt.Printf("Reading %d parameters from device...\n", len(fileConfig.SettingsParameters))
		for key := range fileConfig.SettingsParameters {
			paramsToRead = append(paramsToRead, sens.StringToBytes(key))
			paramKeysFromFile = append(paramKeysFromFile, key)
		}

		res, err = bk.sendLinCommand(devicePNumber, deviceAddressByte, CmdReadInfo, paramsToRead, DriverName)
		if err != nil {
			return nil, fmt.Errorf("initialization failed: read parameters command failed for address %s: %w", fileConfig.Address, err)
		} else {
			expectedDataLength := len(paramsToRead) * 4
			if len(res) < 5+expectedDataLength || int(res[2]) != expectedDataLength {
				fmt.Printf("Warning: read parameters response length mismatch (Addr: %s). Expected data len %d, got %d bytes in response, response len field %d. Attempting partial parse.\n",
					fileConfig.Address, expectedDataLength, len(res)-5, int(res[2]))
			}

			actualDataLength := len(res) - 1 // Индекс до CRC
			for i := 4; i < actualDataLength; i += 4 {
				if i+3 >= actualDataLength {
					fmt.Printf("Warning: incomplete parameter data at end of read response (Addr: %s)\n", fileConfig.Address)
					break
				}
				paramAddrByte := res[i]
				value, err := sens.Convert24BytesToFloat32(res[i+1:i+4], binary.LittleEndian)
				if err != nil {
					fmt.Printf("Warning: ConvertFloat32ToFloat32 error for param Addr 0x%X from device %s: %v\n", paramAddrByte, fileConfig.Address, err)
					continue
				}
				readValuesMap[paramAddrByte] = value
			}
			fmt.Println("Parameters read.")
		}
	} else {
		fmt.Println("No parameters defined in file, skipping parameter read.")
	}

	// Таблицы
	readTablesMap := make(map[string][]float32, len(fileConfig.Tables))
	if len(fileConfig.Tables) > 0 {
		fmt.Printf("Reading %d tables from device...\n", len(fileConfig.Tables))
		for keyTableHex := range fileConfig.Tables {
			fmt.Printf("Reading table %s...\n", keyTableHex)
			tableKeyByte := sens.StringToBytes(keyTableHex)

			readTablePayload := []byte{tableKeyByte, 0x00, 0x00, 0x36, 0x00}

			res, err = bk.sendLinCommand(devicePNumber, deviceAddressByte, 0x0A, readTablePayload, DriverName)
			if err != nil {
				fmt.Printf("Warning: read table %s command failed for address %s: %v. Skipping table.\n", keyTableHex, fileConfig.Address, err)
				continue
			}

			if len(res) < 5 {
				fmt.Printf("Warning: read table %s response is too short (%d bytes). Skipping table.\n", keyTableHex, len(res))
				continue
			}
			tableDataBytes := res[7 : len(res)-1]

			effectiveTableDataBytes := tableDataBytes
			terminatorIndex := -1
			for i := 0; i < len(tableDataBytes)-1; i++ {
				if tableDataBytes[i] == 0x00 && tableDataBytes[i+1] == 0x00 {
					terminatorIndex = i
					break
				}
			}

			if terminatorIndex != -1 {
				fmt.Printf("Info: Found 0x00 0x00 terminator at index %d in table %s data. Using data before terminator.\n", terminatorIndex, keyTableHex)
				effectiveTableDataBytes = tableDataBytes[:terminatorIndex]
			}

			decodedValues, err := decodeTableData(effectiveTableDataBytes)
			if err != nil {
				fmt.Printf("Warning: failed to decode data for table %s (Data: %X): %v. Skipping table.\n", keyTableHex, tableDataBytes, err)
				continue
			}
			readTablesMap[keyTableHex] = decodedValues
			fmt.Printf("Successfully read and processed table %s, stored %d values.\n", keyTableHex, len(decodedValues))
		}
		fmt.Println("Tables read.")
	} else {
		fmt.Println("No tables defined in file, skipping table read.")
	}

	// --- 4. Сохранение прочитанных данных в структуру SensBk2P ---
	currentSettings := configBK.SensBk2P{
		Address:            fileConfig.Address,
		SettingsParameters: make(map[string]configBK.SettingsParameter),
		Tables:             make(map[string]configBK.SettingsTables),
		DriverName:         DriverName,
		DevicePNumber:      devicePNumber,
	}

	// Заполнение параметров
	for key, paramInfo := range fileConfig.SettingsParameters {
		if value, ok := readValuesMap[sens.StringToBytes(key)]; ok {
			currentSettings.SettingsParameters[key] = configBK.SettingsParameter{
				Comment: paramInfo.Comment,
				Value:   value,
			}
		} else {
			fmt.Printf("Warning: Parameter %s not read from device.\n", key)
		}
	}

	// Заполнение таблиц
	for key, tableInfo := range fileConfig.Tables {
		if value, ok := readTablesMap[key]; ok {
			currentSettings.Tables[key] = configBK.SettingsTables{
				Comment: tableInfo.Comment,
				Value:   value,
			}
		} else {
			fmt.Printf("Warning: Table %s not read from device.\n", key)
		}
	}

	// --- 5. Маршалинг структуры в JSON ---
	jsonData, err := json.Marshal(currentSettings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings to JSON: %w", err)
	}

	return jsonData, nil
}

func (bk *BK) Ping(devicePNumber string) error {
	// 1. Загрузить конфиг из файла
	filePath := "bk/" + DriverName + "_" + devicePNumber + ".json"
	fileConfig, err := configBK.ReadConfig(filePath)
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

	deviceAddressByte := sens.StringToBytes(fileConfig.Address)

	// 2. Верификация устройства
	verifyPayload := []byte{DeviceNumber}
	res, err := bk.sendLinCommand(devicePNumber, deviceAddressByte, CmdReadInfo, verifyPayload, DriverName)
	if err != nil {
		return fmt.Errorf("device verification command failed for address %s: %w", fileConfig.Address, err)
	}

	// Проверка ответа верификации
	if len(res) < 9 {
		return fmt.Errorf("device verification response too short (expected >= 9, got %d) from address %s", len(res), fileConfig.Address)
	}
	deviceResponseStr := sens.ByteToHexString(res[6]) + sens.ByteToHexString(res[5])
	_, ok := DeviceList[deviceResponseStr]
	if !ok {
		return fmt.Errorf("unknown device response ID %s from address %s", deviceResponseStr, fileConfig.Address)
	}
	return nil
}
