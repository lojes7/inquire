package main

import (
	"log"
	"os"

	"github.com/lojes7/inquire/internal/router"
	"github.com/lojes7/inquire/pkg/infra"
)

// @title           Inquire API
// @version         1.0
// @description     This is the API documentation for the Inquire backend.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

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

	address := ":" + os.Getenv("PORT")

	err := r.Run(address)
	if err != nil {
		log.Fatalln("路由器出错")
	}
}
