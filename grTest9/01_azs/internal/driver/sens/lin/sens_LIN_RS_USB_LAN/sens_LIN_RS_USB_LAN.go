package sens_LIN_RS_USB_LAN

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"fuelazs/internal/driver/sens"
	"fuelazs/internal/driver/sens/lin/configLIN"
	"fuelazs/internal/driver/sens/sensAdapter"
	"path/filepath"
)

const (
	// DriverName Название устройства
	DriverName = "sens_LIN_RS_USB_LAN"
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
	"B887": DriverName,
	"B888": DriverName,
	"B889": DriverName,
	"B88A": DriverName,
	"B88B": DriverName,
	"B88C": DriverName,
	"B88D": DriverName,
	"B88E": DriverName,
	"B88F": DriverName,
	"B85E": DriverName,
}

type LinDriver struct {
	Adapter *map[string]*sensAdapter.SensAdapter
}

func NewLinDriver(adapter *map[string]*sensAdapter.SensAdapter) *LinDriver {
	return &LinDriver{
		Adapter: adapter,
	}
}

// SetSettings Метод для настройки ЛИН-АДАПТЕРА
func (l *LinDriver) SetSettings(devicePNumber string, data []byte) error {

	filePath := "lin/" + DriverName + "_" + devicePNumber + ".json"
	//filePath := filepath.Join("internal", "driver", "sens", "lin", "configLIN", fmt.Sprintf("sens_LIN_RS_USB_LAN_%s.json", devicePNumber))
	// --- Сценарий 1: Применение новых настроек (data != nil) ---
	if data != nil {
		var dataJSON configLIN.SensLinRsUsbLan
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
		var baseConfig *configLIN.SensLinRsUsbLan
		fileExists := sens.FileExists(filePath)

		if fileExists {
			baseConfig, err = configLIN.ReadConfig(filePath)
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

		// Вычисляем отличающиеся параметры для записи
		writingParamsMap := make(map[string]float32)
		for dataKey, dataParam := range dataJSON.SettingsParameters {
			baseValue, ok := baseConfig.SettingsParameters[dataKey]

			if !ok || baseValue.Value != dataParam.Value {
				writingParamsMap[dataKey] = dataParam.Value
			}
		}

		if len(writingParamsMap) == 0 {
			fmt.Println("No setting changes detected, skipping write.")
			return nil
		}

		// Подготовка к записи
		deviceAddressByte := sens.StringToBytes(baseConfig.Address)

		// 1. Убедиться, что по указанному адресу отвечает нужное устройство
		verifyPayload := []byte{DeviceNumber}
		res, err := l.sendLinCommand(devicePNumber, deviceAddressByte, CmdReadInfo, verifyPayload, DriverName)
		if err != nil {
			return fmt.Errorf("device verification command failed for address %s: %w", baseConfig.Address, err)
		}
		// Проверка ответа
		if len(res) < 9 {
			return fmt.Errorf("device verification response too short from address %s", dataJSON.Address)
		}
		deviceResponseStr := sens.ByteToHexString(res[6]) + sens.ByteToHexString(res[5])
		expectedDeviceName, ok := DeviceList[deviceResponseStr]
		if !ok {
			return fmt.Errorf("unknown device response ID %s from address %s", deviceResponseStr, baseConfig.Address)
		}

		fmt.Printf("Device %s verified at address %s\n", expectedDeviceName, baseConfig.Address)

		// 2. Записываем отличающиеся параметры
		const bytesPerParam = 4
		writePayload := make([]byte, 0, len(writingParamsMap)*bytesPerParam)
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
			writePayload = append(writePayload, paramAddrByte)
			writePayload = append(writePayload, buf...)
			paramAddressesToWrite = append(paramAddressesToWrite, paramHex)
		}

		res, err = l.sendLinCommand(devicePNumber, deviceAddressByte, CmdWrite, writePayload, DriverName)
		if err != nil {
			return fmt.Errorf("write parameters command failed for address %s: %w", baseConfig.Address, err)
		}

		// Проверка ответа записи
		successMap := make(map[byte]bool)
		if len(res) > 4 && len(res)-1 >= int(res[2])+4 {
			for i := 4; i < int(res[2])+4; i++ {
				successMap[res[i]] = true
			}
		} else {
			return fmt.Errorf("invalid write response format or length from address %s", baseConfig.Address)
		}

		for _, paramHex := range paramAddressesToWrite {
			paramAddrByte := sens.StringToBytes(paramHex)
			if _, ok := successMap[paramAddrByte]; !ok {
				return fmt.Errorf("parameter %s (0x%X) write failed (not confirmed in response)", paramHex, paramAddrByte)
			}
		}

		dataJSON.Address = baseConfig.Address

		// 3. Сохраняем успешную конфигурацию в файл
		err = sens.SaveConfig(filePath, dataJSON)
		if err != nil {
			return fmt.Errorf("SaveConfig error after writing parameters: %w", err)
		}

		fmt.Printf("Settings successfully written and saved to %s\n", filePath)
		return nil

	} else {
		// --- Сценарий 2: Синхронизация при запуске (data == nil) ---
		fmt.Printf("Starting settings synchronization for %s\n", devicePNumber)

		// 1. Загрузить конфиг из файла
		fileConfig, err := configLIN.ReadConfig(filePath)
		if err != nil {
			// Если файла нет при запуске - это критическая ошибка инициализации
			return fmt.Errorf("initialization failed, config file %s not found or invalid: %w", filePath, err)
		}

		deviceAddressByte := sens.StringToBytes(fileConfig.Address)

		// 2. Убедиться, что устройство на линии и соответствует конфигу
		verifyPayload := []byte{DeviceNumber}
		res, err := l.sendLinCommand(devicePNumber, deviceAddressByte, CmdReadInfo, verifyPayload, DriverName)
		if err != nil {
			return fmt.Errorf("initialization failed: device verification command failed for address %s: %w", fileConfig.Address, err)
		}
		// Проверка ответа
		if len(res) < 7 {
			return fmt.Errorf("initialization failed: device verification response too short from address %s", fileConfig.Address)
		}
		deviceResponseStr := sens.ByteToHexString(res[6]) + sens.ByteToHexString(res[5])
		expectedDeviceName, ok := DeviceList[deviceResponseStr]
		if !ok {
			return fmt.Errorf("initialization failed: unknown device response ID %s from address %s", deviceResponseStr, fileConfig.Address)
		}
		fmt.Printf("Device %s verified at address %s during init\n", expectedDeviceName, fileConfig.Address)

		// 3. Прочитать текущие значения параметров с устройства
		paramsToRead := make([]byte, 0, len(fileConfig.SettingsParameters))
		paramKeysFromFile := []string{}
		for key := range fileConfig.SettingsParameters {
			paramsToRead = append(paramsToRead, sens.StringToBytes(key))
			paramKeysFromFile = append(paramKeysFromFile, key)
		}

		if len(paramsToRead) == 0 {
			fmt.Println("No parameters configured in file, skipping read/sync.")
			return nil
		}

		res, err = l.sendLinCommand(devicePNumber, deviceAddressByte, CmdReadInfo, paramsToRead, DriverName)
		if err != nil {
			return fmt.Errorf("initialization failed: read parameters command failed for address %s: %w", fileConfig.Address, err)
		}

		// Парсинг ответа чтения
		readValuesMap := make(map[byte]float32) // Map[ParamAddrByte]Value
		expectedDataLength := len(paramsToRead) * 4
		if len(res) < 5+expectedDataLength || int(res[2]) != expectedDataLength {
			fmt.Printf("Warning: read parameters response length mismatch (Addr: %s). Expected data len %d, got %d bytes, response length field %d. Attempting partial parse.\n",
				fileConfig.Address, expectedDataLength, len(res)-5, int(res[2]))
		}

		// Итерация по данным ответа
		actualDataLength := len(res) - 1
		for i := 4; i < actualDataLength; i += 4 {
			if i+3 >= actualDataLength {
				fmt.Printf("Warning: incomplete parameter data at end of read response (Addr: %s)\n", fileConfig.Address)
				break
			}
			paramAddrByte := res[i]
			value, err := sens.Convert24BytesToFloat32(res[i+1:i+4], binary.LittleEndian)
			if err != nil {
				fmt.Printf("Warning: Convert24BytesToFloat32 error for param Addr 0x%X from device %s: %v\n", paramAddrByte, fileConfig.Address, err)
				continue
			}
			readValuesMap[paramAddrByte] = value
		}

		// 4. Сравнить значения из файла со значениями с устройства
		writePayloadMap := make(map[byte]float32)
		writePayloadBytes := make([]byte, 0)
		paramsNotRead := []string{}

		for _, paramKey := range paramKeysFromFile {
			paramAddrByte := sens.StringToBytes(paramKey)
			fileValue := fileConfig.SettingsParameters[paramKey].Value

			deviceValue, ok := readValuesMap[paramAddrByte]
			if !ok {
				paramsNotRead = append(paramsNotRead, paramKey)
				fmt.Printf("Warning: Parameter %s (0x%X) from file was not found in read response from device %s. Scheduling write.\n", paramKey, paramAddrByte, fileConfig.Address)
				writePayloadMap[paramAddrByte] = fileValue
				buf, err := sens.ConvertFloat32To24Bytes(fileValue, binary.LittleEndian)
				if err != nil {
					return fmt.Errorf("init failed: ConvertFloat32To24Bytes error for param %s: %w", paramKey, err)
				}
				writePayloadBytes = append(writePayloadBytes, paramAddrByte)
				writePayloadBytes = append(writePayloadBytes, buf...)
				continue
			}

			// Сравниваем значение из файла со значением с устройства
			if fileValue != deviceValue {
				fmt.Printf("Mismatch found for param %s (0x%X): File=%.4f, Device=%.4f. Scheduling write.\n", paramKey, paramAddrByte, fileValue, deviceValue)
				writePayloadMap[paramAddrByte] = fileValue
				buf, err := sens.ConvertFloat32To24Bytes(fileValue, binary.LittleEndian)
				if err != nil {
					return fmt.Errorf("init failed: ConvertFloat32To24Bytes error for param %s: %w", paramKey, err)
				}
				writePayloadBytes = append(writePayloadBytes, paramAddrByte)
				writePayloadBytes = append(writePayloadBytes, buf...)
			}
		}

		if len(paramsNotRead) > 0 {
			fmt.Printf("Warning: Could not read %d parameters from device %s: %v\n", len(paramsNotRead), fileConfig.Address, paramsNotRead)
		}

		// 5. Записать отличающиеся параметры
		if len(writePayloadMap) == 0 {
			fmt.Printf("Device %s settings are synchronized with file %s.\n", fileConfig.Address, filePath)
			return nil
		}

		fmt.Printf("Writing %d differing parameters to device %s to match file %s\n", len(writePayloadMap), fileConfig.Address, filePath)
		res, err = l.sendLinCommand(devicePNumber, deviceAddressByte, CmdWrite, writePayloadBytes, DriverName)
		if err != nil {
			return fmt.Errorf("initialization sync failed: write parameters command failed for address %s: %w", fileConfig.Address, err)
		}

		// Проверка ответа записи
		successMap := make(map[byte]bool)
		if len(res) > 4 && len(res)-1 >= int(res[2])+4 {
			for i := 4; i < int(res[2])+4; i++ {
				successMap[res[i]] = true
			}
		} else {
			return fmt.Errorf("initialization sync failed: invalid write response format or length from address %s", fileConfig.Address)
		}

		for paramAddrByte := range writePayloadMap {
			if _, ok := successMap[paramAddrByte]; !ok {
				return fmt.Errorf("initialization sync failed: parameter 0x%X write failed (not confirmed in response)", paramAddrByte)
			}
		}

		fmt.Printf("Device %s successfully synchronized with file %s.\n", fileConfig.Address, filePath)
	}

	return nil
}

