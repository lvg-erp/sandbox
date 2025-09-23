package sensAdapter

import (
	"errors"
)

var (
	ErrOpenPort           = errors.New("COM port is not open")
	ErrWriteCommand       = errors.New("failed to write command to COM port")
	ErrReadCommand        = errors.New("failed to read response from COM port")
	ErrReadTimeout        = errors.New("read timeout from COM port")
	ErrReadCRC            = errors.New("received response CRC mismatch")
	ErrInvalidResponse    = errors.New("received invalid response format")
	ErrReadAttemptsLimit  = errors.New("read attempts limit reached with no data")
	ErrInvalidSendCommand = errors.New("invalid command format for sending")
)
