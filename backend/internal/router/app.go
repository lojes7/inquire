package router

import (
	"github.com/gin-gonic/gin"
	_ "github.com/lojes7/inquire/docs"
	"github.com/lojes7/inquire/internal/handler"
	"github.com/lojes7/inquire/internal/ws"
	"github.com/lojes7/inquire/pkg/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Launch() *gin.Engine {
	r := gin.Default()

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 跨域中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")                            // 允许所有域名访问
		c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE")           // 允许的HTTP方法
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization") // 允许的请求头
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204) // 对于预检请求，直接返回成功
			return
		}
		c.Next()
	})

	// RESTful API
	api := r.Group("/api")
	{
		api.POST("/register", handler.Register)               // 注册
		api.POST("/login/uid", handler.LoginByUid)            // 微信号登陆
		api.POST("/login/phone_number", handler.LoginByPhone) // 手机号登陆
		// 刷新 token
		api.POST("/auth/refresh_token", middleware.RefreshAuth(), handler.RefreshToken)

		// /auth 表示需要鉴权的操作
		auth := api.Group("/auth", middleware.JWTAuth())
		{
			// websocket
			auth.GET("/ws", func(c *gin.Context) {
				ws.ServeWs(ws.GetHub(), c)
			})

			// 修改个人信息
			me := auth.Group("/me")
			{
				me.POST("/uid", handler.ReviseUid)           //修改微信号
				me.POST("/password", handler.RevisePassword) // 修改密码
				me.POST("/name", handler.ReviseName)         // 修改用户名
			}

			// 查看他人信息
			info := auth.Group("/info")
			{
				info.GET("/friends/id/:id", handler.FriendInfoByID)        // 根据ID 查看好友信息
				info.GET("/strangers/id/:id", handler.StrangerInfoByID)    // 根据ID 查看陌生人信息
				info.GET("/friends/uid/:uid", handler.FriendInfoByUid)     // 根据Uid 查看好友信息
				info.GET("/strangers/uid/:uid", handler.StrangerInfoByUid) // 根据Uid 查看陌生人信息
			}

			// 好友申请相关
			request := auth.Group("/friendship_requests")
			{
				request.GET("", handler.FriendRequestList)                  //加载好友申请列表
				request.POST("", handler.SendFriendRequest)                 //发送好友申请
				request.POST("/:request_id", handler.FriendRequestAccept)   //同意好友申请
				request.DELETE("/:request_id", handler.FriendRequestDelete) //删除好友申请
			}

			// 好友相关
			friendship := auth.Group("/friendships")
			{
				friendship.GET("", handler.FriendshipList)                  //加载好友列表
				friendship.DELETE("/:friend_id", handler.DeleteFriendship)  //删除好友
				friendship.POST("/remark/:friend_id", handler.ReviseRemark) //修改好友备注
			}

			// 消息相关
			message := auth.Group("/messages")
			{
				message.POST("/text", handler.SendText)          //发送文本消息
				message.POST("/file", handler.SendFile)          // 发送文件
				message.DELETE("/recall", handler.RecallMessage) //撤回消息
				message.DELETE("/delete", handler.DeleteMessage) //删除消息
			}

			// 会话相关
			converse := auth.Group("/conversations")
			{
				converse.GET("", handler.ConversationList)                  // 加载聊天列表
				converse.POST("/private", handler.StartPrivateConversation) // 发起私聊
				converse.POST("/group")                                     //创建群聊
				converse.GET("/:conversation_id", handler.ChatHistoryList)  // 加载聊天记录
			}

			// 文件相关
			file := auth.Group("/files")
			{
				file.GET("/:message_id", handler.DownloadFile)    // 下载文件
				file.POST("/search", handler.SemanticSearchFiles) // 语义化搜索文件
			}
		}
	}
	return r
}
