package routes

import (
	"go-gin-api/controllers/auth"
	"go-gin-api/controllers/rbac"
	"go-gin-api/controllers/todo"
	"go-gin-api/controllers/user"
	"go-gin-api/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.Use(middleware.ErrorHandler())

	authApi := r.Group("/auth")
	{
		authApi.POST("/register", auth.Register)
		authApi.POST("/login", auth.Login)
		authApi.POST("/refresh", auth.Refresh)
		authApi.POST("/logout", auth.Logout)
		authApi.POST("/google/login", auth.GoogleLogin)
	}

	api := r.Group("/services")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/me", auth.Me)

		api.POST("/todos", middleware.RequirePermission("todo:create"), todo.CreateTodo)
		api.GET("/todos", middleware.RequirePermission("todo:read"), todo.GetTodos)
		api.GET("/todos/:id", middleware.RequirePermission("todo:read"), todo.GetTodo)
		api.PUT("/todos/:id", middleware.RequirePermission("todo:update"), todo.UpdateTodo)
		api.DELETE("/todos/:id", middleware.RequirePermission("todo:delete"), todo.DeleteTodo)
	}

	adminApi := r.Group("/admin")
	adminApi.Use(middleware.AuthMiddleware())
	{
		usersApi := adminApi.Group("/users")
		usersApi.Use(middleware.RequirePermission("user:manage"))
		{
			usersApi.GET("/", user.GetUsers)
			usersApi.GET("/:id", user.GetUser)

			usersApi.PUT("/:id", user.UpdateUser)
			usersApi.PUT("/:id/roles", user.UpdateUserRoles)

			usersApi.DELETE("/:user_id/roles/:role_id", user.DeleteRoleFromUser)
		}

		rolesApi := adminApi.Group("/roles")
		rolesApi.Use(middleware.RequirePermission("role:manage"))
		{
			rolesApi.POST("/", rbac.CreateRole)
			rolesApi.GET("/", rbac.GetRoles)
			rolesApi.PUT("/:id", rbac.UpdateRole)
			rolesApi.DELETE("/:id", rbac.DeleteRole)
			rolesApi.PUT("/update/user", rbac.UpdateUserRolesHandler)
		}

		permissionsApi := adminApi.Group("/permissions")
		permissionsApi.Use(middleware.RequirePermission("permission:manage"))
		{
			permissionsApi.POST("/roles/:id/permissions", rbac.CreatePermissionByRoleId)
			permissionsApi.GET("/roles/:id/permissions", rbac.GetPermissionsByRole)
			permissionsApi.DELETE("/roles/:id/permissions/:permission_id", rbac.DeletePermissionFromRole)

			permissionsApi.GET("/", rbac.GetPermissions)
			permissionsApi.GET("/:id", rbac.GetPermission)
			permissionsApi.PUT("/:id", rbac.UpdatePermission)
			permissionsApi.DELETE("/:id", rbac.DeletePermission)

			permissionsApi.POST("/save", rbac.SaveRole)
		}
	}
}
