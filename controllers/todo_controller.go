package controllers

import (
	"go-gin-api/config"
	"go-gin-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CREATE
func CreateTodo(c *gin.Context) {
	var todo models.Todo

	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Create(&todo)
	c.JSON(http.StatusOK, todo)
}

// READ ALL
func GetTodos(c *gin.Context) {
	var todos []models.Todo
	config.DB.Find(&todos)
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
