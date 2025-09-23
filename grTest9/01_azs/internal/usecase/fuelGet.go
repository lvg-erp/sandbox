package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"fuelazs/internal/integration"
	"fuelazs/internal/usecase/models"
	"fyne.io/fyne/v2"
	"log/slog"
	"strconv"
	"time"
)

func (p *Processing) FuelGet(qrInfo models.ScannerResponse) error {
	logger := p.logger.Transaction("fuel_get", qrInfo.TID)
	captureError := CaptureErrorDep{
		tid:        qrInfo.TID,
		operation:  "FuelGet",
		kazsNumber: p.kazsOperator.KazsNumber,
	}

	logger.Info("статус потоков",
		"one_jar_active", p.oneJarActiveProcess,
		"two_jar_active", p.twoJarActiveProcess)

	// Оба потока заняты?
	if p.oneJarActiveProcess && p.twoJarActiveProcess {
		logger.Warn("все емкости заняты.")
		p.CaptureError(fmt.Errorf("all jars is busy"), captureError)

		_ = p.repository.ErrorLogs.InsertError(models.Errors{
			Time:    time.Now().Unix(),
			Handler: "FuelGet",
			Id:      qrInfo.TID,
			Error:   "all jars is busy",
		})

		return nil
	}

	// Блокируем сканер
	logger.Debug("блокировка сканера.")
	p.UpdateQRActive(false)

	logger.Info("GET FuelGetInfo", "timeout_sec", p.appConfig.FuelGetInfoTimeout)
	res := RunWithTimeoutValue(func() (*integration.FuelInfoResponse, error) {
		return p.kazsOperator.FuelGetInfo(&integration.FuelGetInfoRequest{
			TID: qrInfo.TID,
		})
	}, time.Duration(p.appConfig.FuelGetInfoTimeout)*time.Second)

	if res.Err != nil || res.Value.Error {
		var err string
		errMsg := "Нет ответа от сервера"
		if res.Err != nil {
			err = res.Err.Error()
		} else if res.Value.Error {
			err = res.Value.Message
			errMsg = res.Value.Message
		}
		logger.Error("ошибка запроса.", "handler", "FuelGetInfo", "err", err)
		p.CaptureError(fmt.Errorf("kazsOperator.FuelGetInfo error: %s", errMsg), captureError)
		var jarNumber string
		var targetContent *fyne.Container
		switch {
		case !p.oneJarActiveProcess:
			targetContent = p.appGui.LeftSection.Content
			jarNumber = "1"
		case !p.twoJarActiveProcess:
			targetContent = p.appGui.RightSection.Content
			jarNumber = "2"
		default:
			p.UpdateQRActive(true)
			return nil // Если функция возвращает ошибку, убедитесь, что nil здесь уместен
		}
		if targetContent != nil {
			fyne.Do(func() {
				p.appGui.ShowSectionDialog(targetContent, "Ошибка", errMsg, 10, func() {
					p.UpdateQRActive(true)
					p.appGui.CreateDefaultScreen(jarNumber)
				})
			})
		} else {
			p.UpdateQRActive(true)
			p.appGui.CreateDefaultScreen(jarNumber)
		}

		// Записываем лог с ошибкой
		_ = p.repository.ErrorLogs.InsertError(models.Errors{
			Time:    time.Now().Unix(),
			Handler: "FuelGet",
			Id:      qrInfo.TID,
			Error:   fmt.Sprintf("kazsOperator.FuelGetInfo error %v:", errMsg),
		})

		return fmt.Errorf(errMsg)
	}

	captureError.jarNumber = res.Value.Result.JarId
	logger = logger.With("jar_id", res.Value.Result.JarId)

	// Указанный поток свободен?
	if res.Value.Result.JarId == "1" && p.oneJarActiveProcess || res.Value.Result.JarId == "2" && p.twoJarActiveProcess {
		logger.Warn("выбранная ёмкость занята.")
		p.CaptureError(fmt.Errorf("jars №%v busy", res.Value.Result.JarId), captureError)
		// Если выбрана емкость 1, но она занята, проверяем емкость 2
		var jarNumber string
		var targetContent *fyne.Container
		switch {
		case !p.oneJarActiveProcess:
			targetContent = p.appGui.LeftSection.Content
			jarNumber = "1"
		case !p.twoJarActiveProcess:
			targetContent = p.appGui.RightSection.Content
			jarNumber = "2"
		default:
			p.UpdateQRActive(true)
			return nil // Если функция возвращает ошибку, убедитесь, что nil здесь уместен
		}

		if targetContent != nil {
			fyne.Do(func() {
				p.appGui.ShowSectionDialog(targetContent, "Ошибка", fmt.Sprintf("Емкость №%v занята", res.Value.Result.JarId), 10, func() {
					p.UpdateQRActive(true)
					p.appGui.CreateDefaultScreen(jarNumber)
				})
			})
		} else {
			p.UpdateQRActive(true)
			p.appGui.CreateDefaultScreen(jarNumber)
		}

		// Записываем ошибку в локальную БД
		_ = p.repository.ErrorLogs.InsertError(models.Errors{
			Time:    time.Now().Unix(),
			Handler: "FuelGet",
			Id:      qrInfo.TID,
			Error:   fmt.Sprintf("jars №%v busy", res.Value.Result.JarId),
		})
		return nil
	}

	// Блокируем соответствующий процесс
	logger.Debug("процесс заблокирован.")
	p.UpdateJarStatus(res.Value.Result.JarId, true)

	// Разблокирую сканер
	logger.Debug("разблокировка сканера.")
	p.UpdateQRActive(true)

	time.Sleep(500 * time.Millisecond)
	// Показываем экран 0 в соответствующем потоке

	logger.Info("экран 'обработка'.")
	p.appGui.CreateDownloadScreen(res.Value.Result.JarId)

	// 4 на блок схеме
	logger.Info("получение телеметрии до пополнения и проверка незавершенных транзакций.")
	beforeJarInfo, err := p.validationAndCheckLastFuelGet(res.Value.Result.JarId, logger)
	if err != nil {
		logger.Error("ошибка валидации.", "err", err)
		var targetContent *fyne.Container
		if res.Value.Result.JarId == "1" {
			targetContent = p.appGui.LeftSection.Content
		} else {
			targetContent = p.appGui.RightSection.Content
		}

		if targetContent != nil {
			fyne.Do(func() {
				p.appGui.ShowSectionDialog(targetContent, "Ошибка", "Не удалось начать пополнение", 10, func() {
					p.UpdateJarStatus(res.Value.Result.JarId, false)
					p.appGui.CreateDefaultScreen(res.Value.Result.JarId)
				})
			})
		} else {
			p.UpdateJarStatus(res.Value.Result.JarId, false)
			p.appGui.CreateDefaultScreen(res.Value.Result.JarId)
		}

		_ = p.repository.ErrorLogs.InsertError(models.Errors{
			Time:    time.Now().Unix(),
			Handler: "FuelGive",
			Id:      qrInfo.TID,
			Error:   err.Error(),
		})

		return err
	}

	// Подтверждаем с сервером
	logger.Info("GET FuelGetConfirmation", "timeout_sec", p.appConfig.FuelGetConfirmationTimeout)
	resConf := RunWithTimeoutValue(func() (*integration.FuelConfirmationResponse, error) {
		return p.kazsOperator.FuelGetConfirmation(&integration.FuelConfirmationRequest{
			FuelGetID: qrInfo.TID,
		})
	}, time.Duration(p.appConfig.FuelGetConfirmationTimeout)*time.Second)

	if resConf.Err != nil || resConf.Value.Error {
		var resConfErr string
		if resConf.Err != nil {
			resConfErr = resConf.Err.Error()
		} else if resConf.Value.Error {
			resConfErr = resConf.Value.Message
		}
		logger.Error("ошибка запроса.", "handler", "FuelGetConfirmation", "err", resConfErr)

		p.CaptureError(fmt.Errorf("kazsOperator.FuelGetConfirmation error: %v", resConfErr), captureError)
		var targetContent *fyne.Container
		if res.Value.Result.JarId == "1" {
			targetContent = p.appGui.LeftSection.Content
		} else {
			targetContent = p.appGui.RightSection.Content
		}
		errMsg := "Нет соединения с сервером."
		if resConf.Value.Error {
			errMsg = resConf.Value.Message
		}

		if targetContent != nil {
			fyne.Do(func() {
				p.appGui.ShowSectionDialog(targetContent, "Ошибка", errMsg, 10, func() {
					// Разблокируем процесс, если подтверждение не удалось
					p.UpdateJarStatus(res.Value.Result.JarId, false)
					p.appGui.CreateDefaultScreen(res.Value.Result.JarId)
				})
			})
		} else {
			p.UpdateJarStatus(res.Value.Result.JarId, false)
			p.appGui.CreateDefaultScreen(res.Value.Result.JarId)
		}

		// Записываем лог с ошибкой
		_ = p.repository.ErrorLogs.InsertError(models.Errors{
			Time:    time.Now().Unix(),
			Handler: "FuelGet",
			Id:      qrInfo.TID,
			Error:   fmt.Sprintf("kazsOperator.FuelGetConfirmation error: %v", errMsg),
		})

		return fmt.Errorf("kazsOperator.FuelGetConfirmation error: %v", errMsg)
	}

	// Сохраняем в БД начало пополнения
	logger.Info("сохранение транзакции в БД.")
	startTime := time.Now().Unix()
	fuelLiters := float32(0)
	endTime := int64(0)
	monitoringFinishTime := int64(0)
	speed := float32(0)

	err = p.repository.FuelGet.InsertFuelGetTransaction(models.InsertFuelGetTransactionDep{
		FuelGetID:            qrInfo.TID,
		KazsNumber:           p.kazsOperator.KazsNumber,
		JarNumber:            resConf.Value.Result.JarId,
		StartTime:            startTime,
		EndTime:              &endTime,
		MonitoringFinishTime: &monitoringFinishTime,
		Speed:                &speed,
		FuelType:             resConf.Value.Result.FuelType,
		DocNumber:            resConf.Value.Result.DocNumber,
		SensorBeforeGive: &models.SensorInfo{
			H:  float64(beforeJarInfo.H),
			T:  float64(beforeJarInfo.T),
			Pr: float64(beforeJarInfo.Pr),
			U:  float64(beforeJarInfo.U),
			G:  float64(beforeJarInfo.G),
			R:  float64(beforeJarInfo.R),
			U1: float64(beforeJarInfo.U1),
			H2: float64(beforeJarInfo.H2),
			Ut: float64(beforeJarInfo.Ut),
			Rt: float64(beforeJarInfo.Rt),
			Ri: float64(beforeJarInfo.Ri),
			Tr: float64(beforeJarInfo.Tr),
			U2: float64(beforeJarInfo.U2),
			Nt: beforeJarInfo.Nt,
			Dg: float64(beforeJarInfo.Dg),
			Ts: float64(beforeJarInfo.Ts),
		},
		FuelLiterPlan: resConf.Value.Result.FuelLiter,
		FuelLiters:    &fuelLiters,
		SendStatus:    false,
	})

	if err != nil {
		logger.Error("не удалось сохранить транзакцию в БД.", "err", err)
		p.CaptureError(fmt.Errorf("Repo.InsertFuelGetTransaction error: %v", err), captureError)

		var targetContent *fyne.Container
		if res.Value.Result.JarId == "1" {
			targetContent = p.appGui.LeftSection.Content
		} else {
			targetContent = p.appGui.RightSection.Content
		}

		if targetContent != nil {
			fyne.Do(func() {
				p.appGui.ShowSectionDialog(targetContent, "Ошибка", "Не удалось начать пополнение", 10, func() {
					// Разблокируем процесс, если подтверждение не удалось
					p.UpdateJarStatus(res.Value.Result.JarId, false)
					p.appGui.CreateDefaultScreen(res.Value.Result.JarId)
				})
			})
		} else {
			p.UpdateJarStatus(res.Value.Result.JarId, false)
			p.appGui.CreateDefaultScreen(res.Value.Result.JarId)
		}

		// Записываем лог с ошибкой
		_ = p.repository.ErrorLogs.InsertError(models.Errors{
			Time:    time.Now().Unix(),
			Handler: "FuelGet",
			Id:      qrInfo.TID,
			Error:   fmt.Sprintf("Repo.InsertFuelGetTransaction error: %v", err),
		})

		return fmt.Errorf("Repo.InsertFuelGetTransaction error: %v", err)
	}

	// Разблокировка насоса
	logger.Info("включение насоса.")
	err = RunWithTimeout(func() error {
		return p.driver.ControllerDriver.Adapter.EnablePump(res.Value.Result.JarId)
	}, time.Duration(p.driverConfig.ControllerTimeout)*time.Second)
	if err != nil {
		logger.Error("не удалось включить насос.", "err", err)
		p.CaptureError(fmt.Errorf("Controller.EnablePump error: %v", err), captureError)

		err = p.createAndSendFuelGetReport(qrInfo.TID, res.Value.Result.JarId, "Не удалось разблокировать насос", captureError, beforeJarInfo.U*float32(p.kazsConfig.FuelGetConfig.FuelGetUnits), logger)
		if err != nil {
			p.CaptureError(err, captureError)
		}
		var targetContent *fyne.Container
		if res.Value.Result.JarId == "1" {
			targetContent = p.appGui.LeftSection.Content
		} else {
			targetContent = p.appGui.RightSection.Content
		}

		if targetContent != nil {
			fyne.Do(func() {
				p.appGui.ShowSectionDialog(targetContent, "Ошибка", "Насос не отвечает", 10, func() {
					// Разблокируем процесс, если подтверждение не удалось
					p.UpdateJarStatus(res.Value.Result.JarId, false)
					err = p.createAndSendFuelGetReport(qrInfo.TID, res.Value.Result.JarId, "failed enable pump", captureError, beforeJarInfo.U*float32(p.kazsConfig.FuelGetConfig.FuelGetUnits), logger)
					if err != nil {
						p.CaptureError(fmt.Errorf("createAndSendFuelGetReport error: %v", err), captureError)
					}
					p.appGui.CreateDefaultScreen(res.Value.Result.JarId)
				})
			})
		} else {
			p.UpdateJarStatus(res.Value.Result.JarId, false)
			p.appGui.CreateDefaultScreen(res.Value.Result.JarId)
		}

		// Записываем лог с ошибкой
		_ = p.repository.ErrorLogs.InsertError(models.Errors{
			Time:    time.Now().Unix(),
			Handler: "FuelGet",
			Id:      qrInfo.TID,
			Error:   fmt.Sprintf("Repo.InsertFuelGetTransaction error: %v", err),
		})

		return fmt.Errorf("Controller.EnablePump error: %v", resConf.Err)
	}

	// Получаем статус люка
	logger.Info("получение статуса люка.")
	beforeDoorStatus := p.lastControllerDinTelemetry[p.driver.ControllerDriver.Adapter.Maping.Controller.Doors[res.Value.Result.JarId].Number].status

	logger.Info("статус люка.", "door_status", beforeDoorStatus)

	// Открываем замок люка
	logger.Info("открытие замка.")
	err = RunWithTimeout(func() error {
		return p.driver.ControllerDriver.Adapter.OpenLock(res.Value.Result.JarId)
	}, time.Duration(p.driverConfig.ControllerTimeout)*time.Second)
	if err != nil {
		logger.Error("не удалось открыть замок.", "err", err)
		p.CaptureError(fmt.Errorf("Controller.EnablePump error: %v", err), captureError)

		err = p.createAndSendFuelGetReport(qrInfo.TID, res.Value.Result.JarId, "Не удалось открыть замок люка", captureError, beforeJarInfo.U*float32(p.kazsConfig.FuelGetConfig.FuelGetUnits), logger)
		if err != nil {
			p.CaptureError(err, captureError)
		}

		var targetContent *fyne.Container
		if res.Value.Result.JarId == "1" {
			targetContent = p.appGui.LeftSection.Content
		} else {
			targetContent = p.appGui.RightSection.Content
		}

		if targetContent != nil {
			fyne.Do(func() {
				p.appGui.ShowSectionDialog(targetContent, "Ошибка", "Замок не доступен", 10, func() {
					// Разблокируем процесс, если подтверждение не удалось
					p.UpdateJarStatus(res.Value.Result.JarId, false)
					err = p.createAndSendFuelGetReport(qrInfo.TID, res.Value.Result.JarId, "failed open lock", captureError, beforeJarInfo.U*float32(p.kazsConfig.FuelGetConfig.FuelGetUnits), logger)
					if err != nil {
						p.CaptureError(fmt.Errorf("createAndSendFuelGetReport error: %v", err), captureError)
					}
					p.appGui.CreateDefaultScreen(res.Value.Result.JarId)
				})
			})
		} else {
			p.UpdateJarStatus(res.Value.Result.JarId, false)
			p.appGui.CreateDefaultScreen(res.Value.Result.JarId)
		}

		// Записываем лог с ошибкой
		_ = p.repository.ErrorLogs.InsertError(models.Errors{
			Time:    time.Now().Unix(),
			Handler: "FuelGet",
			Id:      qrInfo.TID,
			Error:   fmt.Sprintf("Repo.InsertFuelGetTransaction error: %v", err),
		})

		return fmt.Errorf("Controller.EnablePump error: %v", resConf.Err)
	}

	// Показываем экран 4 в соответствующем потоке
	availableAmount := beforeJarInfo.U * float32(p.kazsConfig.FuelGetConfig.FuelGetUnits) / (beforeJarInfo.Pr / 100) // Максимальный объем емкости
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.kazsConfig.FuelGetConfig.FuelGetStartScreenTimeout)*time.Second)

	go p.setFuelGetStartScreen(ctx, setFuelGetStartScreenDep{
		jarNumber:       res.Value.Result.JarId,
		fuelType:        res.Value.Result.FuelType,
		beforeJarInfo:   beforeJarInfo,
		availableAmount: availableAmount,
		expectedAmount:  float32(res.Value.Result.FuelLiter),
	}, logger)

	go p.monitoringFuelGetStart(ctx, cancel, monitoringFuelGetStartDep{
		jarNumber:        res.Value.Result.JarId,
		fuelType:         res.Value.Result.FuelType,
		expectedLiters:   float32(res.Value.Result.FuelLiter),
		beforeFuelGet:    beforeJarInfo.U * float32(p.kazsConfig.FuelGetConfig.FuelGetUnits),
		documentNumber:   resConf.Value.Result.DocNumber,
		fuelGetID:        qrInfo.TID,
		startTime:        startTime,
		jarVolume:        availableAmount,
		captureError:     captureError,
		beforeDoorStatus: beforeDoorStatus,
	}, logger)

	return nil
}

