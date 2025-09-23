package models

type JarInfo struct {
	Address       string    `json:"Address"`
	H             float32   // Уровень жидкости
	T             float32   // Температура жидкости
	Pr            float32   // Процентное заполнение
	U             float32   // Объем жидкости
	G             float32   // Масса
	R             float32   // Плотность
	U1            float32   // Объем основного продукта
	H2            float32   // Уровень раздела сред
	Ut            float32   // Объем продукта, приведенный к стандартным условиям
	Rt            float32   // Плотность, приведенная к стандартным условиям
	Ri            float32   // Измеренная плотность
	Tr            float32   // Температура измерения плотности
	U2            float32   // Объем жидкости под разделом сред
	Dg            float32   // Абсолютная погрешность измерения массы
	Ts            float32   // Начальная температура
	Nt            []float32 // Показания датчиков температуры
	DriverName    string
	DevicePNumber string
}

type FuelGetReport struct {
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

type Errors struct {
	Time    int64  `json:"Time"`
	Handler string `json:"Handler"`
	Id      string `json:"Id"`
	Error   string `json:"Error"`
}

type InsertFuelGetTransactionDep struct {
	FuelGetID            string
	KazsNumber           string
	JarNumber            string
	StartTime            int64
	EndTime              *int64
	MonitoringFinishTime *int64
	FuelType             string
	DocNumber            string
	SensorBeforeGive     *SensorInfo
	SensorAfterGive      *SensorInfo
	FuelLiterPlan        float64
	FuelLiters           *float32
	SendStatus           bool
	Speed                *float32
	Errors               *string
}

type UpdateFuelGetTransactionDep struct {
	FuelGetID            string
	KazsNumber           *string
	JarNumber            *string
	StartTime            *int64
	EndTime              *int64
	MonitoringFinishTime *int64
	FuelType             *string
	DocNumber            *string
	SensorBeforeGive     *SensorInfo
	SensorAfterGive      *SensorInfo
	FuelLiters           *float32
	FuelLitersPlans      *float64
	SendStatus           *bool
	Speed                *float32
	Errors               *string
}

type FuelGetReceipt struct {
	KazsNumber           string     `json:"KazsNumber"`
	JarId                string     `json:"JarId"`
	FuelType             string     `json:"FuelType"`
	StartTime            int64      `json:"StartTime"`
	EndTime              int64      `json:"EndTime"`
	MonitoringFinishTime int64      `json:"MonitoringFinishTime"`
	DocNumber            string     `json:"DocNumber"`
	FuelLiter            float64    `json:"FuelLiter"`
	FuelLitersPlan       float64    `json:"FuelLitersPlan"`
	Speed                float64    `json:"Speed"`
	SensorBeforeGet      SensorInfo `json:"SensorBeforeGet"`
	SensorAfterGet       SensorInfo `json:"SensorAfterGet"`
	SendStatus           bool       `json:"SendStatus"`
	Errors               string     `json:"Errors"`
}

type LastFuelGetReceipt struct {
	FuelGetID            string     `json:"FuelGiveID"`
	KazsNumber           string     `json:"KazsNumber"`
	JarId                string     `json:"JarId"`
	FuelType             string     `json:"FuelType"`
	StartTime            int64      `json:"StartTime"`
	EndTime              int64      `json:"EndTime"`
	MonitoringFinishTime int64      `json:"MonitoringFinishTime"`
	DocNumber            string     `json:"DocNumber"`
	FuelLiter            float64    `json:"FuelLiter"`
	FuelLitersPlan       float64    `json:"FuelLitersPlan"`
	Speed                float64    `json:"Speed"`
	SensorBeforeGive     SensorInfo `json:"SensorBeforeGive"`
	SensorAfterGive      SensorInfo `json:"SensorAfterGive"`
	SendStatus           bool       `json:"SendStatus"`
	Errors               string     `json:"Errors"`
}

type JarsInfo struct {
	JarNumber     string
	JarLockStatus int
	TrkStatus     int
	H             float32
	T             float32
	Pr            float32
	U             float32
	G             float32
	R             float32
	U1            float32
	H2            float32
	Ut            float32
	Rt            float32
	Ri            float32
	Tr            float32
	U2            float32
	Dg            float32
	Ts            float32
	Nt            string
}
