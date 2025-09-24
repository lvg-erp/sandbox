package usecase

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"fuelazs/internal/driver/sens"
	"fuelazs/internal/driver/sens/lc/sens_PMP_118_Modbus"
	"fuelazs/internal/integration"
	"fuelazs/internal/usecase/models"
	"strconv"
	"strings"
	"time"
)

const (
	UnknownNozzleStatus       = 98
	DriverProblemNozzleStatus = 97
)

func (p *Processing) Telemetry() {
	logger := p.logger.BaseError("telemetry")
	ticker := time.NewTicker(time.Duration(p.kazsConfig.Telemetry.SendTelemetry) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !p.kazsActivation || !p.startProgramm {
				continue
			}
			telemetry := p.generateTelemetryReport()

			logger.Info("POST Telemetry", "timeout_sec", p.appConfig.TelemetryTimeout)
			sendErr := RunWithTimeout(func() error {
				return p.kazsOperator.Telemetry(&telemetry)
			}, time.Duration(p.appConfig.TelemetryTimeout)*time.Second)
			if sendErr != nil {
				logger.Error("ошибка запроса.", "handler", "telemetry", "err", sendErr)
				p.CaptureTelemetryError(fmt.Errorf("send telemetry error: %w", sendErr), TelemetryErrorDep{
					handler:    "Telemetry",
					operation:  "kazsOperator.Telemetry",
					kazsNumber: p.kazsOperator.KazsNumber,
				})

				_ = p.repository.ErrorLogs.InsertError(models.Errors{
					Time:    time.Now().Unix(),
					Handler: "Telemetry",
					Id:      "",
					Error:   fmt.Sprintf("send telemetry error: %v", sendErr),
				})

				// Записываем в локальную БД
				insertErr := p.repository.Telemetry.InsertTelemetry(&telemetry)
				if insertErr != nil {
					p.CaptureTelemetryError(fmt.Errorf("insert error after send failure: %v", insertErr), TelemetryErrorDep{
						handler:    "Repo.InsertTelemetry",
						operation:  "Telemetry",
						kazsNumber: p.kazsOperator.KazsNumber,
						jarNumber:  "",
					})

					_ = p.repository.ErrorLogs.InsertError(models.Errors{
						Time:    time.Now().Unix(),
						Handler: "Telemetry",
						Id:      "",
						Error:   fmt.Sprintf("insert error after send failure: %v", insertErr),
					})

				}
			} else {
				logger.Info("телеметрия успешно отправлена.")
			}
		}
	}
}

