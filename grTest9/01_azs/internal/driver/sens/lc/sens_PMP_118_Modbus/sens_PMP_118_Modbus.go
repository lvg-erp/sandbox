package sens_PMP_118_Modbus

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"path/filepath"

	"fuelazs/internal/driver/sens"
	"fuelazs/internal/driver/sens/lc/configLC"
	"fuelazs/internal/driver/sens/sensAdapter"
)

const (
	// DriverName Название устройства
	DriverName = "sens_PMP_118_Modbus"
	// SyncByte Стартовый байт
	SyncByte byte = 0xB5
	// ZeroByte Служебный байт
	ZeroByte byte = 0x00
	// CmdReadInfo Команда для чтения параметров
	CmdReadInfo byte = 0x0F
	// CmdWrite Команда для записи параметров
	CmdWrite byte = 0x11
	// CmdReadTable Команда для чтения таблиц
	CmdReadTable byte = 0x0A
	// CmdWriteTable Команда для записи таблиц
	CmdWriteTable byte = 0x1A
	// DeviceNumber Адрес номера устройства
	DeviceNumber byte = 0xF2

	H  = "0x01"
	T  = "0x02"
	Pr = "0x03"
	U  = "0x04"
	G  = "0x05"
	R  = "0x06"
	U1 = "0x07"
	H2 = "0x08"
	Ut = "0x11"
	Rt = "0x13"
	Ri = "0x12"
	Tr = "0x14"
	U2 = "0x15"
	Dg = "0x3D"
	Ts = "0x2D"
	Nt = "0xA7"
)

// DeviceList Список устройств, поддерживаемых драйвером
var DeviceList = map[string]string{
	"A260": DriverName,
	"A261": DriverName,
	"A262": DriverName,
	"A263": DriverName,
	"A264": DriverName,
	"A32E": DriverName,
}

type LCDriver struct {
	Adapter *map[string]*sensAdapter.SensAdapter
}

func NewLCDriver(adapter *map[string]*sensAdapter.SensAdapter) *LCDriver {
	return &LCDriver{
		Adapter: adapter,
	}
}

// GetMainStatus Получение основных параметров уровнемера
func (lc *LCDriver) GetMainStatus(devicePNumber string) ([]byte, error) {
	// 1. Загрузка конфига из файла
	filePath := filepath.Join("internal", "driver", "sens", "lc", "configLC", fmt.Sprintf("sens_PMP_118_Modbus_%s.json", devicePNumber))
	fileConfig, err := configLC.ReadConfig(filePath)
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
	res, err := lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdReadInfo}, verifyPayload, DriverName)
	if err != nil {
		return nil, fmt.Errorf("device verification command failed for address %s: %w", fileConfig.Address, err)
	}

	// Проверка ответа верификации
	if len(res) < 9 {
		return nil, fmt.Errorf("device verification response too short (expected >= 9, got %d) from address %s", len(res), fileConfig.Address)
	}
	deviceResponseStr := sens.ByteToHexString(res[6]) + sens.ByteToHexString(res[5])
	_, ok := DeviceList[deviceResponseStr]
	if !ok {
		return nil, fmt.Errorf("unknown device response ID %s from address %s", deviceResponseStr, fileConfig.Address)
	}

	// --- 3. Чтение с устройства ---
	// Основные параметры
	readValuesMap := make(map[string]float32)
	paramsToRead := make([]byte, 0, len(fileConfig.MainRead.Parameters))
	if len(fileConfig.MainRead.Parameters) > 0 {
		for keyHex := range fileConfig.MainRead.Parameters {
			paramsToRead = append(paramsToRead, sens.StringToBytes(keyHex))
		}

		res, err = lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdReadInfo}, paramsToRead, DriverName)
		if err != nil {
			return nil, fmt.Errorf("failed to read main parameters for address %s: %w", fileConfig.Address, err)
		}
		expectedDataLength := len(paramsToRead) * 4
		if len(res) < 5+expectedDataLength || int(res[2]) != expectedDataLength {
			fmt.Printf("Warning: read main parameters response length mismatch (Addr: %s). Expected data len %d, got %d bytes in response, response len field %d.\n",
				fileConfig.Address, expectedDataLength, len(res)-5, int(res[2]))
		}

		actualDataLength := len(res) - 1
		for i := 4; i < actualDataLength; i += 4 {
			if i+3 >= actualDataLength {
				break
			}
			paramAddrByte := res[i]
			value, err := sens.Convert24BytesToFloat32(res[i+1:i+4], binary.LittleEndian)
			if err != nil {
				fmt.Printf("Warning: Convert24BytesToFloat32 error for main param Addr 0x%X from device %s: %v\n", paramAddrByte, fileConfig.Address, err)
				continue
			}
			readValuesMap[sens.ByteToHexStringSimple(paramAddrByte)] = value
		}
	} else {
		fmt.Println("No main parameters defined in file, skipping read.")
	}

	// Таблицы
	readTablesMap := make(map[string][]float32, len(fileConfig.MainRead.Tables))
	if len(fileConfig.MainRead.Tables) > 0 {
		for keyTableHex := range fileConfig.MainRead.Tables {
			tableKeyByte := sens.StringToBytes(keyTableHex)
			readTablePayload := []byte{tableKeyByte, ZeroByte, ZeroByte, 0x36, ZeroByte}

			res, err = lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdReadTable}, readTablePayload, DriverName)
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
				if tableDataBytes[i] == ZeroByte && tableDataBytes[i+1] == ZeroByte {
					terminatorIndex = i
					break
				}
			}

			if terminatorIndex != -1 {
				effectiveTableDataBytes = tableDataBytes[:terminatorIndex]
			}

			var decodedValuesSlice []float32
			for i := 0; i < len(effectiveTableDataBytes); i += 3 {
				if i+3 > len(effectiveTableDataBytes) {
					break
				}
				decodedValues, err := sens.Convert24BytesToFloat32(effectiveTableDataBytes[i:i+3], binary.LittleEndian)
				if err != nil {
					fmt.Printf("Warning: Convert24BytesToFloat32 error for table %s: %v\n", keyTableHex, err)
				}
				decodedValuesSlice = append(decodedValuesSlice, decodedValues)
			}

			readTablesMap[keyTableHex] = decodedValuesSlice
		}
	} else {
		fmt.Println("No main tables defined in file, skipping read.")
	}

	// --- 4. Сохранение прочитанных данных в структуре
	mainParameters := SensPMP118ModbusMainStatus{
		Address:       fileConfig.Address,
		H:             readValuesMap[H],
		T:             readValuesMap[T],
		Pr:            readValuesMap[Pr],
		U:             readValuesMap[U],
		G:             readValuesMap[G],
		R:             readValuesMap[R],
		U1:            readValuesMap[U1],
		H2:            readValuesMap[H2],
		Ut:            readValuesMap[Ut],
		Rt:            readValuesMap[Rt],
		Ri:            readValuesMap[Ri],
		Tr:            readValuesMap[Tr],
		U2:            readValuesMap[U2],
		Dg:            readValuesMap[Dg],
		Ts:            readValuesMap[Ts],
		Nt:            readTablesMap[Nt],
		DriverName:    DriverName,
		DevicePNumber: devicePNumber,
	}

	// --- 5. Маршалинг структуры в JSON ---
	jsonData, err := json.Marshal(mainParameters)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings to JSON: %w", err)
	}

	return jsonData, nil
}

