package handler

import (
	"errors"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/internal/service"
	"github.com/lojes7/inquire/pkg/response"
	"gorm.io/gorm"
)

// FriendshipList 获取好友列表
// @Summary      获取好友列表
// @Description  获取当前用户的好友列表
// @Tags         friendship
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Success      200  {array}   model.FriendshipListResp "获取成功"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/friendships [get]
func FriendshipList(c *gin.Context) {
	id := c.GetUint64("id")
	resp, err := service.FriendshipList(id)

	if err != nil {
		response.Fail(c, 500, "服务器错误")
		return
	}

	response.Success(c, 200, "success", resp)
}

// DeleteFriendship 删除好友
// @Summary      删除好友
// @Description  根据好友ID删除好友
// @Tags         friendship
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        friend_id path      string  true  "好友ID"
// @Success      201  {object}  response.Response   "删除成功"
// @Failure      400  {object}  response.Response   "ID错误或好友不存在"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/friendships/{friend_id} [delete]
func DeleteFriendship(c *gin.Context) {
	userID := c.GetUint64("id")
	friendId := c.Param("friend_id")

	friendID, err := strconv.ParseUint(friendId, 10, 64)
	if err != nil {
		response.Fail(c, 400, "friend_id不合法")
		return
	}

	err = service.DeleteFriendship(userID, friendID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, 400, "好友不存在")
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}

	response.Success(c, 201, "success", nil)
}

// ReviseRemark 修改好友备注
// @Summary      修改好友备注
// @Description  修改好友的备注信息
// @Tags         friendship
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        friend_id path      string  true  "好友ID"
// @Param        req  body      model.RemarkReq  true  "修改备注请求体"
// @Success      200  {object}  response.Response   "修改成功"
// @Failure      400  {object}  response.Response   "ID错误或输入不合法"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/friendships/remark/{friend_id} [post]
func ReviseRemark(c *gin.Context) {
	userID := c.GetUint64("id")
	id := c.Param("friend_id")
	friendID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		log.Println("修改备注时，好友id转为uint64类型时错误")
		response.Fail(c, 400, "修改失败")
		return
	}

	var req model.RemarkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("json解析错误")
		response.Fail(c, 400, "输入不合法")
		return
	}

	err = service.ReviseRemark(userID, friendID, req.Remark)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	response.Success(c, 200, "success", nil)
}