func (p *Processing) TelemetryHistory() {
	logger := p.logger.BaseError("telemetry_history")
	ticker := time.NewTicker(time.Duration(p.kazsConfig.Telemetry.TelemetryHistory) * time.Second)
	defer ticker.Stop()
	var sendFailureCount int
	var maxSendFailureCount = p.kazsConfig.Telemetry.MaxFailureCount
	for range ticker.C {
		if !p.kazsActivation || !p.startProgramm {
			continue
		}

		logger.Info("запущен процесс отправки неотправленной телеметрии.")

		sendFailureCount = 0

		// Вспомогательная функция для повторяющейся обработки ошибок
		handleAndLogError := func(err error, op, jar string) {
			if err != nil {
				telemetryDep := TelemetryErrorDep{
					handler:    "Telemetry",
					operation:  op,
					kazsNumber: p.kazsOperator.KazsNumber,
					jarNumber:  jar,
				}
				p.CaptureTelemetryError(err, telemetryDep)

				_ = p.repository.ErrorLogs.InsertError(models.Errors{
					Time:    time.Now().Unix(),
					Handler: "Telemetry",
					Id:      "",
					Error:   fmt.Sprintf("%s error for jar %s: %v", op, jar, err),
				})

			}
		}

		reportCount, err := p.repository.Telemetry.GetTelemetryRows()
		if err != nil {
			logger.Error("ошибка получения неотправленной телеметрии из БД.", "err", err)
			handleAndLogError(err, "Repo.GetTelemetryRows", "")
			continue
		}

		if reportCount == 0 {
			continue
		}

		logger.Warn("обнаружена неотправленная телеметрия.", "report_count", reportCount)

		for i := int64(1); i <= reportCount; i++ {
			reports, err := p.repository.Telemetry.GetAllTelemetry(1)
			if err != nil {
				logger.Error("ошибка получения неотправленной телеметрии из БД.", "err", err)
				handleAndLogError(err, "Repo.GetAllTelemetry", "")
				break
			}

			if sendFailureCount >= maxSendFailureCount {
				logger.Warn("превышено число попыток отправить телеметрию.", "failure_count", sendFailureCount)
				break
			}

			if len(reports) == 0 {
				break
			}

			logger.Info("POST Telemetry", "timeout_sec", p.appConfig.TelemetryTimeout)
			sendErr := RunWithTimeout(func() error {
				return p.kazsOperator.Telemetry(&reports[0])
			}, time.Duration(p.appConfig.TelemetryTimeout)*time.Second)

			if sendErr != nil {
				logger.Error("ошибка запроса.", "handler", "Telemetry", "err", sendErr)
				sendFailureCount++
				handleAndLogError(sendErr, "kazsOperator.Telemetry", "")
				continue
			}

			// Удаляем из локальной БД одну запись телеметрии
			err = p.repository.Telemetry.DeleteOneTelemetry(reports[0].StatusTime)
			if err != nil {
				logger.Error("не удалось удалить телеметрию из БД.", "err", err)
				handleAndLogError(err, "Repo.DeleteOneTelemetry", "")
			}
			time.Sleep(10 * time.Millisecond)
		}
		logger.Info("процесс отправки неотправленной телеметрии завершен.")
	}
}

func (p *Processing) GetTelemetry() {
	logger := p.logger.BaseError("get_sens-telemetry")
	ticker := time.NewTicker(time.Duration(p.kazsConfig.Telemetry.GetTelemetry) * time.Second)
	defer ticker.Stop()
	captureError := CaptureErrorDep{
		operation:  "GetTelemetry",
		kazsNumber: p.kazsOperator.KazsNumber,
	}

	for range ticker.C {
		if !p.kazsActivation || !p.startProgramm {
			continue
		}

		// Получаем основные параметры для первой емкости
		p.driver.MuSENS.Lock()
		err := RunWithTimeout(func() error {
			return p.driver.SensDriver.LCDriver.GetMainParameters("1")
		}, time.Duration(p.driverConfig.GetMainStatusTimeout)*time.Second)
		p.driver.MuSENS.Unlock()
		captureError.jarNumber = "1"
		if err != nil {
			logger.Error("не удалось получить основные параметры.", "jar_id", "1", "err", err)
			p.CaptureError(fmt.Errorf("LCDriver.GetMainParameters error: %w", err), captureError)
		}
		time.Sleep(500 * time.Millisecond)

		// Получаем основные параметры для второй емкости
		p.driver.MuSENS.Lock()
		err = RunWithTimeout(func() error {
			return p.driver.SensDriver.LCDriver.GetMainParameters("2")
		}, time.Duration(p.driverConfig.GetMainStatusTimeout)*time.Second)
		p.driver.MuSENS.Unlock()
		captureError.jarNumber = "2"
		if err != nil {
			logger.Error("не удалось получить основные параметры.", "jar_id", "2", "err", err)
			p.CaptureError(fmt.Errorf("LCDriver.GetMainParameters error: %w", err), captureError)
		}

		time.Sleep(p.kazsConfig.Telemetry.GetTelemetryJarTimeout * time.Second)

		// Получаем таблицу температур для первой емкости
		p.driver.MuSENS.Lock()
		err = RunWithTimeout(func() error {
			return p.driver.SensDriver.LCDriver.GetTemperature("1")
		}, time.Duration(p.driverConfig.GetMainStatusTimeout)*time.Second)
		p.driver.MuSENS.Unlock()
		captureError.jarNumber = "1"
		if err != nil {
			logger.Error("не удалось получить таблицу температур.", "jar_id", "1", "err", err)
			p.CaptureError(fmt.Errorf("LCDriver.GetMainParameters error: %w", err), captureError)
		}
		time.Sleep(500 * time.Millisecond)

		// Получаем таблицу температур для второй емкости
		p.driver.MuSENS.Lock()
		err = RunWithTimeout(func() error {
			return p.driver.SensDriver.LCDriver.GetTemperature("2")
		}, time.Duration(p.driverConfig.GetMainStatusTimeout)*time.Second)
		p.driver.MuSENS.Unlock()
		captureError.jarNumber = "2"
		if err != nil {
			logger.Error("не удалось получить таблицу температур.", "jar_id", "2", "err", err)
			p.CaptureError(fmt.Errorf("LCDriver.GetMainParameters error: %w", err), captureError)
		}
	}

}

