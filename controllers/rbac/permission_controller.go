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

// GetPermissionsByRole godoc
// @Summary      Get permissions for a role
// @Description  Get all permissions assigned to a specific role
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Role ID"
// @Success      200  {array}   models.Permission
// @Failure      404  {object}  map[string]string
// @Router       /admin/permissions/roles/{id}/permissions [get]
func GetPermissionsByRole(c *gin.Context) {
	var role models.Role
	id := c.Param("id")

	if err := config.DB.Preload("Permissions").First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	c.JSON(http.StatusOK, role.Permissions)
}

// SaveRoleWithPermissions finds or creates a role by name, finds or creates
// each named permission, and replaces the role's permission associations
// with that set. Not a route handler — internal helper used by SaveRole.
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

// SaveRoleRequest is the request body for SaveRole.
type SaveRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	Permissions []string `json:"permissions" binding:"required"`
}

// SaveRole godoc
// @Summary      Save a role with permissions
// @Description  Find or create a role by name and set its permissions (creates permissions that don't yet exist)
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Param        input  body      SaveRoleRequest  true  "Role name and permission names"
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Router       /admin/permissions/roles/save [post]
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
