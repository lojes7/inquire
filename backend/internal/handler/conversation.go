package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/internal/service"
	"github.com/lojes7/inquire/pkg/response"
)

// StartPrivateConversation 新建私聊
// @Summary      发起私聊
// @Description  Create a new private conversation with a friend
// @Tags         conversation
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        req  body      model.IDReq  true  "好友ID"
// @Success      201  {object}  response.Response   "创建成功"
// @Failure      400  {object}  response.Response   "json解析出错"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/conversations/private [post]
func StartPrivateConversation(c *gin.Context) {
	userID := c.GetUint64("id")
	var req model.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "json 解析出错")
		return
	}
	friendID := req.ID

	conversationID, err := service.StartPrivateConversation(userID, friendID)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, 201, "success", conversationID)
}

// ChatHistoryList 加载聊天记录
// @Summary      获取聊天记录
// @Description  Get chat history for a conversation
// @Tags         conversation
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        conversation_id path      string  true  "会话ID"
// @Success      200  {array}   model.ChatHistoryResp "获取成功"
// @Failure      400  {object}  response.Response   "conversation_id参数错误"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/conversations/{conversation_id} [get]
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
// @Summary      获取会话列表
// @Description  Get the list of active conversations
// @Tags         conversation
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Success      200  {array}   model.ConversationListResp "获取成功"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/conversations [get]
func ConversationList(c *gin.Context) {
	userID := c.GetUint64("id")

	resp, err := service.ConversationList(userID)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Success(c, 200, "success", resp)
}