func (p *Processing) ReadTelemetry() {
	logger := p.logger.BaseError("read_sens_telemetry")
	captureError := CaptureErrorDep{
		operation:  "ReadTelemetry",
		kazsNumber: p.kazsOperator.KazsNumber,
	}
	if !(*p.driver.SensDriver.LCDriver.Adapter)["1"].IsOpen() {
		logger.Warn("порт закрыт.")
		p.CaptureError(fmt.Errorf("LCDriver.Adapter isNotOpen"), captureError)
		return
	}

	buf := make([]byte, 128)

	for {
		if !p.kazsActivation || !p.startProgramm {
			continue
		}
		n, err := (*p.driver.SensDriver.LCDriver.Adapter)["1"].Port.Read(buf)
		if err != nil {
			logger.Error("ошибка чтения порта.", "err", err)
			p.CaptureError(fmt.Errorf("LCDriver.Port.Read error: %w", err), captureError)
			err = (*p.driver.SensDriver.LCDriver.Adapter)["1"].Reopen()
			if err != nil {
				p.CaptureError(fmt.Errorf("LCDriver.Reopen error: %w", err), captureError)
			}
			continue
		}

		if n > 0 {
			p.ProcessIncomingData(buf[:n])
		}
	}
}

func (p *Processing) GetTRKTelemetry() {
	logger := p.logger.BaseError("get_trk_telemetry")
	captureError := CaptureErrorDep{
		operation:  "GetTRKTelemetry",
		kazsNumber: p.kazsOperator.KazsNumber,
	}

	for {
		if !p.kazsActivation || !p.startProgramm {
			time.Sleep(1 * time.Second)
			continue
		}

		for i := range p.driver.TopazDriver.Adapter {
			trkStatusRes := p.TRKRequest(GetTRKStatus, i, 0, logger)

			var currentTRKStatus int
			if trkStatusRes.Err != nil {
				if errors.Is(trkStatusRes.Err, context.DeadlineExceeded) {
					currentTRKStatus = UnknownNozzleStatus
				} else {
					currentTRKStatus = DriverProblemNozzleStatus
				}
				p.CaptureError(fmt.Errorf("TopazDriver.GetTRKStatus error: %w", trkStatusRes.Err), captureError)
			} else {
				currentTRKStatus = nozzleStatus(trkStatusRes.ValueStr)
			}
			p.UpdateLastTRKTelemetry(i, currentTRKStatus)
		}
		time.Sleep(p.kazsConfig.Telemetry.GetTRKTelemetry * time.Second)
	}
}

func Search(data []byte) []int {
	var res []int
	messageMinLength := 4
	for i := 0; i <= len(data)-messageMinLength; i++ {
		if data[i] == 0xB5 {
			if data[i+1] == 0x01 || data[i+1] == 0x02 {
				if data[i+3] == 0x8F || data[i+3] == 0x8A {
					res = append(res, i)
				}
			}
		}
	}
	return res
}

