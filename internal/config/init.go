package config

import (
	"log"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitConfig() error {
	viper.AddConfigPath("conf")
	viper.SetConfigName("database")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	log.Println("config:", viper.Get("database"))
	return nil
}

func InitDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(viper.GetString("postgres.dsn")), nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}
