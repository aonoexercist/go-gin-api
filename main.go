package main

import (
	"go-gin-api/config"
	"go-gin-api/models"
	"go-gin-api/routes"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// These variables are populated at build time using -ldflags
var (
	BuildVersion = "dev"
	BuildTime    = "unknown"
	GitCommit    = "unknown"
)

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		// Add both your dev and local URLs here
		AllowOrigins: []string{
			"https://xercisdev.theworkpc.com",
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		},
		AllowCredentials: true,
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Health/Version check endpoint
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":    BuildVersion,
			"build_time": BuildTime,
			"commit":     GitCommit,
		})
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
