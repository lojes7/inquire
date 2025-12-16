package infra

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbOnce sync.Once

	db *gorm.DB
)

func InitDatabase() error {
	var dbInitErr error

	dbOnce.Do(func() {
		newLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: time.Second,
				LogLevel:      logger.Info,
				Colorful:      true,
			})

		var err error
		db, err = gorm.Open(postgres.Open(viper.GetString("postgres.dsn")), &gorm.Config{Logger: newLogger})
		if err != nil {
			dbInitErr = err
		}
	})

	return dbInitErr
}

func GetDB() *gorm.DB {
	return db
}