// GetMainParameters Получение основных параметров уровнемера
func (lc *LCDriver) GetMainParameters(devicePNumber string) error {
	filePath := filepath.Join("internal", "driver", "sens", "lc", "configLC", fmt.Sprintf("sens_PMP_118_Modbus_%s.json", devicePNumber))
	fileConfig, err := configLC.ReadConfig(filePath)
	if err != nil {
		return fmt.Errorf("failed to load config file %s: %w", filePath, err)
	}

	if fileConfig.DriverName != DriverName {
		return fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, fileConfig.DriverName)
	}
	if fileConfig.DevicePNumber != devicePNumber {
		return fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, fileConfig.DevicePNumber)
	}

	deviceAddressByte := sens.StringToBytes(fileConfig.Address)

	paramsToRead := make([]byte, 0, len(fileConfig.MainRead.Parameters))
	if len(fileConfig.MainRead.Parameters) > 0 {
		for keyHex := range fileConfig.MainRead.Parameters {
			paramsToRead = append(paramsToRead, sens.StringToBytes(keyHex))
		}

		err := lc.sendLCCommandWithoutRead(devicePNumber, deviceAddressByte, CmdReadInfo, paramsToRead, DriverName)
		if err != nil {
			return fmt.Errorf("failed to read main parameters for address %s: %w", fileConfig.Address, err)
		}
	} else {
		fmt.Println("No main parameters defined in file, skipping read.")
	}

	return nil
}

// GetTemperature Получение таблицы температур
func (lc *LCDriver) GetTemperature(devicePNumber string) error {
	filePath := filepath.Join("internal", "driver", "sens", "lc", "configLC", fmt.Sprintf("sens_PMP_118_Modbus_%s.json", devicePNumber))
	fileConfig, err := configLC.ReadConfig(filePath)
	if err != nil {
		return fmt.Errorf("failed to load config file %s: %w", filePath, err)
	}

	if fileConfig.DriverName != DriverName {
		return fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, fileConfig.DriverName)
	}
	if fileConfig.DevicePNumber != devicePNumber {
		return fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, fileConfig.DevicePNumber)
	}

	deviceAddressByte := sens.StringToBytes(fileConfig.Address)

	if len(fileConfig.MainRead.Tables) > 0 {
		for keyTableHex := range fileConfig.MainRead.Tables {
			tableKeyByte := sens.StringToBytes(keyTableHex)
			readTablePayload := []byte{tableKeyByte, ZeroByte, ZeroByte, 0x36, ZeroByte}

			err = lc.sendLCCommandWithoutRead(devicePNumber, deviceAddressByte, CmdReadTable, readTablePayload, DriverName)
			if err != nil {
				fmt.Printf("Warning: read table %s command failed for address %s: %v. Skipping table.\n", keyTableHex, fileConfig.Address, err)
				continue
			}
		}
	} else {
		fmt.Println("No main tables defined in file, skipping read.")
	}

	return nil
}

