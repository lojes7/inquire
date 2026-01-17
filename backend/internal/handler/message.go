package handler

import (
	"strconv"
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
	conversationID := req.ConversationID
	content := req.Content

	msgID, err := service.SendText(senderID, conversationID, content)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, 201, "success", msgID)
}

func SendFile(c *gin.Context) {
	userID := c.GetUint64("id")

	conversationIDStr := c.PostForm("conversation_id")
	if conversationIDStr == "" {
		response.Fail(c, 400, "conversation_id 是空的")
		return
	}
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 64)
	if err != nil {
		response.Fail(c, 400, "conversation_id 格式错误")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, 400, "没有接收到文件")
		return
	}

	resp, err := service.SendFile(userID, conversationID, file)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, 201, "success", resp)
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
