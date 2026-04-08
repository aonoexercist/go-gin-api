package main

import (
	"go-gin-api/config"
	"go-gin-api/models"
	"go-gin-api/routes"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		log.Fatal("FRONTEND_URL not set")
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL}, // frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true, // IMPORTANT for cookies
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

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
