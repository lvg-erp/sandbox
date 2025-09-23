package config

import (
	"fmt"
	"github.com/go-playground/validator"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
)

const (
	_defaultConfigurationsPath = "./config.yaml"
)

type Config struct {
	Logger       Logger       `yaml:"Logger"`
	KazsConfig   KazsConfig   `yaml:"KazsConfig" validate:"required"`
	DriverConfig DriverConfig `yaml:"DriverConfig" validate:"required"`
	AppConfig    AppConfig    `yaml:"AppConfig" validate:"required"`
}

type AppConfig struct {
	FuelGetInfoTimeout          int           `yaml:"FuelGetInfoTimeout" validate:"required"`
	FuelGiveInfoTimeout         int           `yaml:"FuelGiveInfoTimeout" validate:"required"`
	FuelGetConfirmationTimeout  int           `yaml:"FuelGetConfirmationTimeout" validate:"required"`
	FuelGiveConfirmationTimeout int           `yaml:"FuelGiveConfirmationTimeout" validate:"required"`
	FuelGetReceiptTimeout       int           `yaml:"FuelGetReceiptTimeout" validate:"required"`
	FuelGiveReceiptTimeout      int           `yaml:"FuelGiveReceiptTimeout" validate:"required"`
	TelemetryTimeout            int           `yaml:"TelemetryTimeout" validate:"required"`
	GetConfigTimeout            int           `yaml:"GetConfigTimeout" validate:"required"`
	NTPOffset                   time.Duration `yaml:"NTPOffset" validate:"required"`
	NTPFailedTimeout            int64         `yaml:"NTPFailedTimeout" validate:"required"`
}

type Logger struct {
	Level logrus.Level
	Type  string
}

type QR struct {
	COMPort     string  `json:"COMPort" validate:"required"`
	BaudRate    string  `json:"BaudRate" validate:"required,numeric"`
	DataBits    int     `json:"DataBits" validate:"required,oneof=5 6 7 8"`
	StopBits    float64 `json:"StopBits" validate:"required,oneof=1 1.5 2"`
	Parity      string  `json:"Parity" validate:"required"`
	ReadTimeout int     `json:"ReadTimeout" validate:"gte=0"`
}

type KazsConfig struct {
	FuelGetConfig  FuelGetConfig  `yaml:"FuelGetConfig" validate:"required"`
	Telemetry      Telemetry      `yaml:"Telemetry" validate:"required"`
	FuelGiveConfig FuelGiveConfig `yaml:"FuelGiveConfig" validate:"required"`
}

type Telemetry struct {
	GetTelemetry           int           `yaml:"GetTelemetry" validate:"required"`
	GetTelemetryJarTimeout time.Duration `yaml:"GetTelemetryJarTimeout" validate:"required"`
	GetTRKTelemetry        time.Duration `yaml:"GetTRKTelemetry" validate:"required"`
	SendTelemetry          int           `yaml:"SendTelemetry" validate:"required"`
	TelemetryHistory       int           `yaml:"TelemetryHistory" validate:"required"`
	MaxFailureCount        int           `yaml:"MaxFailureCount" validate:"required"`
	TelemetryUnits         int           `yaml:"TelemetryUnits" validate:"required,oneof=1 1000"`
	ActualTimeSENS         int64         `yaml:"ActualTimeSENS" validate:"required"`
	ActualTimeTRK          int64         `yaml:"ActualTimeTRK" validate:"required"`
	ActualTemp             int64         `yaml:"ActualTemp" validate:"required"`
}

type FuelGetConfig struct {
	FuelGetStartScreenTimeout          int           `yaml:"FuelGetStartScreenTimeout" validate:"required"`
	FuelLiterStart                     int           `yaml:"FuelLiterStart" validate:"required"`
	FuelGetInProgressScreenTimeout     int           `yaml:"FuelGetInProgressScreenTimeout" validate:"required"`
	FuelGetCompleteScreenTimeout       int           `yaml:"FuelGetCompleteScreenTimeout" validate:"required"`
	FuelGetCompleteScreenNoFuelTimeout int           `yaml:"FuelGetCompleteScreenNoFuelTimeout" validate:"required"`
	FuelGetUnits                       int           `yaml:"FuelGetUnits" validate:"required,oneof=1 1000"`
	CheckUnsentTransactions            int           `yaml:"CheckUnsentTransactions" validate:"required"`
	DoorCloseTimeout                   time.Duration `yaml:"DoorCloseTimeout" validate:"required"`
	StallTimeout                       int64         `yaml:"StallTimeout" validate:"required"`
	StallDeltaLiters                   float32       `yaml:"StallDeltaLiters" validate:"required"`
}

type FuelGiveConfig struct {
	FuelGiveStartScreenTimeout    int     `yaml:"FuelGiveStartScreenTimeout" validate:"required"`
	FuelGiveCompleteScreenTimeout int     `yaml:"FuelGiveCompleteScreenTimeout" validate:"required"`
	FuelGiveTimeout               int     `yaml:"FuelGiveTimeout" validate:"required"`
	CheckUnsentTransactions       int     `yaml:"CheckUnsentTransactions" validate:"required"`
	ActualTimeSENS                int64   `yaml:"ActualTimeSENS" validate:"required"`
	FuelGiveEndTimeout            int64   `yaml:"FuelGiveEndTimeout" validate:"required"`
	FailedTRKResponse             int64   `yaml:"FailedTRKResponse" validate:"required"`
	CountFuelGiveEnd              int     `yaml:"CountFuelGiveEnd" validate:"required"`
	StopPumpTimeout               float32 `yaml:"StopPumpTimeout" validate:"required"`
	FuelGiveSpeedStart            float32 `yaml:"FuelGiveSpeedStart" validate:"required"`
	FuelGiveSpeedMin              float32 `yaml:"FuelGiveSpeedMin" validate:"required"`
}

type DriverConfig struct {
	GetMainStatusTimeout         int `yaml:"GetMainStatusTimeout" validate:"required"`
	GetTRKStatusTimeout          int `yaml:"GetTRKStatusTimeout" validate:"required"`
	SetFuelGiveTimeout           int `yaml:"SetFuelGiveTimeout" validate:"required"`
	ApprovalTRKTimeout           int `yaml:"ApprovalTRKTimeout" validate:"required"`
	GetFuelGiveStatusTimeout     int `yaml:"GetFuelGiveStatusTimeout" validate:"required"`
	GetFullFuelGiveStatusTimeout int `yaml:"GetFullFuelGiveStatusTimeout" validate:"required"`
	FuelGiveSuccessTimeout       int `yaml:"FuelGiveSuccessTimeout" validate:"required"`
	ControllerTimeout            int `yaml:"ControllerTimeout" validate:"required"`
	RetryAttempts                int `yaml:"RetryAttempts" validate:"required"`
}

func NewConfig() (*Config, error) {
	vp := viper.New()

	vp.SetConfigFile(_defaultConfigurationsPath)

	if err := vp.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("viper.ReadInConfig error: %v", err)
	}

	var config Config
	if err := vp.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("viper.Unmarshal error: %v", err)
	}

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return nil, fmt.Errorf("validate config error: %v", err)
	}

	return &config, nil

}
