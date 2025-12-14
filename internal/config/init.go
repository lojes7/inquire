package config

import (
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	newLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		})
	db, err := gorm.Open(postgres.Open(viper.GetString("postgres.dsn")), &gorm.Config{Logger: newLogger})
	if err != nil {
		return nil, err
	}
	return db, nil
}
