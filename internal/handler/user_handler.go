package handler

import (
	"net/http"
	"wechat/internal/model"
	"wechat/internal/service"
	"wechat/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	serve *service.UserService
}

func NewUserHandler(serve *service.UserService) *UserHandler {
	return &UserHandler{serve}
}

type registerReq struct {
	uid string
	pwd string
}

func (h *UserHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(req); err != nil {
		response.Fail(c, http.StatusBadRequest, "请求出错")
		return
	}

	user := model.NewUser(req.uid, req.pwd)
	err := h.serve.Register(user)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
	} else {
		response.Success(c, nil)
	}
}
