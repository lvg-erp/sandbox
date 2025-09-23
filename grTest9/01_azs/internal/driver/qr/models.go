package qr

type QRMaping struct {
	QR Settings `json:"QR"`
}

type Settings struct {
	IsUSB       bool    `json:"IsUSB"`
	IsACM       bool    `json:"IsACM"`
	COMPort     int     `json:"COMPort"`
	BaudRate    int     `json:"BaudRate"`
	StopBits    float32 `json:"StopBits"`
	DataBits    int     `json:"DataBits"`
	Parity      string  `json:"Parity"`
	ReadTimeout int     `json:"ReadTimeout"`
}
