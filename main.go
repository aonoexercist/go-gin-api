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
	config.DB.AutoMigrate(
		&models.Todo{},
		&models.User{},
		&models.Session{},
		&models.Role{},
		&models.Permission{},
	)

	// Seed RBAC data
	config.SeedRBAC(config.DB)

	// Remove Expired Sessions
	config.CleanupSessions(config.DB)

	routes.SetupRoutes(r)

	r.Run(":8080")
}
