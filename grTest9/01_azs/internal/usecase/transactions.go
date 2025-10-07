package usecase

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"fuelazs/internal/driver/topaz/trk"
	"fuelazs/internal/integration"
	"fuelazs/internal/usecase/models"
	myntp "fuelazs/pkg/ntp"
	"github.com/beevik/ntp"
	"log/slog"
	"strconv"
	"time"
)

type ComMessage map[string]json.RawMessage

type RelayItem struct {
	Index string `json:"i"`
	State string `json:"s"`
}

// CheckUnsentFuelGetTransactions Метод для периодической проверки незавершенных транзакций
func (p *Processing) CheckUnsentFuelGetTransactions() {
	logger := p.logger.BaseError("check_unfinished_fuel_get_transactions")
	ticker := time.NewTicker(time.Duration(p.kazsConfig.FuelGetConfig.CheckUnsentTransactions) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if !p.kazsActivation || !p.startProgramm {
				continue
			}
			logger.Info("начало процедуры проверки незавершенных транзакций")
			for i := range p.trkSettings {
				if i == "1" && p.oneJarActiveProcess {
					continue
				}
				if i == "2" && p.twoJarActiveProcess {
					continue
				}
				p.UpdateJarStatus(i, true)
				_, _ = p.validationAndCheckLastFuelGet(i, logger)
				p.UpdateJarStatus(i, false)
			}
			logger.Info("процедура проверки незавершенных транзакций завершена")
		}
	}
}

// CheckUnsentFuelGiveTransactions Метод для периодической проверки незавершенных транзакций
func (p *Processing) CheckUnsentFuelGiveTransactions() {
	logger := p.logger.BaseError("check_unfinished_fuel_give_transactions")
	ticker := time.NewTicker(time.Duration(p.kazsConfig.FuelGiveConfig.CheckUnsentTransactions) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if !p.kazsActivation || !p.startProgramm {
				continue
			}
			logger.Info("начало процедуры проверки незавершенных транзакций.")
			for i := range p.trkSettings {
				if i == "1" && p.oneJarActiveProcess {
					continue
				}
				if i == "2" && p.twoJarActiveProcess {
					continue
				}
				p.UpdateJarStatus(i, true)
				_ = p.CheckTRKStatus(i, logger)
				p.UpdateJarStatus(i, false)
			}
			logger.Info("процедура проверки незавершенных транзакций завершена.")
		}
	}
}

