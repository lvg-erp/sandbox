package processor

//
//import (
//	"database/sql"
//	"encoding/json"
//	"fuelstation/internal/db"
//	"fuelstation/internal/gui"
//	"fuelstation/internal/model"
//	"io/ioutil"
//	"log"
//	"time"
//
//	"fyne.io/fyne/v2"
//)
//
//func ProcessJSONFile(dbConn *sql.DB, gui *gui.Gui, filePath string) {
//	ticker := time.NewTicker(5 * time.Second)
//	defer ticker.Stop()
//
//	for range ticker.C {
//		data, err := ioutil.ReadFile(filePath)
//		if err != nil {
//			log.Printf("Error reading JSON file: %v", err)
//			continue
//		}
//
//		var operations []model.FuelOperation
//		if err := json.Unmarshal(data, &operations); err != nil {
//			log.Printf("Error unmarshaling JSON: %v", err)
//			continue
//		}
//
//		for _, op := range operations {
//			op.ColumnID = "1" // А-92
//			if op.FuelType == "diesel" {
//				op.ColumnID = "2" // Дизель
//			}
//
//			if op.Action == "drain" {
//				op.DrainTimestamp = sql.NullInt64{Int64: op.Timestamp, Valid: true}
//			} else if op.Action == "fill" {
//				op.FillTimestamp = sql.NullInt64{Int64: op.Timestamp, Valid: true}
//			}
//
//			if err := db.InsertOperation(dbConn, op); err != nil {
//				continue
//			}
//
//			fyne.Do(func() {
//				if op.Action == "fill" {
//					// Показываем начальный экран налива
//					gui.CreateFuelGiveStartScreen(op.ColumnID, float32(op.Liters), op.FuelType, 30)
//					// Имитация процесса через 5 секунд
//					go func() {
//						time.Sleep(5 * time.Second)
//						fyne.Do(func() {
//							gui.CreateFuelGiveInProgressScreen(op.ColumnID, op.FuelType, float32(op.Liters)/2, float32(op.Liters))
//						})
//						// Имитация завершения через 10 секунд
//						time.Sleep(5 * time.Second)
//						fyne.Do(func() {
//							gui.CreateFuelGiveCompleteScreen(op.ColumnID, op.FuelType, "DOC123", float32(op.Liters), float32(op.Liters), op.Timestamp, op.Timestamp+300)
//						})
//					}()
//				} else if op.Action == "drain" {
//					// Показываем начальный экран слива
//					gui.CreateFuelGetStartScreen(op.ColumnID, op.FuelType, 0, 1000, float32(op.Liters), 30)
//					// Имитация процесса через 5 секунд
//					go func() {
//						time.Sleep(5 * time.Second)
//						fyne.Do(func() {
//							gui.CreateFuelGetInProgressScreen(op.ColumnID, float32(op.Liters), float32(op.Liters)/2, 0, 1000, 5)
//						})
//						// Имитация завершения через 10 секунд
//						time.Sleep(5 * time.Second)
//						fyne.Do(func() {
//							gui.CreateFuelGetCompleteScreen(op.ColumnID, op.FuelType, "DOC124", 0, float32(op.Liters), float32(op.Liters), float32(op.Liters), op.Timestamp, op.Timestamp+300, 30)
//						})
//					}()
//				}
//			})
//		}
//	}
//}
