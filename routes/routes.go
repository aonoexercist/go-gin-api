package routes

import (
	"go-gin-api/controllers"
	"go-gin-api/controllers/auth"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	authApi := r.Group("/auth")
	{
		authApi.POST("/register", auth.Register)
		authApi.POST("/login", auth.Login)
	}

	api := r.Group("/api")
	{
		api.POST("/todos", controllers.CreateTodo)
		api.GET("/todos", controllers.GetTodos)
		api.GET("/todos/:id", controllers.GetTodo)
		api.PUT("/todos/:id", controllers.UpdateTodo)
		api.DELETE("/todos/:id", controllers.DeleteTodo)
	}
}