// CheckUnsentTransactions Метод для периодической проверки неотправленных транзакций
func (p *Processing) CheckUnsentTransactions() {
	logger := p.logger.BaseError("check_unsent_transactions")
	ticker := time.NewTicker(time.Duration(p.kazsConfig.FuelGiveConfig.CheckUnsentTransactions) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if !p.kazsActivation || !p.startProgramm {
				continue
			}
			logger.Info("начало процедуры проверки неотправленных транзакций.")
			captureError := CaptureErrorDep{
				operation:  "CheckUnsentTransactions",
				kazsNumber: p.kazsOperator.KazsNumber,
			}

			logger.Info("получение неотправленных транзакций из БД.", "transaction", "fuel_give")
			unsentFuelGiveTransactions, err := p.repository.FuelGive.GetUnsentFuelGiveTransactions()
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				logger.Error("не удалось получить неотправленные транзакции", "transaction", "fuel_give", "err", err)
				p.CaptureError(fmt.Errorf("Repo.GetUnsentFuelGiveTransactions error: %v", err), captureError)
			}

			if len(unsentFuelGiveTransactions) != 0 {
				logger.Info("найдены неотправленные транзакции", "transaction", "fuel_give", "count", len(unsentFuelGiveTransactions))
				for _, v := range unsentFuelGiveTransactions {
					var status = false
					logger.Info("POST FuelGiveReceipt", "transaction", "fuel_give", "timeout_sec", p.appConfig.FuelGiveReceiptTimeout, "tid", v.FuelGiveID)
					err = RunWithTimeout(func() error {
						return p.kazsOperator.FuelGiveReceipt(
							&integration.FuelGiveReceipt{
								FuelGiveID: v.FuelGiveID,
							},
							&integration.KazsFuelGiveReceiptRequest{
								KazsNumber: v.KazsNumber,
								JarId:      v.JarId,
								FuelType:   v.FuelType,
								StartTime:  v.StartTime,
								EndTime:    v.EndTime,
								DocNumber:  v.DocNumber,
								FuelLiter:  v.FuelLiter,
								SensorBeforeGive: integration.SensorInfoGive{
									T:  v.SensorBeforeGive.T,
									U:  v.SensorBeforeGive.U,
									R:  v.SensorBeforeGive.R,
									U1: v.SensorBeforeGive.U1,
									Ri: v.SensorBeforeGive.Ri,
									Tr: v.SensorBeforeGive.Tr,
									U2: v.SensorBeforeGive.U2,
								},
								SensorAfterGive: integration.SensorInfoGive{
									T:  v.SensorAfterGive.T,
									U:  v.SensorAfterGive.U,
									R:  v.SensorAfterGive.R,
									U1: v.SensorAfterGive.U1,
									Ri: v.SensorAfterGive.Ri,
									Tr: v.SensorAfterGive.Tr,
									U2: v.SensorAfterGive.U2,
								},
								Errors: v.Errors,
							})
					}, time.Duration(p.appConfig.FuelGiveReceiptTimeout)*time.Second)
					if err != nil {
						logger.Error("ошибка запроса.", "transaction", "fuel_give", "handler", "FuelGiveReceipt", "tid", v.FuelGiveID, "err", err)
						p.CaptureError(fmt.Errorf("KazsOperator.FuelGiveReceipt error: %v", err), captureError)
						continue
					} else {
						status = true
						logger.Info("транзакция успешно отправлена на сервер.", "transaction", "fuel_give", "tid", v.FuelGiveID)
					}

					logger.Info("обновление транзакции в БД.", "transaction", "fuel_give", "tid", v.FuelGiveID)
					updateErr := p.repository.FuelGive.UpdateFuelGiveTransaction(models.UpdateFuelGiveTransactionDep{
						FuelGiveID: v.FuelGiveID,
						SendStatus: &status,
					})
					if updateErr != nil {
						logger.Error("не удалось обновить транзакцию в БД.", "transaction", "fuel_give", "tid", v.FuelGiveID)
						p.CaptureError(fmt.Errorf("FuelGive.UpdateFuelGiveTransaction error: %v", updateErr), captureError)
					}
				}
			} else {
				logger.Info("неотправленных транзакций не найдено.", "transaction", "fuel_give")
			}

			logger.Info("получение неотправленных транзакций из БД.", "transaction", "fuel_get")
			unsentFuelGetTransactions, err := p.repository.FuelGet.GetUnsentFuelGetTransactions()
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				logger.Error("не удалось получить неотправленные транзакции из БД.", "transaction", "fuel_get", "err", err)
				p.CaptureError(fmt.Errorf("Repo.GetUnsentFuelGiveTransactions error: %v", err), captureError)
			}

			if len(unsentFuelGetTransactions) != 0 {
				logger.Info("обнаружены неотправленные транзакции.", "transaction", "fuel_get", "count", len(unsentFuelGetTransactions))
				for _, v := range unsentFuelGetTransactions {
					var status = false
					logger.Info("POST FuelGetReceipt", "transaction", "fuel_get", "timeout_sec", v.FuelGetID)
					err = RunWithTimeout(func() error {
						return p.kazsOperator.FuelGetReceipt(
							&integration.FuelGetReceipt{
								FuelGetID: v.FuelGetID,
							},
							&integration.FuelGetReceiptRequest{
								KazsNumber: v.KazsNumber,
								JarId:      v.JarId,
								FuelType:   v.FuelType,
								StartTime:  v.StartTime,
								EndTime:    v.EndTime,
								DocNumber:  v.DocNumber,
								FuelLiter:  v.FuelLiter,
								SensorBeforeGet: integration.SensorInfo{
									H:  v.SensorBeforeGive.H,
									T:  v.SensorBeforeGive.T,
									Pr: v.SensorBeforeGive.Pr,
									U:  v.SensorBeforeGive.U,
									G:  v.SensorBeforeGive.G,
									R:  v.SensorBeforeGive.R,
									U1: v.SensorBeforeGive.U1,
									H2: v.SensorBeforeGive.H2,
									Ut: v.SensorBeforeGive.Ut,
									Rt: v.SensorBeforeGive.Rt,
									Ri: v.SensorBeforeGive.Ri,
									Tr: v.SensorBeforeGive.Tr,
									U2: v.SensorBeforeGive.U2,
									Nt: v.SensorBeforeGive.Nt,
									Dg: v.SensorBeforeGive.Dg,
									Ts: v.SensorBeforeGive.Ts,
								},
								SensorAfterGet: integration.SensorInfo{
									H:  v.SensorAfterGive.H,
									T:  v.SensorAfterGive.T,
									Pr: v.SensorAfterGive.Pr,
									U:  v.SensorAfterGive.U,
									G:  v.SensorAfterGive.G,
									R:  v.SensorAfterGive.R,
									U1: v.SensorAfterGive.U1,
									H2: v.SensorAfterGive.H2,
									Ut: v.SensorAfterGive.Ut,
									Rt: v.SensorAfterGive.Rt,
									Ri: v.SensorAfterGive.Ri,
									Tr: v.SensorAfterGive.Tr,
									U2: v.SensorAfterGive.U2,
									Nt: v.SensorAfterGive.Nt,
									Dg: v.SensorAfterGive.Dg,
									Ts: v.SensorAfterGive.Ts,
								},
								Errors: v.Errors,
							})
					}, time.Duration(p.appConfig.FuelGiveReceiptTimeout)*time.Second)
					if err != nil {
						logger.Error("ошибка запроса.", "transaction", "fuel_get", "handler", "FuelGetReceipt", "tid", v.FuelGetID, "err", err)
						p.CaptureError(fmt.Errorf("KazsOperator.FuelGetReceipt error: %v", err), captureError)
						continue
					} else {
						logger.Info("транзакция успешно отправлена на сервер.", "transaction", "fuel_get", "tid", v.FuelGetID)
						status = true
					}

					logger.Info("обновление транзакции в БД.", "transaction", "fuel_get", "tid", v.FuelGetID)
					updateErr := p.repository.FuelGet.UpdateFuelGetTransaction(models.UpdateFuelGetTransactionDep{
						FuelGetID:  v.FuelGetID,
						SendStatus: &status,
					})
					if updateErr != nil {
						logger.Error("не удалось обновить транзакцию в БД.", "transaction", "fuel_get", "tid", v.FuelGetID)
						p.CaptureError(fmt.Errorf("FuelGet.UpdateFuelGetTransaction error: %v", updateErr), captureError)
					}
				}
			} else {
				logger.Info("неотправленных транзакций не найдено", "transaction", "fuel_get")
			}
		}
	}
}

