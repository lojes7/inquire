package infra

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/lojes7/inquire/pkg/secure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 若修改文件保存路径 请同时修改Dockerfile
const fileStoragePath = "/assets"

func Init() {
	err := InitDatabase()
	if err != nil {
		log.Fatalln(err)
	}

	err = secure.InitJWT()
	if err != nil {
		log.Fatalln(err)
	}
}

var (
	dbOnce sync.Once
	db     *gorm.DB
)

func GetFilePath() string {
	return fileStoragePath
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
