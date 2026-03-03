package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/internal/service"
	"github.com/lojes7/inquire/pkg/judge"
	"github.com/lojes7/inquire/pkg/response"
	"gorm.io/gorm"
)

// SendFriendRequest 发送好友申请操作
// @Summary      发送好友申请
// @Description  发送好友申请给指定用户
// @Tags         friendship_request
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        req  body      model.AddFriendReq  true  "添加好友请求体"
// @Success      201  {object}  response.Response   "发送成功"
// @Failure      400  {object}  response.Response   "发送失败"
// @Failure      409  {object}  response.Response   "请勿重复发送"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/friendship_requests [post]
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
// @Summary      获取好友申请列表
// @Description  获取当前收到的好友申请列表
// @Tags         friendship_request
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Success      200  {array}   model.FriendRequestListResp "获取成功"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/friendship_requests [get]
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
// @Summary      接受好友申请
// @Description  接受指定的好友申请
// @Tags         friendship_request
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        request_id path      string  true  "申请ID"
// @Success      201  {object}  response.Response   "接受成功"
// @Failure      400  {object}  response.Response   "requestID错误"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/friendship_requests/{request_id} [post]
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
// @Summary      删除好友申请
// @Description  删除或拒绝指定的好友申请
// @Tags         friendship_request
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        request_id path      string  true  "申请ID"
// @Success      201  {object}  response.Response   "删除成功"
// @Failure      400  {object}  response.Response   "requestID错误"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/friendship_requests/{request_id} [delete]
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
