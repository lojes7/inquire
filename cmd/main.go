package main

import (
	"log"
	"vvechat/internal/router"
	"vvechat/pkg/infra"
	"vvechat/pkg/secure"
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
	err = secure.InitJWT()
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

	/*infra.GetDB().AutoMigrate(&model.User{})
	infra.GetDB().AutoMigrate(&model.Friendship{})
	infra.GetDB().AutoMigrate(&model.FriendshipRequest{})
	infra.GetDB().AutoMigrate(&model.Message{})
	infra.GetDB().AutoMigrate(&model.Conversation{})
	infra.GetDB().AutoMigrate(&model.MessageUser{})
	infra.GetDB().AutoMigrate(&model.ConversationUser{})*/
	r := router.Launch()
	err = r.Run(":8080")
	if err != nil {
		log.Fatalln("路由器出错")
	}
}
