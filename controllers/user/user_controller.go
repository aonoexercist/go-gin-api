package user

import (
	"go-gin-api/config"
	dto "go-gin-api/controllers"
	"go-gin-api/models"
	"net/http"
	"strconv"

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

// UpdateUser godoc
// @Summary      Update a user's name and email
// @Description  Update basic profile fields for a user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      int                      true  "User ID"
// @Param        user  body      dto.UpdateUserRequestDTO true  "Updated user fields"
// @Success      200   {object}  dto.UserResponseDTO
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /admin/users/{id} [put]
func UpdateUser(c *gin.Context) {
	var user models.User
	id := c.Param("id")

	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var body dto.UpdateUserRequestDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Optional: check email isn't already taken by another user
	if body.Email != user.Email {
		var existing models.User
		if err := config.DB.Where("email = ? AND id != ?", body.Email, user.ID).First(&existing).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already in use"})
			return
		}
	}

	user.Name = body.Name
	user.Email = body.Email

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reload with roles+permissions so the response DTO stays consistent with GetUsers/UpdateUserRoles
	if err := config.DB.Preload("Roles.Permissions").First(&user, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToUserDTO(user))
}

// UpdateUserRoles godoc
// @Summary      Update a user's roles
// @Description  Replace the roles assigned to a user by role IDs
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id      path      int                        true  "User ID"
// @Param        roleIds body      dto.AssignRolesRequestDTO true  "List of role IDs to assign"
// @Success      200     {object}  dto.UserResponseDTO
// @Failure      400     {object}  map[string]string
// @Failure      404     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /admin/users/{id}/roles [put]
func UpdateUserRoles(c *gin.Context) {
	var user models.User
	id := c.Param("id")

	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var body dto.AssignRolesRequestDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Look up the actual Role records for the given IDs
	var roles []models.Role
	if len(body.RoleIDs) > 0 {
		if err := config.DB.Where("id IN ?", body.RoleIDs).Find(&roles).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Optional: guard against typos/nonexistent IDs by making sure
		// every requested ID actually resolved to a role
		if len(roles) != len(body.RoleIDs) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "One or more role IDs are invalid"})
			return
		}
	}

	if err := config.DB.Model(&user).Association("Roles").Replace(&roles); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reload user with roles+permissions preloaded so the response DTO is accurate
	if err := config.DB.Preload("Roles.Permissions").First(&user, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToUserDTO(user))
}

// DeleteRoleFromUser godoc
// @Summary Remove role from user
// @Description Detaches a specific role from a user using User ID and Role ID.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "User ID"
// @Param role_id path int true "Role ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/users/{user_id}/roles/{role_id} [delete]
func DeleteRoleFromUser(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	roleId, err := strconv.Atoi(c.Param("role_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	// 1. Verify User exists
	var user models.User
	if err := config.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// 2. Verify Role exists
	var role models.Role
	if err := config.DB.First(&role, roleId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	// 3. Remove association in the many-to-many join table
	if err := config.DB.Model(&user).Association("Roles").Delete(&role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role removed from user successfully"})
}
