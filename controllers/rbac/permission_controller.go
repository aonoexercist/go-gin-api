package rbac

import (
	"go-gin-api/config"
	"go-gin-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreatePermission(c *gin.Context) {
	var permission models.Permission

	if err := c.ShouldBindJSON(&permission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Create(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, permission)
}

func GetPermissions(c *gin.Context) {
	var permissions []models.Permission

	if err := config.DB.Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

func GetPermission(c *gin.Context) {
	var permission models.Permission
	id := c.Param("id")

	if err := config.DB.First(&permission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
		return
	}

	c.JSON(http.StatusOK, permission)
}

func UpdatePermission(c *gin.Context) {
	var permission models.Permission
	id := c.Param("id")

	if err := config.DB.First(&permission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
		return
	}

	if err := c.ShouldBindJSON(&permission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Save(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permission)
}

func DeletePermission(c *gin.Context) {
	var permission models.Permission
	id := c.Param("id")

	if err := config.DB.First(&permission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
		return
	}

	if err := config.DB.Delete(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func GetPermissionsByRole(c *gin.Context) {
	var role models.Role
	id := c.Param("id")

	if err := config.DB.Preload("Permissions").First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	c.JSON(http.StatusOK, role.Permissions)
}

func SaveRoleWithPermissions(roleName string, permNames []string) error {
	var role models.Role

	// 1. Find or create role
	err := config.DB.Where("name = ?", roleName).First(&role).Error
	if err != nil {
		// Role not found → create
		role = models.Role{Name: roleName}
		if err := config.DB.Create(&role).Error; err != nil {
			return err
		}
	}

	// 2. Get permissions from DB
	var permissions []models.Permission
	// if err := config.DB.Where("name IN ?", permNames).Find(&permissions).Error; err != nil {
	// 	return err
	// }
	for _, name := range permNames {
		var p models.Permission
		if err := config.DB.Where(models.Permission{Name: name}).FirstOrCreate(&p).Error; err != nil {
			return err
		}
		permissions = append(permissions, p)
	}

	// 3. Replace permissions (🔥 works for both create/update)
	if err := config.DB.Model(&role).Association("Permissions").Replace(permissions); err != nil {
		return err
	}

	return nil
}

type SaveRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	Permissions []string `json:"permissions" binding:"required"`
}

func SaveRole(c *gin.Context) {
	var req SaveRoleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := SaveRoleWithPermissions(req.Name, req.Permissions)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "role saved successfully",
	})
}
