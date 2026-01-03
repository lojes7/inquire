package infra

import (
	"errors"
	"sync"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

var (
	redisOnce sync.Once

	red *redis.Client
)

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
