package auth

import (
	"fmt"
	"go-gin-api/models"

	"gorm.io/gorm"
)

func FindOrCreateUser(db *gorm.DB, info models.GoogleUserInfo) (*models.User, error) {
	var user models.User

	// 1️⃣ Try find by Google ID
	if err := db.Where("google_id = ?", info.ID).First(&user).Error; err == nil {
		// SaveUserData(&info, &user)

		return &user, nil
	}

	// 2️⃣ Try find by email
	if err := db.Where("email = ?", info.Email).First(&user).Error; err == nil {
		// 🔗 Link Google account
		user.GoogleID = &info.ID
		user.Provider = "google"

		// SaveUserData(&info, &user)
		if err := db.Save(&user).Error; err != nil {
			return nil, err
		}

		return &user, nil
	}

	// 3️⃣ Create new user
	user = models.User{
		Email:    info.Email,
		Name:     info.Name,
		GoogleID: &info.ID,
		Provider: "google",
	}

	fmt.Printf("Attempting to create user: %+v\n", user)

	if err := db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// func SaveUserData(info *models.GoogleUserInfo, user *models.User) error {
// 	// Check if name is empty, then update it with Google's data
// 	if user.Name == "" {
// 		user.Name = info.Name
// 	}

// 	if err := config.DB.Save(&user).Error; err != nil {
// 		return err
// 	}
// 	return nil
// }
