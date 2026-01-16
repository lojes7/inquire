package handler

import (
	"vvechat/internal/model"
	"vvechat/internal/service"
	"vvechat/pkg/response"

	"github.com/gin-gonic/gin"
)

func SendText(c *gin.Context) {
	senderID := c.GetUint64("id")
	var req model.SendTextReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, 400, "json 解析出错")
		return
	}

	msgID, err := service.SendText(senderID, req.ConversationID, req.Content)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, 201, "success", msgID)
}

func RecallMessage(c *gin.Context) {
	userID := c.GetUint64("id")
	var req model.IDReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, 400, "json 解析错误")
		return
	}
	// 撤回消息会创建一个系统级消息，这里拿到该消息的ID，返回给前端
	systemMsgID, err := service.RecallMessage(userID, req.ID)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, 201, "success", systemMsgID)
}

func DeleteMessage(c *gin.Context) {
	userID := c.GetUint64("id")
	var req model.IDReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, 400, "json 解析错误")
		return
	}

	err := service.DeleteMessage(userID, req.ID)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, 200, "success", nil)
}
