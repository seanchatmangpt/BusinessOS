// Package handlers provides HTTP handlers for calendar scheduling features.
//
// This file implements the NEW calendar endpoints from the architecture doc:
//
// 1. GET  /api/calendar/availability  - Check free/busy slots for attendees
// 2. POST /api/calendar/schedule      - Agent proposes optimal meeting times
// 3. GET  /api/calendar/suggestions   - Calendar-based task recommendations
//
// These endpoints complement the EXISTING calendar endpoints:
// - GET/POST /api/integrations/google_calendar/events
// - POST /api/integrations/google_calendar/sync
//
// The key design principle: these endpoints PROPOSE and RECOMMEND,
// they do NOT auto-create events or tasks. User approval is always required.
//
// Reference: docs/CALENDAR_VOICE_ARCHITECTURE.md
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/integrations/google"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)

// CalendarSchedulingHandler handles calendar availability and scheduling endpoints.
type CalendarSchedulingHandler struct {
	schedulingSvc  *services.SchedulingService
	suggestionSvc  *services.SuggestionService
	calendarSvc    *google.CalendarService
}

// NewCalendarSchedulingHandler creates a new calendar scheduling handler.
func NewCalendarSchedulingHandler(
	pool *pgxpool.Pool,
	googleProvider *google.Provider,
) *CalendarSchedulingHandler {
	calendarSvc := google.NewCalendarService(googleProvider)
	
	return &CalendarSchedulingHandler{
		schedulingSvc:  services.NewSchedulingService(calendarSvc),
		suggestionSvc:  services.NewSuggestionService(pool),
		calendarSvc:    calendarSvc,
	}
}

// RegisterRoutes registers the calendar scheduling routes.
// These are mounted under /api/calendar/
func (h *CalendarSchedulingHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	calendar := rg.Group("/calendar")
	calendar.Use(authMiddleware)
	{
		// Availability checking
		calendar.GET("/availability", h.GetAvailability)
		
		// Agent-driven scheduling (proposals)
		calendar.POST("/schedule", h.ProposeSchedule)
		
		// Task suggestions based on calendar
		calendar.GET("/suggestions", h.GetTaskSuggestions)
	}
}

// ============================================================================
// Endpoint: GET /api/calendar/availability
// ============================================================================

// GetAvailabilityRequest represents query parameters for availability check.
type GetAvailabilityRequest struct {
	Attendees       []string `form:"attendees"`        // Comma-separated or multiple params
	DurationMinutes int      `form:"duration_minutes"` // Default: 60
	StartDate       string   `form:"start_date"`       // YYYY-MM-DD (required)
	EndDate         string   `form:"end_date"`         // YYYY-MM-DD (required)
	WorkingStart    string   `form:"working_start"`    // HH:MM, default 09:00
	WorkingEnd      string   `form:"working_end"`      // HH:MM, default 17:00
}