func MessageValidation(data []byte) ([]byte, bool) {
	if len(data) < 5 { // Минимальная длина сообщения для проверки data[2] и CRC
		return nil, false
	}

	expectedLength := int(data[2]) + 5
	if len(data) < expectedLength {
		// Сообщение неполное, нужно дождаться больше байтов
		return nil, false
	}
	if len(data) > expectedLength {
		data = data[:expectedLength]
	}

	expectedCRC := (data)[len(data)-1]
	realCRC := sens.SENSCalculateCRC(data[:len(data)-1])

	if expectedCRC == realCRC {
		return data, true
	} else {
		return nil, false
	}
}

func (p *Processing) ProcessIncomingData(data []byte) {
	// 1. Добавляем новые байты в буфер
	p.telemetryBuffer = append(p.telemetryBuffer, data...)

	// 2. Цикл для обработки всех возможных сообщений в буфере
	for {
		// 3. Поиск первого сообщения
		starts := Search(p.telemetryBuffer)

		if len(starts) == 0 {
			break
		}

		// Берем первое найденное начало сообщения
		firstMessageStart := starts[0]

		// 4. Отбрасывание мусора
		if firstMessageStart > 0 {
			p.telemetryBuffer = p.telemetryBuffer[firstMessageStart:]
			starts = Search(p.telemetryBuffer)
			if len(starts) == 0 || starts[0] != 0 {
				break
			}
		}

		// 5. Проверка полной длины (минимум 3 байта для B5, Type, Length)
		if len(p.telemetryBuffer) < 3 {
			break
		}

		// 6. Вычисляем ожидаемую полную длину сообщения (data[2] + 5)
		expectedFullMessageLength := int(p.telemetryBuffer[2]) + 5

		// 7. Проверяем, достаточно ли байтов в буфере для полного сообщения
		if len(p.telemetryBuffer) < expectedFullMessageLength {
			break
		}

		// 8. Извлекаем потенциальное сообщение
		potentialMessage := p.telemetryBuffer[:expectedFullMessageLength]

		// 9. Валидируем сообщение
		validatedMessage, isValid := MessageValidation(potentialMessage)

		if isValid {
			// 10. Удаляем обработанное сообщение из буфера
			p.telemetryBuffer = p.telemetryBuffer[expectedFullMessageLength:]
			err := p.validationSENSMessage(validatedMessage)
			if err != nil {
				p.logger.Error(err.Error())
			}
			// 11. Продолжаем цикл, чтобы проверить, есть ли еще сообщения в оставшемся буфере
		} else {
			// Если сообщение невалидно, отбрасываем его и ищем следующее начало
			p.telemetryBuffer = p.telemetryBuffer[expectedFullMessageLength:]
			// Продолжаем цикл, чтобы найти следующее сообщение
		}
	}
}

