package todo

import (
	"go-gin-api/config"
	"go-gin-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CREATE
func CreateTodo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in session"})
		return
	}

	var todo models.Todo

	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	todo.UserID = userID.(uint)
	config.DB.Create(&todo)
	c.JSON(http.StatusOK, todo)
}

// READ ALL
func GetTodos(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in session"})
		return
	}

	var todos []models.Todo
	config.DB.Where("user_id = ?", userID.(uint)).Find(&todos)
	c.JSON(http.StatusOK, todos)
}

// READ ONE
func GetTodo(c *gin.Context) {
	var todo models.Todo
	id := c.Param("id")

	if err := config.DB.First(&todo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	c.JSON(http.StatusOK, todo)
}

// UPDATE
func UpdateTodo(c *gin.Context) {
	var todo models.Todo
	id := c.Param("id")

	if err := config.DB.First(&todo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	c.ShouldBindJSON(&todo)
	config.DB.Save(&todo)

	c.JSON(http.StatusOK, todo)
}

// DELETE
func DeleteTodo(c *gin.Context) {
	id := c.Param("id")
	config.DB.Delete(&models.Todo{}, id)

	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}