// GetAvailability handles GET /api/calendar/availability
//
// This endpoint queries Google Calendar's FreeBusy API to find time slots
// where ALL specified attendees (plus the current user) are available.
//
// Query Parameters:
//   - attendees: Email addresses of attendees (can repeat or comma-separate)
//   - duration_minutes: Length of meeting (default: 60)
//   - start_date: Start of search range (YYYY-MM-DD, required)
//   - end_date: End of search range (YYYY-MM-DD, required)
//   - working_start: Start of working hours (HH:MM, default: 09:00)
//   - working_end: End of working hours (HH:MM, default: 17:00)
//
// Response:
//
//	{
//	  "available_slots": [
//	    {"start": "2026-01-20T10:00:00Z", "end": "2026-01-20T11:00:00Z"},
//	    ...
//	  ],
//	  "attendees_checked": ["alice@co.com", "bob@co.com", "primary"],
//	  "working_hours": {"start": "09:00", "end": "17:00"},
//	  "total_slots": 15
//	}
func (h *CalendarSchedulingHandler) GetAvailability(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Parse query parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	
	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "start_date and end_date are required (YYYY-MM-DD format)",
		})
		return
	}

	// Parse attendees (support both repeated params and comma-separated)
	attendees := c.QueryArray("attendees")
	if len(attendees) == 0 {
		attendeesStr := c.Query("attendees")
		if attendeesStr != "" {
			attendees = splitAndTrim(attendeesStr, ",")
		}
	}

	// Parse duration (default 60 minutes)
	durationMinutes := 60
	if d := c.Query("duration_minutes"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			durationMinutes = parsed
		}
	}

	// Parse working hours
	var workingHours *services.TimeRange
	workingStart := c.DefaultQuery("working_start", "09:00")
	workingEnd := c.DefaultQuery("working_end", "17:00")
	workingHours = &services.TimeRange{
		Start: workingStart,
		End:   workingEnd,
	}

	// Build request
	req := services.AvailabilityRequest{
		Attendees:       attendees,
		DurationMinutes: durationMinutes,
		DateRange: services.DateRange{
			Start: startDate,
			End:   endDate,
		},
		WorkingHours: workingHours,
	}

	// Get availability
	resp, err := h.schedulingSvc.GetAvailability(c.Request.Context(), user.ID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check availability",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// Endpoint: POST /api/calendar/schedule
// ============================================================================

// ProposeScheduleRequest represents the request body for scheduling proposals.
type ProposeScheduleRequest struct {
	Title           string                       `json:"title" binding:"required"`
	Attendees       []string                     `json:"attendees"`
	DurationMinutes int                          `json:"duration_minutes"` // Default: 60
	MeetingType     string                       `json:"meeting_type,omitempty"`
	TimePreference  string                       `json:"time_preference,omitempty"` // morning, afternoon, any, next_week
	Preferences     *services.SchedulingPreferences `json:"preferences,omitempty"`
	DateRange       *services.DateRange          `json:"date_range,omitempty"`
}

// ProposeSchedule handles POST /api/calendar/schedule
//
// This endpoint is used by the agent to intelligently propose meeting times.
// It does NOT create the event - it returns proposals for user approval.
//
// Request Body:
//
//	{
//	  "title": "Design Team Kickoff",
//	  "attendees": ["designer1@co.com", "designer2@co.com"],
//	  "duration_minutes": 60,
//	  "meeting_type": "kickoff",
//	  "time_preference": "morning",
//	  "preferences": {
//	    "prefer_morning": true,
//	    "avoid_days": ["Monday"]
//	  }
//	}
//
// Response:
//
//	{
//	  "proposals": [
//	    {
//	      "slot": {"start": "2026-01-21T10:00:00Z", "end": "2026-01-21T11:00:00Z"},
//	      "score": 0.95,
//	      "reason": "All attendees free, morning slot"
//	    },
//	    ...
//	  ],
//	  "event_draft": {
//	    "title": "Design Team Kickoff",
//	    "attendees": ["designer1@co.com", "designer2@co.com"],
//	    "duration_minutes": 60
//	  }
//	}
func (h *CalendarSchedulingHandler) ProposeSchedule(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req ProposeScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	// Default duration
	if req.DurationMinutes <= 0 {
		req.DurationMinutes = 60
	}

	// Build scheduling request
	schedReq := services.ScheduleRequest{
		Title:           req.Title,
		Attendees:       req.Attendees,
		DurationMinutes: req.DurationMinutes,
		MeetingType:     req.MeetingType,
		TimePreference:  req.TimePreference,
		Preferences:     req.Preferences,
		DateRange:       req.DateRange,
	}

	// Get proposals
	proposals, err := h.schedulingSvc.ProposeSchedule(c.Request.Context(), user.ID, schedReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate schedule proposals",
			"message": err.Error(),
		})
		return
	}

	// Check if we found any available times
	if len(proposals.Proposals) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"proposals":   []interface{}{},
			"event_draft": proposals.EventDraft,
			"message":     "No available time slots found for all attendees in the specified date range",
		})
		return
	}

	c.JSON(http.StatusOK, proposals)
}

// ============================================================================
// Endpoint: GET /api/calendar/suggestions
// ============================================================================

// GetTaskSuggestions handles GET /api/calendar/suggestions
//
// This endpoint analyzes the user's calendar and returns task suggestions:
// - Meeting prep tasks for upcoming meetings
// - Follow-up tasks for recent meetings
// - Checks against existing tasks to avoid duplicates
//
// Query Parameters:
//   - limit: Maximum number of suggestions (default: 10, max: 25)
//
// Response:
//
//	{
//	  "suggestions": [
//	    {
//	      "type": "meeting_prep",
//	      "title": "Prepare agenda for Design Kickoff",
//	      "related_event_id": "evt_123",
//	      "related_event": "Design Kickoff",
//	      "suggested_due": "2026-01-21T09:00:00Z",
//	      "priority": "high",
//	      "reason": "Meeting in 2 days, no agenda attached"
//	    },
//	    ...
//	  ],
//	  "generated_at": "2026-01-18T14:30:00Z"
//	}
func (h *CalendarSchedulingHandler) GetTaskSuggestions(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Get suggestions
	resp, err := h.suggestionSvc.GetTaskSuggestions(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate task suggestions",
			"message": err.Error(),
		})
		return
	}

	// Apply limit if specified
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			if limit > 25 {
				limit = 25
			}
			if len(resp.Suggestions) > limit {
				resp.Suggestions = resp.Suggestions[:limit]
			}
		}
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================================================
// Helpers
// ============================================================================

// splitAndTrim splits a string by delimiter and trims whitespace.
func splitAndTrim(s string, delimiter string) []string {
	parts := []string{}
	for _, part := range splitString(s, delimiter) {
		trimmed := trimWhitespace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	result := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimWhitespace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
