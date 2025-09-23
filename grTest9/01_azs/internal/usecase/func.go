package usecase

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"log/slog"
	"time"
)

const (
	GetTRKStatus          = "GetTRKStatus"
	ApprovalTRK           = "ApprovalTRK"
	GetFuelGiveStatus     = "GetFuelGiveStatus"
	GetFullFuelGiveStatus = "GetFullFuelGiveStatus"
	FuelGiveSuccess       = "FuelGiveSuccess"
	SetFuelGive           = "SetFuelGive"
)

type Result[T any] struct {
	Value T
	Err   error
}

func RunWithTimeoutValue[T any](fn func() (T, error), timeout time.Duration) Result[T] {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resultChan := make(chan Result[T], 1)

	go func() {
		val, err := fn()
		resultChan <- Result[T]{Value: val, Err: err}
	}()

	select {
	case res := <-resultChan:
		return res
	case <-ctx.Done():
		var zeroValue T
		return Result[T]{Value: zeroValue, Err: ctx.Err()}
	}
}

func RetryWithTimeoutValue[T any](operation func() Result[T], attempts int, timeout int) Result[T] {
	var lastResult Result[T]
	if attempts <= 0 {
		attempts = 1
	}

	for i := 0; i < attempts; i++ {
		result := operation()
		if result.Err == nil {
			return result // Успешное выполнение
		}
		lastResult = result
		if i < attempts-1 {
			time.Sleep(time.Duration(timeout) * time.Second)
		}
	}

	return lastResult
}

func RunWithTimeout(fn func() error, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // Важно вызвать cancel для освобождения ресурсов

	resultChan := make(chan error, 1)

	go func() {
		resultChan <- fn()
	}()

	select {
	case err := <-resultChan:
		return err // Функция завершилась
	case <-ctx.Done():
		return ctx.Err() // Произошел таймаут
	}
}

func RetryWithTimeout(operation func() error, attempts int, timeout int) error {
	var lastErr error
	if attempts <= 0 {
		attempts = 1
	}

	for i := 0; i < attempts; i++ {
		err := operation()
		if err == nil {
			return nil
		}
		lastErr = err

		if i < attempts-1 {
			time.Sleep(time.Duration(timeout) * time.Second)
		}
	}

	return lastErr
}

type TRKResponse struct {
	ValueStr   string
	ValueFloat float32
	Err        error
}