// GetOtherStatus Получение остальных параметров уровнемера
func (lc *LCDriver) GetOtherStatus(devicePNumber string) ([]byte, error) {
	filePath := filepath.Join("internal", "driver", "sens", "lc", "configLC", fmt.Sprintf("sens_PMP_118_Modbus_%s.json", devicePNumber))
	fileConfig, err := configLC.ReadConfig(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file %s: %w", filePath, err)
	}

	if fileConfig.DriverName != DriverName {
		return nil, fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, fileConfig.DriverName)
	}
	if fileConfig.DevicePNumber != devicePNumber {
		return nil, fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, fileConfig.DevicePNumber)
	}

	deviceAddressByte := sens.StringToBytes(fileConfig.Address)

	verifyPayload := []byte{DeviceNumber}
	res, err := lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdReadInfo}, verifyPayload, DriverName)
	if err != nil {
		return nil, fmt.Errorf("device verification command failed for address %s: %w", fileConfig.Address, err)
	}

	if len(res) < 9 {
		return nil, fmt.Errorf("device verification response too short (expected >= 9, got %d) from address %s", len(res), fileConfig.Address)
	}
	deviceResponseStr := sens.ByteToHexString(res[6]) + sens.ByteToHexString(res[5])
	expectedDeviceName, ok := DeviceList[deviceResponseStr]
	if !ok {
		return nil, fmt.Errorf("unknown device response ID %s from address %s", deviceResponseStr, fileConfig.Address)
	}
	fmt.Printf("Device %s verified at address %s\n", expectedDeviceName, fileConfig.Address)

	readValuesMap := make(map[string]float32)
	otherParams := fileConfig.OtherRead.Parameters
	if len(otherParams) > 0 {
		fmt.Printf("Reading %d other parameters from device (max 15 per command)...\n", len(otherParams))
		paramKeys := make([]string, 0, len(otherParams))
		for key := range otherParams {
			paramKeys = append(paramKeys, key)
		}

		for i := 0; i < len(paramKeys); i += 15 {
			batchSize := 15
			if i+batchSize > len(paramKeys) {
				batchSize = len(paramKeys) - i
			}
			paramsToRead := make([]byte, 0, batchSize)
			paramKeysBatch := paramKeys[i : i+batchSize]
			fmt.Printf("Reading batch of %d parameters: %v\n", batchSize, paramKeysBatch)

			for _, keyHex := range paramKeysBatch {
				paramsToRead = append(paramsToRead, sens.StringToBytes(keyHex))
			}

			res, err = lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdReadInfo}, paramsToRead, DriverName)
			if err != nil {
				fmt.Printf("failed to read other parameters batch for address %s: %w", fileConfig.Address, err)
				continue
			}
			expectedDataLength := len(paramsToRead) * 4
			if len(res) < 5+expectedDataLength || int(res[2]) != expectedDataLength {
				fmt.Printf("Warning: read other parameters response length mismatch (Addr: %s). Expected data len %d, got %d bytes in response, response len field %d.\n",
					fileConfig.Address, expectedDataLength, len(res)-5, int(res[2]))
			}

			actualDataLength := len(res) - 1
			for j := 0; j < len(paramsToRead); j++ {
				paramAddrByte := paramsToRead[j]
				responseIndex := 4 + j*4
				if responseIndex+3 < actualDataLength {
					value, err := sens.Convert24BytesToFloat32(res[responseIndex+1:responseIndex+4], binary.LittleEndian)
					if err != nil {
						fmt.Printf("Warning: Convert24BytesToFloat32 error for other param Addr 0x%X from device %s: %v\n", paramAddrByte, fileConfig.Address, err)
						continue
					}
					readValuesMap[sens.ByteToHexStringSimple(paramAddrByte)] = value
				} else {
					fmt.Printf("Warning: Incomplete response for other parameter 0x%X.\n", paramAddrByte)
				}
			}
		}
		fmt.Println("Other parameters read.")
	} else {
		fmt.Println("No other parameters defined in file, skipping read.")
	}

	readTablesMap := make(map[string][]float32, len(fileConfig.OtherRead.Tables))
	if len(fileConfig.OtherRead.Tables) > 0 {
		fmt.Printf("Reading %d tables from device...\n", len(fileConfig.OtherRead.Tables))
		for keyTableHex := range fileConfig.OtherRead.Tables {
			fmt.Printf("Reading main table %s...\n", keyTableHex)
			tableKeyByte := sens.StringToBytes(keyTableHex)
			readTablePayload := []byte{tableKeyByte, ZeroByte, ZeroByte, 0x36, ZeroByte}

			res, err = lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdReadTable}, readTablePayload, DriverName)
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

			var decodedValuesSlice []float32
			for i := 0; i < len(effectiveTableDataBytes); i += 3 {
				if i+3 > len(effectiveTableDataBytes) {
					break
				}
				decodedValues, err := sens.Convert24BytesToFloat32(effectiveTableDataBytes[i:i+3], binary.LittleEndian)
				if err != nil {
					fmt.Printf("Warning: Convert24BytesToFloat32 error for table %s: %v\n", keyTableHex, err)
				}
				decodedValuesSlice = append(decodedValuesSlice, decodedValues)
			}

			readTablesMap[keyTableHex] = decodedValuesSlice
			fmt.Printf("Successfully read and processed table %s, stored %d values.\n", keyTableHex, len(decodedValuesSlice))
		}
		fmt.Println("Other tables read.")
	} else {
		fmt.Println("No other tables defined in file, skipping read.")
	}

	response := SensPMP118ModbusOtherStatus{
		Address:         fileConfig.Address,
		OtherParameters: readValuesMap,
		OtherTables:     readTablesMap,
		DriverName:      DriverName,
		DevicePNumber:   devicePNumber,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal readValuesMap: %w", err)
	}

	err = sens.SaveConfig("LCDataOther.json", response)
	if err != nil {
		return nil, fmt.Errorf("failed to save main parameters config: %w", err)
	}

	return jsonData, nil
}

