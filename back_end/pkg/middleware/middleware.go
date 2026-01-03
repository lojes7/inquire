package middleware

import (
	"strings"
	"vvechat/pkg/response"
	"vvechat/pkg/secure"

	"github.com/gin-gonic/gin"
)

// RefreshAuth RefreshToken专属中间件
func RefreshAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从请求头取 Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, 401, "Header错误")
			c.Abort()
			return
		}

		// 2. 检查 Bearer 前缀
		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			response.Fail(c, 401, "Type错误")
			c.Abort()
			return
		}

		// 3. 拿到 token
		tokenString := strings.TrimPrefix(authHeader, prefix)

		// 4. 解析 token
		claims, err := secure.ParseToken(tokenString)
		if err != nil {
			response.Fail(c, 401, err.Error())
			c.Abort()
			return
		}
		//5.验证是否是refresh_token
		if claims.Type != "refresh" {
			response.Fail(c, 403, "不是refresh_token")
			c.Abort()
			return
		}
		// 6. 把信息放进 Context
		c.Set("id", claims.ID)

		// 7. 放行
		c.Next()
	}
}

func JWTAuth() gin.HandlerFunc {
	//goland:noinspection DuplicatedCode
	return func(c *gin.Context) {
		// 1. 从请求头取 Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, 401, "Header错误")
			c.Abort()
			return
		}

		// 2. 检查 Bearer 前缀
		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			response.Fail(c, 401, "Type错误")
			c.Abort()
			return
		}

		// 3. 拿到 token
		tokenString := strings.TrimPrefix(authHeader, prefix)

		// 4. 解析 token
		claims, err := secure.ParseToken(tokenString)
		if err != nil {
			response.Fail(c, 401, err.Error())
			c.Abort()
			return
		}
		//5.验证是否是token
		if claims.Type != "access" {
			response.Fail(c, 403, "不是token")
			c.Abort()
			return
		}
		// 6. 把信息放进 Context
		c.Set("id", claims.ID)

		// 7. 放行
		c.Next()
	}
}
