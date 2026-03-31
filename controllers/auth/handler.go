package auth

import (
	"go-gin-api/config"
	"go-gin-api/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Register(c *gin.Context) {
	var input RegisterDTO
	var user models.User

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user = models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password, // This will be hashed by the GORM Hook
	}

	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	user.Password = hashedPassword

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created"})
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	session := models.Session{
		UserID:    user.ID,
		UserAgent: c.Request.UserAgent(),
		IPAddress: c.ClientIP(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	config.DB.Create(&session)

	refreshToken, _ := GenerateRefreshToken(session.ID)
	accessToken, _ := GenerateAccessToken(user.ID)

	session.RefreshToken = refreshToken
	config.DB.Save(&session)

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

func Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in session"})
		return
	}

	var user models.User

	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User record not found in database"})
		return
	}

	c.JSON(http.StatusOK, user)
}
