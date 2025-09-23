package integration

import "net/http"

var methods methodsStruct

func init() {
	methods = methodsStruct{
		FuelGetInfo: MethodsKazsOperator{
			Path:   "/v1/kazs/%s/fuelget/%s",
			Method: http.MethodGet,
		},
		FuelGetConfirmation: MethodsKazsOperator{
			Path:   "/v1/kazs/%s/fuelget/%s/confirmation",
			Method: http.MethodGet,
		},
		FuelGetReceipt: MethodsKazsOperator{
			Path:   "/v1/kazs/%s/fuelget/%s/receipt",
			Method: http.MethodPost,
		},
		Telemetry: MethodsKazsOperator{
			Path:   "/v1/kazs/%s/telemetry",
			Method: http.MethodPost,
		},
		FuelGiveInfo: MethodsKazsOperator{
			Path:   "/v1/kazs/%s/fuelgive/%s",
			Method: http.MethodGet,
		},
		FuelGiveConfirmation: MethodsKazsOperator{
			Path:   "/v1/kazs/%s/fuelgive/%s/confirmation",
			Method: http.MethodGet,
		},
		FuelGiveReceipt: MethodsKazsOperator{
			Path:   "/v1/kazs/%s/fuelgive/%s/receipt",
			Method: http.MethodPost,
		},
		Activation: MethodsKazsOperator{
			Path:   "/v1/kazs/%s/activation",
			Method: http.MethodGet,
		},
		GetConfig: MethodsKazsOperator{
			Path:   "/v1/kazs/%s/config/%s",
			Method: http.MethodGet,
		},
	}
}

type MethodsKazsOperator struct {
	Path   string
	Method string
}

type methodsStruct struct {
	FuelGetInfo          MethodsKazsOperator
	FuelGetConfirmation  MethodsKazsOperator
	FuelGetReceipt       MethodsKazsOperator
	Telemetry            MethodsKazsOperator
	FuelGiveInfo         MethodsKazsOperator
	FuelGiveConfirmation MethodsKazsOperator
	FuelGiveReceipt      MethodsKazsOperator
	Activation           MethodsKazsOperator
	GetConfig            MethodsKazsOperator
}
