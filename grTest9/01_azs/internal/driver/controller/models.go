package controller

import "encoding/json"

type ControllerMaping struct {
	Controller Controller `json:"Controller"`
}
type Controller struct {
	Settings Settings        `json:"Settings"`
	Doors    map[string]Door `json:"Doors"`
	Lock     map[string]Lock `json:"Lock"`
	Pump     map[string]Pump `json:"Pump"`
}

type Settings struct {
	IsUSB       bool    `json:"IsUSB"`
	COMPort     int     `json:"COMPort"`
	BaudRate    int     `json:"BaudRate"`
	StopBits    float32 `json:"StopBits"`
	DataBits    int     `json:"DataBits"`
	Parity      string  `json:"Parity"`
	ReadTimeout int     `json:"ReadTimeout"`
	Token       string  `json:"Token"`
}

type Door struct {
	Number string `json:"Number"`
	Relay  string `json:"Relay"`
	Close  string `json:"Close"`
	Open   string `json:"Open"`
}

type Pump struct {
	Number  string `json:"Number"`
	Relay   string `json:"Relay"`
	Enable  string `json:"Enable"`
	Disable string `json:"Disable"`
}

type Lock struct {
	Number string `json:"Number"`
	Relay  string `json:"Relay"`
	Close  string `json:"Close"`
	Open   string `json:"Open"`
}

type Response struct {
	I string `json:"I"`
	S string `json:"S"`
}

type Ping struct {
	Ping Response `json:"Ping"`
}

type Din struct {
	Din []Response `json:"Din"`
}

type Dout struct {
	Dout []Response `json:"Dout"`
}

type DoutSet struct {
	Dout []Response `json:"dout_set"`
}

type ComMessage map[string]json.RawMessage

type RelayItem struct {
	Index string `json:"i"`
	State string `json:"s"`
}
