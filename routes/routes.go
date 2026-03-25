package routes

import (
	"go-gin-api/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")

	{
		api.POST("/todos", controllers.CreateTodo)
		api.GET("/todos", controllers.GetTodos)
		api.GET("/todos/:id", controllers.GetTodo)
		api.PUT("/todos/:id", controllers.UpdateTodo)
		api.DELETE("/todos/:id", controllers.DeleteTodo)
	}
}