func (p *Processing) TRKRequest(operation string, jarNumber string, liter float32, logger *slog.Logger) TRKResponse {
	var trkResponse TRKResponse
	var bigNum uint64
	err := binary.Read(rand.Reader, binary.BigEndian, &bigNum)
	if err != nil {
		bigNum = 0
	}
	logger.Info("получение команды.", "driver", "trk", "command", operation, "trace_id", bigNum)
	defer logger.Info("выполнение команды завершено.", "driver", "trk", "command", operation, "trace_id", bigNum)

	p.driver.MuTRK.Lock()
	logger.Info("началось выполнение команды.", "driver", "trk", "command", operation, "trace_id", bigNum)
	defer p.driver.MuTRK.Unlock()

	switch operation {
	case GetTRKStatus:
		res := RetryWithTimeoutValue(func() Result[string] {
			return RunWithTimeoutValue(func() (string, error) {
				return p.driver.TopazDriver.TopazDriver.GetTRKStatus(jarNumber)
			}, time.Duration(p.driverConfig.GetTRKStatusTimeout)*time.Second)
		}, 1, 0)

		if res.Err == nil {
			logger.Info("статус ТРК.", "driver", "trk", "command", operation, "trk_status", res.Value, "trace_id", bigNum)
		}

		trkResponse = TRKResponse{
			ValueStr: res.Value,
			Err:      res.Err,
		}

	case ApprovalTRK:
		res := RetryWithTimeout(func() error {
			return RunWithTimeout(func() error {
				return p.driver.TopazDriver.TopazDriver.ApprovalTRK(jarNumber)
			}, time.Duration(p.driverConfig.ApprovalTRKTimeout)*time.Second)
		}, 1, 0)

		if res == nil {
			logger.Info("успешное санкционирование.", "driver", "trk", "command", operation, "trace_id", bigNum)
		}

		trkResponse = TRKResponse{
			Err: res,
		}

	case GetFuelGiveStatus:
		res := RetryWithTimeoutValue(func() Result[float32] {
			return RunWithTimeoutValue(func() (float32, error) {
				return p.driver.TopazDriver.TopazDriver.GetFuelGiveStatus(jarNumber)
			}, time.Duration(p.driverConfig.GetFuelGiveStatusTimeout)*time.Second)
		}, 1, 0)

		if res.Err == nil {
			logger.Info("заправлено литров.", "driver", "trk", "command", operation, "fueling_liters", res.Value, "trace_id", bigNum)
		}

		trkResponse = TRKResponse{
			ValueFloat: res.Value,
			Err:        res.Err,
		}

	case GetFullFuelGiveStatus:
		res := RetryWithTimeoutValue(func() Result[float32] {
			return RunWithTimeoutValue(func() (float32, error) {
				return p.driver.TopazDriver.TopazDriver.GetFullFuelGiveStatus(jarNumber)
			}, time.Duration(p.driverConfig.GetFullFuelGiveStatusTimeout)*time.Second)
		}, 1, 0)

		if res.Err == nil {
			logger.Info("заправлено литров.", "driver", "trk", "command", operation, "fueling_liters", res.Value, "trace_id", bigNum)
		}

		trkResponse = TRKResponse{
			ValueFloat: res.Value,
			Err:        res.Err,
		}

	case FuelGiveSuccess:
		res := RetryWithTimeout(func() error {
			return RunWithTimeout(func() error {
				return p.driver.TopazDriver.TopazDriver.FuelGiveSuccess(jarNumber)
			}, time.Duration(p.driverConfig.FuelGiveSuccessTimeout)*time.Second)
		}, 1, 0)

		if res == nil {
			logger.Info("успешное завершение заправки.", "driver", "trk", "command", operation, "trace_id", bigNum)
		}

		trkResponse = TRKResponse{
			Err: res,
		}

	case SetFuelGive:
		res := RetryWithTimeout(func() error {
			return RunWithTimeout(func() error {
				return p.driver.TopazDriver.TopazDriver.SetFuelGive(jarNumber, liter)
			}, time.Duration(p.driverConfig.SetFuelGiveTimeout)*time.Second)
		}, 1, 0)

		if res == nil {
			logger.Info("успешное задание количества топлива для выдачи.", "driver", "trk", "command", operation, "expected_liter", liter, "trace_id", bigNum)
		}

		trkResponse = TRKResponse{
			Err: res,
		}

	default:
		logger.Info("операция не поддерживается.", "driver", "trk", "command", operation, "trace_id", bigNum)
		return TRKResponse{
			Err: fmt.Errorf("command '%s' is not supported", operation),
		}
	}

	if trkResponse.Err != nil {
		logger.Error("ошибка выполнения операции.", "driver", "trk", "command", operation, "err", trkResponse.Err, "trace_id", bigNum)
		reopenErr := p.ReopenTopaz(jarNumber)
		if reopenErr != nil {
			log.Println("не удалось переоткрыть порт.", "driver", "trk", "command", operation, "err", reopenErr, "trace_id", bigNum)
		}
	}

	time.Sleep(500 * time.Millisecond)
	return trkResponse
}

type CaptureErrorDep struct {
	tid        string
	operation  string
	kazsNumber string
	jarNumber  string
}

type TelemetryErrorDep struct {
	handler    string
	operation  string
	kazsNumber string
	jarNumber  string
}

func (p *Processing) CaptureError(err error, dep CaptureErrorDep) {
	//exception := p.errproc.NewException(err)
	//exception.SetTag("TID", dep.tid)
	//exception.SetTag("Operation", dep.operation)
	//exception.SetTag("KazsNumber", dep.kazsNumber)
	//exception.SetTag("JarNumber", dep.jarNumber)
	//exception.Capture()

}

func (p *Processing) CaptureTelemetryError(err error, dep TelemetryErrorDep) {
	//exception := p.errproc.NewException(err)
	//exception.SetTag("Handler", dep.handler)
	//exception.SetTag("Operation", dep.operation)
	//exception.SetTag("KazsNumber", dep.kazsNumber)
	//exception.SetTag("JarNumber", dep.jarNumber)
	//exception.Capture()
}

func (p *Processing) ReopenSens() error {
	p.driver.MuSENS.Lock()
	err := p.driver.SensDriver.Adapter["1"].Reopen()
	p.driver.MuSENS.Unlock()
	if err != nil {
		return err
	}
	return nil
}

func (p *Processing) ReopenTopaz(jarNumber string) error {
	err := p.driver.TopazDriver.Adapter[jarNumber].Reopen()
	if err != nil {
		return err
	}
	return nil
}