// GetSettings Метод для получения настроек уровнемера
func (lc *LCDriver) GetSettings(devicePNumber string) ([]byte, error) {
	filePath := filepath.Join("internal", "driver", "sens", "lc", "configLC", fmt.Sprintf("sens_PMP_118_Modbus_%s.json", devicePNumber))
	fileConfig, err := configLC.ReadConfig(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file %s: %w", filePath, err)
	}

	if fileConfig.DriverName != DriverName {
		return nil, fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, fileConfig.DriverName)
	}
	if fileConfig.DevicePNumber != devicePNumber {
		return nil, fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, fileConfig.DevicePNumber)
	}

	deviceAddressByte := sens.StringToBytes(fileConfig.Address)

	verifyPayload := []byte{DeviceNumber}
	res, err := lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdReadInfo}, verifyPayload, DriverName)
	if err != nil {
		return nil, fmt.Errorf("device verification command failed for address %s: %w", fileConfig.Address, err)
	}

	if len(res) < 9 {
		return nil, fmt.Errorf("device verification response too short (expected >= 9, got %d) from address %s", len(res), fileConfig.Address)
	}
	deviceResponseStr := sens.ByteToHexString(res[6]) + sens.ByteToHexString(res[5])
	expectedDeviceName, ok := DeviceList[deviceResponseStr]
	if !ok {
		return nil, fmt.Errorf("unknown device response ID %s from address %s", deviceResponseStr, fileConfig.Address)
	}
	fmt.Printf("Device %s verified at address %s\n", expectedDeviceName, fileConfig.Address)

	readValuesMap := make(map[string]float32)
	otherParams := fileConfig.SettingsParameters
	if len(otherParams) > 0 {
		fmt.Printf("Reading %d other parameters from device (max 15 per command)...\n", len(otherParams))
		paramKeys := make([]string, 0, len(otherParams))
		for key := range otherParams {
			paramKeys = append(paramKeys, key)
		}

		for i := 0; i < len(paramKeys); i += 15 {
			batchSize := 15
			if i+batchSize > len(paramKeys) {
				batchSize = len(paramKeys) - i
			}
			paramsToRead := make([]byte, 0, batchSize)
			paramKeysBatch := paramKeys[i : i+batchSize]
			fmt.Printf("Reading batch of %d parameters: %v\n", batchSize, paramKeysBatch)

			for _, keyHex := range paramKeysBatch {
				paramsToRead = append(paramsToRead, sens.StringToBytes(keyHex))
			}

			res, err = lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdReadInfo}, paramsToRead, DriverName)
			if err != nil {
				fmt.Printf("failed to read other parameters batch for address %s: %w", fileConfig.Address, err)
				continue
			}
			expectedDataLength := len(paramsToRead) * 4
			if len(res) < 5+expectedDataLength || int(res[2]) != expectedDataLength {
				fmt.Printf("Warning: read other parameters response length mismatch (Addr: %s). Expected data len %d, got %d bytes in response, response len field %d.\n",
					fileConfig.Address, expectedDataLength, len(res)-5, int(res[2]))
			}

			actualDataLength := len(res) - 1
			for j := 0; j < len(paramsToRead); j++ {
				paramAddrByte := paramsToRead[j]
				responseIndex := 4 + j*4
				if responseIndex+3 < actualDataLength {
					value, err := sens.Convert24BytesToFloat32(res[responseIndex+1:responseIndex+4], binary.LittleEndian)
					if err != nil {
						fmt.Printf("Warning: Convert24BytesToFloat32 error for other param Addr 0x%X from device %s: %v\n", paramAddrByte, fileConfig.Address, err)
						continue
					}
					readValuesMap[sens.ByteToHexStringSimple(paramAddrByte)] = value
				} else {
					fmt.Printf("Warning: Incomplete response for other parameter 0x%X.\n", paramAddrByte)
				}
			}
		}
		fmt.Println("Other parameters read.")
	} else {
		fmt.Println("No other parameters defined in file, skipping read.")
	}

	readTablesMap := make(map[string][]float32, len(fileConfig.Tables))
	if len(fileConfig.Tables) > 0 {
		fmt.Printf("Reading %d tables from device...\n", len(fileConfig.Tables))
		for keyTableHex := range fileConfig.Tables {
			fmt.Printf("Reading main table %s...\n", keyTableHex)
			tableKeyByte := sens.StringToBytes(keyTableHex)
			readTablePayload := []byte{tableKeyByte, ZeroByte, ZeroByte, 0x36, ZeroByte}

			res, err = lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdReadTable}, readTablePayload, DriverName)
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

			var decodedValuesSlice []float32
			for i := 0; i < len(effectiveTableDataBytes); i += 3 {
				if i+3 > len(effectiveTableDataBytes) {
					break
				}
				decodedValues, err := sens.Convert24BytesToFloat32(effectiveTableDataBytes[i:i+3], binary.LittleEndian)
				if err != nil {
					fmt.Printf("Warning: Convert24BytesToFloat32 error for table %s: %v\n", keyTableHex, err)
				}
				decodedValuesSlice = append(decodedValuesSlice, decodedValues)
			}

			readTablesMap[keyTableHex] = decodedValuesSlice
			fmt.Printf("Successfully read and processed table %s, stored %d values.\n", keyTableHex, len(decodedValuesSlice))
		}
		fmt.Println("Other tables read.")
	} else {
		fmt.Println("No other tables defined in file, skipping read.")
	}

	response := SensPMP118ModbusOtherStatus{
		Address:         fileConfig.Address,
		OtherParameters: readValuesMap,
		OtherTables:     readTablesMap,
		DriverName:      DriverName,
		DevicePNumber:   devicePNumber,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal readValuesMap: %w", err)
	}

	err = sens.SaveConfig("LCSettings.json", response)
	if err != nil {
		return nil, fmt.Errorf("failed to save main parameters config: %w", err)
	}

	return jsonData, nil
}

