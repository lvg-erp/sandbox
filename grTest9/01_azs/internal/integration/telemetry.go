package integration

import "fmt"

type TelemetryRequest struct {
	StatusTime     int64                     `json:"StatusTime"`
	PowerStatus    int64                     `json:"PowerStatus"`
	BatteryStatus  float64                   `json:"BatteryStatus"`
	OutTemperature float64                   `json:"OutTemperature"`
	KazsErrors     string                    `json:"KazsErrors"`
	Jars           []KazsTelemetryNestedJars `json:"Jars"`
}

type KazsTelemetryNestedJars struct {
	JarId         string  `json:"JarId"`
	JarLockStatus int     `json:"JarLockStatus"`
	NozzleStatus  int     `json:"NozzleStatus"`
	H             float32 `json:"H"`
	T             float32 `json:"T"`
	Pr            float32 `json:"Pr"`
	U             float32 `json:"U"`
	G             float32 `json:"G"`
	R             float32 `json:"R"`
	U1            float32 `json:"U1"`
	H2            float32 `json:"H2"`
	Ut            float32 `json:"Ut"`
	Rt            float32 `json:"Rt"`
	Ri            float32 `json:"Ri"`
	Tr            float32 `json:"Tr"`
	U2            float32 `json:"U2"`
	Nt            string  `json:"Nt"`
	Dg            float32 `json:"Dg"`
	Ts            float32 `json:"Ts"`
	JarErrors     string  `json:"JarErrors"`
}

func (o *KazsOperator) Telemetry(body *TelemetryRequest) error {

	method := methods.Telemetry
	method.Path = fmt.Sprintf(method.Path, o.KazsID)
	err := o.sendRequestKazsOperator(sendRequestArgs{
		methodsKazsOperator: method,
		urlValues:           nil,
		requestBody:         body,
		resultStruct:        nil,
	})
	if err != nil {
		return err
	}

	return nil
}
