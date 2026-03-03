package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/internal/service"
	"github.com/lojes7/inquire/pkg/response"
	"github.com/lojes7/inquire/pkg/secure"
)

// Register 注册操作
// @Summary      用户注册
// @Description  用户注册接口
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        req  body      model.RegisterReq  true  "注册请求体"
// @Success      201  {object}  response.Response   "注册成功"
// @Failure      400  {object}  response.Response   "请求参数错误或手机号重复"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /register [post]
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
		var myErr *secure.MyError

		if errors.As(err, &myErr) {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "转换为 myErr 时错误")
		}

		return
	}
	response.Success(c, 201, "注册成功", nil)

}

// LoginByUid 微信号登陆操作
// @Summary      微信号登录
// @Description  使用微信号和密码登录
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        req  body      model.LoginByUidReq  true  "微信号登录请求体"
// @Success      200  {object}  response.Response{data=model.LoginResp} "登录成功"
// @Failure      400  {object}  response.Response   "请求参数错误"
// @Router       /login/uid [post]
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
// @Summary      手机号登录
// @Description  使用手机号和密码登录
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        req  body      model.LoginByPhoneReq  true  "手机号登录请求体"
// @Success      200  {object}  response.Response{data=model.LoginResp} "登录成功"
// @Failure      400  {object}  response.Response   "请求参数错误"
// @Router       /login/phone_number [post]
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
// @Summary      修改微信号
// @Description  修改当前用户的微信号
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        req  body      model.UidReq  true  "修改微信号请求体"
// @Success      201  {object}  response.Response   "修改成功"
// @Failure      400  {object}  response.Response   "请求参数错误或微信号重复"
// @Router       /auth/me/uid [post]
func ReviseUid(c *gin.Context) {
	id := c.GetUint64("id")

	var req model.UidReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "json解析出错")
		return
	}
	newUid := req.Uid

	err := service.ReviseUid(id, newUid)
	if err != nil {
		var myErr *secure.MyError

		if errors.As(err, &myErr) {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "转换为 myErr 时错误")
		}
		return
	}

	response.Success(c, 201, "success", nil)
}

// RevisePassword 修改密码
// @Summary      修改密码
// @Description  修改当前用户的密码
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        req  body      model.RevisePasswordReq  true  "修改密码请求体"
// @Success      201  {object}  response.Response   "修改成功"
// @Failure      400  {object}  response.Response   "请求参数错误"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/me/password [post]
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
// @Summary      修改用户名
// @Description  修改当前用户的用户名
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        req  body      model.NameReq  true  "修改用户名请求体"
// @Success      201  {object}  response.Response   "修改成功"
// @Failure      400  {object}  response.Response   "请求参数错误"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/me/name [post]
func ReviseName(c *gin.Context) {
	id := c.GetUint64("id")
	var req model.NameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("json解析失败")
		response.Fail(c, 400, "修改失败")
		return
	}

	err := service.ReviseName(id, req.Name)
	if err != nil {
		response.Fail(c, 500, "数据库错误")
		return
	}

	response.Success(c, 201, "success", nil)
}
