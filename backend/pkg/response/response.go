package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func Success(c *gin.Context, code int, msg string, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code":    code,
		"message": msg,
		"data":    data,
	})
}

func Fail(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{
		"code":    code,
		"message": msg,
	})
}
