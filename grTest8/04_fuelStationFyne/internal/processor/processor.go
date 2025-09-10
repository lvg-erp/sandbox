package processor

import (
	_ "context"
	"database/sql"
	"encoding/json"
	"errors"
	dbpkg "fuelstation/internal/db"
	"fuelstation/internal/gui"
	"fyne.io/fyne/v2/dialog"
	"io/ioutil"
	"log"
)

func ProcessJSONFile(db *sql.DB, gui *gui.Gui, filename string, ready chan<- struct{}) {
	log.Println("ProcessJSONFile: Начало обработки JSON файла")
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		gui.ShowError(err)
		close(ready)
		return
	}

	var operations []struct {
		Type      string `json:"type"`
		ColumnID  int    `json:"column_id"`
		FuelType  string `json:"fuel_type"`
		Liters    int    `json:"liters"`
		Action    string `json:"action"`
		Timestamp int64  `json:"timestamp"`
	}
	if err := json.Unmarshal(data, &operations); err != nil {
		gui.ShowError(err)
		close(ready)
		return
	}

	for _, op := range operations {
		if op.Type != "operation" {
			gui.ShowError(errors.New("неверный тип операции: " + op.Type))
			continue
		}

		if op.Action == "fill" {
			log.Println("ProcessJSONFile: Начало операции заправки")
			gui.FuelGiveScreen() // Отображаем диалог с прогресс-баром
			if err := dbpkg.SaveFuelOperation(db, op.ColumnID, op.FuelType, op.Liters, op.Action, op.Timestamp); err != nil {
				gui.ShowError(err)
				continue
			}
			// Прогресс-бар и отмена обрабатываются в FuelGiveScreen
			log.Println("ProcessJSONFile: Операция заправки завершена")
			gui.updateChan <- func() {
				dialog.ShowInformation("Успех", "Заправка завершена", gui.Window)
			}
		} else if op.Action == "drain" {
			log.Println("ProcessJSONFile: Начало операции слива")
			gui.FuelGetScreen() // Отображаем диалог с прогресс-баром
			if err := dbpkg.SaveFuelOperation(db, op.ColumnID, op.FuelType, op.Liters, op.Action, op.Timestamp); err != nil {
				gui.ShowError(err)
				continue
			}
			// Прогресс-бар и отмена обрабатываются в FuelGetScreen
			log.Println("ProcessJSONFile: Операция слива завершена")
			gui.updateChan <- func() {
				dialog.ShowInformation("Успех", "Слив завершён", gui.Window)
			}
		} else {
			gui.ShowError(errors.New("неизвестное действие: " + op.Action))
		}
	}

	log.Println("ProcessJSONFile: Обработка JSON файла завершена")
	close(ready)
}