func (p *Processing) validationSENSMessage(data []byte) error {
	readValuesMap := make(map[string]float32)
	jarNumberInt := int(data[1])
	jarNumberStr := strconv.Itoa(jarNumberInt)

	if data[3] == 0x8F {
		for i := 4; i < len(data)-1; i += 4 {
			if i+3 > len(data)-1 {
				break
			}
			paramAddressByte := data[i]
			value, err := sens.Convert24BytesToFloat32(data[i+1:i+4], binary.LittleEndian)
			if err != nil {
				continue
			}
			readValuesMap[sens.ByteToHexStringSimple(paramAddressByte)] = value
		}

		lastTime := time.Now().Unix()
		mainParameters := SENSStatus{
			timeAS: lastTime,
			H:      readValuesMap[sens_PMP_118_Modbus.H],
			T:      readValuesMap[sens_PMP_118_Modbus.T],
			Pr:     readValuesMap[sens_PMP_118_Modbus.Pr],
			U:      readValuesMap[sens_PMP_118_Modbus.U],
			G:      readValuesMap[sens_PMP_118_Modbus.G],
			R:      readValuesMap[sens_PMP_118_Modbus.R],
			U1:     readValuesMap[sens_PMP_118_Modbus.U1],
			H2:     readValuesMap[sens_PMP_118_Modbus.H2],
			Ut:     readValuesMap[sens_PMP_118_Modbus.Ut],
			Rt:     readValuesMap[sens_PMP_118_Modbus.Rt],
			Ri:     readValuesMap[sens_PMP_118_Modbus.Ri],
			Tr:     readValuesMap[sens_PMP_118_Modbus.Tr],
			U2:     readValuesMap[sens_PMP_118_Modbus.U2],
			Dg:     readValuesMap[sens_PMP_118_Modbus.Dg],
			Ts:     readValuesMap[sens_PMP_118_Modbus.Ts],
		}

		p.UpdateLastSENSTelemetry(jarNumberStr, mainParameters)
	}

	if data[3] == 0x8A {
		tableDataBytes := data[7 : len(data)-1]
		effectiveTableDataBytes := tableDataBytes
		terminatorIndex := -1
		for i := 0; i < len(tableDataBytes)-1; i++ {
			if tableDataBytes[i] == sens_PMP_118_Modbus.ZeroByte && tableDataBytes[i+1] == sens_PMP_118_Modbus.ZeroByte {
				terminatorIndex = i
				break
			}
		}

		if terminatorIndex != -1 {
			effectiveTableDataBytes = tableDataBytes[:terminatorIndex]
		}

		var decodedValueSlice []float32
		for i := 0; i < len(effectiveTableDataBytes); i += 3 {
			if i+3 > len(effectiveTableDataBytes) {
				break
			}
			decodedValues, err := sens.Convert24BytesToFloat32(effectiveTableDataBytes[i:i+3], binary.LittleEndian)
			if err != nil {
				return err
			}
			decodedValueSlice = append(decodedValueSlice, decodedValues)
		}
		decodeValueSliceJSON, err := json.Marshal(decodedValueSlice)
		decodeValueSliceStr := string(decodeValueSliceJSON)
		decodeValueSliceStr = strings.ReplaceAll(decodeValueSliceStr, "[", "")
		decodeValueSliceStr = strings.ReplaceAll(decodeValueSliceStr, "]", "")
		if err != nil {
			return err
		}
		lastTime := time.Now().Unix()
		mainParameters := TempStatus{
			timeAS: lastTime,
			nt:     decodeValueSliceStr,
		}

		p.UpdateLastTempTelemetry(jarNumberStr, mainParameters)
	}

	return nil
}

