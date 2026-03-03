package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lojes7/inquire/internal/service"
	"github.com/lojes7/inquire/pkg/response"
	"github.com/lojes7/inquire/pkg/secure"
)

// RefreshToken 刷新Token
// @Summary      刷新Token
// @Description  使用 Refresh Token 获取新的 Access Token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Success      201  {object}  response.Response{data=model.TokenResp} "刷新成功"
// @Failure      500  {object}  response.Response   "Token问题"
// @Router       /auth/refresh_token [post]
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
// @Summary      根据ID查看好友信息
// @Description  根据用户ID获取好友详细信息
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        id   path      int     true  "好友用户ID"
// @Success      200  {object}  response.Response{data=model.FriendInfoResp} "获取成功"
// @Failure      400  {object}  response.Response   "ID错误或找不到好友"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/info/friends/id/{id} [get]
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
		var myErr *secure.MyError

		if errors.As(err, &myErr) {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "转换为 myErr 时错误")
		}
		return
	}
	response.Success(c, 200, "success", resp)
}

// StrangerInfoByID 查看陌生人信息
// @Summary      根据ID查看陌生人信息
// @Description  根据用户ID获取陌生人信息
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        id   path      int     true  "陌生人用户ID"
// @Success      200  {object}  response.Response{data=model.StrangerInfoResp} "获取成功"
// @Failure      400  {object}  response.Response   "ID错误或找不到此人"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/info/strangers/id/{id} [get]
func StrangerInfoByID(c *gin.Context) {
	id := c.Param("id")
	strangerID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		response.Fail(c, 400, "陌生人 ID 错误")
		return
	}

	resp, err := service.StrangerInfoByID(strangerID)
	if err != nil {
		var myErr *secure.MyError

		if errors.As(err, &myErr) {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "转换为 myErr 时错误")
		}
		return
	}
	response.Success(c, 200, "success", resp)
}

// FriendInfoByUid 查看好友信息 通过Uid
// @Summary      根据Uid查看好友信息
// @Description  根据用户Uid获取好友详细信息
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        uid  path      string  true  "好友Uid"
// @Success      200  {object}  response.Response{data=model.FriendInfoResp} "获取成功"
// @Failure      400  {object}  response.Response   "Uid为空或找不到好友"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/info/friends/uid/{uid} [get]
func FriendInfoByUid(c *gin.Context) {
	userID := c.GetUint64("id")
	uid := c.Param("uid")
	if uid == "" {
		response.Fail(c, 400, "Uid 不能为空")
		return
	}

	resp, err := service.FriendInfoByUid(userID, uid)
	if err != nil {
		var myErr *secure.MyError

		if errors.As(err, &myErr) {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "转换为 myErr 时错误")
		}
		return
	}
	response.Success(c, 200, "success", resp)
}

// StrangerInfoByUid 查看陌生人信息通过Uid
// @Summary      根据Uid查看陌生人信息
// @Description  根据用户Uid获取陌生人信息
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        uid  path      string  true  "陌生人Uid"
// @Success      200  {object}  response.Response{data=model.StrangerInfoResp} "获取成功"
// @Failure      400  {object}  response.Response   "Uid为空或找不到此人"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/info/strangers/uid/{uid} [get]
func StrangerInfoByUid(c *gin.Context) {
	uid := c.Param("uid")
	if uid == "" {
		response.Fail(c, 400, "Uid 不能为空")
		return
	}

	resp, err := service.StrangerInfoByUid(uid)
	if err != nil {
		var myErr *secure.MyError

		if errors.As(err, &myErr) {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "转换为 myErr 时错误")
		}
		return
	}
	response.Success(c, 200, "success", resp)
}
