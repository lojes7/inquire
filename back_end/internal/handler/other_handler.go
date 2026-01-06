package handler

import (
	"strconv"
	"vvechat/internal/service"
	"vvechat/pkg/judge"
	"vvechat/pkg/response"

	"github.com/gin-gonic/gin"
)

// RefreshToken 刷新Token
func RefreshToken(c *gin.Context) {
	id := c.GetUint64("id")

	resp, err := service.RefreshToken(id)
	if err != nil {
		response.Fail(c, 500, "token出现问题"+err.Error())
		return
	}

	response.Success(c, 201, "success", resp)
}

// FriendInfoByID 查看好友信息
func FriendInfoByID(c *gin.Context) {
	userID := c.GetUint64("id")
	id := c.Param("id")
	friendID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		response.Fail(c, 400, "好友 ID 错误")
		return
	}

	resp, err := service.FriendInfoByID(userID, friendID)
	if err != nil {
		if judge.IsUniqueConflict(err) {
			response.Fail(c, 400, "找不到好友")
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}
	response.Success(c, 200, "success", resp)
}

// StrangerInfoByID 查看陌生人信息
func StrangerInfoByID(c *gin.Context) {
	id := c.Param("id")
	strangerID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		response.Fail(c, 400, "陌生人 ID 错误")
		return
	}

	resp, err := service.StrangerInfoByID(strangerID)
	if err != nil {
		if judge.IsUniqueConflict(err) {
			response.Fail(c, 400, "找不到此人")
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}
	response.Success(c, 200, "success", resp)
}

// FriendInfoByUid 查看好友信息 通过Uid
func FriendInfoByUid(c *gin.Context) {
	userID := c.GetUint64("id")
	uid := c.Param("uid")
	if uid == "" {
		response.Fail(c, 400, "Uid 不能为空")
		return
	}

	resp, err := service.FriendInfoByUid(userID, uid)
	if err != nil {
		if judge.IsUniqueConflict(err) {
			response.Fail(c, 400, "找不到好友")
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}
	response.Success(c, 200, "success", resp)
}

// StrangerInfoByUid 查看陌生人信息通过Uid
func StrangerInfoByUid(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		response.Fail(c, 400, "Uid 不能为空")
		return
	}

	resp, err := service.StrangerInfoByUid(uid)
	if err != nil {
		if judge.IsUniqueConflict(err) {
			response.Fail(c, 400, "找不到此人")
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}
	response.Success(c, 200, "success", resp)
}
