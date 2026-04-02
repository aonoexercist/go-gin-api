package config

import (
	"go-gin-api/models"
	"log"

	"gorm.io/gorm"
)

func SeedRBAC(db *gorm.DB) {
	permissions := []string{"create", "read", "update", "delete"}

	var permModels []models.Permission

	// 1. Create permissions if not exist
	for _, p := range permissions {
		var perm models.Permission

		err := db.Where("name = ?", p).First(&perm).Error
		if err != nil {
			perm = models.Permission{Name: p}
			if err := db.Create(&perm).Error; err != nil {
				log.Println("failed to create permission:", p)
				continue
			}
		}

		permModels = append(permModels, perm)
	}

	// 2. Create super_admin role if not exist
	var role models.Role
	err := db.Where("name = ?", "super_admin").First(&role).Error
	if err != nil {
		role = models.Role{Name: "super_admin"}
		if err := db.Create(&role).Error; err != nil {
			log.Println("failed to create role: super_admin")
			return
		}
	}

	// 3. Assign all permissions to super_admin
	if err := db.Model(&role).Association("Permissions").Replace(permModels); err != nil {
		log.Println("failed to assign permissions to super_admin")
	}
}
