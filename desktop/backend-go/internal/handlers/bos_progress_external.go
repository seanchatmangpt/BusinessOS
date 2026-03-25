package handlers

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/services"
)

// ReceiveExternalProgressEventRequest represents an incoming progress event from pm4py-rust
type ReceiveExternalProgressEventRequest struct {
	Progress  uint32 `json:"progress" binding:"required"`
	Algorithm string `json:"algorithm" binding:"required"`
	ElapsedMs uint64 `json:"elapsed_ms" binding:"required"`
	SessionID string `json:"session_id"` // Optional: group events by session
}

var (
	// globalStreamingService is a package-level streaming service for receiving external events
	// This is initialized once and reused
	globalStreamingService *services.BOSStreamingService
	streamingServiceMu     sync.Mutex
)

// GetGlobalStreamingService returns the global streaming service, initializing if needed
func GetGlobalStreamingService() *services.BOSStreamingService {
	streamingServiceMu.Lock()
	defer streamingServiceMu.Unlock()

	if globalStreamingService == nil {
		globalStreamingService = services.NewBOSStreamingService(slog.Default())
	}

	return globalStreamingService
}

// ReceiveExternalProgressEventHandler handles progress events POSTed from pm4py-rust
// POST /api/bos/progress
//
// This endpoint receives progress events emitted during discovery/conformance operations
// and broadcasts them to connected SSE clients.
//
// Expected JSON body:
//
//	{
//	  "progress": 50,
//	  "algorithm": "alpha",
//	  "elapsed_ms": 2500,
//	  "session_id": "550e8400-e29b-41d4-a716-446655440000" (optional)
//	}
func ReceiveExternalProgressEventHandler(c *gin.Context) {
	var req ReceiveExternalProgressEventRequest
	logger := slog.Default().With("component", "bos_progress_external")

	// Parse JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid progress event request",
			"error", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Validate progress percentage
	if req.Progress > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Progress must be between 0 and 100",
		})
		return
	}

	// Generate session ID if not provided
	sessionIDStr := req.SessionID
	if sessionIDStr == "" {
		sessionIDStr = uuid.New().String()
	}

	// Parse session ID
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		logger.Warn("Invalid session ID in progress event",
			"session_id", sessionIDStr,
			"error", err,
		)
		// Use generated UUID as fallback
		sessionID = uuid.New()
	}

	// Create BOS stream event from progress data
	event := &services.BOSStreamEvent{
		ID:        uuid.New().String(),
		EventType: "discovery_progress",
		SessionID: sessionID.String(),
		Progress: &services.BOSProgressMetrics{
			EventsProcessed: int64(req.Progress), // Using progress as stand-in
			PercentComplete: int32(req.Progress),
			CurrentStep:     req.Algorithm,
		},
		TimestampMs: time.Now().UnixMilli(),
	}

	logger.Info("Received external progress event",
		"progress", req.Progress,
		"algorithm", req.Algorithm,
		"elapsed_ms", req.ElapsedMs,
		"session_id", sessionID,
	)

	// Get streaming service and publish event
	streamingService := GetGlobalStreamingService()
	streamingService.PublishEvent(event)

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"status":     "received",
		"session_id": sessionID.String(),
	})
}
