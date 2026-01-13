package handler

import (
	"vvechat/internal/model"
	"vvechat/internal/service"
	"vvechat/pkg/response"

	"github.com/gin-gonic/gin"
)

func SendMessage(c *gin.Context) {
	senderID := c.GetUint64("id")
	var req model.SendMessageReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, 400, "json 解析出错")
		return
	}

	msgID, err := service.SendMessage(senderID, req.ConversationID, req.Content)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, 201, "success", msgID)
}
