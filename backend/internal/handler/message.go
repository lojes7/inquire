package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/internal/service"
	"github.com/lojes7/inquire/pkg/response"
	"github.com/lojes7/inquire/pkg/secure"
)

// SendText 发送文本消息
// @Summary      发送文本消息
// @Description  Send a text message to a conversation
// @Tags         message
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        req  body      model.SendTextReq  true  "发送文本消息请求体"
// @Success      201  {object}  response.Response   "发送成功"
// @Failure      400  {object}  response.Response   "json解析出错"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/messages/text [post]
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
		if myErr := secure.Unwrap(err); myErr != nil {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}
	response.Success(c, 201, "success", msgID)
}

// SendFile 发送文件
// @Summary      发送文件
// @Description  Send a file to a conversation
// @Tags         message
// @Accept       multipart/form-data
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        conversation_id formData string true "会话ID"
// @Param        file formData file true "文件"
// @Success      201  {object}  response.Response   "发送成功"
// @Failure      400  {object}  response.Response   "参数错误"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/messages/file [post]
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

	resp, err := service.SendFile(c.Request.Context(), userID, conversationID, file)
	if err != nil {
		if myErr := secure.Unwrap(err); myErr != nil {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}
	response.Success(c, 201, "success", resp)
}

// RecallMessage 撤回消息
// @Summary      撤回消息
// @Description  Recall a sent message
// @Tags         message
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        req  body      model.IDReq  true  "消息ID"
// @Success      201  {object}  response.Response   "撤回成功"
// @Failure      400  {object}  response.Response   "json解析出错"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/messages/recall [delete]
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
		if myErr := secure.Unwrap(err); myErr != nil {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}
	response.Success(c, 201, "success", systemMsgID)
}

// DeleteMessage 删除消息
// @Summary      删除消息
// @Description  Delete a message
// @Tags         message
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        req  body      model.IDReq  true  "消息ID"
// @Success      200  {object}  response.Response   "删除成功"
// @Failure      400  {object}  response.Response   "json解析出错"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/messages/delete [delete]
func DeleteMessage(c *gin.Context) {
	userID := c.GetUint64("id")
	var req model.IDReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, 400, "json 解析错误")
		return
	}

	err := service.DeleteMessage(userID, req.ID)
	if err != nil {
		if myErr := secure.Unwrap(err); myErr != nil {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}
	response.Success(c, 201, "success", nil)
}

// SendWorkspaceFile 从工作区发送文件
// @Summary      从工作区发送文件
// @Description  从用户工作区中选择已有文件发送到指定会话
// @Tags         message
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        req  body      model.SendWorkspaceFileReq  true  "发送工作区文件请求体"
// @Success      201  {object}  response.Response   "发送成功"
// @Failure      400  {object}  response.Response   "json解析出错"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/messages/workspace-file [post]
func SendWorkspaceFile(c *gin.Context) {
	userID := c.GetUint64("id")

	var req model.SendWorkspaceFileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "json 解析出错")
		return
	}

	resp, err := service.SendWorkspaceFile(userID, req.ConversationID, req.FileID)
	if err != nil {
		if myErr := secure.Unwrap(err); myErr != nil {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}
	response.Success(c, 201, "success", resp)
}
