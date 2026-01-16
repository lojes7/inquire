package handler

import (
	"strconv"
	"vvechat/internal/model"
	"vvechat/internal/service"
	"vvechat/pkg/response"

	"github.com/gin-gonic/gin"
)

// CreatePrivateConversation 新建私聊
func CreatePrivateConversation(c *gin.Context) {
	userID := c.GetUint64("id")
	var req model.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "json 解析出错")
		return
	}

	err := service.CreatePrivateConversation(userID, req.ID)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, 201, "success", nil)
}

// ChatHistoryList 加载聊天记录
func ChatHistoryList(c *gin.Context) {
	userID := c.GetUint64("id")
	id := c.Param("conversation_id")
	conversationID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		response.Fail(c, 400, "conversation_id参数错误")
		return
	}

	resp, err := service.ChatHistoryList(userID, conversationID)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, 200, "success", resp)
}

// ConversationList 会话列表
func ConversationList(c *gin.Context) {
	userID := c.GetUint64("id")

	resp, err := service.ConversationList(userID)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, 200, "success", resp)
}
