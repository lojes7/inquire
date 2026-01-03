package handler

import (
	"vvechat/internal/model"
	"vvechat/internal/service"
	"vvechat/pkg/judge"
	"vvechat/pkg/response"

	"github.com/gin-gonic/gin"
)

func RefreshToken(c *gin.Context) {
	id := c.GetUint64("id")

	resp, err := service.RefreshToken(id)
	if err != nil {
		response.Fail(c, 500, "token出现问题"+err.Error())
		return
	}

	response.Success(c, 201, "success", resp)
}

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

func FriendInfo(c *gin.Context) {

}

func StrangerInfo(c *gin.Context) {

}
