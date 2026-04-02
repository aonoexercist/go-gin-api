package rbac

import (
	"fmt"
	"go-gin-api/config"
	"go-gin-api/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateRole(c *gin.Context) {
	var role models.Role

	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Create(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, role)
}

func GetRoles(c *gin.Context) {
	var roles []models.Role

	if err := config.DB.Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, roles)
}

func GetRole(c *gin.Context) {
	var role models.Role
	id := c.Param("id")

	if err := config.DB.First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	c.JSON(http.StatusOK, role)
}

func UpdateRole(c *gin.Context) {
	var role models.Role
	id := c.Param("id")

	if err := config.DB.First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Save(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, role)
}

func DeleteRole(c *gin.Context) {
	var role models.Role
	id := c.Param("id")

	if err := config.DB.First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	if err := config.DB.Delete(&role, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

func UpdateUserRoles(userID uint, roleNames []string) error {
	var user models.User

	// 1. Fetch the user first to ensure they exist
	if err := config.DB.First(&user, userID).Error; err != nil {
		return err
	}

	// 2. Convert role names (strings) into actual Role structs
	var roles []models.Role
	for _, name := range roleNames {
		var r models.Role
		// FirstOrCreate ensures the role exists in the 'roles' table
		// so we have a valid ID for the join table link.
		if err := config.DB.Where("name = ?", name).FirstOrCreate(&r, models.Role{Name: name}).Error; err != nil {
			fmt.Printf("Could not find role: %s\n", name)
			return err
		}
		roles = append(roles, r)
	}

	// 3. Sync the 'user_roles' join table
	// Replace() will:
	// - Add new links for new roles
	// - Delete links for roles not in the slice
	// - Keep existing links untouched
	return config.DB.Model(&user).Association("Roles").Replace(roles)
}

func UpdateUserRolesHandler(c *gin.Context) {
	var req struct {
		UserID    uint     `json:"user_id"`
		RoleNames []string `json:"role_names"`
	}
	log.Printf("user info: %+v\n", req)
	fmt.Printf("Received UpdateUserRoles request: %+v\n", req)

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := UpdateUserRoles(req.UserID, req.RoleNames); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User roles updated successfully"})
}
