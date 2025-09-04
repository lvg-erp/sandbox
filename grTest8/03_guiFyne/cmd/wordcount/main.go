package main

import (
	"context"
	"wordcount/internal/cache"
	"wordcount/internal/gui"

	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	if logger == nil {
		os.Exit(1)
	}

	// Создаём кэш с воркером
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wordCache := cache.NewWordCache(logger)
	go wordCache.Run(ctx)

	// Создаём GUI
	appGui, err := gui.NewAppGui(wordCache, logger)
	if err != nil {
		logger.Error("Failed to create GUI", "err", err)
		os.Exit(1)
	}

	// Запускаем приложение
	logger.Info("Starting WordCount application")
	go func() {
		<-time.After(1 * time.Second)
		logger.Info("Application running, enter a string in the GUI")
	}()
	appGui.FyneWindow.ShowAndRun()
}
