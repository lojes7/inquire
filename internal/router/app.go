package router

import (
	"vvechat/internal/handler"

	"github.com/gin-gonic/gin"
)

func Launch() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/register", handler.Register)
		api.POST("/login/uid", handler.LoginByUid)
		api.POST("/login/phone", handler.LoginByPhone)
	}

	return r
}
