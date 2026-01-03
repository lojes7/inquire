package infra

import (
	"log"
	"sync"

	"github.com/spf13/viper"
)

var (
	configOnce sync.Once
)

func InitConfig() error {
	var configInitErr error

	configOnce.Do(func() {
		viper.AddConfigPath("conf")
		viper.SetConfigName("config")

		err := viper.ReadInConfig()
		if err != nil {
			configInitErr = err
		}

		log.Println("config:", viper.Get("config"))
	})

	return configInitErr
}
