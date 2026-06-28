package handler

import (
	"net/http"

	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/gin-gonic/gin"
)

func TestDB(c *gin.Context) {
	if database.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	var count int64
	database.DB.Raw("SELECT COUNT(*) FROM users").Count(&count)

	c.JSON(http.StatusOK, gin.H{
		"message":    "database test ok",
		"user_count": count,
	})
}

func TestCreateUser(c *gin.Context) {
	if database.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	result := database.DB.Exec(`
		INSERT INTO users (id, username, email, password, role, created_at, updated_at)
		VALUES (gen_random_uuid(), ?, ?, ?, 'user', NOW(), NOW())
	`, req.Username, req.Email, req.Password)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "user created",
		"username": req.Username,
		"email":    req.Email,
	})
}
