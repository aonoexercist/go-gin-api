package auth

import (
	"fmt"
	"go-gin-api/config"
	"go-gin-api/models"

	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/api/idtoken"
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

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !CheckPassword(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := userLogin(c, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not complete login"})
		return
	}

	c.JSON(200, gin.H{"message": "Login successful"})
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

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(401, gin.H{"error": "Invalid token claims"})
		return
	}

	sessionID, ok := claims["session_id"].(float64)
	if !ok {
		c.JSON(401, gin.H{"error": "Invalid session ID claim"})
		return
	}

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
	newRefreshToken, err := GenerateRefreshToken(session.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not generate refresh token"})
		return
	}

	newAccessToken, err := GenerateAccessToken(session.UserID)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not generate access token"})
		return
	}

	// rotate refresh token
	session.RefreshToken = newRefreshToken
	if err := config.DB.Save(&session).Error; err != nil {
		c.JSON(500, gin.H{"error": "could not rotate refresh token"})
		return
	}

	SetAuthCookies(c, newAccessToken, newRefreshToken)

	c.JSON(200, gin.H{"message": "Refreshed"})
}

func Logout(c *gin.Context) {
	token, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(401, gin.H{"error": "No token"})
		return
	}

	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !parsed.Valid {
		c.JSON(401, gin.H{"error": "Invalid token"})
		return
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(401, gin.H{"error": "Invalid token claims"})
		return
	}

	sessionIDf, ok := claims["session_id"].(float64)
	if !ok {
		c.JSON(401, gin.H{"error": "Invalid session id in token"})
		return
	}

	sessionID := uint(sessionIDf)

	if err := config.DB.Delete(&models.Session{}, sessionID).Error; err != nil {
		c.JSON(500, gin.H{"error": "Could not delete session"})
		return
	}

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

	if err := config.DB.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User record not found in database"})
		return
	}

	c.JSON(http.StatusOK, ToUserDTO(user))
}

func userLogin(c *gin.Context, user models.User) error {
	session := models.Session{
		UserID:    user.ID,
		UserAgent: c.Request.UserAgent(),
		IPAddress: c.ClientIP(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := config.DB.Create(&session).Error; err != nil {
		return err
	}

	refreshToken, err := GenerateRefreshToken(session.ID)
	if err != nil {
		return err
	}

	accessToken, err := GenerateAccessToken(user.ID)
	if err != nil {
		return err
	}

	session.RefreshToken = refreshToken
	if err := config.DB.Save(&session).Error; err != nil {
		return err
	}

	SetAuthCookies(c, accessToken, refreshToken)
	return nil
}

func GoogleLogin(c *gin.Context) {
	var req models.GoogleAuthRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	ctx := c.Request.Context()

	// 🔥 Verify Google ID Token
	payload, err := idtoken.Validate(ctx, req.Token, os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid google token"})
		return
	}

	fmt.Printf("Google token payload: %+v\n", payload) // Debugging line

	// Extract data safely
	emailI, ok := payload.Claims["email"].(string)
	if !ok {
		c.JSON(400, gin.H{"error": "invalid google token payload (email)"})
		return
	}
	nameI, ok := payload.Claims["name"].(string)
	if !ok {
		c.JSON(400, gin.H{"error": "invalid google token payload (name)"})
		return
	}
	googleID := payload.Subject

	// Build your struct
	userInfo := models.GoogleUserInfo{
		ID:    googleID,
		Email: emailI,
		Name:  nameI,
	}

	// 🔥 Your existing logic
	user, err := FindOrCreateUser(config.DB, userInfo)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to process user"})
		return
	}

	if err := userLogin(c, *user); err != nil {
		c.JSON(500, gin.H{"error": "could not complete login"})
		return
	}

	c.JSON(200, gin.H{"message": "login successful"})
}
