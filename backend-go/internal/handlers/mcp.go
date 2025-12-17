package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)

// ListMCPTools returns all available MCP tools
func (h *Handlers) ListMCPTools(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	mcpService := services.NewMCPService(h.pool, user.ID)
	tools := mcpService.GetAllTools()

	c.JSON(http.StatusOK, tools)
}

// ExecuteMCPTool executes an MCP tool
func (h *Handlers) ExecuteMCPTool(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Tool      string                 `json:"tool" binding:"required"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mcpService := services.NewMCPService(h.pool, user.ID)
	result, err := mcpService.ExecuteTool(c.Request.Context(), req.Tool, req.Arguments)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"result":  result,
	})
}

// MCPHealth returns the health status of MCP services
func (h *Handlers) MCPHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "mcp",
	})
}
