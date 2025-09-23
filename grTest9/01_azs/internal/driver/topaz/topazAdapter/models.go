package topazAdapter

type TRKMaping struct {
	TRK map[string]TRKMapingConfig `json:"TRK"`
}

type TRKMapingConfig struct {
	IsUSB       bool    `json:"IsUSB"`
	COMPort     int     `json:"COMPort"`
	BaudRate    int     `json:"BaudRate"`
	DataBits    int     `json:"DataBits"`
	StopBits    float32 `json:"StopBits"`
	Parity      string  `json:"Parity"`
	ReadTimeout int     `json:"ReadTimeout"`
}