// 4 на блок-схеме
func (p *Processing) validationAndCheckLastFuelGet(jarNumber string, logger *slog.Logger) (models.JarsInfo, error) {
	captureError := CaptureErrorDep{
		operation:  "FuelGet",
		kazsNumber: p.kazsOperator.KazsNumber,
		jarNumber:  jarNumber,
	}

	// Читаем телеметрию
	errMsg := ""

	beforeTelemetry := p.lastSENSTelemetry[jarNumber]
	beforeTemp := p.lastTempTelemetry[jarNumber]

	if time.Now().Unix()-beforeTelemetry.timeAS > p.kazsConfig.FuelGiveConfig.ActualTimeSENS {
		p.CaptureError(fmt.Errorf("old sens telemetry"), captureError)

		errMsg += fmt.Sprintf("Неактуальная телеметрия после пополнения. ")
	}

	lastFuelGet, err := p.repository.FuelGet.GetLastFuelGetTransaction(jarNumber)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			p.CaptureError(fmt.Errorf("Repo.GetLastFuelGetTransaction error: %v", err), captureError)
			return models.JarsInfo{}, fmt.Errorf("Repo.GetLastFuelGetTransaction error: %v", err)
		} else {
			p.CaptureError(fmt.Errorf("Незавершенных пополнений нет"), captureError)
		}
	} else {
		// Нашли незавершенное пополнение
		p.logger.Warn("Найдено незавершенное пополнение")
		captureError.tid = lastFuelGet.FuelGetID
		endTime := time.Now().Unix()
		fuelLiters := (beforeTelemetry.U - float32(lastFuelGet.SensorBeforeGive.U)) * float32(p.kazsConfig.FuelGetConfig.FuelGetUnits)
		updateErr := p.repository.FuelGet.UpdateFuelGetTransaction(models.UpdateFuelGetTransactionDep{
			FuelGetID:  lastFuelGet.FuelGetID,
			EndTime:    &endTime,
			FuelLiters: &fuelLiters,
			SensorAfterGive: &models.SensorInfo{
				H:  float64(beforeTelemetry.H),
				T:  float64(beforeTelemetry.T),
				Pr: float64(beforeTelemetry.Pr),
				U:  float64(beforeTelemetry.U),
				G:  float64(beforeTelemetry.G),
				R:  float64(beforeTelemetry.R),
				U1: float64(beforeTelemetry.U1),
				H2: float64(beforeTelemetry.H2),
				Ut: float64(beforeTelemetry.Ut),
				Rt: float64(beforeTelemetry.Rt),
				Ri: float64(beforeTelemetry.Ri),
				Tr: float64(beforeTelemetry.Tr),
				U2: float64(beforeTelemetry.U2),
				Nt: beforeTemp.nt,
				Dg: float64(beforeTelemetry.Dg),
				Ts: float64(beforeTelemetry.Ts),
			},
			Errors: &errMsg,
		})
		if updateErr != nil {
			p.CaptureError(fmt.Errorf("Repo.GetLastFuelGetTransaction error: %v", updateErr), captureError)
		}

		receipt, err := p.repository.FuelGet.GetFuelGetTransaction(lastFuelGet.FuelGetID)
		if err != nil {
			p.CaptureError(fmt.Errorf("FuelGet.GetFuelGetTransaction error: %v", err), captureError)
			return models.JarsInfo{}, fmt.Errorf("FuelGet.GetFuelGetTransaction error: %v", err)
		}

		err = RunWithTimeout(func() error {
			return p.kazsOperator.FuelGetReceipt(
				&integration.FuelGetReceipt{
					FuelGetID: lastFuelGet.FuelGetID,
				},
				&integration.FuelGetReceiptRequest{
					KazsNumber: p.kazsOperator.KazsNumber,
					JarId:      jarNumber,
					FuelType:   receipt.FuelType,
					StartTime:  receipt.StartTime,
					EndTime:    endTime,
					DocNumber:  receipt.DocNumber,
					FuelLiter:  float64(fuelLiters),
					SensorBeforeGet: integration.SensorInfo{
						H:  receipt.SensorBeforeGet.H,
						T:  receipt.SensorBeforeGet.T,
						Pr: receipt.SensorBeforeGet.Pr,
						U:  receipt.SensorBeforeGet.U,
						G:  receipt.SensorBeforeGet.G,
						R:  receipt.SensorBeforeGet.R,
						U1: receipt.SensorBeforeGet.U1,
						H2: receipt.SensorBeforeGet.H2,
						Ut: receipt.SensorBeforeGet.Ut,
						Rt: receipt.SensorBeforeGet.Rt,
						Ri: receipt.SensorBeforeGet.Ri,
						Tr: receipt.SensorBeforeGet.Tr,
						U2: receipt.SensorBeforeGet.U2,
						Nt: receipt.SensorBeforeGet.Nt,
						Dg: receipt.SensorBeforeGet.Dg,
						Ts: receipt.SensorBeforeGet.Ts,
					},
					SensorAfterGet: integration.SensorInfo{
						H:  receipt.SensorAfterGet.H,
						T:  receipt.SensorAfterGet.T,
						Pr: receipt.SensorAfterGet.Pr,
						U:  receipt.SensorAfterGet.U,
						G:  receipt.SensorAfterGet.G,
						R:  receipt.SensorAfterGet.R,
						U1: receipt.SensorAfterGet.U1,
						H2: receipt.SensorAfterGet.H2,
						Ut: receipt.SensorAfterGet.Ut,
						Rt: receipt.SensorAfterGet.Rt,
						Ri: receipt.SensorAfterGet.Ri,
						Tr: receipt.SensorAfterGet.Tr,
						U2: receipt.SensorAfterGet.U2,
						Nt: receipt.SensorAfterGet.Nt,
						Dg: receipt.SensorAfterGet.Dg,
						Ts: receipt.SensorAfterGet.Ts,
					},
					Errors: errMsg,
				})
		}, time.Duration(p.appConfig.FuelGetReceiptTimeout)*time.Second)
		if err != nil {
			p.CaptureError(fmt.Errorf("KazsOperator.FuelGetReceipt error: %v", err), captureError)
			return models.JarsInfo{}, fmt.Errorf("KazsOperator.FuelGetReceipt error: %v", err)
		}

		status := true

		updateErr = p.repository.FuelGet.UpdateFuelGetTransaction(models.UpdateFuelGetTransactionDep{
			FuelGetID:  lastFuelGet.FuelGetID,
			SendStatus: &status,
		})
		if updateErr != nil {
			p.CaptureError(fmt.Errorf("FuelGet.UpdateFuelGetTransaction error: %v", updateErr), captureError)
			return models.JarsInfo{}, fmt.Errorf("FuelGet.UpdateFuelGetTransaction error: %v", updateErr)
		}
	}

	// Проверяем контроллер
	err = RunWithTimeout(func() error {
		return p.driver.ControllerDriver.Adapter.Ping()
	}, time.Duration(p.driverConfig.ControllerTimeout)*time.Second)
	if err != nil {
		p.CaptureError(fmt.Errorf("Controller.Ping error: %v", err), captureError)

		return models.JarsInfo{}, fmt.Errorf("Controller.Ping error: %v", err)
	}

	jarLockStatus, err := strconv.Atoi(p.lastControllerDinTelemetry[p.driver.ControllerDriver.Adapter.Maping.Controller.Doors[jarNumber].Number].status)
	if err != nil {
		jarLockStatus = 2
	}

	return models.JarsInfo{
		JarNumber:     jarNumber,
		JarLockStatus: jarLockStatus,
		TrkStatus:     p.lastTRKTelemetry[jarNumber].status,
		H:             beforeTelemetry.H,
		T:             beforeTelemetry.T,
		Pr:            beforeTelemetry.Pr,
		U:             beforeTelemetry.U,
		G:             beforeTelemetry.U,
		R:             beforeTelemetry.R,
		U1:            beforeTelemetry.U1,
		H2:            beforeTelemetry.H2,
		Ut:            beforeTelemetry.Ut,
		Rt:            beforeTelemetry.Rt,
		Ri:            beforeTelemetry.Ri,
		Tr:            beforeTelemetry.Tr,
		U2:            beforeTelemetry.U2,
		Dg:            beforeTelemetry.Dg,
		Ts:            beforeTelemetry.Ts,
		Nt:            beforeTemp.nt,
	}, nil
}

