package pdf

//import (
//	"bytes"
//	"fmt"
//
//	"github.com/johnfercher/maroto/v2"
//	"github.com/johnfercher/maroto/v2/pkg/consts/align"
//	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
//	"github.com/johnfercher/maroto/v2/pkg/core"
//	"github.com/johnfercher/maroto/v2/pkg/props"
//	"github.com/skip2/go-qrcode"
//)
//
//type TicketData struct {
//	BookingID  int
//	FilmTitle  string
//	HallNumber int
//	Row        int
//	Seat       int
//	Date       string
//	Time       string
//	QRContent  string
//}
//
//func GenerateTicket(data TicketData) ([]byte, error) {
//	m := maroto.New()
//	m.AddPages(
//		maroto.NewPage(
//			core.WithMargins(20, 15, 20, 15),
//			core.WithSize(core.A4),
//		),
//	)
//
//	// Заголовок
//	m.AddRow(20,
//		maroto.NewCol(12, maroto.NewText("ЭЛЕКТРОННЫЙ БИЛЕТ",
//			props.Text{
//				Size:  24,
//				Style: fontstyle.Bold,
//				Align: align.Center,
//				Color: props.Blue,
//			},
//		)),
//	)
//
//	m.AddLine()
//
//	// Информация
//	m.AddRow(12,
//		maroto.NewCol(12, maroto.NewText(fmt.Sprintf("Фильм: %s", data.FilmTitle),
//			props.Text{Size: 16, Style: fontstyle.Bold},
//		)),
//	)
//
//	m.AddRow(10,
//		maroto.NewCol(6, maroto.NewText(fmt.Sprintf("Зал: %d", data.HallNumber))),
//		maroto.NewCol(6, maroto.NewText(fmt.Sprintf("Ряд: %d | Место: %d", data.Row, data.Seat))),
//	)
//
//	m.AddRow(10,
//		maroto.NewCol(12, maroto.NewText(fmt.Sprintf("Дата: %s | Время: %s", data.Date, data.Time))),
//	)
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
//	m.AddRow(80,
//		maroto.NewCol(12, maroto.NewImageFromBytes(qrPNG, props.Rect{
//			Center:  true,
//			Percent: 60,
//		})),
//	)
//
//	m.AddRow(12,
//		maroto.NewCol(12, maroto.NewText("Покажите QR-код на входе",
//			props.Text{
//				Align: align.Center,
//				Size:  12,
//				Style: fontstyle.Italic,
//			},
//		)),
//	)
//
//	// Генерация PDF
//	var buf bytes.Buffer
//	err = m.Generate(&buf)
//	if err != nil {
//		return nil, err
//	}
//
//	return buf.Bytes(), nil
//}
