package main

import (
	"log"
	"wechat/internal/router"
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

	r := router.Launch(db)
	r.Run(":8080")
}