type setFuelGetStartScreenDep struct {
	jarNumber       string          // Номер емкости
	fuelType        string          // Тип тополя
	beforeJarInfo   models.JarsInfo // Показания уровнемера в начале пополнения
	availableAmount float32         // Максимальный объем емкости
	expectedAmount  float32         // Количество литров для пополнения
}

// setFuelGetStartScreen Метод отображает экран 4 (пока не закончится таймер или не начнется пополнение)
func (p *Processing) setFuelGetStartScreen(ctx context.Context, dep setFuelGetStartScreenDep, logger *slog.Logger) {
	logger.Info("начало процесса отрисовки экрана 'ожидание начала пополнения'.")
	deadline, ok := ctx.Deadline()
	var timeout int
	if ok {
		timeout = int(time.Until(deadline).Seconds())
	} else {
		timeout = p.kazsConfig.FuelGetConfig.FuelGetStartScreenTimeout
	}

	logger.Info("экран 'ожидание начала пополнения.'", "timeout_sec", timeout)
	p.appGui.CreateFuelGetStartScreen(dep.jarNumber, dep.fuelType, dep.beforeJarInfo.U*float32(p.kazsConfig.FuelGetConfig.FuelGetUnits), dep.availableAmount-dep.beforeJarInfo.U*float32(p.kazsConfig.FuelGetConfig.FuelGetUnits), dep.expectedAmount, timeout)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for i := timeout - 1; i >= 0; i-- {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.appGui.CreateFuelGetStartScreen(dep.jarNumber, dep.fuelType, dep.beforeJarInfo.U*float32(p.kazsConfig.FuelGetConfig.FuelGetUnits), dep.availableAmount-dep.beforeJarInfo.U*float32(p.kazsConfig.FuelGetConfig.FuelGetUnits), dep.expectedAmount, i)
		}
	}
}

