package usecase

import (
	"fmt"
	"fuelazs/internal/integration"
	"fuelazs/internal/usecase/convert"
	"fuelazs/internal/usecase/models"
	"time"
)

func (p *Processing) Activation(qrInfo models.ScannerResponse) error {
	tempLogger := p.logger.Transaction("activation", qrInfo.TID)
	captureError := CaptureErrorDep{
		tid:       qrInfo.TID,
		operation: "Activation",
	}

	tempLogger.Info("начало процесса активации.")
	tempLogger.Info("GET Activation")
	// === Отправляем запрос на сервер ===
	response, err := p.kazsOperator.Activation(qrInfo.TID, qrInfo.ADR)
	if err != nil || response.Error {
		errMsg := err.Error()
		if response != nil && response.Error {
			errMsg = response.Message
		}
		tempLogger.Error("ошибка запроса.", "handler", "Activation", "err", errMsg)
		p.CaptureError(fmt.Errorf("kazsOperator.Activation error: %v", errMsg), captureError)
		return fmt.Errorf("kazsOperator.Activation error: %v", errMsg)
	}

	// Записываем в БД
	repo := convert.ConvertActivation(response)
	err = p.repository.Activation.InsertActivation(repo)
	if err != nil {
		tempLogger.Error("не удалось записать в БД.", "err", err)
		p.CaptureError(fmt.Errorf("Repo.InsertActivation error: %v", err), captureError)
		return fmt.Errorf("Repo.InsertActivation error: %v", err)
	}

	p.appGui.SetUI(response.Result.Logo, response.Result.SupportNumber, response.Result.KazsNumber, response.Result.KazsTimezone)
	p.appGui.CreateDefaultScreen("1")
	p.appGui.CreateDefaultScreen("2")
	p.appGui.CreateHeader()
	err = p.kazsOperator.SetConfig(response.Result.URL, response.Result.KazsApiKey, response.Result.KazsID, response.Result.KazsNumber, response.Result.ConfigHash)
	if err != nil {
		tempLogger.Error("kazsOperator.SetConfig", "err", err)
		p.CaptureError(fmt.Errorf("error setting kazs config: %s", err.Error()), CaptureErrorDep{})
		return fmt.Errorf("setConfig error: %v", err)
	}

	p.kazsActivation = true
	tempLogger.Info("казс успешно активирована.")
	return nil
}

func (p *Processing) GetConfig() {
	logger := p.logger.BaseError("GetConfig")
	captureError := CaptureErrorDep{
		operation:  "GetConfig",
		kazsNumber: p.kazsOperator.KazsNumber,
	}

	ticker := time.NewTicker(120 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !p.kazsActivation || !p.startProgramm {
			continue
		}
		logger.Info("GET GetConfig")
		res := RunWithTimeoutValue(func() (*integration.KazsGetConfigResponse, error) {
			return p.kazsOperator.GetConfig(&integration.GetConfigRequest{
				p.kazsOperator.ConfigHash,
			})
		}, time.Duration(p.appConfig.GetConfigTimeout)*time.Second)

		if res.Err != nil || res.Value.Error {
			var errMsg string
			if res.Err != nil {
				errMsg = res.Err.Error()
			} else if res.Value.Error {
				errMsg = res.Value.Message
			}
			p.CaptureError(fmt.Errorf("kazsOperator.GetConfig error: %v", errMsg), captureError)
			logger.Error("ошибка запроса.", "handler", "GetConfig", "err", errMsg)
			continue
		}

		err := p.repository.Activation.UpdateActivation(models.UpdateActivationData{
			KazsID:        p.kazsOperator.KazsID,
			ConfigHash:    &res.Value.Result.ConfigHash,
			KazsNumber:    &res.Value.Result.KazsNumber,
			KazsTimezone:  &res.Value.Result.KazsTimezone,
			NtpServer:     &res.Value.Result.NtpServer,
			SupportNumber: &res.Value.Result.SupportNumber,
			Logo:          &res.Value.Result.Logo,
		})

		if err != nil {
			p.CaptureError(fmt.Errorf("kazsOperator.UpdateActivation error: %v", err), captureError)
			logger.Error("не удалось обновить БД.", "err", err)
			continue
		}

		p.appGui.SetUI(res.Value.Result.Logo, res.Value.Result.SupportNumber, res.Value.Result.KazsNumber, res.Value.Result.KazsTimezone)
	}
}
