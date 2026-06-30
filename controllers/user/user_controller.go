package user

import (
	"go-gin-api/config"
	dto "go-gin-api/controllers"
	"go-gin-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetUsers godoc
// @Summary      List all users
// @Description  Get all users along with their roles and permissions
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {array}   dto.UserResponseDTO
// @Failure      500  {object}  map[string]string
// @Router       /admin/users [get]
func GetUsers(c *gin.Context) {
	var users []models.User

	if err := config.DB.Preload("Roles.Permissions").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. Initialize the DTO slice with the same length for performance
	userDTOs := make([]dto.UserResponseDTO, len(users))

	// 3. Loop and convert
	for i, user := range users {
		userDTOs[i] = dto.ToUserDTO(user)
	}

	c.JSON(http.StatusOK, userDTOs)
}

// GetUser godoc
// @Summary      Get a single user
// @Description  Get a user by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  dto.UserResponseDTO
// @Failure      404  {object}  map[string]string
// @Router       /admin/users/{id} [get]
func GetUser(c *gin.Context) {
	var user models.User
	id := c.Param("id")

	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, dto.ToUserDTO(user))
}

// UpdateUserRoles godoc
// @Summary      Update a user's roles
// @Description  Replace the roles assigned to a user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id     path      int           true  "User ID"
// @Param        roles  body      []models.Role true  "List of roles to assign"
// @Success      200    {object}  dto.UserResponseDTO
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /admin/users/{id}/roles [put]
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

	c.JSON(http.StatusOK, dto.ToUserDTO(user))
}
