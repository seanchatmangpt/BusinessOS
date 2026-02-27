package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	osa "github.com/Miosa-osa/sdk-go"
	"github.com/rhl/businessos-backend/internal/services"
)

// OSAHandler handles HTTP requests for the OSA integration endpoints.
// It holds both the higher-level OSAService (for Health/ListSkills) and the
// raw osa.Client for operations that OSAService does not wrap
// (Orchestrate with arbitrary payloads, Stream, Classify).
type OSAHandler struct {
	osaSvc *services.OSAService
	client osa.Client
}

// NewOSAHandler creates a new OSAHandler.
// The client is extracted from the service so the handler can use it directly
// for methods not surfaced by OSAService (Orchestrate, Stream, Classify).
func NewOSAHandler(osaSvc *services.OSAService) *OSAHandler {
	return &OSAHandler{
		osaSvc: osaSvc,
		client: osaSvc.Client(),
	}
}

// RegisterOSARoutes registers all OSA routes on the provided router group.
// Routes:
//
//	POST /osa/orchestrate        - run the full ReAct agent loop
//	GET  /osa/stream/:sessionID  - SSE stream for live agent events
//	POST /osa/classify           - classify a message with Signal Theory
//	GET  /osa/health             - OSA health check
//	GET  /osa/skills             - list available skills
func (h *OSAHandler) RegisterOSARoutes(rg *gin.RouterGroup) {
	osaGroup := rg.Group("/osa")
	{
		osaGroup.POST("/orchestrate", h.Orchestrate)
		osaGroup.GET("/stream/:sessionID", h.Stream)
		osaGroup.POST("/classify", h.Classify)
		osaGroup.GET("/health", h.Health)
		osaGroup.GET("/skills", h.ListSkills)
	}
}

// orchestrateRequest is the JSON body for POST /osa/orchestrate.
type orchestrateRequest struct {
	Input       string         `json:"input" binding:"required"`
	UserID      string         `json:"user_id"`
	WorkspaceID string         `json:"workspace_id"`
	SessionID   string         `json:"session_id"`
	Context     map[string]any `json:"context"`
}

// Orchestrate handles POST /osa/orchestrate.
// Runs the full ReAct agent loop and returns the final agent response.
func (h *OSAHandler) Orchestrate(c *gin.Context) {
	var req orchestrateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	sdkReq := osa.OrchestrateRequest{
		Input:       req.Input,
		UserID:      req.UserID,
		WorkspaceID: req.WorkspaceID,
		SessionID:   req.SessionID,
		Context:     req.Context,
	}

	resp, err := h.client.Orchestrate(c.Request.Context(), sdkReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"code":    "OSA_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// Stream handles GET /osa/stream/:sessionID.
// Opens an SSE connection and forwards events from the OSA event channel until
// the session completes or the client disconnects.
func (h *OSAHandler) Stream(c *gin.Context) {
	sessionID := c.Param("sessionID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "sessionID is required",
			},
		})
		return
	}

	eventCh, err := h.client.Stream(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"code":    "OSA_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	// Set SSE response headers.
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // disable nginx proxy buffering

	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-eventCh:
			if !ok {
				return false
			}
			data, jsonErr := json.Marshal(event)
			if jsonErr != nil {
				return false
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}

// classifyRequest is the JSON body for POST /osa/classify.
type classifyRequest struct {
	Message string `json:"message" binding:"required"`
	Channel string `json:"channel"`
}

// Classify handles POST /osa/classify.
// Classifies a message using Signal Theory 5-tuple without processing it.
func (h *OSAHandler) Classify(c *gin.Context) {
	var req classifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	channel := osa.Channel(req.Channel)
	if channel == "" {
		channel = osa.ChannelHTTP
	}

	signal, err := h.client.Classify(c.Request.Context(), req.Message, channel)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"code":    "OSA_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": signal})
}

// Health handles GET /osa/health.
// Returns the health status of the connected OSA instance.
func (h *OSAHandler) Health(c *gin.Context) {
	status, err := h.osaSvc.Health(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"code":    "OSA_UNHEALTHY",
				"message": err.Error(),
			},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": status})
}

// ListSkills handles GET /osa/skills.
// Returns the list of skills available in the connected OSA instance.
func (h *OSAHandler) ListSkills(c *gin.Context) {
	skills, err := h.osaSvc.ListSkills(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"code":    "OSA_ERROR",
				"message": err.Error(),
			},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": skills})
}
