package integration

import "fmt"

type KazsActivationResponse struct {
	Error       bool                         `json:"Error"`
	Description string                       `json:"Description"`
	Message     string                       `json:"Message"`
	Result      KazsActivationNestedResponse `json:"Result"`
}

type KazsActivationNestedResponse struct {
	KazsApiKey    string                   `json:"KazsApiKey"`
	ResetPass     string                   `json:"ResetPass"`
	KazsID        string                   `json:"KazsID"`
	URL           string                   `json:"URL"`
	ConfigHash    string                   `json:"ConfigHash"`
	KazsNumber    string                   `json:"KazsNumber"`
	KazsTimezone  string                   `json:"KazsTimezone"`
	NtpServer     string                   `json:"NtpServer"`
	SupportNumber string                   `json:"SupportNumber"`
	Logo          string                   `json:"Logo"`
	Jars          []KazsNestedResponseJars `json:"Jars"`
}

type KazsNestedResponseJars struct {
	JarId    string `json:"JarId"`
	FuelType string `json:"FuelType"`
}

func (o *KazsOperator) Activation(tid string, url string) (*KazsActivationResponse, error) {

	method := methods.Activation
	method.Path = fmt.Sprintf(method.Path, tid)

	var result KazsActivationResponse

	err := o.setUrl(url)
	if err != nil {
		return &result, err
	}

	err = o.sendRequestKazsOperator(sendRequestArgs{
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
