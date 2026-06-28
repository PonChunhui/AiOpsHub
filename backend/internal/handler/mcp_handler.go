package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListMCPServers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	servers, total, err := mcpService.List(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"servers":  servers,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func GetMCPServer(c *gin.Context) {
	id := c.Param("id")

	server, err := mcpService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "MCP server not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   200,
		"server": server,
	})
}

func CreateMCPServer(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		URL         string `json:"url" binding:"required"`
		AuthType    string `json:"auth_type"`
		AuthToken   string `json:"auth_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, _ := c.Get("username")
	createdBy := "unknown"
	if username != nil {
		createdBy = username.(string)
	}

	server, err := mcpService.Create(c.Request.Context(), req.Name, req.Description, req.URL, req.AuthType, req.AuthToken, createdBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "MCP Server created successfully",
		"server":  server,
	})
}

func UpdateMCPServer(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		URL         string `json:"url"`
		AuthType    string `json:"auth_type"`
		AuthToken   string `json:"auth_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, _ := c.Get("username")
	updatedBy := "unknown"
	if username != nil {
		updatedBy = username.(string)
	}

	server, err := mcpService.Update(c.Request.Context(), id, req.Name, req.Description, req.URL, req.AuthType, req.AuthToken, updatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "MCP Server updated successfully",
		"server":  server,
	})
}

func DeleteMCPServer(c *gin.Context) {
	id := c.Param("id")

	if err := mcpService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "MCP Server deleted successfully",
	})
}

func TestMCPServer(c *gin.Context) {
	id := c.Param("id")

	err := mcpService.TestConnection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "Connection failed: " + err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Connection successful",
		"success": true,
	})
}

func GetMCPServerTools(c *gin.Context) {
	id := c.Param("id")

	tools, err := mcpService.GetTools(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"tools": tools,
	})
}
