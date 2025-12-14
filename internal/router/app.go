package router

import (
	"wechat/internal/service"

	"github.com/gin-gonic/gin"
)

func Launch(serve *service.Serve) *gin.Engine {
	r := gin.Default()

	r.POST("/user/register", serve.RegisterUser)

	return r
}
