package handler

import (
	"net/http"
	"vvechat/internal/model"
	"vvechat/internal/service"
	"vvechat/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	serve *service.UserService
}

func NewUserHandler(serve *service.UserService) *UserHandler {
	return &UserHandler{serve}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req RequestBasicInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "json解析出错")
		return
	}

	user, err := model.NewUser(req.Name, req.Password, req.PhoneNumber)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
	}

	err = h.serve.Register(user)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
	} else {
		response.Success(c, nil)
	}
}

func (h *UserHandler) LoginByUid(c *gin.Context) {
	var req RequestBasicInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "json解析出错")
		return
	}

	err := h.serve.LoginByUid(req.Uid, req.Password)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
	} else {
		response.Success(c, nil)
	}
}