type monitoringFuelGetStartDep struct {
	jarNumber        string  // Номер емкости
	fuelType         string  // Тип топлива
	expectedLiters   float32 // Количество литров для пополнения
	beforeFuelGet    float32 // Показания уровнемера до пополнения
	documentNumber   string  // Номер документа
	fuelGetID        string  // Идентификатор транзакции
	startTime        int64   // Время начала транзакции
	jarVolume        float32 // Максимальный объем емкости
	beforeDoorStatus string  // Статус люка
	captureError     CaptureErrorDep
}

// monitoringFuelGetStart Метод для мониторинга начала пополнения и переходит на экран 5
func (p *Processing) monitoringFuelGetStart(ctx context.Context, cancel context.CancelFunc, dep monitoringFuelGetStartDep, logger *slog.Logger) {
	logger.Info("запущен процесс мониторинга начала пополнения.")

	defer cancel()
	var afterFuelGet float32
	for {
		select {
		case <-ctx.Done(): // Отмена по таймауту
			logger.Warn("вышел таймаут начала пополнения.")
			logger.Info("закрытие замка.")
			err := p.driver.ControllerDriver.Adapter.CloseLock(dep.jarNumber)
			if err != nil {
				logger.Error("не удалось закрыть замок.", "err", err)
				p.CaptureError(fmt.Errorf("close lock error: %w", err), dep.captureError)
			}

			timer := p.kazsConfig.FuelGetConfig.FuelGetCompleteScreenNoFuelTimeout

			completeCtx, _ := context.WithTimeout(context.Background(), time.Duration(p.kazsConfig.FuelGetConfig.FuelGetCompleteScreenNoFuelTimeout)*time.Second)
			go p.startFuelGetComplete(completeCtx, startFuelGetCompleteDep{
				jarNumber:      dep.jarNumber,
				fuelType:       dep.fuelType,
				docNumber:      dep.documentNumber,
				beforeFuelGet:  dep.beforeFuelGet,
				afterFuelGet:   afterFuelGet,
				fuelGetPlan:    dep.expectedLiters,
				startTime:      dep.startTime,
				fuelGetID:      dep.fuelGetID,
				timer:          timer,
				startSpeedTime: 0,
				endSpeedTime:   0,
				captureError:   dep.captureError,
			}, logger)
			return
		default:
		}

		latestJarInfo := p.lastSENSTelemetry[dep.jarNumber]

		afterFuelGet = latestJarInfo.U * float32(p.kazsConfig.FuelGetConfig.FuelGetUnits)

		// Логика перехода
		afterDoorStatus := p.lastControllerDinTelemetry[p.driver.ControllerDriver.Adapter.Maping.Controller.Doors[dep.jarNumber].Number].status

		if latestJarInfo.U*float32(p.kazsConfig.FuelGetConfig.FuelGetUnits) > dep.beforeFuelGet+float32(p.kazsConfig.FuelGetConfig.FuelLiterStart) ||
			(dep.beforeDoorStatus == p.driver.ControllerDriver.Adapter.Maping.Controller.Doors[dep.jarNumber].Close && afterDoorStatus == p.driver.ControllerDriver.Adapter.Maping.Controller.Doors[dep.jarNumber].Open) {

			logger.Info("закрытие замка.")
			err := p.driver.ControllerDriver.Adapter.CloseLock(dep.jarNumber)
			if err != nil {
				logger.Error("не удалось закрыть замок.", "err", err)
				p.CaptureError(fmt.Errorf("Driver.CloseLock error: %w", err), dep.captureError)
			}
			drainedAmount := afterFuelGet - dep.beforeFuelGet
			// Переход к фазе InProgress
			go p.startFuelGetInProgress(startFuelGetInProgressDep{
				jarNumber:         dep.jarNumber,
				fuelType:          dep.fuelType,
				documentNumber:    dep.documentNumber,
				expectedFuelLiter: dep.expectedLiters,
				beforeFuelGet:     dep.beforeFuelGet,
				fuelVolume:        afterFuelGet,
				fuelGetID:         dep.fuelGetID,
				jarVolume:         dep.jarVolume,
				startTime:         dep.startTime,
				drainedAmount:     drainedAmount,
				beforeDoorStatus:  afterDoorStatus,
			}, logger)

			return
		}
		time.Sleep(1 * time.Second)
	}
}

