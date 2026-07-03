package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/auth"
	"github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager/internal/response"
)

func AuthRequired(secret, deployKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Error(c, 401, "missing token")
			c.Abort()
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := auth.ParseToken(token, secret, deployKey)
		if err != nil {
			response.Error(c, 401, "invalid token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != "admin" {
			response.Error(c, 403, "admin required")
			c.Abort()
			return
		}
		c.Next()
	}
}