// SetSettings Метод для установки настроек уровнемера
func (lc *LCDriver) SetSettings(devicePNumber string, data []byte) error {
	filePath := filepath.Join("internal", "driver", "sens", "lc", "configLC", fmt.Sprintf("sens_PMP_118_Modbus_%s.json", devicePNumber))

	if data != nil {
		var dataJSON configLC.SensPMP118Modbus
		err := json.Unmarshal(data, &dataJSON)
		if err != nil {
			return fmt.Errorf("data.json parse error: %w", err)
		}

		if dataJSON.DriverName != DriverName {
			return fmt.Errorf("data.json driver name mismatch. Expect: %s, Have: %s", DriverName, dataJSON.DriverName)
		}
		if dataJSON.DevicePNumber != devicePNumber {
			return fmt.Errorf("data.json device number mismatch. Expect: %s, Have: %s", devicePNumber, dataJSON.DevicePNumber)
		}

		var baseConfig *configLC.SensPMP118Modbus
		fileExists := sens.FileExists(filePath)

		if fileExists {
			baseConfig, err = configLC.ReadConfig(filePath)
			if err != nil {
				return fmt.Errorf("failed to load existing config %s: %w", filePath, err)
			}
		} else {
			return fmt.Errorf("initialization failed, config file %s not found or invalid: %w", filePath, err)
		}

		if err := validateSettingsStructure(baseConfig, &dataJSON); err != nil {
			return fmt.Errorf("data.json structure validation failed: %w", err)
		}

		deviceAddressByte := sens.StringToBytes(baseConfig.Address)
		writingParamsMap := make(map[string]float32)
		for dataKey, dataParam := range dataJSON.SettingsParameters {
			baseValue, ok := baseConfig.SettingsParameters[dataKey]
			if !ok || baseValue.Value != dataParam.Value {
				writingParamsMap[dataKey] = dataParam.Value
			}
		}

		writingTablesMap := make(map[string][]float32)
		for dataKey, dataTable := range dataJSON.Tables {
			baseTable, ok := baseConfig.Tables[dataKey]
			if !ok || !sens.AreSlicesEqual(baseTable.Value, dataTable.Value) {
				writingTablesMap[dataKey] = dataTable.Value
			}
		}

		if len(writingParamsMap) == 0 && len(writingTablesMap) == 0 {
			fmt.Println("No setting or table changes detected, skipping write.")
			return nil
		}

		verifyPayload := []byte{DeviceNumber}
		res, err := lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdReadInfo}, verifyPayload, DriverName)
		if err != nil {
			return fmt.Errorf("device verification command failed for address %s: %w", dataJSON.Address, err)
		}

		if len(res) < 9 {
			return fmt.Errorf("device verification response too short (expected >= 9, got %d) from address %s", len(res), dataJSON.Address)
		}
		deviceResponseStr := sens.ByteToHexString(res[6]) + sens.ByteToHexString(res[5])
		expectedDeviceName, ok := DeviceList[deviceResponseStr]
		if !ok {
			return fmt.Errorf("unknown device response ID %s from address %s", deviceResponseStr, baseConfig.Address)
		}
		fmt.Printf("Device %s verified at address %s\n", expectedDeviceName, dataJSON.Address)

		if len(writingParamsMap) > 0 {
			fmt.Printf("Writing %d parameters in batches of 15...\n", len(writingParamsMap))
			paramKeysToWrite := make([]string, 0, len(writingParamsMap))
			for key := range writingParamsMap {
				paramKeysToWrite = append(paramKeysToWrite, key)
			}

			for i := 0; i < len(paramKeysToWrite); i += 15 {
				batchSize := 15
				if i+batchSize > len(paramKeysToWrite) {
					batchSize = len(paramKeysToWrite) - i
				}
				paramsToWriteBatch := paramKeysToWrite[i : i+batchSize]
				fmt.Printf("Writing batch of %d parameters: %v\n", batchSize, paramsToWriteBatch)

				writeParamsPayload := make([]byte, 0, batchSize*4)
				paramAddressesInBatch := make([]byte, 0, batchSize)

				for _, paramHex := range paramsToWriteBatch {
					paramValue := writingParamsMap[paramHex]
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
					paramAddressesInBatch = append(paramAddressesInBatch, paramAddrByte)
				}

				res, err = lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdWrite}, writeParamsPayload, DriverName)
				if err != nil {
					return fmt.Errorf("write parameters command failed for address %s (batch %d): %w", dataJSON.Address, i/15+1, err)
				}
				successParamsMap := make(map[byte]bool)
				if len(res) >= 5 && len(res) == int(res[2])+5 {
					for j := 4; j < len(res)-1; j++ {
						successParamsMap[res[j]] = true
					}
				} else {
					return fmt.Errorf("invalid write parameters response format or length from address %s (batch %d, got %d bytes, len field %d)", dataJSON.Address, i/15+1, len(res), res[2])
				}

				for _, addr := range paramAddressesInBatch {
					if _, ok := successParamsMap[addr]; !ok {
						paramHex := sens.ByteToHexString(addr)
						return fmt.Errorf("parameter %s (0x%X) write failed in batch %d (not confirmed in response)", paramHex, addr, i/15+1)
					}
				}
				fmt.Printf("Batch %d of parameters written successfully.\n", i/15+1)
			}
			fmt.Println("All parameters written successfully.")
		} else {
			fmt.Println("No parameters to write.")
		}

		if len(writingTablesMap) > 0 {
			fmt.Printf("Writing %d tables...\n", len(writingTablesMap))
			for tableKeyHex, tableValues := range writingTablesMap {
				encodeDataSlice := make([]byte, 0, len(tableValues)*3)
				fmt.Printf("Writing table %s for sync...\n", tableKeyHex)
				tableKeyByte := sens.StringToBytes(tableKeyHex)
				for _, tableValue := range tableValues {
					encodeData, err := sens.ConvertFloat32To24Bytes(tableValue, binary.LittleEndian)
					if err != nil {
						return fmt.Errorf("sync failed: failed to encode data for table %s: %w", tableKeyHex, err)
					}
					encodeDataSlice = append(encodeDataSlice, encodeData...)
				}
				writeTablePayload := []byte{tableKeyByte, ZeroByte, ZeroByte}
				writeTablePayload = append(writeTablePayload, encodeDataSlice...)

				res, err = lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdWriteTable}, writeTablePayload, DriverName)
				if err != nil {
					return fmt.Errorf("sync failed: write table %s command failed for address %s: %w", tableKeyHex, dataJSON.Address, err)
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

		dataJSON.MainRead = baseConfig.MainRead
		dataJSON.OtherRead = baseConfig.OtherRead
		dataJSON.Address = baseConfig.Address

		fmt.Printf("Saving configuration with potentially shorter tables back to %s...\n", filePath)
		err = sens.SaveConfig(filePath, dataJSON)
		if err != nil {
			return fmt.Errorf("settings/tables written to device, but SaveConfig failed for %s: %w", filePath, err)
		}

		fmt.Printf("Settings successfully written and saved to %s\n", filePath)
		return nil
	}

	// --- Сценарий 2: Синхронизация при запуске (data == nil) ---
	fmt.Printf("Starting settings synchronization for %s\n", devicePNumber)

	fileConfig, err := configLC.ReadConfig(filePath)
	if err != nil {
		return fmt.Errorf("initialization failed, config file %s not found or invalid: %w", filePath, err)
	}

	deviceAddressByte := sens.StringToBytes(fileConfig.Address)

	verifyPayload := []byte{DeviceNumber}
	res, err := lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdReadInfo}, verifyPayload, DriverName)
	if err != nil {
		return fmt.Errorf("initialization failed: device verification command failed for address %s: %w", fileConfig.Address, err)
	}
	if len(res) < 9 {
		return fmt.Errorf("initialization failed: device verification response too short from address %s", fileConfig.Address)
	}
	deviceResponseStr := sens.ByteToHexString(res[6]) + sens.ByteToHexString(res[5])
	expectedDeviceName, ok := DeviceList[deviceResponseStr]
	if !ok {
		return fmt.Errorf("initialization failed: unknown device response ID %s from address %s", deviceResponseStr, fileConfig.Address)
	}
	fmt.Printf("Device %s verified at address %s during init\n", expectedDeviceName, fileConfig.Address)

	readValuesMap := make(map[byte]float32)
	paramKeysFromFile := make([]string, 0, len(fileConfig.SettingsParameters))
	for key := range fileConfig.SettingsParameters {
		paramKeysFromFile = append(paramKeysFromFile, key)
	}

	fmt.Printf("Reading %d parameters from device in batches of 15...\n", len(paramKeysFromFile))
	for i := 0; i < len(paramKeysFromFile); i += 15 {
		batchSize := 15
		if i+batchSize > len(paramKeysFromFile) {
			batchSize = len(paramKeysFromFile) - i
		}
		paramsToReadBatch := paramKeysFromFile[i : i+batchSize]
		fmt.Printf("Reading batch %d of %d parameters: %v\n", i/15+1, len(paramKeysFromFile)/15+1, paramsToReadBatch)

		paramsToReadPayload := make([]byte, 0, batchSize)
		for _, key := range paramsToReadBatch {
			paramsToReadPayload = append(paramsToReadPayload, sens.StringToBytes(key))
		}

		res, err = lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdReadInfo}, paramsToReadPayload, DriverName)
		if err != nil {
			fmt.Printf("initialization failed: read parameters command failed for address %s (batch %d): %w", fileConfig.Address, i/15+1, err)
			continue
		}

		expectedDataLength := len(paramsToReadPayload) * 4
		if len(res) < 5+expectedDataLength || int(res[2]) != expectedDataLength {
			fmt.Printf("Warning: read parameters response length mismatch (Addr: %s, batch %d). Expected data len %d, got %d bytes in response, response len field %d.\n",
				fileConfig.Address, i/15+1, expectedDataLength, len(res)-5, int(res[2]))
		}

		actualDataLength := len(res) - 1
		for j := 0; j < len(paramsToReadPayload); j++ {
			paramKey := paramsToReadBatch[j]
			paramAddrByte := paramsToReadPayload[j]
			responseIndex := 4 + j*4
			if responseIndex+3 < actualDataLength {
				value, err := sens.Convert24BytesToFloat32(res[responseIndex+1:responseIndex+4], binary.LittleEndian)
				if err != nil {
					fmt.Printf("Warning: Convert24BytesToFloat32 error for param Addr 0x%X from device %s (batch %d): %v\n", paramAddrByte, fileConfig.Address, i/15+1, err)
					continue
				}
				readValuesMap[paramAddrByte] = value
			} else {
				fmt.Printf("Warning: Incomplete response for parameter %s (0x%X) in batch %d.\n", paramKey, paramAddrByte, i/15+1)
			}
		}
		fmt.Printf("Batch %d of parameters read.\n", i/15+1)
	}
	fmt.Println("All parameters read for sync.")

	readTablesMap := make(map[byte][]float32, len(fileConfig.Tables))
	if len(fileConfig.Tables) > 0 {
		fmt.Printf("Reading %d tables from device...\n", len(fileConfig.Tables))
		for keyTableHex := range fileConfig.Tables {
			fmt.Printf("Reading main table %s...\n", keyTableHex)
			tableKeyByte := sens.StringToBytes(keyTableHex)
			readTablePayload := []byte{tableKeyByte, ZeroByte, ZeroByte, 0x36, ZeroByte}

			res, err = lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdReadTable}, readTablePayload, DriverName)
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

			var decodedValuesSlice []float32
			for i := 0; i < len(effectiveTableDataBytes); i += 3 {
				if i+3 > len(effectiveTableDataBytes) {
					break
				}
				decodedValues, err := sens.Convert24BytesToFloat32(effectiveTableDataBytes[i:i+3], binary.LittleEndian)
				if err != nil {
					fmt.Printf("Warning: Convert24BytesToFloat32 error for table %s: %v\n", keyTableHex, err)
				}
				decodedValuesSlice = append(decodedValuesSlice, decodedValues)
			}

			readTablesMap[sens.StringToBytes(keyTableHex)] = decodedValuesSlice
			fmt.Printf("Successfully read and processed table %s, stored %d values.\n", keyTableHex, len(decodedValuesSlice))
		}
		fmt.Println("Other tables read.")
	} else {
		fmt.Println("No other tables defined in file, skipping read.")
	}

	writeParamsPayloadMap := make(map[byte]float32)
	paramsNotRead := []string{}

	for _, paramKey := range paramKeysFromFile {
		paramAddrByte := sens.StringToBytes(paramKey)
		fileValue := fileConfig.SettingsParameters[paramKey].Value

		deviceValue, ok := readValuesMap[paramAddrByte]
		if !ok {
			paramsNotRead = append(paramsNotRead, paramKey)
			fmt.Printf("Warning: Parameter %s (0x%X) from file was not found in read response from device %s. Scheduling write from file.\n", paramKey, paramAddrByte, fileConfig.Address)
			writeParamsPayloadMap[paramAddrByte] = fileValue
			continue
		}

		if fileValue != deviceValue {
			fmt.Printf("Param mismatch %s (0x%X): File=%.4f, Device=%.4f. Scheduling write.\n", paramKey, paramAddrByte, fileValue, deviceValue)
			writeParamsPayloadMap[paramAddrByte] = fileValue
		}
	}

	if len(paramsNotRead) > 0 {
		fmt.Printf("Warning: Could not read %d parameters from device %s: %v\n", len(paramsNotRead), fileConfig.Address, paramsNotRead)
	}

	writingTablesMap := make(map[string][]float32)
	for fileTableKey, fileTable := range fileConfig.Tables {
		readTable, tableExistedInBase := readTablesMap[sens.StringToBytes(fileTableKey)]
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

	if len(writeParamsPayloadMap) > 0 {
		fmt.Printf("Writing %d differing parameters to device %s in batches of 15...\n", len(writeParamsPayloadMap), fileConfig.Address)
		paramsToWrite := make(map[byte]float32)
		for addr, val := range writeParamsPayloadMap {
			paramsToWrite[addr] = val
		}

		var batch []byte
		batchCounter := 1
		for addr, val := range paramsToWrite {
			buf, err := sens.ConvertFloat32To24Bytes(val, binary.LittleEndian)
			if err != nil {
				return fmt.Errorf("sync failed: ConvertFloat32To24Bytes error for param 0x%X: %w", addr, err)
			}
			batch = append(batch, addr)
			batch = append(batch, buf...)

			if len(batch)/4 == 15 || len(paramsToWrite) == len(batch)/4 {
				fmt.Printf("Writing sync batch %d with %d parameters.\n", batchCounter, len(batch)/4)
				res, err = lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdWrite}, batch, DriverName)
				if err != nil {
					return fmt.Errorf("sync failed: write parameters command failed for address %s (batch %d): %w", fileConfig.Address, batchCounter, err)
				}

				successParamsMap := make(map[byte]bool)
				if len(res) >= 5 && len(res) == int(res[2])+5 {
					for i := 4; i < len(res)-1; i++ {
						successParamsMap[res[i]] = true
					}
				} else {
					return fmt.Errorf("sync failed: invalid write parameters response format or length from address %s (batch %d, got %d bytes, len field %d)", fileConfig.Address, batchCounter, len(res), res[2])
				}

				for i := 0; i < len(batch); i += 4 {
					paramAddr := batch[i]
					if _, ok := successParamsMap[paramAddr]; !ok {
						return fmt.Errorf("sync failed: parameter 0x%X write failed in batch %d (not confirmed in response)", paramAddr, batchCounter)
					}
				}
				batch = nil
				batchCounter++
			}
		}
		fmt.Println("Parameters written for sync.")
	} else {
		fmt.Println("No parameters to write for sync.")
	}

	if len(writingTablesMap) > 0 {
		fmt.Printf("Writing %d differing tables to device %s...\n", len(writingTablesMap), fileConfig.Address)
		for tableKeyHex, tableValues := range writingTablesMap {
			encodedDataSlice := make([]byte, 0, len(tableValues)*3)
			fmt.Printf("Writing table %s for sync...\n", tableKeyHex)
			tableKeyByte := sens.StringToBytes(tableKeyHex)

			for _, val := range tableValues {
				encodedData, err := sens.ConvertFloat32To24Bytes(val, binary.LittleEndian)
				if err != nil {
					return fmt.Errorf("sync failed: failed to encode data for table %s: %w", tableKeyHex, err)
				}
				encodedDataSlice = append(encodedDataSlice, encodedData...)
			}

			writeTablePayload := []byte{tableKeyByte, ZeroByte, ZeroByte}
			writeTablePayload = append(writeTablePayload, encodedDataSlice...)

			res, err = lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdWriteTable}, writeTablePayload, DriverName)
			if err != nil {
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

	fmt.Printf("Device %s successfully synchronized with file %s.\n", fileConfig.Address, filePath)
	return nil
}

