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
	"strings"
	"time"
)

const (
	StatusIdle              = "30"
	StatusNozzleLifted      = "31"
	StatusFuelingComplete   = "34"
	StatusFuelingInProgress = "33"
	StatusAuthorized        = "32"

	StopPumpTimeParameter = "0x3B"
)

// FuelGive Метод реализующий алгоритм заправки
func (p *Processing) FuelGive(qrInfo models.ScannerResponse) error {
	// === Инициализация логгера и sentry ===
	logger := p.logger.Transaction("fuel_give", qrInfo.TID)
	captureError := CaptureErrorDep{
		tid:        qrInfo.TID,
		operation:  "FuelGive",
		kazsNumber: p.kazsOperator.KazsNumber,
	}

	logger.Info("статус потоков.",
		"one_jar_active", p.oneJarActiveProcess,
		"two_jar_active", p.twoJarActiveProcess)

	// === Проверка на занятость ёмкостей ===
	if p.oneJarActiveProcess && p.twoJarActiveProcess {
		logger.Warn("все емкости заняты.")
		p.CaptureError(fmt.Errorf("all jars is busy"), captureError)

		_ = p.repository.ErrorLogs.InsertError(models.Errors{
			Time:    time.Now().Unix(),
			Handler: "FuelGive",
			Id:      qrInfo.TID,
			Error:   "all jars is busy",
		})

		return nil
	}

	// === Блокировка сканера ===
	logger.Info("блокировка сканера.")
	p.UpdateQRActive(false)

	// === Получение информации с сервера ===
	logger.Info("GET FuelGiveInfo", "timeout_sec", p.appConfig.FuelGetInfoTimeout)
	res := RunWithTimeoutValue(func() (*integration.FuelInfoResponse, error) {
		return p.kazsOperator.FuelGiveInfo(&integration.FuelGiveInfoRequest{
			TID: qrInfo.TID,
		})
	}, time.Duration(p.appConfig.FuelGiveInfoTimeout)*time.Second)

	// === Обработка ошибок FuelGiveInfo ===
	if res.Err != nil || res.Value.Error {
		var err string
		errMsg := "Нет ответа от сервера"
		if res.Err != nil {
			err = res.Err.Error()
		} else if res.Value.Error {
			err = res.Value.Message
			errMsg = res.Value.Message
		}
		logger.Error("ошибка запроса.", "handler", "FuelGiveInfo", "err", err)
		p.CaptureError(fmt.Errorf("kazsOperator.FuelGiveInfo error %v", err), captureError)

		// === Выводим ошибку в свободном потоке ===
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
			return nil
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

		_ = p.repository.ErrorLogs.InsertError(models.Errors{
			Time:    time.Now().Unix(),
			Handler: "FuelGive",
			Id:      qrInfo.TID,
			Error:   fmt.Sprintf("kazsOperator.FuelGiveInfo error %v", res.Err),
		})
		return res.Err
	}

	captureError.jarNumber = res.Value.Result.JarId
	logger = logger.With("jar_id", res.Value.Result.JarId)

	// === Проверка, что указанный поток свободен ===
	if res.Value.Result.JarId == "1" && p.oneJarActiveProcess || res.Value.Result.JarId == "2" && p.twoJarActiveProcess {
		logger.Warn("выбранная ёмкость занята.")
		p.CaptureError(fmt.Errorf("jars №%v busy", res.Value.Result.JarId), captureError)
		// Выводим ошибку в свободном потоке
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
			return nil
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
			Handler: "FuelGive",
			Id:      qrInfo.TID,
			Error:   fmt.Sprintf("jars №%v busy", res.Value.Result.JarId),
		})
		return nil
	}

	// === Блокировка соответствующего процесса ===
	logger.Info("процесс заблокирован.")
	p.UpdateJarStatus(res.Value.Result.JarId, true)

	// === Разблокировка сканера ===
	logger.Info("разблокировка сканера.")
	p.UpdateQRActive(true)
	time.Sleep(500 * time.Millisecond)

	// === Отображение экрана 'обработка' ===
	logger.Info("экран 'обработка'.")
	p.appGui.CreateDownloadScreen(res.Value.Result.JarId)

	// === Проверка статуса ТРК ===
	logger.Info("проверка на незавершенные заправки.")
	err := p.CheckTRKStatus(res.Value.Result.JarId, logger)
	if err != nil {
		logger.Error("не удалось проверить на незавершенные заправки.", "err", err)
		errMsg := fmt.Errorf("TRK status err: %v", err)

		var targetContent *fyne.Container
		if res.Value.Result.JarId == "1" {
			targetContent = p.appGui.LeftSection.Content
		} else {
			targetContent = p.appGui.RightSection.Content
		}

		if targetContent != nil {
			fyne.Do(func() {
				p.appGui.ShowSectionDialog(targetContent, "Ошибка", "ТРК не отвечает", 10, func() {
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
			Error:   errMsg.Error(),
		})
		return errMsg
	}

	// === Получение телеметрии до заправки ===
	logger.Info("получение телеметрии до заправка.")
	var sensorBeforeGive models.SensorInfoGive
	errMsg := ""
	beforeTelemetry := p.lastSENSTelemetry[res.Value.Result.JarId]

	// === Проверка на актуальность телеметрии ===
	if time.Now().Unix()-beforeTelemetry.timeAS > p.kazsConfig.FuelGiveConfig.ActualTimeSENS {
		logger.Warn("неактуальная телеметрия до заправки.", "age_sec", time.Now().Unix()-beforeTelemetry.timeAS)
		p.CaptureError(fmt.Errorf("old sens telemetry"), captureError)

		_ = p.repository.ErrorLogs.InsertError(models.Errors{
			Time:    time.Now().Unix(),
			Handler: "FuelGive",
			Id:      qrInfo.TID,
			Error:   fmt.Sprintf("GetLastTelemetry error %v", err),
		})

		errMsg += fmt.Sprint("Неактуальная телеметрия до заправки. ", time.Now().Unix(), beforeTelemetry.timeAS, p.kazsConfig.FuelGiveConfig.ActualTimeSENS)
	}

	sensorBeforeGive = models.SensorInfoGive{
		T:  float64(beforeTelemetry.T),
		U:  float64(beforeTelemetry.U),
		R:  float64(beforeTelemetry.R),
		U1: float64(beforeTelemetry.U1),
		Ri: float64(beforeTelemetry.Ri),
		Tr: float64(beforeTelemetry.Tr),
		U2: float64(beforeTelemetry.U2),
		H:  float64(beforeTelemetry.H),
		G:  float64(beforeTelemetry.G),
	}

	// === Подтверждение с сервера ===
	logger.Info("GET FuelGiveConfirmation", "timeout_sec", p.appConfig.FuelGiveConfirmationTimeout)
	var resErr = errors.New("")
	resConf := RunWithTimeoutValue(func() (*integration.FuelConfirmationResponse, error) {
		return p.kazsOperator.FuelGiveConfirmation(&integration.FuelGiveConfirmationRequest{
			FuelGiveID: qrInfo.TID,
		})
	}, time.Duration(p.appConfig.FuelGiveConfirmationTimeout)*time.Second)

	if resConf.Err != nil || resConf.Value.Error {
		var resConfErr string
		if resConf.Err != nil {
			resConfErr = resConf.Err.Error()
		} else if resConf.Value.Error {
			resConfErr = resConf.Value.Message
		}
		logger.Error("ошибка запроса.", "handler", "FuelGiveConfirmation", "err", resConfErr)

		resErr = fmt.Errorf("kazs.Operator.FuelGiveConfirmation error: %v", resConf.Err)
		if resConf.Value.Error {
			resErr = fmt.Errorf("kazsOperator.FuelGiveConfirmation error: %v", resConf.Value.Message)
		}
		p.CaptureError(resErr, captureError)

		var targetContent *fyne.Container
		if res.Value.Result.JarId == "1" {
			targetContent = p.appGui.LeftSection.Content
		} else {
			targetContent = p.appGui.RightSection.Content
		}

		errMsg1 := "Нет ответа от сервера"
		if resConf.Value.Error {
			errMsg1 = resConf.Value.Message
		}

		if targetContent != nil {
			fyne.Do(func() {
				p.appGui.ShowSectionDialog(targetContent, "Ошибка", errMsg1, 10, func() {
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
			Error:   resErr.Error(),
		})
		return resErr
	}

	// === Сохранение данных о начале транзакции в БД ===
	logger.Info("сохранение транзакции в БД.")
	startTime := time.Now().Unix()
	fuelLiters := float32(0)
	endTime := int64(0)
	err = p.repository.FuelGive.InsertFuelGiveTransaction(models.InsertFuelGiveTransactionDep{
		FuelGiveID:       qrInfo.TID,
		KazsNumber:       p.kazsOperator.KazsNumber,
		JarNumber:        resConf.Value.Result.JarId,
		StartTime:        startTime,
		FuelType:         resConf.Value.Result.FuelType,
		DocNumber:        resConf.Value.Result.DocNumber,
		FuelLitersPlan:   resConf.Value.Result.FuelLiter,
		FuelLiters:       &fuelLiters,
		SensorBeforeGive: &sensorBeforeGive,
		Errors:           &errMsg,
		EndTime:          &endTime,
		SendStatus:       false,
	})
	if err != nil {
		logger.Error("не удалось сохранить транзакцию в БД.", "err", err)
		resErr = fmt.Errorf("failed to insert fuel give transaction: %v", err)

		p.CaptureError(resErr, captureError)

		var targetContent *fyne.Container
		if res.Value.Result.JarId == "1" {
			targetContent = p.appGui.LeftSection.Content
		} else {
			targetContent = p.appGui.RightSection.Content
		}

		if targetContent != nil {
			fyne.Do(func() {
				p.appGui.ShowSectionDialog(targetContent, "Ошибка", "Не удалось начать заправку", 10, func() {
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
			Error:   resErr.Error(),
		})

		return resErr

	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.kazsConfig.FuelGiveConfig.FuelGiveStartScreenTimeout)*time.Second)

	// === Отрисовки экрана с обратным отчетом ===
	go p.setFuelGiveStartScreen(ctx, setFuelGiveStartScreenDep{
		jarNumber:      res.Value.Result.JarId,
		fuelType:       res.Value.Result.FuelType,
		beforeJarInfo:  beforeTelemetry,
		expectedAmount: float32(res.Value.Result.FuelLiter),
	}, logger)

	// === Мониторинг начала заправки ===
	go p.monitoringFuelGiveStart(ctx, cancel, monitoringFuelGiveStartDep{
		fuelGiveID:    qrInfo.TID,
		jarNumber:     res.Value.Result.JarId,
		fuelType:      resConf.Value.Result.FuelType,
		docNumber:     resConf.Value.Result.DocNumber,
		expectedLiter: float32(res.Value.Result.FuelLiter),
		captureError:  captureError,
		startTime:     startTime,
		beforeJarInfo: beforeTelemetry,
		err:           errMsg,
	}, logger)

	return nil
}

type setFuelGiveStartScreenDep struct {
	jarNumber      string     // Номер емкости
	fuelType       string     // Тип тополя
	beforeJarInfo  SENSStatus // Показания уровнемера в начале пополнения
	expectedAmount float32    // Количество литров для заправки
}

// setFuelGiveStartScreen Метод отрисовки экрана ожидания начала заправки
func (p *Processing) setFuelGiveStartScreen(ctx context.Context, dep setFuelGiveStartScreenDep, logger *slog.Logger) {
	logger.Info("начало процесса отрисовки экрана 'ожидание начала заправки'.")
	deadline, ok := ctx.Deadline()
	var timeout int

	if ok {
		timeout = int(time.Until(deadline).Seconds())
	} else {
		timeout = p.kazsConfig.FuelGiveConfig.FuelGiveStartScreenTimeout
	}

	logger.Info("экран 'ожидание начала заправки'", "timeout_sec", timeout)
	p.appGui.CreateFuelGiveStartScreen(dep.jarNumber, dep.expectedAmount, dep.fuelType, timeout)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for i := timeout - 1; i >= 0; i-- {
		select {
		case <-ctx.Done(): // Контекст отменен, либо по таймауту, либо вручную
			return
		case <-ticker.C:
			p.appGui.CreateFuelGiveStartScreen(dep.jarNumber, dep.expectedAmount, dep.fuelType, i)
		}
	}
}

type monitoringFuelGiveStartDep struct {
	fuelGiveID    string
	jarNumber     string
	fuelType      string
	docNumber     string
	expectedLiter float32 // Количество литров для заправки
	captureError  CaptureErrorDep
	beforeJarInfo SENSStatus
	startTime     int64
	err           string
}

// monitoringFuelGiveStart Метод мониторинга начала заправки
func (p *Processing) monitoringFuelGiveStart(ctx context.Context, cancel context.CancelFunc, dep monitoringFuelGiveStartDep, logger *slog.Logger) {
	logger.Info("запущен процесс мониторинга начала заправки.")
	defer cancel()
	errMsg1 := ""
	for {
		select {
		case <-ctx.Done():
			// === Вышел таймаут снятия пистолета ===
			errMsg := fmt.Sprintf("Информационная. Пистолет не был снят более %d секунд после подтверждения с сервера.", p.kazsConfig.FuelGiveConfig.FuelGiveStartScreenTimeout)
			logger.Warn(errMsg)

			// === Завершение заправки ===
			completeCtx, _ := context.WithTimeout(context.Background(), time.Duration(p.kazsConfig.FuelGiveConfig.FuelGiveCompleteScreenTimeout)*time.Second)

			go p.startFuelGiveComplete(completeCtx, startFuelGiveCompleteDep{
				fuelGiveID:    dep.fuelGiveID,
				jarNumber:     dep.jarNumber,
				fuelType:      dep.fuelType,
				docNumber:     dep.docNumber,
				expectedLiter: dep.expectedLiter,
				factLiter:     0,
				startTime:     dep.startTime,
				captureError:  dep.captureError,
				errors:        errMsg,
				avgSpeed:      0,
			}, logger)
			return
		default:
		}

		trkStatus := p.TRKRequest(GetTRKStatus, dep.jarNumber, 0, logger)

		if trkStatus.Err != nil {
			if errMsg1 == "" {
				errMsg1 = fmt.Sprint("Не удалось получить статус пистолета на ТРК.")
				p.CaptureError(fmt.Errorf(errMsg1), dep.captureError)

				_ = p.repository.ErrorLogs.InsertError(models.Errors{
					Time:    time.Now().Unix(),
					Handler: "FuelGive",
					Id:      dep.fuelGiveID,
					Error:   errMsg1,
				})
			}
			continue
		}

		// === Пистолет висит ===
		if trkStatus.ValueStr == StatusIdle {
			continue
		}

		// === Пистолет снят ===
		if trkStatus.ValueStr == StatusNozzleLifted {
			// === Задание количества литров на ТРК ===
			res := p.TRKRequest(SetFuelGive, dep.jarNumber, dep.expectedLiter, logger)

			if res.Err == nil {
				// === Санкционирование ТРК ===
				res = p.TRKRequest(ApprovalTRK, dep.jarNumber, 0, logger)
			}

			if res.Err != nil {
				errMsg := fmt.Errorf("Критическая. Не удалось санкционировать ТРК. Заправка завершена с 0 литров.")
				logger.Error(errMsg.Error(), "err", res.Err.Error())
				p.CaptureError(errMsg, dep.captureError)

				_ = p.repository.ErrorLogs.InsertError(models.Errors{
					Time:    time.Now().Unix(),
					Handler: "FuelGive",
					Id:      dep.fuelGiveID,
					Error:   errMsg.Error(),
				})

				// === Завершение заправки ===
				completeCtx, _ := context.WithTimeout(context.Background(), time.Duration(p.kazsConfig.FuelGiveConfig.FuelGiveCompleteScreenTimeout)*time.Second)

				go p.startFuelGiveComplete(completeCtx, startFuelGiveCompleteDep{
					fuelGiveID:    dep.fuelGiveID,
					jarNumber:     dep.jarNumber,
					fuelType:      dep.fuelType,
					docNumber:     dep.docNumber,
					expectedLiter: dep.expectedLiter,
					factLiter:     0,
					startTime:     dep.startTime,
					captureError:  dep.captureError,
					errors:        errMsg.Error(),
					avgSpeed:      0,
				}, logger)

				return
			}

			// === Начало процедуры заправки ===
			logger.Info("экран 'процедура заправки'.")
			p.appGui.CreateFuelGiveInProgressScreen(dep.jarNumber, dep.fuelType, 0, dep.expectedLiter)

			go p.startFuelGiveInProgress(startFuelGiveInProgressDep{
				fuelGiveID:    dep.fuelGiveID,
				jarNumber:     dep.jarNumber,
				fuelType:      dep.fuelType,
				docNumber:     dep.docNumber,
				expectedLiter: dep.expectedLiter,
				startTime:     dep.startTime,
				liter:         0,
				sensorBeforeInfo: models.SensorInfoGive{
					T:  float64(dep.beforeJarInfo.T),
					U:  float64(dep.beforeJarInfo.U),
					R:  float64(dep.beforeJarInfo.R),
					U1: float64(dep.beforeJarInfo.U1),
					Ri: float64(dep.beforeJarInfo.Ri),
					Tr: float64(dep.beforeJarInfo.Tr),
					U2: float64(dep.beforeJarInfo.U2),
				},
				captureError: dep.captureError,
			}, logger)

			return
		}
	}
}

type startFuelGiveInProgressDep struct {
	fuelGiveID       string
	jarNumber        string
	fuelType         string
	docNumber        string
	expectedLiter    float32 // Количество литров для заправки
	startTime        int64
	liter            float32
	captureError     CaptureErrorDep
	sensorBeforeInfo models.SensorInfoGive
}

// startFuelGiveInProgress Метод для процесса заправки (подразумевает вечный цикл, в случае нарушения взаимодействия с ТРК)
func (p *Processing) startFuelGiveInProgress(dep startFuelGiveInProgressDep, logger *slog.Logger) {
	logger.Info("начало мониторинга процесса заправки.")
	lastChangeTime := time.Now().Unix()                                    // Последнее изменение количества литров на ТРК
	lastFuelGive := dep.liter                                              // Последнее количество литров
	lastSuccessTime := time.Now().Unix()                                   // Время последней удачной попытки получения количества литров и статуса ТРК
	maxFailedTimeout := p.kazsConfig.FuelGiveConfig.FailedTRKResponse      // 5 min Таймаут ожидания удачной попытки получения количества литров и статуса ТРК, после которого выводится окно технической неисправности
	countFuelGiveEnd := p.kazsConfig.FuelGiveConfig.CountFuelGiveEnd       // Количество попыток для завершения заправки, после которой выводится окно технической неисправности
	addTimeTrkFinish := int64(p.kazsConfig.FuelGiveConfig.FuelGiveTimeout) // Добавочное время к настройкам ТРК автоматического завершения заправки, после которого выводится окно технической неисправности

	var (
		isTimeoutReported          bool
		isInvalidStatusReported    bool
		isAutoFinishFailedReported bool
		isCompleteFailedReported   bool
		isFuelGetErrorReported     bool
		isStatusGetErrorReported   bool
		errorMessages              []string
	)

	prevStatus := ""

	// === Проверяем, что получили настройки ТРК ===
	var stopPumpTimeout float32
	timeout, ok := p.trkSettings[dep.jarNumber].GeneralParameters[StopPumpTimeParameter]
	if !ok {
		stopPumpTimeout = p.kazsConfig.FuelGiveConfig.StopPumpTimeout
	} else {
		stopPumpTimeout = timeout
	}

	// === Подсчет количества литров
	var timeStart int64
	var literStart float32
	var reachedStart bool

	var fuelPoints []struct {
		liters float32
		ts     int64
	}

	for {
		fuelChanged := false
		var newFuelGive float32

		// === ТРК не отвечает в сумме больше N секунд ===
		if time.Now().Unix()-lastSuccessTime > maxFailedTimeout {
			logger.Warn("трк не отвечает более N секунд", "failed_timeout", time.Now().Unix()-lastSuccessTime)
			if !isTimeoutReported {
				errMsg := fmt.Sprintf("Не удалось получить количество литров и статус ТРК более чем %d секунд", maxFailedTimeout)
				p.CaptureError(fmt.Errorf(errMsg), dep.captureError)
				_ = p.repository.ErrorLogs.InsertError(models.Errors{
					Time:    time.Now().Unix(),
					Handler: "FuelGive",
					Id:      dep.fuelGiveID,
					Error:   errMsg,
				})

				errorMessages = append(errorMessages, errMsg)
				isTimeoutReported = true
			}

			// === Отображаем экран технических неполадок
			logger.Info("Экран 'тех. неполадки'")
			p.appGui.CreateTechnicalErrorScreen(dep.jarNumber)
		}

		if countFuelGiveEnd < 1 {
			// === Отображаем экран технических неполадок
			logger.Warn("вышло количество попыток завершить заправку.")
			logger.Info("экран 'тех. неполадки.'")
			p.appGui.CreateTechnicalErrorScreen(dep.jarNumber)
		}

		// === Получение литров ===
		fuelGive := p.TRKRequest(GetFuelGiveStatus, dep.jarNumber, 0, logger)

		if fuelGive.Err != nil {
			if !isFuelGetErrorReported {
				errMsg := fmt.Sprintf("Ошибка получения количества заправленных литров на ТРК: %s", fuelGive.Err)
				p.CaptureError(fmt.Errorf(errMsg), dep.captureError)
				_ = p.repository.ErrorLogs.InsertError(models.Errors{
					Time:    time.Now().Unix(),
					Handler: "FuelGive",
					Id:      dep.fuelGiveID,
					Error:   errMsg,
				})
				errorMessages = append(errorMessages, errMsg)
				isFuelGetErrorReported = true
			}
			continue
		} else if fuelGive.ValueFloat > lastFuelGive {
			newFuelGive = fuelGive.ValueFloat
			fuelChanged = true
		}

		// === Получение статуса ТРК ===
		trkStatus := p.TRKRequest(GetTRKStatus, dep.jarNumber, 0, logger)
		if trkStatus.Err == nil {
			if prevStatus == StatusAuthorized && trkStatus.ValueStr == StatusFuelingInProgress {
				lastChangeTime = time.Now().Unix()
			}
			prevStatus = trkStatus.ValueStr
		}

		// === Изменилось количество топлива ===
		if fuelChanged && trkStatus.Err == nil && trkStatus.ValueStr != StatusAuthorized {
			lastFuelGive = newFuelGive
			lastChangeTime = time.Now().Unix()

			fuelPoints = append(fuelPoints, struct {
				liters float32
				ts     int64
			}{liters: lastFuelGive, ts: lastChangeTime})

			if !reachedStart && lastFuelGive >= p.kazsConfig.FuelGiveConfig.FuelGiveSpeedStart {
				timeStart = lastChangeTime
				literStart = lastFuelGive
				reachedStart = true
			}

			p.appGui.CreateFuelGiveInProgressScreen(dep.jarNumber, dep.fuelType, fuelGive.ValueFloat, dep.expectedLiter)
		}

		switch {
		case trkStatus.Err != nil || trkStatus.ValueStr == "":
			if !isStatusGetErrorReported {
				errMsg := fmt.Sprintf("Ошибка получения статуса ТРК: %s", trkStatus.Err)
				p.CaptureError(fmt.Errorf(errMsg), dep.captureError)
				_ = p.repository.ErrorLogs.InsertError(models.Errors{
					Time:    time.Now().Unix(),
					Handler: "FuelGive",
					Id:      dep.fuelGiveID,
					Error:   errMsg,
				})
				errorMessages = append(errorMessages, errMsg)
				isStatusGetErrorReported = true
			}
			continue

		// === Заправка завершена на ТРК ===
		case trkStatus.ValueStr == StatusFuelingComplete:
			endTime := time.Now().Unix()
			logger.Info("ТРК завершило заправку. Получение отчета и завершение заправки.", "trk_status", trkStatus.ValueStr, "end_time", endTime)

			litter, completeErr := p.completeFuelGive(dep, logger)

			if completeErr != nil {
				logger.Warn("не удалось завершить заправку на ТРК.", "err", completeErr)
				countFuelGiveEnd--
				if !isCompleteFailedReported {
					errMsg := fmt.Sprintf("Не удалось завершить заправку на ТРК: %s", completeErr)
					p.CaptureError(fmt.Errorf(errMsg), dep.captureError)
					_ = p.repository.ErrorLogs.InsertError(models.Errors{
						Time:    time.Now().Unix(),
						Handler: "FuelGive",
						Id:      dep.fuelGiveID,
						Error:   errMsg,
					})
					errorMessages = append(errorMessages, errMsg)
					isCompleteFailedReported = true
				}
				continue
			}

			// Отрисовка экрана ожидания (Экран 0)
			logger.Info("экран 'обработка'.")
			p.appGui.CreateDownloadScreen(dep.jarNumber)

			var avgSpeed float64
			if reachedStart && litter > p.kazsConfig.FuelGiveConfig.FuelGiveSpeedMin {
				half := litter / 2
				var timeEnd int64

				var literEnd float32

				for _, point := range fuelPoints {
					if point.liters >= half {
						timeEnd = point.ts
						literEnd = point.liters
						break
					}
				}
				if timeEnd > timeStart {
					avgSpeed = (float64(literEnd-literStart) / float64(timeEnd-timeStart)) * 60
					logger.Info("средняя скорость заправки.", "avg_speed_l_per_m", avgSpeed, "half_liters", half, "time_start", timeStart, "time_end", timeEnd, "liter_start", literStart, "liter_end", literEnd)
				}
			} else {
				logger.Info("средняя скорость не учитывается.")
			}

			// === Завершение заправки ===
			ctx, _ := context.WithTimeout(context.Background(), time.Duration(p.kazsConfig.FuelGiveConfig.FuelGiveCompleteScreenTimeout)*time.Second)

			go p.startFuelGiveComplete(ctx, startFuelGiveCompleteDep{
				fuelGiveID:    dep.fuelGiveID,
				jarNumber:     dep.jarNumber,
				fuelType:      dep.fuelType,
				docNumber:     dep.docNumber,
				expectedLiter: dep.expectedLiter,
				factLiter:     litter,
				startTime:     dep.startTime,
				captureError:  dep.captureError,
				errors:        strings.Join(errorMessages, "; "),
				avgSpeed:      avgSpeed,
			}, logger)
			return

		// === Неожиданный статус ТРК во время заправки ===
		case trkStatus.ValueStr != StatusFuelingInProgress && trkStatus.ValueStr != StatusAuthorized:
			if !isInvalidStatusReported {
				logger.Warn("невалидный статус ТРК после заправки.", "trk_status", trkStatus.ValueStr)
				errMsg := fmt.Sprintf("Критическая. Невозможный статус ТРК в процессе заправки: %s", trkStatus.ValueStr)
				p.CaptureError(fmt.Errorf(errMsg), dep.captureError)
				_ = p.repository.ErrorLogs.InsertError(models.Errors{
					Time:    time.Now().Unix(),
					Handler: "FuelGive",
					Id:      dep.fuelGiveID,
					Error:   errMsg,
				})
				errorMessages = append(errorMessages, errMsg)
				isInvalidStatusReported = true
			}
		}

		// === Успешное получение литров и статуса ===
		lastSuccessTime = time.Now().Unix()

		// === Проверка, что ТРК автоматически завершило заправку, если топливо не поступает N секунд ===
		it := time.Now().Unix() - lastChangeTime
		if it > int64(stopPumpTimeout)+addTimeTrkFinish && trkStatus.ValueStr != StatusAuthorized {
			logger.Warn("ТРК автоматически не завершила заправку.", "trk_status", trkStatus.ValueStr)
			if !isAutoFinishFailedReported {
				errMsg := fmt.Sprintf("Критическая. ТРК не завершило заправку. Топливо не шло более %d секунд. В настройках ТРК время отсечки %f", it, stopPumpTimeout)
				p.CaptureError(fmt.Errorf(errMsg), dep.captureError)
				_ = p.repository.ErrorLogs.InsertError(models.Errors{
					Time:    time.Now().Unix(),
					Handler: "FuelGive",
					Id:      dep.fuelGiveID,
					Error:   errMsg,
				})

				errorMessages = append(errorMessages, errMsg)
				isAutoFinishFailedReported = true

				// === Отображаем экран технических неполадок
				logger.Info("экран 'тех. неполадки'.")
				p.appGui.CreateTechnicalErrorScreen(dep.jarNumber)
			}
		}

		time.Sleep(1 * time.Second)

	}
}

func (p *Processing) completeFuelGive(dep startFuelGiveInProgressDep, logger *slog.Logger) (float32, error) {

	// === Получаем полный отчет о заправке ===
	fuelGive := p.TRKRequest(GetFullFuelGiveStatus, dep.jarNumber, 0, logger)
	if fuelGive.Err != nil {
		return 0, fuelGive.Err
	}

	// === Завершаем заправку на ТРК ===
	res := p.TRKRequest(FuelGiveSuccess, dep.jarNumber, 0, logger)
	if res.Err != nil {
		return 0, res.Err
	}

	return fuelGive.ValueFloat, nil

}

type startFuelGiveCompleteDep struct {
	fuelGiveID    string
	jarNumber     string
	fuelType      string
	docNumber     string
	expectedLiter float32
	factLiter     float32
	startTime     int64
	captureError  CaptureErrorDep
	errors        string
	avgSpeed      float64
}

// startFuelGiveComplete Метод для завершения процесса заправки
func (p *Processing) startFuelGiveComplete(ctx context.Context, dep startFuelGiveCompleteDep, logger *slog.Logger) {
	logger.Info("начало процедуры завершения заправки.")

	// === Инициализация таймера ===
	timer := p.kazsConfig.FuelGiveConfig.FuelGiveCompleteScreenTimeout
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// === Время окончания заправки ===
	endTime := time.Now().Unix()

	for {
		select {
		case <-ctx.Done():
			logger.Info("процесс завершения заправки завершен.")
			logger.Info("формирование и отправка отчета на сервер.")
			err := p.createAndSendReport(dep.fuelGiveID, dep.jarNumber, dep.errors, dep.captureError, dep.factLiter, endTime, logger, dep.avgSpeed)
			if err != nil {
				logger.Error("не удалось отправить отчет на сервер.", "err", err)
				p.CaptureError(fmt.Errorf("createAndSendReport error: %v", err), dep.captureError)

				_ = p.repository.ErrorLogs.InsertError(models.Errors{
					Time:    time.Now().Unix(),
					Handler: "FuelGive",
					Id:      dep.fuelGiveID,
					Error:   fmt.Sprintf("createAndSendReport error: %v", err),
				})
			}

			logger.Info("разблокировка потока.")
			p.UpdateJarStatus(dep.jarNumber, false)

			logger.Info("экран 0")
			p.appGui.CreateDefaultScreen(dep.jarNumber)

			logger.Info("процесс заправки завершен.")
			return
		case <-ticker.C:
			timer--
			p.appGui.CreateFuelGiveCompleteScreen(
				dep.jarNumber,
				dep.fuelType,
				dep.docNumber,
				dep.expectedLiter,
				dep.factLiter,
				dep.startTime,
				endTime,
				timer,
			)
		}
	}
}

// createAndSendReport Метод для формирования и отправки отчета на сервер
func (p *Processing) createAndSendReport(fuelGiveID, jarNumber, errors string, captureError CaptureErrorDep, factLiter float32, endTime int64, logger *slog.Logger, avgSpeed float64) error {
	logger.Info("начало процесса формирования отчета.")
	// Получаем показания с уровнемера
	var sensorAfterGive models.SensorInfoGive
	errMsg := errors

	logger.Info("получение телеметрии после заправки.")
	afterTelemetry := p.lastSENSTelemetry[jarNumber]

	if time.Now().Unix()-afterTelemetry.timeAS > int64(p.kazsConfig.FuelGiveConfig.FuelGiveCompleteScreenTimeout)-p.kazsConfig.FuelGiveConfig.ActualTimeSENS {
		logger.Warn("неактуальная телеметрия после заправки.", "age_sec", time.Now().Unix()-afterTelemetry.timeAS)
		p.CaptureError(fmt.Errorf("old sens telemetry"), captureError)
		errMsg += "Неактуальная телеметрия после заправки. "
	}

	sensorAfterGive = models.SensorInfoGive{
		T:  float64(afterTelemetry.T),
		U:  float64(afterTelemetry.U),
		R:  float64(afterTelemetry.R),
		U1: float64(afterTelemetry.U1),
		Ri: float64(afterTelemetry.Ri),
		Tr: float64(afterTelemetry.Tr),
		U2: float64(afterTelemetry.U2),
		H:  float64(afterTelemetry.H),
		G:  float64(afterTelemetry.G),
	}

	logger.Info("обновление транзакции в БД.")
	updateErr := p.repository.FuelGive.UpdateFuelGiveTransaction(models.UpdateFuelGiveTransactionDep{
		FuelGiveID:      fuelGiveID,
		EndTime:         &endTime,
		FuelLiters:      &factLiter,
		SensorAfterGive: &sensorAfterGive,
	})
	if updateErr != nil {
		logger.Error("не удалось обновить транзакцию в БД.", "err", updateErr)
		p.CaptureError(updateErr, captureError)
	}

	logger.Info("получение транзакции из БД.")
	receipt, err := p.repository.FuelGive.GetFuelGiveTransactionByID(fuelGiveID)
	if err != nil {
		logger.Error("не удалось получить транзакцию из БД.", "err", err)
		return fmt.Errorf("FuelGive.GetFuelGiveTransaction error: %v", err)
	}

	if receipt.SendStatus {
		logger.Info("отчет уже был отправлен на сервер.")
		return nil
	}

	errMsg += receipt.Errors

	var status = false

	logger.Info("POST FuelGiveReceipt", "timeout_sec", p.appConfig.FuelGiveReceiptTimeout)
	err = RunWithTimeout(func() error {
		return p.kazsOperator.FuelGiveReceipt(
			&integration.FuelGiveReceipt{
				FuelGiveID: fuelGiveID,
			},
			&integration.KazsFuelGiveReceiptRequest{
				KazsNumber: receipt.KazsNumber,
				JarId:      receipt.JarId,
				FuelType:   receipt.FuelType,
				StartTime:  receipt.StartTime,
				EndTime:    receipt.EndTime,
				DocNumber:  receipt.DocNumber,
				FuelLiter:  receipt.FuelLiter,
				SensorBeforeGive: integration.SensorInfoGive{
					T:  receipt.SensorBeforeGive.T,
					U:  receipt.SensorBeforeGive.U,
					R:  receipt.SensorBeforeGive.R,
					U1: receipt.SensorBeforeGive.U1,
					Ri: receipt.SensorBeforeGive.Ri,
					Tr: receipt.SensorBeforeGive.Tr,
					U2: receipt.SensorBeforeGive.U2,
					H:  receipt.SensorBeforeGive.H,
					G:  receipt.SensorBeforeGive.G,
				},
				SensorAfterGive: integration.SensorInfoGive{
					T:  receipt.SensorAfterGive.T,
					U:  receipt.SensorAfterGive.U,
					R:  receipt.SensorAfterGive.R,
					U1: receipt.SensorAfterGive.U1,
					Ri: receipt.SensorAfterGive.Ri,
					Tr: receipt.SensorAfterGive.Tr,
					U2: receipt.SensorAfterGive.U2,
					H:  receipt.SensorAfterGive.H,
					G:  receipt.SensorAfterGive.G,
				},
				Errors:   errMsg,
				AvgSpeed: avgSpeed,
			})
	}, time.Duration(p.appConfig.FuelGiveReceiptTimeout)*time.Second)
	if err != nil {
		logger.Error("ошибка запроса.", "handler", "FuelGiveReceipt", "err", err)
		return fmt.Errorf("KazsOperator.FuelGiveReceipt error: %v", err)
	} else {
		logger.Info("отчет успешно отправлен на сервер.")
		status = true
	}

	logger.Info("обновление транзакции в БД.")
	updateErr = p.repository.FuelGive.UpdateFuelGiveTransaction(models.UpdateFuelGiveTransactionDep{
		FuelGiveID: fuelGiveID,
		SendStatus: &status,
	})
	if updateErr != nil {
		logger.Error("не удалось обновить транзакцию в БД.", "err", err)
		return fmt.Errorf("FuelGive.UpdateFuelGiveTransaction error: %v", updateErr)
	}

	return nil
}

// CheckTRKStatus Проверка статуса ТРК
func (p *Processing) CheckTRKStatus(jarNumber string, logger *slog.Logger) error {
	logger.Info("начало процесса проверки ТРК.")
	captureError := CaptureErrorDep{
		operation:  "FuelGive",
		kazsNumber: p.kazsOperator.KazsNumber,
		jarNumber:  jarNumber,
	}
	trkStatus := p.TRKRequest(GetTRKStatus, jarNumber, 0, logger)

	if trkStatus.Err != nil {
		p.CaptureError(fmt.Errorf("TRKDriver.GetTRKStatus error: %v", trkStatus.Err), captureError)
		return fmt.Errorf("TRKDriver.GetTRKStatus error: %v", trkStatus.Err)
	}

	if trkStatus.ValueStr == StatusIdle || trkStatus.ValueStr == StatusNozzleLifted {
		return nil
	}

	if trkStatus.ValueStr != StatusFuelingComplete {
		logger.Warn("неожиданный статус ТРК.", "trk_status", trkStatus.ValueStr)
		p.CaptureError(fmt.Errorf("Критическая. Неожиданный статус ТРК №%v: %v", jarNumber, trkStatus.ValueStr), captureError)
		return fmt.Errorf("Критическая. Неожиданный статус ТРК №%v: %v", jarNumber, trkStatus.ValueStr)
	}

	fuelGive := p.TRKRequest(GetFullFuelGiveStatus, jarNumber, 0, logger)

	if fuelGive.Err != nil {
		p.CaptureError(fmt.Errorf("TRKDriver.GetFullFuelGiveStatus error: %v", fuelGive.Err), captureError)
		return fmt.Errorf("TRKDriver.GetFullFuelGiveStatus error: %v", fuelGive.Err)
	}

	success := p.TRKRequest(FuelGiveSuccess, jarNumber, 0, logger)

	if success.Err != nil {
		p.CaptureError(fmt.Errorf("TRKDriver.FuelGiveSuccess error: %v", success), captureError)
		return fmt.Errorf("TRKDriver.FuelGiveSuccess error: %v", success)
	}

	logger.Info("получение последней транзакции из БД.")
	lastFuelGive, err := p.repository.FuelGive.GetLastFuelGiveTransaction(jarNumber)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("заправка не была завершена на ТРК. Отчет отправлен.")
			p.CaptureError(fmt.Errorf("Заправка не была завершена на ТРК №%v. Отчет отправлен.", jarNumber), captureError)
			return nil
		}
		logger.Error("не удалось получить последнюю транзакции из БД.", "err", err)
		p.CaptureError(fmt.Errorf("Repo.GetLastFuelGiveTransaction error: %v", err), captureError)
		return fmt.Errorf("Repo.GetLastFuelGiveTransaction error: %v", err)
	}

	captureError.tid = lastFuelGive.FuelGiveID

	logger.Info("обновление транзакции в БД.")
	err = p.repository.FuelGive.UpdateFuelGiveTransaction(models.UpdateFuelGiveTransactionDep{
		FuelGiveID: lastFuelGive.FuelGiveID,
		FuelLiters: &fuelGive.ValueFloat,
	})
	if err != nil {
		logger.Error("не удалось обновить транзакцию в БД.", "err", err)
		p.CaptureError(err, captureError)
	}
	endTime := time.Now().Unix()

	logger.Info("формирование и отправка отчета на сервер.")
	err = p.createAndSendReport(lastFuelGive.FuelGiveID, jarNumber, lastFuelGive.Errors, captureError, fuelGive.ValueFloat, endTime, logger, 0)
	if err != nil {
		logger.Error("не удалось отправить отчет.", "err", err)
		p.CaptureError(fmt.Errorf("CreateAndSendReport error: %v", err), captureError)
		return fmt.Errorf("CreateAndSendReport error: %v", err)
	}

	return nil
}
