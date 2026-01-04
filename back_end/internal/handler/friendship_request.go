package handler

import (
	"errors"
	"strconv"
	"vvechat/internal/model"
	"vvechat/internal/service"
	"vvechat/pkg/judge"
	"vvechat/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SendFriendRequest 发送好友申请操作
func SendFriendRequest(c *gin.Context) {
	senderID := c.GetUint64("id")
	var req model.AddFriendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "json解析出错")
		return
	}

	err := service.SendFriendRequest(senderID, req.ReceiverID, req.VerificationMessage, req.SenderName)
	if err != nil {
		if errors.Is(err, gorm.ErrInvalidData) {
			response.Fail(c, 400, "发送失败")
		} else if judge.IsUniqueConflict(err) {
			response.Fail(c, 409, "发送失败，请勿重复发送")
		} else {
			response.Fail(c, 500, "服务器出错")
		}
		return
	}

	response.Success(c, 201, "发送成功", nil)
}

// FriendRequestList 加载好友申请列表操作
func FriendRequestList(c *gin.Context) {
	receiverID := c.GetUint64("id")

	respSlice, err := service.FriendRequestList(receiverID)
	if err != nil {
		response.Fail(c, 500, "服务器错误")
		return
	}

	response.Success(c, 200, "success", respSlice)
}

// FriendRequestAccept 通过好友申请操作
func FriendRequestAccept(c *gin.Context) {
	requestID := c.Param("request_id")
	id, err := strconv.ParseUint(requestID, 10, 64)
	if err != nil {
		response.Fail(c, 400, "requestID错误")
		return
	}

	err = service.FriendRequestAccept(id)
	if err != nil {
		response.Fail(c, 500, "服务器错误")
		return
	}

	response.Success(c, 201, "success", nil)
}

// FriendRequestDelete 删除好友申请
func FriendRequestDelete(c *gin.Context) {
	requestID := c.Param("request_id")
	id, err := strconv.ParseUint(requestID, 10, 64)
	if err != nil {
		response.Fail(c, 400, "requestID错误")
		return
	}

	err = service.FriendRequestDelete(id)
	if err != nil {
		response.Fail(c, 500, "服务器错误")
		return
	}

	response.Success(c, 201, "success", nil)
}
