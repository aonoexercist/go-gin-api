package main

import (
	"go-gin-api/config"
	"go-gin-api/models"
	"go-gin-api/routes"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	config.ConnectDB()

	// Auto migrate
	if err := config.DB.AutoMigrate(
		&models.Todo{},
		&models.User{},
		&models.Session{},
		&models.Role{},
		&models.Permission{},
	); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	// Seed RBAC data
	config.SeedRBAC(config.DB)

	// Remove Expired Sessions
	config.CleanupSessions(config.DB)

	routes.SetupRoutes(r)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
