package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Launch(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	return r
}
