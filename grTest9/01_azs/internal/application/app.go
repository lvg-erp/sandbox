package application

import (
	"context"
	"fmt"
	"fuelazs/internal/repository/postgres"

	//"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
	//"gl.iteco.com/technology/go_general/errproc"
	"fuelazs/config"
	"fuelazs/internal/logger"
)

type App struct {
	Repo   *postgres.Registry
	AppGui *AppGui
	//Cron   *Cron
}

type AppDep struct {
	Version string
	DbPath  string
	Config  *config.Config
	Cred    *config.Credentials
	Logger  *logger.Logger
}

func NewApp(appDep *AppDep) (*App, error) {

	sentryLogger := logrus.New()
	if sentryLogger == nil {
		return nil, fmt.Errorf("init logrus error")
	}

	// Инициализация обработчика ошибок
	//errorProc, err := errproc.NewErrProc(sentryLogger)
	//if err != nil {
	//	return nil, err
	//}

	// Инициализация Sentry
	//err = sentry.Init(sentry.ClientOptions{
	//	Dsn:              appDep.Cred.Sentry.Dsn,
	//	TracesSampleRate: appDep.Cred.Sentry.TracesSampleRate, // Настройка выборки трассировок
	//	Release:          appDep.Version,                      // Версия приложения
	//	Debug:            false,                               // Отключение режима отладки
	//	AttachStacktrace: true,                                // Включение прикрепления трассировок стека
	//	EnableTracing:    true,                                // Включение трассировки производительности
	//	//SampleRate:       1.0,                         // Частота отправки событий
	//	BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	//		if hint.Request != nil {
	//			body, err := io.ReadAll(hint.Request.Body)
	//			if err == nil {
	//				hint.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	//				event.Request.Data = string(body)
	//			}
	//		}
	//		return event
	//	},
	//})
	//if err != nil {
	//	appDep.Logger.Error("init sentry error", "err", err)
	//}
	//defer sentry.Flush(2 * time.Second)

	// Инициализация БД
	repoRegistry, err := NewRegistry(appDep.DbPath, appDep.Logger)
	if err != nil {
		return nil, err
	}

	// Инициализация драйверов
	drivers, err := NewDrivers()
	if err != nil {
		return nil, err
	}

	// Инициализация интеграций
	integration, err := NewIntegration()
	if err != nil {
		return nil, err
	}

	// Инициализация графического интерфейса
	appGui := NewAppGui(repoRegistry.Activation)

	// Инициализация usecase
	useCases, err := NewUseCases(UseCasesDep{
		Config:       appDep.Config,
		Logger:       appDep.Logger,
		KazsOperator: integration.KazsOperator,
		AppGui:       appGui.MainContent,
		Driver:       drivers,
		Repository:   repoRegistry,
		//ErrProc:      errorProc,
	})
	if err != nil {
		return nil, err
	}

	// Инициализация Cron
	//appCron, err := NewCron(sentryLogger, errorProc)
	//if err != nil {
	//	return nil, err
	//}

	readCtx, readCancel := context.WithCancel(context.Background())

	// Чтение порта контроллера
	//TODO: Эмуляция!!!!!!!!
	go useCases.Processing.ReadPort(readCancel)

	<-readCtx.Done()

	// Запускаем основной процесс программы
	go useCases.Processing.MainProcess()

	// Запускаем сбор телеметрии
	go useCases.Processing.GetTelemetry()

	// Запускаем чтение телеметрии
	go useCases.Processing.ReadTelemetry()

	// Запускаем сбор телеметрии ТРК
	go useCases.Processing.GetTRKTelemetry()

	// Запускаем отправку телеметрии
	go useCases.Processing.Telemetry()

	// Запускаем отправку телеметрии
	go useCases.Processing.TelemetryHistory()

	// Запускаем проверку незавершенных заправок
	go useCases.Processing.CheckUnsentFuelGiveTransactions()

	// Запускаем проверку незавершенны пополнений
	go useCases.Processing.CheckUnsentFuelGetTransactions()

	// Запускаем проверку неотправленных транзакций
	go useCases.Processing.CheckUnsentTransactions()

	// Обновление информации о КАЗС
	go useCases.Processing.GetConfig()

	return &App{
		Repo:   repoRegistry,
		AppGui: appGui,
		//Cron:   appCron,
	}, nil
}
