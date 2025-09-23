package models

type FuelGiveStart struct {
	FuelGiveID       string
	StartTime        int64
	JarNumber        string
	FuelType         string
	DocNumber        string
	SensorBeforeGive SensorInfoGive
	Liters           float32
}

type FuelGiveReports struct {
	FuelGiveID string
	JarNumber  string
	StartTime  int64
	EndTime    int64
	Liters     float32
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

type InsertFuelGiveTransactionDep struct {
	FuelGiveID       string
	KazsNumber       string
	JarNumber        string
	StartTime        int64
	EndTime          *int64
	FuelType         string
	DocNumber        string
	SensorBeforeGive *SensorInfoGive
	SensorAfterGive  *string
	FuelLiters       *float32
	FuelLitersPlan   float64
	SendStatus       bool
	Errors           *string
}

type UpdateFuelGiveTransactionDep struct {
	FuelGiveID       string
	KazsNumber       *string
	JarNumber        *string
	StartTime        *int64
	EndTime          *int64
	FuelType         *string
	DocNumber        *string
	SensorBeforeGive *SensorInfoGive
	SensorAfterGive  *SensorInfoGive
	FuelLiters       *float32
	FuelLitersPlan   *float64
	SendStatus       *bool
	Errors           *string
}

type FuelGiveReceipt struct {
	KazsNumber       string         `json:"KazsNumber"`
	JarId            string         `json:"JarId"`
	FuelType         string         `json:"FuelType"`
	StartTime        int64          `json:"StartTime"`
	EndTime          int64          `json:"EndTime"`
	DocNumber        string         `json:"DocNumber"`
	FuelLiter        float64        `json:"FuelLiter"`
	FuelLitersPlan   float64        `json:"FuelLitersPlan"`
	SensorBeforeGive SensorInfoGive `json:"SensorBeforeGive"`
	SensorAfterGive  SensorInfoGive `json:"SensorAfterGive"`
	SendStatus       bool           `json:"SendStatus"`
	Errors           string         `json:"Errors"`
}

type LastFuelGiveReceipt struct {
	FuelGiveID       string         `json:"FuelGiveID"`
	KazsNumber       string         `json:"KazsNumber"`
	JarId            string         `json:"JarId"`
	FuelType         string         `json:"FuelType"`
	StartTime        int64          `json:"StartTime"`
	EndTime          int64          `json:"EndTime"`
	DocNumber        string         `json:"DocNumber"`
	FuelLiter        float64        `json:"FuelLiter"`
	FuelLitersPlan   float64        `json:"FuelLitersPlan"`
	SensorBeforeGive SensorInfoGive `json:"SensorBeforeGive"`
	SensorAfterGive  SensorInfoGive `json:"SensorAfterGive"`
	SendStatus       bool           `json:"SendStatus"`
	Errors           string         `json:"Errors"`
}
