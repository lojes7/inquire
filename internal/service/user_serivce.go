package service

import (
	"net/http"
	"wechat/internal/class"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Serve struct {
	DB *gorm.DB
}

func (s *Serve) RegisterUser(c *gin.Context) {
	uid := c.Query("uid")
	pwd := c.Query("password")
	re_pwd := c.Query("re_password")

	if pwd != re_pwd {
		c.String(http.StatusBadRequest, "两次密码输入不一致！")
		return
	}
	user := class.NewUser(uid, pwd)
	tx := class.AddUserToDB(user, s.DB)
	if tx.Error != nil {
		c.String(http.StatusConflict, "微信号重复！")
		return
	}
	c.String(http.StatusOK, "注册成功！")
}
