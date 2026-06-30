package todo

import (
	"go-gin-api/config"
	"go-gin-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateTodo godoc
// @Summary      Create a todo
// @Description  Create a new todo item owned by the currently authenticated user
// @Tags         todos
// @Accept       json
// @Produce      json
// @Param        todo  body      models.Todo  true  "Todo to create"
// @Success      200   {object}  models.Todo
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /services/todos [post]
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

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in session"})
		return
	}

	todo.UserID = uid
	if err := config.DB.Create(&todo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create todo"})
		return
	}

	c.JSON(http.StatusOK, todo)
}

// GetTodos godoc
// @Summary      List todos
// @Description  Get all todos belonging to the currently authenticated user
// @Tags         todos
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.Todo
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /services/todos [get]
func GetTodos(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in session"})
		return
	}

	var todos []models.Todo

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in session"})
		return
	}

	if err := config.DB.Where("user_id = ?", uid).Find(&todos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve todos"})
		return
	}

	c.JSON(http.StatusOK, todos)
}

// GetTodo godoc
// @Summary      Get a single todo
// @Description  Get a todo by ID
// @Tags         todos
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Todo ID"
// @Success      200  {object}  models.Todo
// @Failure      404  {object}  map[string]string
// @Router       /services/todos/{id} [get]
func GetTodo(c *gin.Context) {
	var todo models.Todo
	id := c.Param("id")

	if err := config.DB.First(&todo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	c.JSON(http.StatusOK, todo)
}

// UpdateTodo godoc
// @Summary      Update a todo
// @Description  Update an existing todo by ID
// @Tags         todos
// @Accept       json
// @Produce      json
// @Param        id    path      int          true  "Todo ID"
// @Param        todo  body      models.Todo  true  "Updated todo fields"
// @Success      200   {object}  models.Todo
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /services/todos/{id} [put]
func UpdateTodo(c *gin.Context) {
	var todo models.Todo
	id := c.Param("id")

	if err := config.DB.First(&todo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Save(&todo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update todo"})
		return
	}

	c.JSON(http.StatusOK, todo)
}

// DeleteTodo godoc
// @Summary      Delete a todo
// @Description  Delete a todo by ID
// @Tags         todos
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Todo ID"
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /services/todos/{id} [delete]
func DeleteTodo(c *gin.Context) {
	id := c.Param("id")
	if err := config.DB.Delete(&models.Todo{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete todo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}
