package main

import (
	"log"
	"wechat/internal/config"
	"wechat/internal/router"
)

func main() {
	err := config.InitConfig()
	if err != nil {
		log.Fatalln(err)
	}
	db, err := config.InitDatabase()
	if err != nil {
		log.Fatal(err)
	}

	r := router.Launch(db)
	r.Run(":8080")
}