type startFuelGetInProgressDep struct {
	jarNumber         string  // Номер емкости
	fuelType          string  // Тип топлива
	documentNumber    string  // Номер документа
	expectedFuelLiter float32 // Количество литров для пополнения
	beforeFuelGet     float32 // Первоначальные показания уровнемера
	fuelVolume        float32 // Последние показания уровнемера
	drainedAmount     float32 // Количество залитого топлива
	fuelGetID         string  // Идентификатор транзакции
	jarVolume         float32 // Объем емкости
	startTime         int64
	beforeDoorStatus  string
	captureError      CaptureErrorDep
}

// startFuelGetInProgress Метод 6 на блок-схеме
func (p *Processing) startFuelGetInProgress(dep startFuelGetInProgressDep, logger *slog.Logger) {
	logger.Info("начало мониторинга процесса пополнения.")
	startSpeedTime := time.Now().Unix()
	var fueledLiters float32
	var currentFuelVolume float32
	remainingTime := p.kazsConfig.FuelGetConfig.FuelGetInProgressScreenTimeout
	stallTimeout := p.kazsConfig.FuelGetConfig.StallTimeout

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(remainingTime)*time.Second)
	defer cancel()

	deadline, _ := ctx.Deadline()

	logger.Info("экран 'процесса пополнения'.")
	p.appGui.CreateFuelGetInProgressScreen(dep.jarNumber, dep.expectedFuelLiter, dep.drainedAmount, dep.fuelVolume, dep.jarVolume, remainingTime/60)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	startFuelGetTime := time.Now()

	lastFuelChangeTime := time.Now().Unix()
	lastVolume := dep.beforeFuelGet

	for {
		select {
		case <-ctx.Done():
			logger.Warn("вышло время пополнения.", "timeout_sec", p.kazsConfig.FuelGetConfig.FuelGetInProgressScreenTimeout)

			timer := p.kazsConfig.FuelGetConfig.FuelGetCompleteScreenTimeout

			completeCtx, _ := context.WithTimeout(context.Background(), time.Duration(p.kazsConfig.FuelGetConfig.FuelGetCompleteScreenTimeout)*time.Second)
			go p.startFuelGetComplete(completeCtx, startFuelGetCompleteDep{
				jarNumber:      dep.jarNumber,
				fuelType:       dep.fuelType,
				docNumber:      dep.documentNumber,
				beforeFuelGet:  dep.beforeFuelGet,
				afterFuelGet:   currentFuelVolume,
				fuelGetPlan:    dep.expectedFuelLiter,
				startTime:      dep.startTime,
				fuelGetID:      dep.fuelGetID,
				timer:          timer,
				startSpeedTime: startSpeedTime,
				endSpeedTime:   lastFuelChangeTime,
				captureError:   dep.captureError,
			}, logger)
			return

		case <-ticker.C:
			remaining := time.Until(deadline)
			minutesLeft := int(remaining.Minutes())
			if minutesLeft < 0 {
				minutesLeft = 0
			}

			currentJarInfo := p.lastSENSTelemetry[dep.jarNumber]
			currentFuelVolume = currentJarInfo.U * float32(p.kazsConfig.FuelGetConfig.FuelGetUnits)

			fueledLiters = currentFuelVolume - dep.beforeFuelGet
			if fueledLiters < 0 {
				fueledLiters = 0
			}

			p.appGui.CreateFuelGetInProgressScreen(dep.jarNumber, dep.expectedFuelLiter, fueledLiters, currentJarInfo.U*float32(p.kazsConfig.FuelGetConfig.FuelGetUnits), dep.jarVolume, minutesLeft)

			// === Контроль изменения объема топлива с погрешностью ===
			delta := p.kazsConfig.FuelGetConfig.StallDeltaLiters
			if abs(currentFuelVolume-lastVolume) > delta {
				lastFuelChangeTime = time.Now().Unix()
				lastVolume = currentFuelVolume
			} else if time.Now().Unix()-lastFuelChangeTime > stallTimeout {
				logger.Info("топливо не поступает более N секунд.", "no_change_time", time.Now().Unix()-lastFuelChangeTime)

				timer := p.kazsConfig.FuelGetConfig.FuelGetCompleteScreenTimeout

				completeCtx, _ := context.WithTimeout(context.Background(), time.Duration(p.kazsConfig.FuelGetConfig.FuelGetCompleteScreenTimeout)*time.Second)
				go p.startFuelGetComplete(completeCtx, startFuelGetCompleteDep{
					jarNumber:      dep.jarNumber,
					fuelType:       dep.fuelType,
					docNumber:      dep.documentNumber,
					beforeFuelGet:  dep.beforeFuelGet,
					afterFuelGet:   currentFuelVolume,
					fuelGetPlan:    dep.expectedFuelLiter,
					startTime:      dep.startTime,
					fuelGetID:      dep.fuelGetID,
					timer:          timer,
					startSpeedTime: startSpeedTime,
					endSpeedTime:   lastFuelChangeTime,
					captureError:   dep.captureError,
				}, logger)
				return
			}

			// Люк изменил состояние на закрыт
			afterDoorStatus := p.lastControllerDinTelemetry[p.driver.ControllerDriver.Adapter.Maping.Controller.Doors[dep.jarNumber].Number].status

			isDoorNowClosed := afterDoorStatus == p.driver.ControllerDriver.Adapter.Maping.Controller.Doors[dep.jarNumber].Close
			wasDoorInitiallyOpen := dep.beforeDoorStatus == p.driver.ControllerDriver.Adapter.Maping.Controller.Doors[dep.jarNumber].Open

			if time.Since(startFuelGetTime) > p.kazsConfig.FuelGetConfig.DoorCloseTimeout*time.Second {
				if wasDoorInitiallyOpen && isDoorNowClosed {
					logger.Info("люк изменил состояние на закрыт.", "door_before_status", dep.beforeDoorStatus, "door_after_status", afterDoorStatus)

					timer := p.kazsConfig.FuelGetConfig.FuelGetCompleteScreenTimeout

					completeCtx, _ := context.WithTimeout(context.Background(), time.Duration(p.kazsConfig.FuelGetConfig.FuelGetCompleteScreenTimeout)*time.Second)
					go p.startFuelGetComplete(completeCtx, startFuelGetCompleteDep{
						jarNumber:      dep.jarNumber,
						fuelType:       dep.fuelType,
						docNumber:      dep.documentNumber,
						beforeFuelGet:  dep.beforeFuelGet,
						afterFuelGet:   currentFuelVolume,
						fuelGetPlan:    dep.expectedFuelLiter,
						startTime:      dep.startTime,
						fuelGetID:      dep.fuelGetID,
						timer:          timer,
						startSpeedTime: startSpeedTime,
						endSpeedTime:   lastFuelChangeTime,
						captureError:   dep.captureError,
					}, logger)
					return
				}
			}

		}
	}
}

