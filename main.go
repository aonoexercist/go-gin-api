package main

import (
	"go-gin-api/config"
	"go-gin-api/models"
	"go-gin-api/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	config.ConnectDB()

	// Auto migrate
	config.DB.AutoMigrate(&models.Todo{})

	routes.SetupRoutes(r)

	r.Run(":8080")
}
