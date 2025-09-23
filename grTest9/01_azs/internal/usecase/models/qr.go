package models

type ScannerResponse struct {
	TYPE int    `json:"TYPE"`
	TID  string `json:"TID"`
	ADR  string `json:"ADR"`
}
