package qr

import (
	"bufio"
	"encoding/json"
	"fmt"
	"fuelazs/internal/usecase/models"
	"go.bug.st/serial"
	"log/slog"
	"strings"
	"sync"
	"time"
)

type QRAdapter struct {
	Port              serial.Port
	QRAdapterSettings QRAdapterSettings
	mutex             sync.Mutex
	Maping            *QRMaping
}

type QRAdapterSettings struct {
	PortName    string
	BaudRate    int
	DataBits    int
	Parity      serial.Parity
	StopBits    serial.StopBits
	ReadTimeout int
	Token       string
}

func NewQRAdapter() (*QRAdapter, error) {
	QRInfo, err := getConnectionInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get controller connection info: %w", err)
	}

	comName := getOSPortName(QRInfo.QR.COMPort, QRInfo.QR.IsUSB, QRInfo.QR.IsACM)

	Port, err := serial.Open(comName, &serial.Mode{
		BaudRate: QRInfo.QR.BaudRate,
		DataBits: QRInfo.QR.DataBits,
		Parity:   parity(QRInfo.QR.Parity),
		StopBits: stopBits(QRInfo.QR.StopBits),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open serial Port: %w", err)
	}

	err = Port.SetReadTimeout(time.Duration(QRInfo.QR.ReadTimeout) * time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to set read timeout: %w", err)
	}

	return &QRAdapter{
		Port: Port,
		QRAdapterSettings: QRAdapterSettings{
			PortName:    comName,
			BaudRate:    QRInfo.QR.BaudRate,
			DataBits:    QRInfo.QR.DataBits,
			Parity:      parity(QRInfo.QR.Parity),
			StopBits:    stopBits(QRInfo.QR.StopBits),
			ReadTimeout: QRInfo.QR.ReadTimeout,
		},
		Maping: QRInfo,
		mutex:  sync.Mutex{},
	}, nil
}

func (r *QRAdapter) Read(dataChan chan<- models.ScannerResponse, stopChan <-chan struct{}, activeChan <-chan bool) {
	if r.Port == nil {
		fmt.Println("QRReader.Read: Последовательный порт не инициализирован.")
		return
	}

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	scanner := bufio.NewScanner(r.Port)
	isPaused := false
	for {
		if isPaused {
			select {
			case <-stopChan:
				fmt.Println("QRReader.Read: Получен сигнал остановки во время паузы.")
				return
			case newActiveState := <-activeChan:
				if newActiveState {
					isPaused = false // Снимаем с паузы
					fmt.Println("QRReader.Read: Работа возобновлена.")
				}
				continue // Возвращаемся в начало цикла for
			case <-ticker.C:
				_ = r.Reopen()
			}
		}

		// Основной select, который ждёт событий
		select {
		case <-stopChan:
			fmt.Println("QRReader.Read: Получен сигнал остановки.")
			return
		case newActiveState := <-activeChan:
			if !newActiveState {
				isPaused = true // Ставим на паузу
				fmt.Println("QRReader.Read: Работа приостановлена.")
			}
		case <-ticker.C:
			_ = r.Reopen()
			scanner = bufio.NewScanner(r.Port)
		default:
			scanSucces := scanner.Scan()

			if !scanSucces {
				err := scanner.Err()
				if strings.Contains(err.Error(), "multiple Read calls return no data or error") {
					continue
				}
				continue
			} else {
				line := scanner.Text()
				slog.Info("QR Read:", "qr_info", line)
				cleanedLine := strings.TrimSpace(line)
				if len(cleanedLine) == 0 {
					continue // Пропускаем пустые строки
				}

				var response models.ScannerResponse
				if err := json.Unmarshal([]byte(cleanedLine), &response); err != nil {
					fmt.Printf("QRReader.Read: Ошибка десериализации JSON '%s': %v\n", cleanedLine, err)
					continue
				}

				select {
				case dataChan <- response:
					// Успешно отправлено
				case <-stopChan:
					fmt.Println("QRReader.Read: Получен сигнал остановки при отправке данных.")
					return
				}
			}
		}
	}
}

func (c *QRAdapter) Reopen() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.IsOpen() {
		err := c.Port.Close()
		if err != nil {
			return fmt.Errorf("close port error: %s", err)
		}
		c.Port = nil
	}

	newPort, err := serial.Open(c.QRAdapterSettings.PortName, &serial.Mode{
		BaudRate: c.QRAdapterSettings.BaudRate,
		DataBits: c.QRAdapterSettings.DataBits,
		Parity:   c.QRAdapterSettings.Parity,
		StopBits: c.QRAdapterSettings.StopBits,
	})
	if err != nil {
		return fmt.Errorf("open port error: %s", err)
	}

	c.Port = newPort

	err = newPort.SetReadTimeout(time.Duration(c.QRAdapterSettings.ReadTimeout) * time.Second)
	if err != nil {
		newPort.Close()
		c.Port = nil
		return fmt.Errorf("set read timeout error: %s", err)
	}

	return nil
}

func (c *QRAdapter) IsOpen() bool {
	return c.Port != nil
}
