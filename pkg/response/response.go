package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, msg string, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
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
