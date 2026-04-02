package routes

import (
	"go-gin-api/controllers/auth"
	"go-gin-api/controllers/rbac"
	"go-gin-api/controllers/todo"
	"go-gin-api/controllers/user"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	authApi := r.Group("/auth")
	{
		authApi.POST("/register", auth.Register)
		authApi.POST("/login", auth.Login)
		authApi.POST("/refresh", auth.Refresh)
		authApi.POST("/logout", auth.Logout)
		authApi.POST("/google/login", auth.GoogleLogin)
	}

	api := r.Group("/services")
	api.Use(auth.AuthMiddleware())
	{
		api.GET("/me", auth.Me)

		api.POST("/todos", todo.CreateTodo)
		api.GET("/todos", todo.GetTodos)
		api.GET("/todos/:id", todo.GetTodo)
		api.PUT("/todos/:id", todo.UpdateTodo)
		api.DELETE("/todos/:id", todo.DeleteTodo)
	}

	adminApi := r.Group("/admin")
	adminApi.Use(auth.AuthMiddleware(), rbac.RequireRole("super_admin"))
	{
		usersApi := adminApi.Group("/users")
		{
			usersApi.GET("/", user.GetUsers)
			usersApi.GET("/:id", user.GetUser)
			usersApi.PUT("/:id/roles", user.UpdateUserRoles)
		}

		rolesApi := adminApi.Group("/roles")
		{
			rolesApi.POST("/", rbac.CreateRole)
			rolesApi.GET("/", rbac.GetRoles)
			rolesApi.PUT("/:id", rbac.UpdateRole)
			rolesApi.DELETE("/:id", rbac.DeleteRole)
			rolesApi.PUT("/update/user", rbac.UpdateUserRolesHandler)
		}

		permissionsApi := adminApi.Group("/permissions")
		{
			permissionsApi.GET("/role/:id", rbac.GetPermissionsByRole)
			permissionsApi.POST("/save", rbac.SaveRole)
		}
	}
}
