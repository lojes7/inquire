package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/internal/service"
	"github.com/lojes7/inquire/pkg/response"
	"github.com/lojes7/inquire/pkg/secure"
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
		if myErr := secure.Unwrap(err); myErr != nil {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}

	resp := model.IDResp{
		ID: conversationID,
	}

	response.Success(c, 201, "success", resp)
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
		if myErr := secure.Unwrap(err); myErr != nil {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "服务器错误")
		}
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
		if myErr := secure.Unwrap(err); myErr != nil {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}
	response.Success(c, 200, "success", resp)
}

// CreateGroup 创建群聊
// @Summary      创建群聊
// @Description  Create a new group conversation
// @Tags         conversation
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        req  body      model.CreateGroupReq  true  "群成员ID列表"
// @Success      201  {object}  response.Response   "创建成功"
// @Failure      400  {object}  response.Response   "json解析出错"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/conversations/group [post]
func CreateGroup(c *gin.Context) {
	userID := c.GetUint64("id")
	var req model.CreateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "json 解析出错")
		return
	}

	var memberIDsUint []uint64
	for _, idStr := range req.MemberIDs {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			response.Fail(c, 400, "成员ID格式错误")
			return
		}
		memberIDsUint = append(memberIDsUint, id)
	}

	// 必须包含群主自己
	memberIDsUint = append(memberIDsUint, userID)
	// 去重
	uniqueIDs := make(map[uint64]bool)
	// 最终的成员ID列表，保持原有顺序但去重
	finalIDs := make([]uint64, 0)
	for _, id := range memberIDsUint {
		if !uniqueIDs[id] {
			uniqueIDs[id] = true
			finalIDs = append(finalIDs, id)
		}
	}

	conversationID, err := service.CreateGroupConversation(userID, req.GroupName, finalIDs)
	if err != nil {
		if myErr := secure.Unwrap(err); myErr != nil {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}
	response.Success(c, 201, "success", conversationID)
}

// BanUser 禁言用户
// @Summary      禁言用户
// @Description  Ban a user from a conversation (Private block or Group mute)
// @Tags         conversation
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        conversation_id path string true "会话ID"
// @Param        req  body      model.IDReq  true  "目标用户ID"
// @Success      200  {object}  response.Response   "操作成功"
// @Failure      400  {object}  response.Response   "参数错误"
// @Failure      403  {object}  response.Response   "无权限"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/conversations/{conversation_id}/ban [post]
func BanUser(c *gin.Context) {
	userID := c.GetUint64("id")
	idStr := c.Param("conversation_id")
	conversationID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Fail(c, 400, "会话ID格式错误")
		return
	}

	var req model.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "json 解析出错")
		return
	}

	err = service.BanUser(userID, conversationID, req.ID)
	if err != nil {
		if myErr := secure.Unwrap(err); myErr != nil {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}
	response.Success(c, 200, "success", nil)
}

// KickUser 踢出用户
// @Summary      踢出用户
// @Description  Kick a user from a group conversation
// @Tags         conversation
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        conversation_id path string true "会话ID"
// @Param        req  body      model.IDReq  true  "目标用户ID"
// @Success      200  {object}  response.Response   "操作成功"
// @Failure      400  {object}  response.Response   "参数错误"
// @Failure      403  {object}  response.Response   "无权限"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/conversations/{conversation_id}/kick [post]
func KickUser(c *gin.Context) {
	userID := c.GetUint64("id")
	idStr := c.Param("conversation_id")
	conversationID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Fail(c, 400, "会话ID格式错误")
		return
	}

	var req model.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "json 解析出错")
		return
	}

	err = service.KickUser(userID, conversationID, req.ID)
	if err != nil {
		if myErr := secure.Unwrap(err); myErr != nil {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}
	response.Success(c, 200, "success", nil)
}

// LeaveGroup 退出群聊
// @Summary      退出群聊
// @Description  Leave a group conversation
// @Tags         conversation
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        conversation_id path string true "会话ID"
// @Success      200  {object}  response.Response   "操作成功"
// @Failure      400  {object}  response.Response   "参数错误"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/conversations/{conversation_id}/leave [post]
func LeaveGroup(c *gin.Context) {
	userID := c.GetUint64("id")
	idStr := c.Param("conversation_id")
	conversationID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Fail(c, 400, "会话ID格式错误")
		return
	}

	err = service.LeaveGroup(userID, conversationID)
	if err != nil {
		if myErr := secure.Unwrap(err); myErr != nil {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}
	response.Success(c, 200, "success", nil)
}
