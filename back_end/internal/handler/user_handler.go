package handler

import (
	"log"
	"net/http"
	"vvechat/internal/model"
	"vvechat/internal/service"

	"vvechat/pkg/judge"
	"vvechat/pkg/response"

	"github.com/gin-gonic/gin"
)

// Register 注册操作
func Register(c *gin.Context) {
	var req model.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "json解析出错")
		return
	}

	user, err := model.NewUser(req.Name, req.Password, req.PhoneNumber)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	err = service.Register(user)
	if err != nil {
		if judge.IsUniqueConflict(err) {
			response.Fail(c, http.StatusBadRequest, "手机号已存在")
		} else {
			response.Fail(c, 500, "数据库错误")
		}
	} else {
		response.Success(c, 201, "注册成功", nil)
	}
}

// LoginByUid 微信号登陆操作
func LoginByUid(c *gin.Context) {
	var req model.LoginByUidReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "json解析出错")
		return
	}

	loginResp, err := service.LoginByUid(req.Uid, req.Password)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
	} else {
		response.Success(c, 200, "登陆成功", loginResp)
	}
}

// LoginByPhone 手机号登陆操作
func LoginByPhone(c *gin.Context) {
	var req model.LoginByPhoneReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "json解析出错")
		return
	}

	loginResp, err := service.LoginByPhone(req.PhoneNumber, req.Password)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
	} else {
		response.Success(c, 200, "登陆成功", loginResp)
	}
}

// ReviseUid 修改微信号
func ReviseUid(c *gin.Context) {
	id := c.GetUint64("id")

	var req model.ReviseUidReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "json解析出错")
		return
	}
	newUid := req.NewUid

	err := service.ReviseUid(id, newUid)
	if err != nil {
		if judge.IsUniqueConflict(err) {
			response.Fail(c, 400, "微信号重复")
		} else {
			response.Fail(c, 500, "数据库错误")
		}
		return
	}

	response.Success(c, 201, "success", nil)
}

// RevisePassword 修改密码
func RevisePassword(c *gin.Context) {
	id := c.GetUint64("id")
	var req model.RevisePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("json解析失败")
		response.Fail(c, 400, "修改失败")
		return
	}

	err := service.RevisePassword(id, req.PrevPassword, req.NewPassword)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}

	response.Success(c, 201, "success", nil)
}

// ReviseName 修改用户名
func ReviseName(c *gin.Context) {
	id := c.GetUint64("id")
	var req model.ReviseNameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("json解析失败")
		response.Fail(c, 400, "修改失败")
		return
	}

	err := service.ReviseName(id, req.NewName)
	if err != nil {
		response.Fail(c, 500, "数据库错误")
		return
	}

	response.Success(c, 201, "success", nil)
}
