package handler

import (
	"net/http"
	"vvechat/internal/model"
	"vvechat/internal/service"
	"vvechat/pkg/response"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	var req RegisterRequest
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
		response.Fail(c, http.StatusBadRequest, err.Error())
	} else {
		response.Success(c, "注册成功", nil)
	}
}

func LoginByUid(c *gin.Context) {
	var req LoginByUidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "json解析出错")
		return
	}

	err := service.LoginByUid(req.Uid, req.Password)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
	} else {
		response.Success(c, "登陆成功", nil)
	}
}

func LoginByPhone(c *gin.Context) {
	var req LoginByPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "json解析出错")
		return
	}

	err := service.LoginByPhone(req.PhoneNumber, req.Password)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
	} else {
		response.Success(c, "登陆成功", nil)
	}
}
