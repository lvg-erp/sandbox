package logger

import (
	"log/slog"
	"os"
)

type Logger struct {
	base *slog.Logger
}

func NewSlog() *Logger {
	var handler slog.Handler
	handler = slog.NewTextHandler(os.Stdout, nil)

	return &Logger{
		base: slog.New(handler),
	}
}

func (l *Logger) WithKazsNumber(kazsNumber string) {
	l.base = l.base.With("kazsNumber", kazsNumber)
}

func (l *Logger) Info(msg string, attrs ...any) {
	l.base.Info(msg, attrs...)
}

func (l *Logger) Error(msg string, attrs ...any) {
	l.base.Error(msg, attrs...)
}

func (l *Logger) Warn(msg string, attrs ...any) {
	l.base.Warn(msg, attrs...)
}

func (l *Logger) BaseError(operation string) *slog.Logger {
	return l.base.With(
		slog.String("operation", operation),
	)
}

func (l *Logger) Transaction(transaction, transactionID string) *slog.Logger {
	return l.base.With(
		slog.String("transaction", transaction),
		slog.String("tid", transactionID),
	)
}
