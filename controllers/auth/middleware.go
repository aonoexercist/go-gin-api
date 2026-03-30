package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie("access_token")
		if err != nil {
			c.AbortWithStatus(401)
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatus(401)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// "user_id" must match the key you used when creating the token
			c.Set("user_id", claims["user_id"])
		}

		c.Next()
	}
}
