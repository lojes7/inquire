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

type registerRequest struct {
	Uid      string `json:"name" binding:"required,max=64"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(req); err != nil {
		response.Fail(c, http.StatusBadRequest, "json解析出错")
		return
	}

	user := model.NewUser(req.Uid, req.Password)
	err := h.serve.Register(user)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
	} else {
		response.Success(c, nil)
	}
}
