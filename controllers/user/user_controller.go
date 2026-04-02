package user

import (
	"go-gin-api/config"
	"go-gin-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	var users []models.User

	if err := config.DB.Preload("Roles.Permissions").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. Initialize the DTO slice with the same length for performance
	userDTOs := make([]UserResponseDTO, len(users))

	// 3. Loop and convert
	for i, user := range users {
		userDTOs[i] = ToUserDTO(user)
	}

	c.JSON(http.StatusOK, userDTOs)
}

func GetUser(c *gin.Context) {
	var user models.User
	id := c.Param("id")

	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, ToUserDTO(user))
}

func UpdateUserRoles(c *gin.Context) {
	var user models.User
	id := c.Param("id")

	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var roles []models.Role
	if err := c.ShouldBindJSON(&roles); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Model(&user).Association("Roles").Replace(&roles); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ToUserDTO(user))
}
