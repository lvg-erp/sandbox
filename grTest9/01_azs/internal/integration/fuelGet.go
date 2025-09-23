package integration

import "fmt"

type FuelInfoResponse struct {
	Error       bool
	Description string
	Message     string
	Result      Result
}

type Result struct {
	JarId     string
	FuelLiter float64
	FuelType  string
}

type FuelGetInfoRequest struct {
	TID string
}

type FuelConfirmationResponse struct {
	Error       bool                         `json:"Error"`
	Description string                       `json:"Description"`
	Message     string                       `json:"Message"`
	Result      FuelConfirmationNestedResult `json:"Result"`
}

type FuelConfirmationNestedResult struct {
	JarId     string  `json:"JarId"`
	DocNumber string  `json:"DocNumber"`
	FuelLiter float64 `json:"FuelLiter"`
	FuelType  string  `json:"FuelType"`
}

type FuelConfirmationRequest struct {
	FuelGetID string `json:"FuelGetID"`
}

type FuelGetReceipt struct {
	FuelGetID string `json:"FuelGetID"`
}

type FuelGetReceiptRequest struct {
	KazsNumber      string     `json:"KazsNumber"`
	JarId           string     `json:"JarId"`
	FuelType        string     `json:"FuelType"`
	StartTime       int64      `json:"StartTime"`
	EndTime         int64      `json:"EndTime"`
	DocNumber       string     `json:"DocNumber"`
	FuelLiter       float64    `json:"FuelLiter"`
	SensorBeforeGet SensorInfo `json:"SensorBeforeGet"`
	SensorAfterGet  SensorInfo `json:"SensorAfterGet"`
	Errors          string     `json:"Errors"`
}

type SensorInfo struct {
	H  float64 `json:"h"`
	T  float64 `json:"t"`
	Pr float64 `json:"pr"`
	U  float64 `json:"U"`
	G  float64 `json:"G"`
	R  float64 `json:"r"`
	U1 float64 `json:"U1"`
	H2 float64 `json:"h2"`
	Ut float64 `json:"Ut"`
	Rt float64 `json:"rt"`
	Ri float64 `json:"ri"`
	Tr float64 `json:"tr"`
	U2 float64 `json:"U2"`
	Nt string  `json:"nt"`
	Dg float64 `json:"dG"`
	Ts float64 `json:"tS"`
}

func (o *KazsOperator) FuelGetInfo(req *FuelGetInfoRequest) (*FuelInfoResponse, error) {

	method := methods.FuelGetInfo
	method.Path = fmt.Sprintf(method.Path, o.KazsID, req.TID)

	var result FuelInfoResponse

	err := o.sendRequestKazsOperator(sendRequestArgs{
		methodsKazsOperator: method,
		urlValues:           nil,
		requestBody:         nil,
		resultStruct:        &result,
	})
	if err != nil {
		return &result, err
	}

	return &result, nil
}

func (o *KazsOperator) FuelGetConfirmation(req *FuelConfirmationRequest) (*FuelConfirmationResponse, error) {

	method := methods.FuelGetConfirmation
	method.Path = fmt.Sprintf(method.Path, o.KazsID, req.FuelGetID)

	var result FuelConfirmationResponse

	err := o.sendRequestKazsOperator(sendRequestArgs{
		methodsKazsOperator: method,
		urlValues:           nil,
		requestBody:         nil,
		resultStruct:        &result,
	})

	if err != nil {
		return &result, err
	}
	return &result, nil
}

func (o *KazsOperator) FuelGetReceipt(req *FuelGetReceipt, body *FuelGetReceiptRequest) error {

	method := methods.FuelGetReceipt
	method.Path = fmt.Sprintf(method.Path, o.KazsID, req.FuelGetID)

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