// Ping Метод для проверки отвечает ли уровнемер
func (lc *LCDriver) Ping(devicePNumber string) error {
	filePath := filepath.Join("internal", "driver", "sens", "lc", "configLC", fmt.Sprintf("sens_PMP_118_Modbus_%s.json", devicePNumber))
	fileConfig, err := configLC.ReadConfig(filePath)
	if err != nil {
		return fmt.Errorf("failed to load config file %s: %w", filePath, err)
	}

	if fileConfig.DriverName != DriverName {
		return fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, fileConfig.DriverName)
	}
	if fileConfig.DevicePNumber != devicePNumber {
		return fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, fileConfig.DevicePNumber)
	}

	deviceAddressByte := sens.StringToBytes(fileConfig.Address)
	verifyPayload := []byte{DeviceNumber}
	res, err := lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdReadInfo}, verifyPayload, DriverName)
	if err != nil {
		return fmt.Errorf("device verification command failed for address %s: %w", fileConfig.Address, err)
	}

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

// GetFuelVolume Метод для получения объема топлива
func (lc *LCDriver) GetFuelVolume(devicePNumber string) (float32, error) {
	filePath := filepath.Join("internal", "driver", "sens", "lc", "configLC", fmt.Sprintf("sens_PMP_118_Modbus_%s.json", devicePNumber))
	fileConfig, err := configLC.ReadConfig(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to load config file %s: %w", filePath, err)
	}

	if fileConfig.DriverName != DriverName {
		return 0, fmt.Errorf("config file %s has wrong driver name: expected %s, got %s", filePath, DriverName, fileConfig.DriverName)
	}
	if fileConfig.DevicePNumber != devicePNumber {
		return 0, fmt.Errorf("config file %s has wrong device number: expected %s, got %s", filePath, devicePNumber, fileConfig.DevicePNumber)
	}

	deviceAddressByte := sens.StringToBytes(fileConfig.Address)

	verifyPayload := []byte{DeviceNumber}
	res, err := lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdReadInfo}, verifyPayload, DriverName)
	if err != nil {
		return 0, fmt.Errorf("device verification command failed for address %s: %w", fileConfig.Address, err)
	}

	if len(res) < 9 {
		return 0, fmt.Errorf("device verification response too short (expected >= 9, got %d) from address %s", len(res), fileConfig.Address)
	}
	deviceResponseStr := sens.ByteToHexString(res[6]) + sens.ByteToHexString(res[5])
	expectedDeviceName, ok := DeviceList[deviceResponseStr]
	if !ok {
		return 0, fmt.Errorf("unknown device response ID %s from address %s", deviceResponseStr, fileConfig.Address)
	}
	fmt.Printf("Device %s verified at address %s\n", expectedDeviceName, fileConfig.Address)

	res, err = lc.sendLinCommand(devicePNumber, []byte{deviceAddressByte}, []byte{CmdReadInfo}, []byte{0x04}, DriverName)
	if err != nil {
		return 0, fmt.Errorf("failed to verification command for address %s: %w", fileConfig.Address, err)
	}
	if len(res) < 9 {
		return 0, fmt.Errorf("failed to verification response too short from address %s", fileConfig.Address)
	}
	value, err := sens.Convert24BytesToFloat32(res[5:8], binary.LittleEndian)
	if err != nil {
		return 0, fmt.Errorf("failed to convert response to float32 from address %s: %w", fileConfig.Address, err)
	}
	if value == -1 {
		return 0, nil
	}
	return value, nil
}
