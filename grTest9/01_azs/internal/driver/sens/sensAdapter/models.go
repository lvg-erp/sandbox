package sensAdapter

type SENSMaping struct {
	LinAdapter          map[string]SENSMapingConfig `json:"LinAdapter"`
	VirtualLevelControl map[string]VirtualLC        `json:"VirtualLevelControl"`
}

type SENSMapingConfig struct {
	IsUSB       bool    `json:"IsUSB"`
	COMPort     int     `json:"COMPort"`
	BaudRate    int     `json:"BaudRate"`
	StopBits    float32 `json:"StopBits"`
	DataBits    int     `json:"DataBits"`
	Parity      string  `json:"Parity"`
	ReadTimeout int     `json:"ReadTimeout"`
	Bk          []int   `json:"Bk"`
	LC          []int   `json:"LC"`
}

type VirtualLC struct {
	LC []int `json:"LC"`
}
