package main

import (
	"fmt"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"html/template"
	"image/png"
	"log"
	"net/http"
)

type Page struct {
	Title string
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{Title: "QR Code Generator"}
	t, _ := template.ParseFiles("generator.html")
	err := t.Execute(w, p)
	if err != nil {
		return
	}
}

func viewCodeHandler(w http.ResponseWriter, r *http.Request) {
	dataString := r.FormValue("dataString")
	qrCode, _ := qr.Encode(dataString, qr.L, qr.Auto)
	qrCode, _ = barcode.Scale(qrCode, 128, 128)
	err := png.Encode(w, qrCode)
	if err != nil {
		return
	}
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/generator/", viewCodeHandler)
	fmt.Printf("starting server at port 8000\n")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
