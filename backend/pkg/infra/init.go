package infra

import (
	"log"
	"os"
	"sync"
	"time"
	"vvechat/pkg/secure"
	_ "vvechat/pkg/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Init() {
	err := InitDatabase()
	if err != nil {
		log.Fatalln(err)
	}

	err = secure.InitJWT()
	if err != nil {
		log.Fatalln(err)
	}

	InitFileStorage()
}

var (
	dbOnce          sync.Once
	db              *gorm.DB
	fileStoragePath string
)

func GetFilePath() string {
	return fileStoragePath
}

func InitFileStorage() {
	fileStoragePath = os.Getenv("FILE_STORAGE_PATH")
}

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
		dsn := os.Getenv("DATABASE_URL")

		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: newLogger,
		},
		)
		if err != nil {
			dbInitErr = err
		}
	})

	return dbInitErr
}

func GetDB() *gorm.DB {
	return db
}
