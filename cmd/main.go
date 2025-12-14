package main

import (
	"log"
	"wechat/internal/router"
	"wechat/internal/service"
	"wechat/internal/system"
)

func main() {
	err := system.InitConfig()
	if err != nil {
		log.Fatalln(err)
	}
	db, err := system.InitDatabase()
	if err != nil {
		log.Fatal(err)
	}

	serve := service.Serve{
		DB: db,
	}

	r := router.Launch(&serve)
	r.Run(":8080")
}
