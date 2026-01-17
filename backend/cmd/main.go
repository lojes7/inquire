package main

import (
	"log"
	"vvechat/internal/router"
	"vvechat/pkg/infra"
)

func main() {
	infra.Init()
	/*infra.GetDB().AutoMigrate(&model.User{})
	infra.GetDB().AutoMigrate(&model.Friendship{})
	infra.GetDB().AutoMigrate(&model.FriendshipRequest{})
	infra.GetDB().AutoMigrate(&model.Message{})
	infra.GetDB().AutoMigrate(&model.Conversation{})
	infra.GetDB().AutoMigrate(&model.MessageUser{})
	infra.GetDB().AutoMigrate(&model.ConversationUser{})
	infra.GetDB().AutoMigrate(&model.File{})*/
	r := router.Launch()
	err := r.Run(":8080")
	if err != nil {
		log.Fatalln("路由器出错")
	}
}
