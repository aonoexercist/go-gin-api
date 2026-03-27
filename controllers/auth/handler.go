package auth

import (
	"go-gin-api/config"
	"go-gin-api/models"

	// "net/http"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	c.BindJSON(&input)

	hashedPassword, _ := HashPassword(input.Password)

	user := models.User{
		Email:    input.Email,
		Password: hashedPassword,
	}

	// save to DB (gorm)
	config.DB.Create(&user)

	c.JSON(200, gin.H{"message": "User created"})
}

func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	c.BindJSON(&input)

	var user models.User
	config.DB.Where("email = ?", input.Email).First(&user)

	if !CheckPassword(input.Password, user.Password) {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	token, _ := GenerateToken(user.ID)

	c.JSON(200, gin.H{
		"token": token,
	})
}