func (p *Processing) generateTelemetryReport() integration.TelemetryRequest {

	var oneErrorMsg = ""
	var twoErrorMsg = ""

	// === 1 емкость ===
	if time.Now().Unix()-p.lastSENSTelemetry["1"].timeAS > p.kazsConfig.Telemetry.ActualTimeSENS {
		oneErrorMsg += fmt.Sprintf("Ошибка: Неактуальная телеметрия. ")
	}

	if time.Now().Unix()-p.lastTRKTelemetry["1"].timeAS > p.kazsConfig.Telemetry.ActualTimeTRK {
		oneErrorMsg += fmt.Sprintf("Ошибка: Неатуальный статус ТРК. ")
	}

	if time.Now().Unix()-p.lastTempTelemetry["1"].timeAS > p.kazsConfig.Telemetry.ActualTemp {
		oneErrorMsg += fmt.Sprintf("Ошибка: Неактуальная температура. ")
	}

	// === 2 емкость ===
	if time.Now().Unix()-p.lastSENSTelemetry["2"].timeAS > p.kazsConfig.Telemetry.ActualTimeSENS {
		twoErrorMsg += fmt.Sprintf("Ошибка: Неактуальная телеметрия. ")
	}

	if time.Now().Unix()-p.lastTRKTelemetry["2"].timeAS > p.kazsConfig.Telemetry.ActualTimeTRK {
		twoErrorMsg += fmt.Sprintf("Ошибка: Неатуальнаый статус ТРК. ")
	}

	if time.Now().Unix()-p.lastTempTelemetry["2"].timeAS > p.kazsConfig.Telemetry.ActualTemp {
		twoErrorMsg += fmt.Sprintf("Ошибка: Неактуальная температура. ")
	}

	return integration.TelemetryRequest{
		StatusTime:     time.Now().Unix(),
		PowerStatus:    0,
		BatteryStatus:  0,
		OutTemperature: 0,
		KazsErrors:     "",
		Jars: []integration.KazsTelemetryNestedJars{
			{
				JarId:         "1",
				JarLockStatus: doorStatus(p.lastControllerDinTelemetry[p.driver.ControllerDriver.Adapter.Maping.Controller.Doors["1"].Number].status),
				NozzleStatus:  p.lastTRKTelemetry["1"].status,
				H:             p.lastSENSTelemetry["1"].H * float32(p.kazsConfig.Telemetry.TelemetryUnits),
				T:             p.lastSENSTelemetry["1"].T,
				Pr:            p.lastSENSTelemetry["1"].Pr,
				U:             p.lastSENSTelemetry["1"].U * float32(p.kazsConfig.Telemetry.TelemetryUnits),
				G:             p.lastSENSTelemetry["1"].G,
				R:             p.lastSENSTelemetry["1"].R,
				U1:            p.lastSENSTelemetry["1"].U * float32(p.kazsConfig.Telemetry.TelemetryUnits), // TODO:
				H2:            p.lastSENSTelemetry["1"].H2 * float32(p.kazsConfig.Telemetry.TelemetryUnits),
				Ut:            p.lastSENSTelemetry["1"].Ut,
				Rt:            p.lastSENSTelemetry["1"].Rt,
				Ri:            p.lastSENSTelemetry["1"].Ri,
				Tr:            p.lastSENSTelemetry["1"].Tr,
				U2:            p.lastSENSTelemetry["1"].U2,
				Nt:            p.lastTempTelemetry["1"].nt,
				Dg:            p.lastSENSTelemetry["1"].Dg,
				Ts:            p.lastSENSTelemetry["1"].Ts,
				JarErrors:     oneErrorMsg,
			},
			{
				JarId:         "2",
				JarLockStatus: doorStatus(p.lastControllerDinTelemetry[p.driver.ControllerDriver.Adapter.Maping.Controller.Doors["2"].Number].status),
				NozzleStatus:  p.lastTRKTelemetry["2"].status,
				H:             p.lastSENSTelemetry["2"].H * float32(p.kazsConfig.Telemetry.TelemetryUnits),
				T:             p.lastSENSTelemetry["2"].T,
				Pr:            p.lastSENSTelemetry["2"].Pr,
				U:             p.lastSENSTelemetry["2"].U * float32(p.kazsConfig.Telemetry.TelemetryUnits),
				G:             p.lastSENSTelemetry["2"].G,
				R:             p.lastSENSTelemetry["2"].R,
				U1:            p.lastSENSTelemetry["2"].U * float32(p.kazsConfig.Telemetry.TelemetryUnits), // TODO:
				H2:            p.lastSENSTelemetry["2"].H2 * float32(p.kazsConfig.Telemetry.TelemetryUnits),
				Ut:            p.lastSENSTelemetry["2"].Ut,
				Rt:            p.lastSENSTelemetry["2"].Rt,
				Ri:            p.lastSENSTelemetry["2"].Ri,
				Tr:            p.lastSENSTelemetry["2"].Tr,
				U2:            p.lastSENSTelemetry["2"].U2,
				Nt:            p.lastTempTelemetry["2"].nt,
				Dg:            p.lastSENSTelemetry["2"].Dg,
				Ts:            p.lastSENSTelemetry["2"].Ts,
				JarErrors:     twoErrorMsg,
			},
		},
	}
}

func nozzleStatus(status string) int {
	switch status {
	case "30":
		return 30
	case "31":
		return 31
	case "32":
		return 32
	case "33":
		return 33
	case "34":
		return 34
	default:
		return 99
	}
}

func doorStatus(status string) int {
	switch status {
	case "0":
		return 0
	case "1":
		return 1
	default:
		return 2
	}
}
