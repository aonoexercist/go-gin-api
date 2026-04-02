package rbac

import (
	"go-gin-api/config"
	"go-gin-api/models"

	"github.com/gin-gonic/gin"
)

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}

		userID := userIDVal.(uint)

		var user models.User
		if err := config.DB.Preload("Roles").First(&user, userID).Error; err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "user_not_found"})
			return
		}

		for _, role := range user.Roles {
			for _, allowed := range allowedRoles {
				if role.Name == allowed {
					c.Next()
					return
				}
			}
		}

		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
	}
}
