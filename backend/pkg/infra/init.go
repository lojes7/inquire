package infra

import (
	"errors"
	"log"
	"os"
	"sync"
	"time"
	"vvechat/pkg/secure"
	_ "vvechat/pkg/utils"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Init() {
	err := InitConfig()
	if err != nil {
		log.Fatalln(err)
	}

	err = InitDatabase()
	if err != nil {
		log.Fatalln(err)
	}

	err = InitRedis()
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
	dbOnce     sync.Once
	configOnce sync.Once
	redisOnce  sync.Once

	red             *redis.Client
	db              *gorm.DB
	fileStoragePath string
)

func GetFilePath() string {
	return fileStoragePath
}

func InitFileStorage() {
	fileStoragePath = viper.GetString("file_storage.path")
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

func GetRedis() *redis.Client {
	return red
}

func InitRedis() error {
	var redisInitErr error

	redisOnce.Do(func() {
		red = redis.NewClient(&redis.Options{
			Addr:     viper.GetString("redis.addr"),
			Password: viper.GetString("redis.password"),
			DB:       viper.GetInt("redis.db"),
			PoolSize: viper.GetInt("redis.pool_size"),
		})

		// 使用 Ping 来验证连接
		_, err := red.Ping().Result()
		if err != nil {
			redisInitErr = errors.New("redis连接失败:  " + err.Error())
		}
	})

	return redisInitErr
}

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
