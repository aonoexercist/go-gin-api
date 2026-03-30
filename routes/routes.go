package routes

import (
	"go-gin-api/controllers/auth"
	"go-gin-api/controllers/todo"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	authApi := r.Group("/auth")
	{
		authApi.POST("/register", auth.Register)
		authApi.POST("/login", auth.Login)
	}

	api := r.Group("/api")
	api.Use(auth.AuthMiddleware())
	{
		api.GET("/me", auth.Me)
		api.POST("/logout", auth.Logout)

		api.POST("/todos", todo.CreateTodo)
		api.GET("/todos", todo.GetTodos)
		api.GET("/todos/:id", todo.GetTodo)
		api.PUT("/todos/:id", todo.UpdateTodo)
		api.DELETE("/todos/:id", todo.DeleteTodo)
	}
}
