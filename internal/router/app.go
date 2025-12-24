package router

import (
	"vvechat/internal/handler"
	"vvechat/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func Launch() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/register", handler.Register)
		api.POST("/login/uid", handler.LoginByUid)
		api.POST("/login/phone_number", handler.LoginByPhone)
		api.POST("/auth/refresh_token", middleware.RefreshAuth(), handler.RefreshToken)

		auth := api.Group("/auth", middleware.JWTAuth())
		{
			auth.PATCH("/me/uid", handler.ReviseUid) //修改微信号

			converse := auth.Group("/conversations")
			{
				converse.GET("")                  //加载聊天列表
				converse.GET("/:conversation_id") //加载聊天窗口
			}

			request := auth.Group("/friendship_requests")
			{
				request.GET("", handler.FriendRequestList)                //加载好友申请列表
				request.POST("", handler.SendFriendRequest)               //发送好友申请
				request.POST("/:request_id", handler.FriendRequestAction) //同意或拒绝好友申请
			}

			friendship := auth.Group("/friendships")
			{
				friendship.GET("", handler.FriendshipList)                 //加载好友列表
				friendship.DELETE("/:friend_id", handler.DeleteFriendship) //删除好友
			}

			message := auth.Group("/messages")
			{
				message.POST("")          //发送消息
				message.DELETE("/recall") //撤回消息
				message.DELETE("/delete") //删除消息
			}

		}
	}

	return r
}
