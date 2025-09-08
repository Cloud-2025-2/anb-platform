package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(strings.ToLower(h), "bearer ") {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimSpace(h[7:])
		t, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !t.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if claims, ok := t.Claims.(jwt.MapClaims); ok {
			if sub, ok2 := claims["sub"].(string); ok2 {
				c.Set("user_id", sub)
				c.Next()
				return
			}
		}
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}