// GetSettings Метод для получения настроек ЛИН-АДАПТЕРА
func (l *LinDriver) GetSettings(devicePNumber string) ([]byte, error) {
	// --- Шаг 1: Загрузить конфиг  ---
	filePath := "lin/" + DriverName + "_" + devicePNumber + ".json"
	deviceConfig, err := configLIN.ReadConfig(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings for %s: configuration file %s not found or invalid: %w", devicePNumber, filePath, err)
	}

	// Проверка базовых полей конфига
	if deviceConfig.DriverName != DriverName {
		return nil, fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, deviceConfig.DriverName)
	}
	if deviceConfig.DevicePNumber != devicePNumber {
		return nil, fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, deviceConfig.DevicePNumber)
	}

	deviceAddressByte := sens.StringToBytes(deviceConfig.Address)

	// --- Шаг 2: Верификация устройства на линии по его адресу ---
	verifyPayload := []byte{0xF2}
	res, err := l.sendLinCommand(devicePNumber, deviceAddressByte, CmdReadInfo, verifyPayload, DriverName)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings for %s: device verification command failed for address %s: %w", devicePNumber, deviceConfig.Address, err)
	}
	if len(res) < 7 {
		return nil, fmt.Errorf("failed to get settings for %s: device verification response too short from address %s", devicePNumber, deviceConfig.Address)
	}
	deviceResponseStr := sens.ByteToHexString(res[6]) + sens.ByteToHexString(res[5])
	verifiedDeviceName, ok := DeviceList[deviceResponseStr]
	if !ok {
		return nil, fmt.Errorf("failed to get settings for %s: unknown device response ID %s from address %s", devicePNumber, deviceResponseStr, deviceConfig.Address)
	}
	fmt.Printf("Device %s verified at address %s for GetSettings\n", verifiedDeviceName, deviceConfig.Address)

	// --- Шаг 3: Чтение параметров с устройства ---
	paramsToRead := make([]byte, 0, len(deviceConfig.SettingsParameters))
	paramKeyMap := make(map[byte]string)

	if len(deviceConfig.SettingsParameters) == 0 {
		fmt.Printf("Warning: No parameters defined in config file %s for device %s. Returning minimal config.\n", filePath, devicePNumber)
	} else {
		for keyHex := range deviceConfig.SettingsParameters {
			addrByte := sens.StringToBytes(keyHex)
			paramsToRead = append(paramsToRead, addrByte)
			paramKeyMap[addrByte] = keyHex
		}
	}

	if len(paramsToRead) == 0 {
		response := *deviceConfig
		responseJSON, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal empty config for %s: %w", devicePNumber, err)
		}
		return responseJSON, nil
	}

	res, err = l.sendLinCommand(devicePNumber, deviceAddressByte, CmdReadInfo, paramsToRead, DriverName)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings for %s: read parameters command failed for address %s: %w", devicePNumber, deviceConfig.Address, err)
	}

	// --- Шаг 4: Парсинг ответа и формирование результата ---
	readValuesMap := make(map[string]configLIN.SettingsParameter) // Map[ParamHexKey]ParameterData

	// Проверка длины ответа
	expectedDataLength := len(paramsToRead) * 4
	if len(res) < 5+expectedDataLength || int(res[2]) != expectedDataLength {
		fmt.Printf("Warning: read parameters response length mismatch for %s (Addr: %s). Expected data len %d, got response len %d, response length field %d. Attempting partial parse.\n",
			devicePNumber, deviceConfig.Address, expectedDataLength, len(res), int(res[2]))
	}

	actualDataLength := len(res) - 1
	parsedCount := 0
	for i := 4; i < actualDataLength; i += 4 {
		if i+3 >= actualDataLength {
			fmt.Printf("Warning: incomplete parameter data at end of read response for %s (Addr: %s)\n", devicePNumber, deviceConfig.Address)
			break
		}
		paramAddrByte := res[i]
		value, err := sens.Convert24BytesToFloat32(res[i+1:i+4], binary.LittleEndian)
		if err != nil {
			fmt.Printf("Warning: Convert24BytesToFloat32 error for param Addr 0x%X from device %s: %v. Skipping parameter.\n", paramAddrByte, deviceConfig.Address, err)
			continue
		}

		paramKeyHex, ok := paramKeyMap[paramAddrByte]
		if !ok {
			fmt.Printf("Warning: Received data for unexpected parameter address 0x%X from device %s. Ignoring.\n", paramAddrByte, deviceConfig.Address)
			continue
		}

		originalParamData, _ := deviceConfig.SettingsParameters[paramKeyHex]

		// Сохраняем прочитанное значение и метаданные
		readValuesMap[paramKeyHex] = configLIN.SettingsParameter{
			Comment: originalParamData.Comment, // Сохраняем комментарий из файла
			Value:   value,                     // Используем значение с устройства
		}
		parsedCount++
	}
	fmt.Printf("Successfully parsed %d parameters from device %s\n", parsedCount, deviceConfig.Address)

	// Проверяем, все ли запрошенные параметры были получены
	if parsedCount != len(paramsToRead) {
		fmt.Printf("Warning: Requested %d parameters, but successfully parsed only %d for device %s\n", len(paramsToRead), parsedCount, deviceConfig.Address)
	}

	// Формируем финальную структуру ответа
	response := configLIN.SensLinRsUsbLan{
		DriverName:         DriverName,
		DevicePNumber:      devicePNumber,        // Используем запрошенный PNumber
		Address:            deviceConfig.Address, // Используем адрес из файла
		SettingsParameters: readValuesMap,        // Используем карту прочитанных параметров
	}

	// --- Шаг 5: Маршалинг в JSON ---
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response JSON for %s: %w", devicePNumber, err)
	}

	return responseJSON, nil
}

