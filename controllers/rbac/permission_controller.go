package rbac

import (
	"go-gin-api/config"
	"go-gin-api/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CreatePermissionRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreatePermission godoc
// @Summary Create permission by role ID
// @Description Creates (or reuses) a permission and attaches it to a specific role.
// @Tags Permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Role ID"
// @Param permission body CreatePermissionRequest true "Permission Name Payload"
// @Success 201 {object} models.Permission
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/permissions/roles/{id}/permissions [post]
func CreatePermissionByRoleId(c *gin.Context) {
	roleId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	var role models.Role
	if err := config.DB.First(&role, roleId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	// Bind to our input DTO instead of full models.Permission
	var input CreatePermissionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find existing permission by name, or create it if it doesn't exist yet.
	var permission models.Permission
	if err := config.DB.Where("name = ?", input.Name).First(&permission).Error; err != nil {
		permission = models.Permission{Name: input.Name}
		if err := config.DB.Create(&permission).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Attach the permission to the role via the many2many join table.
	if err := config.DB.Model(&role).Association("Permissions").Append(&permission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, permission)
}

// GetPermissions godoc
// @Summary List permissions
// @Description Returns all permissions.
// @Tags Permissions
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Permission
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/permissions [get]
func GetPermissions(c *gin.Context) {
	var permissions []models.Permission

	if err := config.DB.Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

// GetPermission godoc
// @Summary Get permission
// @Description Returns a permission by ID.
// @Tags Permissions
// @Produce json
// @Security BearerAuth
// @Param id path int true "Permission ID"
// @Success 200 {object} models.Permission
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/permissions/{id} [get]
func GetPermission(c *gin.Context) {
	var permission models.Permission
	id := c.Param("id")

	if err := config.DB.First(&permission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
		return
	}

	c.JSON(http.StatusOK, permission)
}

// UpdatePermission godoc
// @Summary Update permission
// @Description Updates an existing permission.
// @Tags Permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Permission ID"
// @Param permission body models.Permission true "Updated Permission"
// @Success 200 {object} models.Permission
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/permissions/{id} [put]
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

// DeletePermission godoc
// @Summary Delete permission
// @Description Deletes a permission by ID.
// @Tags Permissions
// @Produce json
// @Security BearerAuth
// @Param id path int true "Permission ID"
// @Success 204
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/permissions/{id} [delete]
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

	c.JSON(http.StatusOK, gin.H{"message": "Permission deleted successfully"})
}

// DeletePermissionFromRole godoc
// @Summary Remove permission from role
// @Description Detaches a permission from a specific role using Role ID and Permission ID.
// @Tags Permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Role ID"
// @Param permission_id path int true "Permission ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/permissions/roles/{id}/permissions/{permission_id} [delete]
func DeletePermissionFromRole(c *gin.Context) {
	roleId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	permissionId, err := strconv.Atoi(c.Param("permission_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission id"})
		return
	}

	// 1. Verify Role exists
	var role models.Role
	if err := config.DB.First(&role, roleId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	// 2. Verify Permission exists
	var permission models.Permission
	if err := config.DB.First(&permission, permissionId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "permission not found"})
		return
	}

	// 3. Remove association from the join table
	if err := config.DB.Model(&role).Association("Permissions").Delete(&permission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. Check if any other role is still using this permission
	var count int64
	config.DB.Table("role_permissions").Where("permission_id = ?", permissionId).Count(&count)

	// 5. If no other roles use it, delete from permissions table
	if count == 0 {
		config.DB.Delete(&permission)
	}

	c.JSON(http.StatusOK, gin.H{"message": "permission removed from role successfully"})
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
