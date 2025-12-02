package main

import (
	"context"
	"fuelazs/config"
	"fuelazs/internal/application"
	"fuelazs/internal/logger"
	"os"
	"os/signal"
	"syscall"
)

const Version = "v0.0.1"

// const DBPath = "fuelazs.db"
const DBPath = "host=localhost port=5464 user=postgres password=password dbname=fuelazs sslmode=disable"

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	_ = ctx

	// Инициализация логгера

	logger := logger.NewSlog()
	logger.Info("KAZS.App", "Version", Version)

	// Инициализация конфига
	conf, err := config.NewConfig()
	if err != nil {
		logger.Error("config init error", "err", err)
		return
	}

	// Инициализация кредов
	cred, err := config.NewCredentials()
	if err != nil {
		logger.Error("credentials init error", "err", err)
		return
	}

	application, err := application.NewApp(&application.AppDep{
		Version: Version,
		Logger:  logger,
		DbPath:  DBPath,
		Config:  conf,
		Cred:    cred,
	})
	if err != nil {
		logger.Error("application init error", "err", err)
		return
	}

	err = application.AppGui.Show()
	if err != nil {
		logger.Error("application show error", "err", err)
		return
	}

}
