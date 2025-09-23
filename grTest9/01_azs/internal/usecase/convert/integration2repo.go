package convert

import (
	"fuelazs/internal/integration"
	"fuelazs/internal/usecase/models"
	"time"
)

func ConvertFuelGetReport(fg integration.FuelGetReceiptRequest) models.FuelGetReport {
	return models.FuelGetReport{
		KazsNumber: fg.KazsNumber,
		JarId:      fg.JarId,
		FuelType:   fg.FuelType,
		StartTime:  fg.StartTime,
		EndTime:    fg.EndTime,
		DocNumber:  fg.DocNumber,
		FuelLiter:  fg.FuelLiter,
		SensorBeforeGet: models.SensorInfo{
			H:  fg.SensorBeforeGet.H,
			T:  fg.SensorBeforeGet.T,
			Pr: fg.SensorBeforeGet.Pr,
			U:  fg.SensorBeforeGet.U,
			G:  fg.SensorBeforeGet.G,
			R:  fg.SensorBeforeGet.R,
			U1: fg.SensorBeforeGet.U1,
			H2: fg.SensorBeforeGet.H2,
			Ut: fg.SensorBeforeGet.Ut,
			Rt: fg.SensorBeforeGet.Rt,
			Ri: fg.SensorBeforeGet.Ri,
			Tr: fg.SensorBeforeGet.Tr,
			U2: fg.SensorBeforeGet.U2,
			Nt: fg.SensorBeforeGet.Nt,
			Dg: fg.SensorBeforeGet.Dg,
			Ts: fg.SensorBeforeGet.Ts,
		},
		SensorAfterGet: models.SensorInfo{
			H:  fg.SensorAfterGet.H,
			T:  fg.SensorAfterGet.T,
			Pr: fg.SensorAfterGet.Pr,
			U:  fg.SensorAfterGet.U,
			G:  fg.SensorAfterGet.G,
			R:  fg.SensorAfterGet.R,
			U1: fg.SensorAfterGet.U1,
			H2: fg.SensorAfterGet.H2,
			Ut: fg.SensorAfterGet.Ut,
			Rt: fg.SensorAfterGet.Rt,
			Ri: fg.SensorAfterGet.Ri,
			Tr: fg.SensorAfterGet.Tr,
			U2: fg.SensorAfterGet.U2,
			Nt: fg.SensorAfterGet.Nt,
			Dg: fg.SensorAfterGet.Dg,
			Ts: fg.SensorAfterGet.Ts,
		},
		Errors: fg.Errors,
	}
}

func ConvertActivation(data *integration.KazsActivationResponse) models.ActivationData {
	return models.ActivationData{
		LastModification: time.Now(),
		KazsAPIKey:       data.Result.KazsApiKey,
		ResetPassword:    data.Result.ResetPass,
		KazsID:           data.Result.KazsID,
		URL:              data.Result.URL,
		ConfigHash:       data.Result.ConfigHash,
		KazsNumber:       data.Result.KazsNumber,
		KazsTimezone:     data.Result.KazsTimezone,
		NTPServer:        data.Result.NtpServer,
		SupportNumber:    data.Result.SupportNumber,
		Logo:             data.Result.Logo,
	}
}
