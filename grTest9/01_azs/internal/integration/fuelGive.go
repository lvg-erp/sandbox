package integration

import "fmt"

type FuelGiveInfoRequest struct {
	TID string
}

type FuelGiveReceipt struct {
	FuelGiveID string `json:"FuelGetID"`
}

type FuelGiveConfirmationRequest struct {
	FuelGiveID string `json:"FuelGetID"`
}

type KazsFuelGiveReceiptRequest struct {
	KazsNumber       string         `json:"KazsNumber"`
	JarId            string         `json:"JarId"`
	FuelType         string         `json:"FuelType"`
	StartTime        int64          `json:"StartTime"`
	EndTime          int64          `json:"EndTime"`
	DocNumber        string         `json:"DocNumber"`
	FuelLiter        float64        `json:"FuelLiter"`
	SensorBeforeGive SensorInfoGive `json:"SensorBeforeGive"`
	SensorAfterGive  SensorInfoGive `json:"SensorAfterGive"`
	Errors           string         `json:"Errors"`
	AvgSpeed         float64        `json:"AvgSpeed"`
}

type SensorInfoGive struct {
	T  float64 `json:"t"`
	U  float64 `json:"U"`
	R  float64 `json:"r"`
	U1 float64 `json:"U1"`
	Ri float64 `json:"ri"`
	Tr float64 `json:"tr"`
	U2 float64 `json:"U2"`
	H  float64 `json:"H"`
	G  float64 `json:"G"`
}

func (o *KazsOperator) FuelGiveInfo(req *FuelGiveInfoRequest) (*FuelInfoResponse, error) {

	method := methods.FuelGiveInfo
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

func (o *KazsOperator) FuelGiveConfirmation(req *FuelGiveConfirmationRequest) (*FuelConfirmationResponse, error) {

	method := methods.FuelGiveConfirmation
	method.Path = fmt.Sprintf(method.Path, o.KazsID, req.FuelGiveID)

	var result FuelConfirmationResponse

	err := o.sendRequestKazsOperator(sendRequestArgs{
		methodsKazsOperator: method,
		urlValues:           nil,
		requestBody:         nil,
		resultStruct:        &result,
	})

	if err != nil {
		fmt.Println(fmt.Sprintf("Error in FuelGiveConfirmation: %v", err))
		return &result, err
	}

	fmt.Println(result)
	return &result, nil
}

func (o *KazsOperator) FuelGiveReceipt(req *FuelGiveReceipt, body *KazsFuelGiveReceiptRequest) error {

	method := methods.FuelGiveReceipt
	method.Path = fmt.Sprintf(method.Path, o.KazsID, req.FuelGiveID)

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