type startFuelGetCompleteDep struct {
	jarNumber      string  // Номер емкости
	fuelType       string  // Тип топлива
	docNumber      string  // Номер документа
	beforeFuelGet  float32 // Показания уровнемера до пополнения
	afterFuelGet   float32 // Количество топлива после пополнения
	fuelGetPlan    float32 // Пополнили по плану
	startTime      int64   // Время начало пополнения
	fuelGetID      string  // Идентификатор транзакции
	timer          int     // Таймаут
	startSpeedTime int64   // Время начала подсчета скорости пополнения
	endSpeedTime   int64   // Время конца подсчета скорости пополнения
	captureError   CaptureErrorDep
}

func (p *Processing) startFuelGetComplete(ctx context.Context, dep startFuelGetCompleteDep, logger *slog.Logger) {
	logger.Info("начало процесса завершения пополнения.")

	logger.Info("отключение насоса.")
	err := RunWithTimeout(func() error {
		return p.driver.ControllerDriver.Adapter.DisablePump(dep.jarNumber)
	}, time.Duration(p.driverConfig.ControllerTimeout)*time.Second)
	if err != nil {
		logger.Error("не удалось отключить насос.", "err", err)
		p.CaptureError(fmt.Errorf("Controller.DisablePump error: %v", err), dep.captureError)
	}

	monitoringFinishTime := time.Now().Unix()
	timer := dep.timer
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var sensorAfterGive models.SensorInfo
	var errMsg = ""

	afterJarInfo := p.lastSENSTelemetry[dep.jarNumber]
	afterTemp := p.lastTempTelemetry[dep.jarNumber]

	sensorAfterGive = models.SensorInfo{
		H:  float64(afterJarInfo.H),
		T:  float64(afterJarInfo.T),
		Pr: float64(afterJarInfo.Pr),
		U:  float64(afterJarInfo.U),
		G:  float64(afterJarInfo.G),
		R:  float64(afterJarInfo.R),
		U1: float64(afterJarInfo.U1),
		H2: float64(afterJarInfo.H2),
		Ut: float64(afterJarInfo.Ut),
		Rt: float64(afterJarInfo.Rt),
		Ri: float64(afterJarInfo.Ri),
		Tr: float64(afterJarInfo.Tr),
		U2: float64(afterJarInfo.U2),
		Nt: afterTemp.nt,
		Dg: float64(afterJarInfo.Dg),
		Ts: float64(afterJarInfo.Ts),
	}

	factLiter := (afterJarInfo.U * float32(p.kazsConfig.FuelGetConfig.FuelGetUnits)) - dep.beforeFuelGet

	// === Расчет скорости пополнения ===
	var speed float32
	if factLiter == 0 || dep.startSpeedTime == 0 || dep.endSpeedTime == 0 {
		speed = 0
	} else {
		timeInMinutes := float32(dep.endSpeedTime-dep.startSpeedTime) / 60

		if timeInMinutes > 0 {
			speed = factLiter / timeInMinutes
		} else {
			speed = 0
		}
	}
	logger.Info("средняя скорость пополнения.", "avg_speed_l_per_m", speed, "time_start", dep.startSpeedTime, "time_end", dep.endSpeedTime, "liters", factLiter)

	// Указываем время завершения транзакции в БД
	logger.Info("обновление транзакции в БД.")
	err = p.repository.FuelGet.UpdateFuelGetTransaction(models.UpdateFuelGetTransactionDep{
		FuelGetID:            dep.fuelGetID,
		MonitoringFinishTime: &monitoringFinishTime,
		FuelLiters:           &factLiter,
		SensorAfterGive:      &sensorAfterGive,
		Speed:                &speed,
		Errors:               &errMsg,
	})

	if err != nil {
		logger.Error("не удалось обновить транзакцию в БД.", "err", err)
		p.CaptureError(fmt.Errorf("UpdateFuelGetTransaction error: %v", err), dep.captureError)
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("процесс завершения пополнения завершен.")
			logger.Info("формирование и отправка отчета на сервер.")
			err := p.createAndSendFuelGetReport(dep.fuelGetID, dep.jarNumber, "", dep.captureError, dep.beforeFuelGet, logger)
			if err != nil {
				logger.Error("не удалось отправить отчет на сервер.", "err", err)
				p.CaptureError(fmt.Errorf("createAndSendReport error: %v", err), dep.captureError)

				_ = p.repository.ErrorLogs.InsertError(models.Errors{
					Time:    time.Now().Unix(),
					Handler: "FuelGet",
					Id:      dep.fuelGetID,
					Error:   fmt.Sprintf("createAndSendReport error: %v", err),
				})
			}

			logger.Info("разблокировка потока.")
			p.UpdateJarStatus(dep.jarNumber, false)

			logger.Info("экран 0")
			p.appGui.CreateDefaultScreen(dep.jarNumber)

			logger.Info("процесс пополнения завершен.")
			return

		case <-ticker.C:
			timer--
			p.appGui.CreateFuelGetCompleteScreen(
				dep.jarNumber,
				dep.fuelType,
				dep.docNumber,
				dep.beforeFuelGet,
				dep.afterFuelGet,
				dep.fuelGetPlan,
				dep.afterFuelGet-dep.beforeFuelGet,
				dep.startTime,
				monitoringFinishTime,
				timer,
			)
		}
	}
}