func (p *Processing) StartProgram(ntpServer string) error {
	p.appGui.CreateStartScreen("1")
	p.appGui.CreateStartScreen("2")

	logger := p.logger.BaseError("StartProgram")
	captureError := CaptureErrorDep{
		operation:  "StartProgram",
		kazsNumber: p.kazsOperator.KazsNumber,
	}
	p.CaptureError(fmt.Errorf("StartProgram"), captureError)
	logger.Info("Starting program...")

	// === Настройка NTP сервера ===
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p.SettingNTP(ctx, cancel, ntpServer, logger)

	// Пинг оборудования
	// Уровнемер
	logger.Info("проверка уровнемеров.")
	for i := range p.driver.SensDriver.Adapter["1"].LC {
		err := RunWithTimeout(func() error {
			return p.driver.SensDriver.LCDriver.Ping(i)
		}, time.Duration(p.driverConfig.GetMainStatusTimeout)*time.Second)
		if err != nil {
			logger.Error("уровнемер не отвечает.", "№", i, "err", err)
			p.CaptureError(fmt.Errorf("no response from the LC №%s. Error: %w", i, err), captureError)
		}
	}

	// БК
	logger.Info("проверка блоков коммутации.")
	for i := range p.driver.SensDriver.Adapter["1"].BK {
		err := RunWithTimeout(func() error {
			return p.driver.SensDriver.BKDriver.Ping(i)
		}, time.Duration(p.driverConfig.GetMainStatusTimeout)*time.Second)
		if err != nil {
			logger.Error("блок коммутации не отвечает.", "№", i, "err", err)
			p.CaptureError(fmt.Errorf("no response from the BK №%s. Error: %w", i, err), captureError)
		}
	}

	// ЛИН
	logger.Info("проверка лин. адаптера.")
	for i := range *p.driver.SensDriver.LinDriver.Adapter {
		err := RunWithTimeout(func() error {
			return p.driver.SensDriver.LinDriver.Ping(i)
		}, time.Duration(p.driverConfig.GetMainStatusTimeout)*time.Second)
		if err != nil {
			logger.Error("лин. адаптер не отвечает", "№", i, "err", err)
			p.CaptureError(fmt.Errorf("no response from the LIN №%s. Error: %w", i, err), captureError)
		}
	}

	// ТРК
	logger.Info("проверка ТРК.")
	for i := range p.driver.TopazDriver.Adapter {
		res := p.TRKRequest(GetTRKStatus, i, 0, logger)
		if res.Err != nil {
			logger.Error("ТРК не отвечает.", "№", i, "err", res.Err)
			p.CaptureError(fmt.Errorf("no response from the TRK №%s. Error: %w", i, res.Err), captureError)
		}
	}

	// Получаем настройки ТРК
	logger.Info("получение настроек ТРК.")
	TRKSettings := make(map[string]trk.TRKResponse)
	for i := range *p.driver.TopazDriver.TopazDriver.Adapter {
		var settings trk.TRKResponse
		res := RunWithTimeoutValue(func() ([]byte, error) {
			return p.driver.TopazDriver.TopazDriver.GetSettings(i)
		}, time.Duration(p.driverConfig.GetTRKStatusTimeout)*time.Second)
		if res.Err != nil {
			logger.Error("не удалось получить настройки ТРК", "№", i, "err", res.Err)
			p.CaptureError(fmt.Errorf("get TRK settings error:%w", res.Err), captureError)
			continue
		}
		err := json.Unmarshal(res.Value, &settings)
		if err != nil {
			logger.Error("не удалось получить настройки ТРК.", "№", i, "err", err)
			p.CaptureError(fmt.Errorf("get TRK settings error:%w", err), captureError)
			continue
		}

		TRKSettings[i] = settings

		if len(TRKSettings) == 0 {
			logger.Warn("настройки ТРК не найдены.")
		} else {
			logger.Info("настройки ТРК найдены.")
		}
	}

	p.trkSettings = TRKSettings

	// Контроллер
	logger.Info("проверка контроллера.")
	err := RunWithTimeout(func() error {
		return p.driver.ControllerDriver.Adapter.Verify()
	}, time.Duration(p.driverConfig.ControllerTimeout)*time.Second)
	if err != nil {
		logger.Error("контроллер не отвечает.", "err", err)
		p.CaptureError(fmt.Errorf("ControllerDriver.Ping error: %v", err), captureError)
	}
	time.Sleep(200 * time.Millisecond)

	logger.Info("получение статуса люков.")
	err = RunWithTimeout(func() error {
		return p.driver.ControllerDriver.Adapter.GetDin()
	}, time.Duration(p.driverConfig.ControllerTimeout)*time.Second)
	if err != nil {
		logger.Error("не удалось получить статус люков.", "err", err)
		p.CaptureError(fmt.Errorf("ControllerDriver.GetDin error: %v", err), captureError)
	}
	time.Sleep(200 * time.Millisecond)

	logger.Info("получение статусов замков.")
	err = RunWithTimeout(func() error {
		return p.driver.ControllerDriver.Adapter.GetDout()
	}, time.Duration(p.driverConfig.ControllerTimeout)*time.Second)
	if err != nil {
		logger.Error("не удалось получить статус замков", "err", err)
		p.CaptureError(fmt.Errorf("ControllerDriver.GetDout error: %v", err), captureError)
	}

	// TODO: Свет
	// TODO: Сигнал
	// Завершаем заправки
	p.startProgramm = true

	// === Получаем показания с уровнемера ===
	logger.Info("ожидание телеметрии.")
	for {
		if len(p.lastSENSTelemetry) == 0 {
			continue
		} else {
			logger.Info("телеметрия получена.")
			break
		}
	}

	logger.Info("проверка незавершенных транзакций.", "transaction", "fuel_give")
	unfinishedFuelGive := make([]models.LastFuelGiveReceipt, 0, len(*p.driver.TopazDriver.TopazDriver.Adapter))
	for i := range *p.driver.TopazDriver.TopazDriver.Adapter {
		logger.Info("получение незавершенных транзакций из БД.", "transaction", "fuel_give")
		unfinished, err := p.repository.FuelGive.GetLastFuelGiveTransaction(i)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				logger.Error("не удалось получить незавершенные транзакции из БД.", "transaction", "fuel_give", "jar_id", i, "err", err)
				p.CaptureError(fmt.Errorf("Repo.GetLastFuelGiveTransaction error: %v", err), CaptureErrorDep{
					operation:  "StartProgram",
					kazsNumber: p.kazsOperator.KazsNumber,
					jarNumber:  i,
				})

				_ = p.repository.ErrorLogs.InsertError(models.Errors{
					Time:    time.Now().Unix(),
					Handler: "StartProgram",
					Id:      "",
					Error:   fmt.Sprintf("Repo.GetLastFuelGiveTransaction error: %v", err),
				})
			}
			continue
		}
		unfinishedFuelGive = append(unfinishedFuelGive, unfinished)
	}

	if len(unfinishedFuelGive) != 0 {
		logger.Info("найдены незавершенные транзакции", "transaction", "fuel_give")
		for _, v := range unfinishedFuelGive {
			captureError = CaptureErrorDep{
				tid:        v.FuelGiveID,
				operation:  "StartProgram",
				kazsNumber: v.KazsNumber,
				jarNumber:  v.JarId,
			}

			trkStatus := p.TRKRequest(GetTRKStatus, v.JarId, 0, logger)
			if trkStatus.Err != nil {
				p.CaptureError(fmt.Errorf("TRKDriver.GetTRKStatus error: %v", trkStatus.Err), captureError)

				_ = p.repository.ErrorLogs.InsertError(models.Errors{
					Time:    time.Now().Unix(),
					Handler: "FuelGive",
					Id:      "",
					Error:   fmt.Sprintf("TRKDriver.GetTRKStatus error: %v", trkStatus.Err),
				})

				continue
			}
			if trkStatus.ValueStr == StatusIdle || trkStatus.ValueStr == StatusNozzleLifted {
				logger.Warn("после перезагрузки неожиданный статус ТРК.", "transaction", "fuel_give", "jar_id", v.JarId, "trk_status", trkStatus.ValueStr, "tid", v.FuelGiveID)
				p.CaptureError(fmt.Errorf("Информационная. После перезапуска программы при наличие незавершенной транзакции в БД - статус ТРК №%v: %v", v.JarId, trkStatus.ValueStr), captureError)

				_ = p.repository.ErrorLogs.InsertError(models.Errors{
					Time:    time.Now().Unix(),
					Handler: "FuelGive",
					Id:      "",
					Error:   fmt.Sprintf("Информационная. После перезапуска программы при наличие незавершенной транзакции в БД - статус ТРК №%v: %v", v.JarId, trkStatus.ValueStr),
				})

				var sensorAfterGive models.SensorInfoGive
				afterTelemetry := p.lastSENSTelemetry[v.JarId]
				sensorAfterGive = models.SensorInfoGive{
					T:  float64(afterTelemetry.T),
					U:  float64(afterTelemetry.U),
					R:  float64(afterTelemetry.R),
					U1: float64(afterTelemetry.U1),
					Ri: float64(afterTelemetry.Ri),
					Tr: float64(afterTelemetry.Tr),
					U2: float64(afterTelemetry.U2),
				}

				endTime := time.Now().Unix()

				logger.Info("обновление транзакции в БД.", "transaction", "fuel_give", "jar_id", v.JarId, "tid", v.FuelGiveID)
				err = p.repository.FuelGive.UpdateFuelGiveTransaction(models.UpdateFuelGiveTransactionDep{
					FuelGiveID:      v.FuelGiveID,
					EndTime:         &endTime,
					SensorAfterGive: &sensorAfterGive,
				})
				if err != nil {
					logger.Error("не удалось обновить транзакцию в БД.", "transaction", "fuel_give", "jar_id", v.JarId, "tid", v.FuelGiveID, "err", err)
					p.CaptureError(fmt.Errorf("Repo.UpdateFuelGiveTransaction error: %v", err), captureError)
				}

				continue
			}

			p.UpdateJarStatus(v.JarId, true)
			go p.startFuelGiveInProgress(startFuelGiveInProgressDep{
				fuelGiveID:       v.FuelGiveID,
				jarNumber:        v.JarId,
				fuelType:         v.FuelType,
				docNumber:        v.DocNumber,
				startTime:        v.StartTime,
				expectedLiter:    float32(v.FuelLitersPlan),
				liter:            float32(v.FuelLiter),
				captureError:     captureError,
				sensorBeforeInfo: v.SensorBeforeGive,
			}, logger)
		}
	} else {
		logger.Info("незавершенных транзакций не найдено.", "transaction", "fuel_give")
	}

	// Завершаем пополнения
	logger.Info("проверка незавершенных транзакций.", "transaction", "fuel_get")
	unfinishedFuelGet := make([]models.LastFuelGetReceipt, 0, len((*p.driver.SensDriver.LCDriver.Adapter)["1"].LC))
	for i := range (*p.driver.SensDriver.LCDriver.Adapter)["1"].LC {
		logger.Info("получение незавершенных транзакций из БД.", "transaction", "fuel_get")
		unfinished, err := p.repository.FuelGet.GetLastFuelGetTransaction(i)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				logger.Error("не удалось получить незавершенные транзакции из БД.", "transaction", "fuel_get", "jar_id", i, "err", err)
				p.CaptureError(fmt.Errorf("Repo.GetLastFuelGetTransaction error: %v", err), CaptureErrorDep{
					operation:  "StartProgram",
					kazsNumber: p.kazsOperator.KazsNumber,
					jarNumber:  i,
				})

				_ = p.repository.ErrorLogs.InsertError(models.Errors{
					Time:    time.Now().Unix(),
					Handler: "FuelGet",
					Id:      "",
					Error:   fmt.Sprintf("Repo.GetLastFuelGetTransaction error: %v", err),
				})
			}
			continue
		}
		unfinishedFuelGet = append(unfinishedFuelGet, unfinished)
	}

	if len(unfinishedFuelGet) != 0 {
		logger.Info("найдены незавершенные транзакции", "transaction", "fuel_get")
		for _, v := range unfinishedFuelGet {
			p.UpdateJarStatus(v.JarId, true)

			currentFuelVolume := p.lastSENSTelemetry[v.JarId]

			lastPumpStatus := p.lastControllerDoutTelemetry[p.driver.ControllerDriver.Adapter.Maping.Controller.Pump[v.JarId].Number].status

			if lastPumpStatus == p.driver.ControllerDriver.Adapter.Maping.Controller.Pump[v.JarId].Disable {
				logger.Warn("после перезагрузки насос заблокирован, при незавершенном пополнениее", "transaction", "fuel_get", "jar_id", v.JarId, "tid", v.FuelGetID)
				p.CaptureError(fmt.Errorf("После перезагрузки насос заблокирован при незавершенном пополнении"), captureError)
			}

			beforeDoorStatus := p.lastControllerDinTelemetry[p.driver.ControllerDriver.Adapter.Maping.Controller.Doors[v.JarId].Number].status

			drainedAmount := currentFuelVolume.U*float32(p.kazsConfig.FuelGetConfig.FuelGetUnits) - float32(v.SensorBeforeGive.U)*float32(p.kazsConfig.FuelGetConfig.FuelGetUnits)
			jarVolume := float32(v.SensorBeforeGive.U) * float32(p.kazsConfig.FuelGetConfig.FuelGetUnits) / (float32(v.SensorBeforeGive.Pr) / 100)

			if v.MonitoringFinishTime == 0 {
				go p.startFuelGetInProgress(startFuelGetInProgressDep{
					jarNumber:         v.JarId,
					fuelType:          v.FuelType,
					documentNumber:    v.DocNumber,
					expectedFuelLiter: float32(v.FuelLitersPlan),
					beforeFuelGet:     float32(v.SensorBeforeGive.U) * float32(p.kazsConfig.FuelGetConfig.FuelGetUnits),
					fuelVolume:        currentFuelVolume.U * float32(p.kazsConfig.FuelGetConfig.FuelGetUnits),
					drainedAmount:     drainedAmount,
					fuelGetID:         v.FuelGetID,
					jarVolume:         jarVolume,
					startTime:         v.StartTime,
					captureError:      captureError,
					beforeDoorStatus:  beforeDoorStatus,
				}, logger)
			} else if v.EndTime == 0 {
				completeCtx, _ := context.WithTimeout(context.Background(), time.Duration(p.kazsConfig.FuelGetConfig.FuelGetCompleteScreenTimeout)*time.Second)
				go p.startFuelGetComplete(completeCtx, startFuelGetCompleteDep{
					jarNumber:      v.JarId,
					fuelType:       v.FuelType,
					docNumber:      v.DocNumber,
					beforeFuelGet:  float32(v.SensorBeforeGive.U) * float32(p.kazsConfig.FuelGetConfig.FuelGetUnits),
					afterFuelGet:   currentFuelVolume.U * float32(p.kazsConfig.FuelGetConfig.FuelGetUnits),
					fuelGetPlan:    float32(v.FuelLitersPlan),
					startTime:      v.StartTime,
					fuelGetID:      v.FuelGetID,
					timer:          p.kazsConfig.FuelGetConfig.FuelGetCompleteScreenTimeout,
					startSpeedTime: 0,
					endSpeedTime:   0,
					captureError:   captureError,
				}, logger)
			}
		}
	} else {
		logger.Info("незавершенных транзакций не найдено.", "transaction", "fuel_get")
	}

	return nil
}

