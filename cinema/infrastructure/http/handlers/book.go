package handlers

//import (
//	"bytes"
//	"fmt"
//	"github.com/johnfercher/maroto/pkg/color"
//	"github.com/johnfercher/maroto/pkg/consts"
//	"github.com/johnfercher/maroto/pkg/pdf"
//	"github.com/johnfercher/maroto/pkg/props"
//	"github.com/skip2/go-qrcode"
//	_ "image/png"
//)
//
//type TicketData struct {
//	BookingID  int
//	FilmTitle  string
//	HallNumber int
//	Row        int
//	Seat       int
//	Date       string // "2025-11-15"
//	Time       string // "19:30"
//	QRContent  string // например: "TICKET:42"
//}
//
//func GenerateTicket(data TicketData) ([]byte, error) {
//	m := pdf.NewMaroto(consts.Portrait, consts.A4)
//	m.SetPageMargins(20, 15, 20)
//
//	// Заголовок
//	m.Row(20, func() {
//		m.Col(12, func() {
//			m.Text("ЭЛЕКТРОННЫЙ БИЛЕТ", props.Text{
//				Size:  24,
//				Style: consts.Bold,
//				Align: consts.Center,
//				Top:   10,
//				Color: color.NewBlue(),
//			})
//		})
//	})
//
//	m.Line(2)
//
//	// Данные
//	m.Row(10, func() {
//		m.Col(12, func() {
//			m.Text(fmt.Sprintf("Фильм: %s", data.FilmTitle), props.Text{Size: 16, Style: consts.Bold})
//		})
//	})
//
//	m.Row(8, func() {
//		m.Col(6, func() {
//			m.Text(fmt.Sprintf("Зал: %d", data.HallNumber))
//		})
//		m.Col(6, func() {
//			m.Text(fmt.Sprintf("Ряд: %d | Место: %d", data.Row, data.Seat))
//		})
//	})
//
//	m.Row(8, func() {
//		m.Col(12, func() {
//			m.Text(fmt.Sprintf("Дата: %s | Время: %s", data.Date, data.Time))
//		})
//	})
//
//	// QR-код
//	qr, err := qrcode.New(data.QRContent, qrcode.Medium)
//	if err != nil {
//		return nil, err
//	}
//	qr.DisableBorder = true
//	qrPNG, err := qr.PNG(300)
//	if err != nil {
//		return nil, err
//	}
//
//	m.Row(80, func() {
//		m.Col(12, func() {
//			m.ImageFromBytes(qrPNG, props.Rect{
//				Center:  true,
//				Percent: 60,
//			})
//		})
//	})
//
//	m.Row(10, func() {
//		m.Col(12, func() {
//			m.Text("Покажите QR-код на входе", props.Text{
//				Align: consts.Center,
//				Size:  12,
//				Style: consts.Italic,
//			})
//		})
//	})
//
//	var buf bytes.Buffer
//	err = m.Output(&buf)
//	return buf.Bytes(), err
//}
