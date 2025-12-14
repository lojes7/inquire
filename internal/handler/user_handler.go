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

type RegisterRequest struct {
	Uid      string `json:"name" binding:"required,max=64"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
