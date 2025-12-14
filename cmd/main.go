package main

import (
	"log"
	"vvechat/internal/config"
	"vvechat/internal/router"
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