func (l *LinDriver) Ping(devicePNumber string) error {
	// --- Шаг 1: Загрузить конфиг  ---
	//filePath := "lin/" + DriverName + "_" + devicePNumber + ".json"
	filePath := filepath.Join("internal", "driver", "sens", "lin", "configLIN", fmt.Sprintf("sens_LIN_RS_USB_LAN_%s.json", devicePNumber))
	deviceConfig, err := configLIN.ReadConfig(filePath)
	if err != nil {
		return fmt.Errorf("failed to get settings for %s: configuration file %s not found or invalid: %w", devicePNumber, filePath, err)
	}

	// Проверка базовых полей конфига
	if deviceConfig.DriverName != DriverName {
		return fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, deviceConfig.DriverName)
	}
	if deviceConfig.DevicePNumber != devicePNumber {
		return fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, deviceConfig.DevicePNumber)
	}

	deviceAddressByte := sens.StringToBytes(deviceConfig.Address)

	// --- Шаг 2: Верификация устройства на линии по его адресу ---
	verifyPayload := []byte{0xF2}
	res, err := l.sendLinCommand(devicePNumber, deviceAddressByte, CmdReadInfo, verifyPayload, DriverName)
	if err != nil {
		return fmt.Errorf("failed to get settings for %s: device verification command failed for address %s: %w", devicePNumber, deviceConfig.Address, err)
	}
	if len(res) < 7 {
		return fmt.Errorf("failed to get settings for %s: device verification response too short from address %s", devicePNumber, deviceConfig.Address)
	}
	deviceResponseStr := sens.ByteToHexString(res[6]) + sens.ByteToHexString(res[5])
	_, ok := DeviceList[deviceResponseStr]
	if !ok {
		return fmt.Errorf("failed to get settings for %s: unknown device response ID %s from address %s", devicePNumber, deviceResponseStr, deviceConfig.Address)
	}

	return nil
}
