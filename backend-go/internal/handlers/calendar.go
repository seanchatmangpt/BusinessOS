package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
	"google.golang.org/api/calendar/v3"
)

// CalendarHandler handles calendar-related endpoints
type CalendarHandler struct {
	*Handlers
	calendarService *services.GoogleCalendarService
}

// NewCalendarHandler creates a new calendar handler
func NewCalendarHandler(h *Handlers, calendarService *services.GoogleCalendarService) *CalendarHandler {
	return &CalendarHandler{
		Handlers:        h,
		calendarService: calendarService,
	}
}

// ListEvents returns calendar events for a date range
func (h *CalendarHandler) ListEvents(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Parse date range from query params
	startStr := c.DefaultQuery("start", time.Now().Format("2006-01-02"))
	endStr := c.DefaultQuery("end", time.Now().AddDate(0, 1, 0).Format("2006-01-02"))

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		start = time.Now()
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		end = time.Now().AddDate(0, 1, 0)
	}

	queries := sqlc.New(h.pool)
	events, err := queries.ListCalendarEvents(c.Request.Context(), sqlc.ListCalendarEventsParams{
		UserID:    user.ID,
		StartTime: pgtype.Timestamptz{Time: start, Valid: true},
		EndTime:   pgtype.Timestamptz{Time: end.Add(24 * time.Hour), Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list events"})
		return
	}

	c.JSON(http.StatusOK, TransformCalendarEvents(events))
}

// GetEvent returns a single calendar event
func (h *CalendarHandler) GetEvent(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	queries := sqlc.New(h.pool)
	event, err := queries.GetCalendarEvent(c.Request.Context(), sqlc.GetCalendarEventParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(http.StatusOK, TransformCalendarEvent(event))
}

// CreateEvent creates a new calendar event
func (h *CalendarHandler) CreateEvent(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Title         string   `json:"title" binding:"required"`
		Description   *string  `json:"description"`
		StartTime     string   `json:"start_time" binding:"required"`
		EndTime       string   `json:"end_time" binding:"required"`
		AllDay        bool     `json:"all_day"`
		Location      *string  `json:"location"`
		MeetingType   *string  `json:"meeting_type"`
		ContextID     *string  `json:"context_id"`
		ProjectID     *string  `json:"project_id"`
		ClientID      *string  `json:"client_id"`
		RecordingURL  *string  `json:"recording_url"`
		MeetingLink   *string  `json:"meeting_link"`
		ExternalLinks []string `json:"external_links"`
		MeetingNotes  *string  `json:"meeting_notes"`
		SyncToGoogle  bool     `json:"sync_to_google"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start time format"})
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end time format"})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse optional UUIDs
	var contextID, projectID, clientID pgtype.UUID
	if req.ContextID != nil {
		if parsed, err := uuid.Parse(*req.ContextID); err == nil {
			contextID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}
	if req.ProjectID != nil {
		if parsed, err := uuid.Parse(*req.ProjectID); err == nil {
			projectID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}
	if req.ClientID != nil {
		if parsed, err := uuid.Parse(*req.ClientID); err == nil {
			clientID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	// Parse meeting type
	var meetingType sqlc.NullMeetingtype
	if req.MeetingType != nil {
		meetingType = sqlc.NullMeetingtype{
			Meetingtype: stringToMeetingType(*req.MeetingType),
			Valid:       true,
		}
	}

	// External links JSON
	externalLinksJSON := []byte("[]")
	if len(req.ExternalLinks) > 0 {
		externalLinksJSON, _ = json.Marshal(req.ExternalLinks)
	}

	// Sync to Google if requested and connected
	var googleEventID *string
	if req.SyncToGoogle {
		googleEvent := &calendar.Event{
			Summary:     req.Title,
			Description: stringValue(req.Description),
			Location:    stringValue(req.Location),
			Start: &calendar.EventDateTime{
				DateTime: startTime.Format(time.RFC3339),
			},
			End: &calendar.EventDateTime{
				DateTime: endTime.Format(time.RFC3339),
			},
		}

		created, err := h.calendarService.CreateEvent(c.Request.Context(), user.ID, googleEvent)
		if err == nil {
			googleEventID = &created.Id
		}
	}

	event, err := queries.CreateCalendarEvent(c.Request.Context(), sqlc.CreateCalendarEventParams{
		UserID:        user.ID,
		GoogleEventID: googleEventID,
		CalendarID:    stringPtr("primary"),
		Title:         &req.Title,
		Description:   req.Description,
		StartTime:     pgtype.Timestamptz{Time: startTime, Valid: true},
		EndTime:       pgtype.Timestamptz{Time: endTime, Valid: true},
		AllDay:        &req.AllDay,
		Location:      req.Location,
		Attendees:     []byte("[]"),
		Status:        stringPtr("confirmed"),
		Visibility:    stringPtr("default"),
		HtmlLink:      nil,
		Source:        stringPtr("businessos"),
		MeetingType:   meetingType,
		ContextID:     contextID,
		ProjectID:     projectID,
		ClientID:      clientID,
		RecordingUrl:  req.RecordingURL,
		MeetingLink:   req.MeetingLink,
		ExternalLinks: externalLinksJSON,
		MeetingNotes:  req.MeetingNotes,
		ActionItems:   []byte("[]"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, TransformCalendarEvent(event))
}

// UpdateEvent updates a calendar event
func (h *CalendarHandler) UpdateEvent(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var req struct {
		Title         *string  `json:"title"`
		Description   *string  `json:"description"`
		StartTime     *string  `json:"start_time"`
		EndTime       *string  `json:"end_time"`
		AllDay        *bool    `json:"all_day"`
		Location      *string  `json:"location"`
		MeetingType   *string  `json:"meeting_type"`
		ContextID     *string  `json:"context_id"`
		ProjectID     *string  `json:"project_id"`
		ClientID      *string  `json:"client_id"`
		RecordingURL  *string  `json:"recording_url"`
		MeetingLink   *string  `json:"meeting_link"`
		ExternalLinks []string `json:"external_links"`
		MeetingNotes  *string  `json:"meeting_notes"`
		ActionItems   []string `json:"action_items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Get existing event
	existing, err := queries.GetCalendarEvent(c.Request.Context(), sqlc.GetCalendarEventParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Update fields
	title := existing.Title
	if req.Title != nil {
		title = req.Title
	}
	description := existing.Description
	if req.Description != nil {
		description = req.Description
	}
	startTime := existing.StartTime
	if req.StartTime != nil {
		if t, err := time.Parse(time.RFC3339, *req.StartTime); err == nil {
			startTime = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}
	endTime := existing.EndTime
	if req.EndTime != nil {
		if t, err := time.Parse(time.RFC3339, *req.EndTime); err == nil {
			endTime = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}
	allDay := existing.AllDay
	if req.AllDay != nil {
		allDay = req.AllDay
	}
	location := existing.Location
	if req.Location != nil {
		location = req.Location
	}

	meetingType := existing.MeetingType
	if req.MeetingType != nil {
		meetingType = sqlc.NullMeetingtype{
			Meetingtype: stringToMeetingType(*req.MeetingType),
			Valid:       true,
		}
	}

	contextID := existing.ContextID
	if req.ContextID != nil {
		if parsed, err := uuid.Parse(*req.ContextID); err == nil {
			contextID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}
	projectID := existing.ProjectID
	if req.ProjectID != nil {
		if parsed, err := uuid.Parse(*req.ProjectID); err == nil {
			projectID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}
	clientID := existing.ClientID
	if req.ClientID != nil {
		if parsed, err := uuid.Parse(*req.ClientID); err == nil {
			clientID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	recordingURL := existing.RecordingUrl
	if req.RecordingURL != nil {
		recordingURL = req.RecordingURL
	}
	meetingLink := existing.MeetingLink
	if req.MeetingLink != nil {
		meetingLink = req.MeetingLink
	}
	externalLinks := existing.ExternalLinks
	if req.ExternalLinks != nil {
		externalLinks, _ = json.Marshal(req.ExternalLinks)
	}
	meetingNotes := existing.MeetingNotes
	if req.MeetingNotes != nil {
		meetingNotes = req.MeetingNotes
	}
	actionItems := existing.ActionItems
	if req.ActionItems != nil {
		actionItems, _ = json.Marshal(req.ActionItems)
	}

	event, err := queries.UpdateCalendarEvent(c.Request.Context(), sqlc.UpdateCalendarEventParams{
		ID:            pgtype.UUID{Bytes: id, Valid: true},
		UserID:        user.ID,
		Title:         title,
		Description:   description,
		StartTime:     startTime,
		EndTime:       endTime,
		AllDay:        allDay,
		Location:      location,
		Attendees:     existing.Attendees,
		Status:        existing.Status,
		Visibility:    existing.Visibility,
		HtmlLink:      existing.HtmlLink,
		MeetingType:   meetingType,
		ContextID:     contextID,
		ProjectID:     projectID,
		ClientID:      clientID,
		RecordingUrl:  recordingURL,
		MeetingLink:   meetingLink,
		ExternalLinks: externalLinks,
		MeetingNotes:  meetingNotes,
		ActionItems:   actionItems,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	c.JSON(http.StatusOK, TransformCalendarEvent(event))
}

// DeleteEvent deletes a calendar event
func (h *CalendarHandler) DeleteEvent(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Get event to check if it's synced with Google
	event, err := queries.GetCalendarEvent(c.Request.Context(), sqlc.GetCalendarEventParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Delete from Google if synced
	if event.GoogleEventID != nil && *event.GoogleEventID != "" {
		_ = h.calendarService.DeleteEvent(c.Request.Context(), user.ID, *event.GoogleEventID)
	}

	// Delete from database
	err = queries.DeleteCalendarEvent(c.Request.Context(), sqlc.DeleteCalendarEventParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted"})
}

// SyncCalendar syncs events from Google Calendar
func (h *CalendarHandler) SyncCalendar(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Sync last 30 days and next 90 days
	timeMin := time.Now().AddDate(0, -1, 0)
	timeMax := time.Now().AddDate(0, 3, 0)

	if err := h.calendarService.SyncEvents(c.Request.Context(), user.ID, timeMin, timeMax); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync calendar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Calendar synced successfully"})
}

// GetTodayEvents returns today's events
func (h *CalendarHandler) GetTodayEvents(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)
	events, err := queries.GetTodayEvents(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get today's events"})
		return
	}

	c.JSON(http.StatusOK, TransformCalendarEvents(events))
}

// GetUpcomingEvents returns upcoming events
func (h *CalendarHandler) GetUpcomingEvents(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)
	events, err := queries.GetUpcomingEvents(c.Request.Context(), sqlc.GetUpcomingEventsParams{
		UserID: user.ID,
		Limit:  10,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get upcoming events"})
		return
	}

	c.JSON(http.StatusOK, TransformCalendarEvents(events))
}

// Helper functions

func stringToMeetingType(s string) sqlc.Meetingtype {
	typeMap := map[string]sqlc.Meetingtype{
		"team":           sqlc.MeetingtypeTeam,
		"sales":          sqlc.MeetingtypeSales,
		"onboarding":     sqlc.MeetingtypeOnboarding,
		"kickoff":        sqlc.MeetingtypeKickoff,
		"implementation": sqlc.MeetingtypeImplementation,
		"standup":        sqlc.MeetingtypeStandup,
		"retrospective":  sqlc.MeetingtypeRetrospective,
		"planning":       sqlc.MeetingtypePlanning,
		"review":         sqlc.MeetingtypeReview,
		"one_on_one":     sqlc.MeetingtypeOneOnOne,
		"client":         sqlc.MeetingtypeClient,
		"internal":       sqlc.MeetingtypeInternal,
		"external":       sqlc.MeetingtypeExternal,
		"other":          sqlc.MeetingtypeOther,
	}
	if enum, ok := typeMap[strings.ToLower(s)]; ok {
		return enum
	}
	return sqlc.MeetingtypeOther
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
