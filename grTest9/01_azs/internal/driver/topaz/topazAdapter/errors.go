package topazAdapter

import (
	"errors"
)

var (
	ErrOpenPort           = errors.New("COM port is not open")
	ErrWriteCommand       = errors.New("failed to write command to COM port")
	ErrReadCommand        = errors.New("failed to read response from COM port")
	ErrReadTimeout        = errors.New("read timeout from COM port")
	ErrInvalidSendCommand = errors.New("invalid command format for sending")
	ErrIncompleteResponse = errors.New("incomplete response received")
	ErrCrcMismatch        = errors.New("crc mismatch in response")   // Добавим для явного указания
	ErrUnexpectedResponse = errors.New("unexpected response format") // Для других ошибок формата
)