func (c *Processing) ReadPort(cancel context.CancelFunc) {
	logger := c.logger.BaseError("controller_read_port")

	// Используем порт из ControllerAdapter
	port := c.driver.ControllerDriver.Adapter.Port
	if port == nil {
		logger.Error("последовательный порт не инициализирован")
		return
	}

	scanner := bufio.NewScanner(port)
	for scanner.Scan() {
		message := bytes.TrimSpace(scanner.Bytes())
		if len(message) > 0 {
			c.saveMessage(message, logger)
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Error("ошибка чтения порта", "err", err)
		// Пытаемся переоткрыть порт при ошибке
		if err := c.driver.ControllerDriver.Adapter.Reopen(); err != nil {
			logger.Error("не удалось переоткрыть порт", "err", err)
		}
	}

	// Вызываем cancel, чтобы завершить контекст
	cancel()
}

func (c *Processing) saveMessage(message []byte, logger *slog.Logger) {
	var msg ComMessage
	//Исправлено для теста
	relayItems := []RelayItem{
		{Index: "1", State: "0"},
		{Index: "2", State: "1"},
	}

	wrappedMap := make(map[string][]RelayItem)
	wrappedMap["din"] = relayItems
	wrappedJSON, _ := json.Marshal(wrappedMap)

	err := json.Unmarshal([]byte(wrappedJSON), &msg)
	if err != nil {
		c.logger.Error("ошибка десериализации JSON", "err", err)
		return
	}
	//----Исправлено

	for key, rawValue := range msg {
		var relayItem []RelayItem
		err := json.Unmarshal(rawValue, &relayItem)
		if err != nil {
			c.CaptureError(fmt.Errorf("insert din error: %w", err), CaptureErrorDep{
				kazsNumber: c.kazsOperator.KazsNumber,
			})
			continue
		}

		switch key {
		case "din":
			var currentRelays [8]string
			for i := 0; i < 8; i++ {
				currentRelays[i] = "0"
			}
			for _, item := range relayItem {
				idx, _ := strconv.Atoi(item.Index)
				if idx >= 1 && idx <= 8 {
					currentRelays[idx-1] = item.State
					c.UpdateLastControllerDinTelemetry(idx, item.State)
				}
			}
		case "dout":
			var currentRelays [8]string
			for i := 0; i < 8; i++ {
				currentRelays[i] = "0"
			}
			for _, item := range relayItem {
				idx, _ := strconv.Atoi(item.Index)
				if idx >= 1 && idx <= 8 {
					currentRelays[idx-1] = item.State
					c.UpdateLastControllerDoutTelemetry(idx, item.State)
				}
			}
		case "din_set":
			for _, item := range relayItem {
				idx, _ := strconv.Atoi(item.Index)
				if idx >= 1 && idx <= 8 {
					c.UpdateLastControllerDinTelemetry(idx, item.State)
				}
			}
		case "dout_set":
			for _, item := range relayItem {
				idx, _ := strconv.Atoi(item.Index)
				if idx >= 1 && idx <= 8 {
					c.UpdateLastControllerDoutTelemetry(idx, item.State)
				}
			}
		default:
		}
	}
}

func (p *Processing) SettingNTP(ctx context.Context, cancel context.CancelFunc, ntpServer string, logger *slog.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if ntpServer == "" {
			logger.Warn("NTP server is empty")
			cancel()
			return
		}

		// === Получение текущего NTP сервера ===
		currentNTP, err := myntp.GetNTPServer()
		if err != nil {
			logger.Error("не удалось получить NTP сервер.", "err", err)
		}

		// === Установка нового NTP сервера ===
		if currentNTP != ntpServer {
			err = myntp.SetNTPServer(ntpServer)
			if err != nil {
				logger.Error("не удалось установить NTP сервер.", "err", err)
			} else {
				logger.Info("NTP сервер успешно установлен.", "ntp_server", ntpServer)
			}
		}

		// === Проверка расхождения времени ===
		localTime := time.Now().UTC()
		ntpTime, err := ntp.Time(ntpServer)
		if err != nil {
			logger.Error("не удалось получить время с NTP сервера", "err", err)
		} else {
			offset := localTime.Sub(ntpTime.UTC())
			if offset > p.appConfig.NTPOffset*time.Second {
				logger.Warn("критическое расхождение между NTP и local временем.", "local_time", localTime, "ntp_time", ntpTime, "offset", offset)
			} else {
				logger.Info("время синхронизировано.", "local_time", localTime, "ntp_time", ntpTime, "offset", offset)
				cancel()
				return
			}
		}
		time.Sleep(30 * time.Second)
	}
}
