package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Credentials struct {
	//Sentry Sentry
}

//	type Sentry struct {
//		Dsn              string  `yaml:"Dsn"`              // DSN для подключения к Sentry
//		TracesSampleRate float64 `yaml:"TracesSampleRate"` // Частота выборки трассировок
//	}
func NewCredentials() (*Credentials, error) {
	configType := "credentials"

	var cred Credentials

	viper.AddConfigPath(".")
	viper.SetConfigName(configType)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("%s %s", configType, err)
	}

	//err = viper.UnmarshalKey("Sentry", &cred.Sentry)
	//if err != nil {
	//	return nil, fmt.Errorf("%s %s", configType, err)
	//}

	return &cred, nil
}
