package trk

type TRKResponse struct {
	Address           string             `json:"Address"`
	GeneralParameters map[string]float32 `json:"GeneralParameters"`
	DriverName        string             `json:"DriverName"`
	DevicePNumber     string             `json:"DevicePNumber"`
}

type FuelGiveReport struct {
	LiterCount float32 `json:"literCount"` // Конечное количество литров
	StartTime  int64   `json:"startTime"`  // Время начала мониторинга заправки (Unix timestamp)
	EndTime    int64   `json:"endTime"`    // Время завершения заправки (Unix timestamp)
}

type ValueSetting struct {
	Left  int `json:"left"`
	Right int `json:"right"`
}
