package auth

import (
	"go-gin-api/config"
	"go-gin-api/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Register(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.BindJSON(&user)

	hashedPassword, _err := HashPassword(user.Password)
	if _err != nil {
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	user.Password = hashedPassword

	// save to DB (gorm)
	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Could not create user"})
		return
	}

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

	accessToken, _ := GenerateAccessToken(user.ID)
	refreshToken, _ := GenerateRefreshToken(user.ID)

	SetAuthCookies(c, accessToken, refreshToken)

	c.JSON(200, gin.H{
		"message": "Login successful",
	})
}

func Refresh(c *gin.Context) {
	oldToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(401, gin.H{"error": "No token"})
		return
	}

	token, err := jwt.Parse(oldToken, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		c.JSON(401, gin.H{"error": "Invalid token"})
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	sessionID := uint(claims["session_id"].(float64))

	var session models.Session
	config.DB.First(&session, sessionID)

	// 🔥 Detect token reuse (VERY IMPORTANT)
	if session.RefreshToken != oldToken {
		// possible attack → revoke all sessions
		config.DB.Where("user_id = ?", session.UserID).Delete(&models.Session{})
		c.JSON(401, gin.H{"error": "Token reuse detected"})
		return
	}

	// generate new tokens
	newRefreshToken, _ := GenerateRefreshToken(session.ID)
	newAccessToken, _ := GenerateAccessToken(session.UserID)

	// rotate refresh token
	session.RefreshToken = newRefreshToken
	config.DB.Save(&session)

	SetAuthCookies(c, newAccessToken, newRefreshToken)

	c.JSON(200, gin.H{"message": "Refreshed"})
}

func Logout(c *gin.Context) {
	token, _ := c.Cookie("refresh_token")

	parsed, _ := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	claims := parsed.Claims.(jwt.MapClaims)
	sessionID := uint(claims["session_id"].(float64))

	config.DB.Delete(&models.Session{}, sessionID)

	ClearCookies(c)

	c.JSON(200, gin.H{"message": "Logged out"})
}
