package main

import (
	"log"
	"vvechat/internal/router"
	"vvechat/pkg/infra"
	"vvechat/pkg/utils"
)

func initAll() error {
	err := utils.InitSnowflake()
	if err != nil {
		return err
	}

	err = infra.InitConfig()
	if err != nil {
		return err
	}

	err = infra.InitDatabase()
	if err != nil {
		return err
	}

	err = infra.InitRedis()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	err := initAll()
	if err != nil {
		log.Fatalln(err)
	}

	r := router.Launch()
	r.Run(":8080")
}
