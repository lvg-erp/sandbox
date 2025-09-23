package integration

import "fmt"

type GetConfigRequest struct {
	ConfigHash string `json:"configHash"`
}

type KazsGetConfigResponse struct {
	Error       bool                        `json:"Error"`
	Description string                      `json:"Description"`
	Message     string                      `json:"Message"`
	Result      KazsGetConfigNestedResponse `json:"Result"`
}

type KazsGetConfigNestedResponse struct {
	ConfigHash    string                   `json:"ConfigHash"`
	KazsNumber    string                   `json:"KazsNumber"`
	KazsTimezone  string                   `json:"KazsTimezone"`
	NtpServer     string                   `json:"NtpServer"`
	SupportNumber string                   `json:"SupportNumber"`
	Logo          string                   `json:"Logo"`
	Jars          []KazsNestedResponseJars `json:"Jars"`
}

func (o *KazsOperator) GetConfig(req *GetConfigRequest) (*KazsGetConfigResponse, error) {

	method := methods.GetConfig
	method.Path = fmt.Sprintf(method.Path, o.KazsID, req.ConfigHash)

	var result KazsGetConfigResponse

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