// createAndSendReport Метод для формирования и отправки отчета на сервер
func (p *Processing) createAndSendFuelGetReport(fuelGetID, jarNumber, errors string, captureError CaptureErrorDep, beforeFuelGet float32, logger *slog.Logger) error {
	logger.Info("начало процесса формирования отчета.")

	endTime := time.Now().Unix()

	// Получаем показания с уровнемера после пополнения
	var sensorAfterGive models.SensorInfo
	var errMsg string

	logger.Info("получение телеметрии после пополнения.")
	afterJarInfo := p.lastSENSTelemetry[jarNumber]
	afterTemp := p.lastTempTelemetry[jarNumber]

	sensorAfterGive = models.SensorInfo{
		H:  float64(afterJarInfo.H),
		T:  float64(afterJarInfo.T),
		Pr: float64(afterJarInfo.Pr),
		U:  float64(afterJarInfo.U),
		G:  float64(afterJarInfo.G),
		R:  float64(afterJarInfo.R),
		U1: float64(afterJarInfo.U1),
		H2: float64(afterJarInfo.H2),
		Ut: float64(afterJarInfo.Ut),
		Rt: float64(afterJarInfo.Rt),
		Ri: float64(afterJarInfo.Ri),
		Tr: float64(afterJarInfo.Tr),
		U2: float64(afterJarInfo.U2),
		Nt: afterTemp.nt,
		Dg: float64(afterJarInfo.Dg),
		Ts: float64(afterJarInfo.Ts),
	}

	factLiter := (afterJarInfo.U * float32(p.kazsConfig.FuelGetConfig.FuelGetUnits)) - beforeFuelGet

	logger.Info("обновление транзакции в БД.")
	updateErr := p.repository.FuelGet.UpdateFuelGetTransaction(models.UpdateFuelGetTransactionDep{
		FuelGetID:       fuelGetID,
		FuelLiters:      &factLiter,
		EndTime:         &endTime,
		SensorAfterGive: &sensorAfterGive,
		Errors:          &errMsg,
	})
	if updateErr != nil {
		logger.Error("не удалось обновить транзакцию в БД.", "err", updateErr)
		p.CaptureError(updateErr, captureError)
	}

	logger.Info("получение транзакции из БД.")
	receipt, err := p.repository.FuelGet.GetFuelGetTransaction(fuelGetID)
	if err != nil {
		logger.Error("не удалось получить транзакцию из БД.", "err", err)
		return fmt.Errorf("FuelGet.GetFuelGetTransaction error: %v", err)
	}

	if receipt.SendStatus {
		logger.Info("отчет уже был отправлен на сервер.")
		return nil
	}

	errMsg += receipt.Errors

	logger.Info("POST FuelGetReceipt", "timeout_sec", p.appConfig.FuelGetReceiptTimeout)
	var status = false
	err = RunWithTimeout(func() error {
		return p.kazsOperator.FuelGetReceipt(
			&integration.FuelGetReceipt{
				FuelGetID: fuelGetID,
			},
			&integration.FuelGetReceiptRequest{
				KazsNumber: p.kazsOperator.KazsNumber,
				JarId:      jarNumber,
				FuelType:   receipt.FuelType,
				StartTime:  receipt.StartTime,
				EndTime:    endTime,
				DocNumber:  receipt.DocNumber,
				FuelLiter:  float64(factLiter),
				SensorBeforeGet: integration.SensorInfo{
					H:  receipt.SensorBeforeGet.H,
					T:  receipt.SensorBeforeGet.T,
					Pr: receipt.SensorBeforeGet.Pr,
					U:  receipt.SensorBeforeGet.U,
					G:  receipt.SensorBeforeGet.G,
					R:  receipt.SensorBeforeGet.R,
					U1: receipt.SensorBeforeGet.U1,
					H2: receipt.SensorBeforeGet.H2,
					Ut: receipt.SensorBeforeGet.Ut,
					Rt: receipt.SensorBeforeGet.Rt,
					Ri: receipt.SensorBeforeGet.Ri,
					Tr: receipt.SensorBeforeGet.Tr,
					U2: receipt.SensorBeforeGet.U2,
					Nt: receipt.SensorBeforeGet.Nt,
					Dg: receipt.SensorBeforeGet.Dg,
					Ts: receipt.SensorBeforeGet.Ts,
				},
				SensorAfterGet: integration.SensorInfo{
					H:  float64(afterJarInfo.H),
					T:  float64(afterJarInfo.T),
					Pr: float64(afterJarInfo.Pr),
					U:  float64(afterJarInfo.U),
					G:  float64(afterJarInfo.G),
					R:  float64(afterJarInfo.R),
					U1: float64(afterJarInfo.U1),
					H2: float64(afterJarInfo.H2),
					Ut: float64(afterJarInfo.Ut),
					Rt: float64(afterJarInfo.Rt),
					Ri: float64(afterJarInfo.Ri),
					Tr: float64(afterJarInfo.Tr),
					U2: float64(afterJarInfo.U2),
					Nt: afterTemp.nt,
					Dg: float64(afterJarInfo.Dg),
					Ts: float64(afterJarInfo.Ts),
				},
				Errors: errMsg,
			})
	}, time.Duration(p.appConfig.FuelGetReceiptTimeout)*time.Second)
	if err != nil {
		logger.Error("ошибка запроса.", "handler", "FuelGetReceipt", "err", err)
		return fmt.Errorf("KazsOperator.FuelGetReceipt error: %v", err)
	} else {
		logger.Info("отчет успешно отправлен на сервер.")
		status = true
	}

	logger.Info("обновление транзакции в БД.")
	updateErr = p.repository.FuelGet.UpdateFuelGetTransaction(models.UpdateFuelGetTransactionDep{
		FuelGetID:  fuelGetID,
		SendStatus: &status,
	})
	if updateErr != nil {
		logger.Error("не удалось обновить транзакцию в БД.", "err", err)
		return fmt.Errorf("FuelGet.UpdateFuelGetTransaction error: %v", updateErr)
	}

	return nil
}

func abs(f float32) float32 {
	if f < 0 {
		return -f
	}
	return f
}
