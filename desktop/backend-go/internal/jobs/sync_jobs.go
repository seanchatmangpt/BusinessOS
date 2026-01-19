// Package jobs provides background job handlers for data synchronization.
package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/api/calendar/v3"

	"github.com/rhl/businessos-backend/internal/integrations/google"
	"github.com/rhl/businessos-backend/internal/services"
)

// =============================================================================
// SYNC JOBS
// Background jobs for polling providers to ensure no data is missed
// =============================================================================

// SyncJobHandler handles data synchronization jobs.
type SyncJobHandler struct {
	pool              *pgxpool.Pool
	syncService       *services.SyncService
	webhookSubService *services.WebhookSubscriptionService
	logger            *slog.Logger
}

// NewSyncJobHandler creates a new sync job handler.
func NewSyncJobHandler(
	pool *pgxpool.Pool,
	syncService *services.SyncService,
	webhookSubService *services.WebhookSubscriptionService,
	logger *slog.Logger,
) *SyncJobHandler {
	return &SyncJobHandler{
		pool:              pool,
		syncService:       syncService,
		webhookSubService: webhookSubService,
		logger:            logger,
	}
}

// =============================================================================
// JOB TYPE CONSTANTS
// =============================================================================

const (
	JobTypeSyncGoogleCalendar    = "sync_google_calendar"
	JobTypeSyncSlackMessages     = "sync_slack_messages"
	JobTypeSyncLinearIssues      = "sync_linear_issues"
	JobTypeSyncHubSpotContacts   = "sync_hubspot_contacts"
	JobTypeSyncNotionPages       = "sync_notion_pages"
	JobTypeSyncClickUpTasks      = "sync_clickup_tasks"
	JobTypeSyncAirtableRecords   = "sync_airtable_records"
	JobTypeSyncFathomMeetings    = "sync_fathom_meetings"
	JobTypeSyncMicrosoftCalendar = "sync_microsoft_calendar"
)

// =============================================================================
// GOOGLE CALENDAR SYNC
// =============================================================================

// SyncGoogleCalendar syncs Google Calendar events for all connected users.
func (h *SyncJobHandler) SyncGoogleCalendar(ctx context.Context, payload map[string]interface{}) error {
	h.logger.Info("Starting Google Calendar sync job")

	// Get all users with Google Calendar OAuth tokens
	rows, err := h.pool.Query(ctx, `
		SELECT DISTINCT user_id FROM oauth_tokens
		WHERE provider = 'google'
		  AND deleted_at IS NULL
		  AND expires_at > NOW()
	`)
	if err != nil {
		return fmt.Errorf("query google users: %w", err)
	}
	defer rows.Close()

	syncCount := 0
	errorCount := 0

	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			h.logger.Error("Failed to scan user ID", slog.Any("error", err))
			continue
		}

		// Check if recent webhook was received (skip if yes)
		if h.hasRecentWebhook(ctx, userID, "google", "calendar", 10*time.Minute) {
			h.logger.Debug("Skipping Google Calendar sync (recent webhook)",
				slog.String("user_id", userID.String()),
			)
			continue
		}

		// Sync events for this user
		if err := h.syncGoogleCalendarForUser(ctx, userID); err != nil {
			h.logger.Error("Failed to sync Google Calendar",
				slog.String("user_id", userID.String()),
				slog.Any("error", err),
			)
			errorCount++
		} else {
			syncCount++
		}
	}

	h.logger.Info("Google Calendar sync job complete",
		slog.Int("synced", syncCount),
		slog.Int("errors", errorCount),
	)

	return nil
}

func (h *SyncJobHandler) syncGoogleCalendarForUser(ctx context.Context, userID uuid.UUID) error {
	// Get sync token for incremental sync
	syncToken, err := h.syncService.GetSyncToken(ctx, userID, "google", "calendar")
	if err != nil {
		h.logger.Error("Failed to get sync token", slog.Any("error", err))
	}

	// Initialize Google Calendar service
	googleProvider := google.NewProvider(h.pool, []string{"calendar"})
	calendarSrv := google.NewCalendarService(googleProvider)

	// Get Calendar API client
	calAPI, err := calendarSrv.GetCalendarAPI(ctx, userID.String())
	if err != nil {
		return fmt.Errorf("get calendar API: %w", err)
	}

	// Sync events for next 30 days
	timeMin := time.Now()
	timeMax := time.Now().AddDate(0, 0, 30)

	call := calAPI.Events.List("primary").
		TimeMin(timeMin.Format(time.RFC3339)).
		TimeMax(timeMax.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		MaxResults(250)

	// Use sync token if available
	if syncToken != "" {
		call = call.SyncToken(syncToken)
	}

	events, err := call.Do()
	if err != nil {
		return fmt.Errorf("fetch events: %w", err)
	}

	// Sync each event
	syncedCount := 0
	for _, event := range events.Items {
		if event.Status == "cancelled" {
			continue
		}

		syncedEvent, err := h.convertGoogleCalendarEvent(userID, event)
		if err != nil {
			h.logger.Error("Failed to convert event",
				slog.String("event_id", event.Id),
				slog.Any("error", err),
			)
			continue
		}

		if _, err := h.syncService.UpsertCalendarEvent(ctx, *syncedEvent); err != nil {
			h.logger.Error("Failed to upsert event",
				slog.String("event_id", event.Id),
				slog.Any("error", err),
			)
		} else {
			syncedCount++
		}
	}

	// Save new sync token
	if events.NextSyncToken != "" {
		if err := h.syncService.SaveSyncToken(ctx, userID, "google", "calendar", events.NextSyncToken); err != nil {
			h.logger.Error("Failed to save sync token", slog.Any("error", err))
		}
	}

	h.logger.Info("Synced Google Calendar events",
		slog.String("user_id", userID.String()),
		slog.Int("count", syncedCount),
	)

	return nil
}

func (h *SyncJobHandler) convertGoogleCalendarEvent(userID uuid.UUID, event *calendar.Event) (*services.SyncedCalendarEvent, error) {
	// Parse start and end times
	var startTime, endTime time.Time
	var allDay bool

	if event.Start.DateTime != "" {
		startTime, _ = time.Parse(time.RFC3339, event.Start.DateTime)
		endTime, _ = time.Parse(time.RFC3339, event.End.DateTime)
	} else {
		startTime, _ = time.Parse("2006-01-02", event.Start.Date)
		endTime, _ = time.Parse("2006-01-02", event.End.Date)
		allDay = true
	}

	// Extract meeting link
	meetingLink := ""
	if event.ConferenceData != nil && len(event.ConferenceData.EntryPoints) > 0 {
		for _, ep := range event.ConferenceData.EntryPoints {
			if ep.EntryPointType == "video" {
				meetingLink = ep.Uri
				break
			}
		}
	} else if event.HangoutLink != "" {
		meetingLink = event.HangoutLink
	}

	// Parse attendees
	var attendees []services.CalendarAttendee
	if event.Attendees != nil {
		for _, a := range event.Attendees {
			attendees = append(attendees, services.CalendarAttendee{
				Email:          a.Email,
				Name:           a.DisplayName,
				ResponseStatus: a.ResponseStatus,
			})
		}
	}

	// Extract organizer
	organizerEmail := ""
	if event.Organizer != nil {
		organizerEmail = event.Organizer.Email
	}

	return &services.SyncedCalendarEvent{
		UserID:           userID,
		Provider:         "google",
		ExternalID:       event.Id,
		Title:            event.Summary,
		Description:      event.Description,
		StartTime:        startTime,
		EndTime:          endTime,
		AllDay:           allDay,
		Location:         event.Location,
		Attendees:        attendees,
		OrganizerEmail:   organizerEmail,
		MeetingLink:      meetingLink,
		RecurringEventID: event.RecurringEventId,
		RawData: map[string]interface{}{
			"status":       event.Status,
			"htmlLink":     event.HtmlLink,
			"colorId":      event.ColorId,
			"visibility":   event.Visibility,
			"transparency": event.Transparency,
		},
	}, nil
}

// =============================================================================
// SLACK MESSAGES SYNC
// =============================================================================

// SyncSlackMessages syncs Slack messages for all connected users.
func (h *SyncJobHandler) SyncSlackMessages(ctx context.Context, payload map[string]interface{}) error {
	h.logger.Info("Starting Slack messages sync job")

	// Get all users with Slack OAuth tokens
	rows, err := h.pool.Query(ctx, `
		SELECT DISTINCT user_id FROM oauth_tokens
		WHERE provider = 'slack'
		  AND deleted_at IS NULL
		  AND expires_at > NOW()
	`)
	if err != nil {
		return fmt.Errorf("query slack users: %w", err)
	}
	defer rows.Close()

	syncCount := 0
	errorCount := 0

	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			h.logger.Error("Failed to scan user ID", slog.Any("error", err))
			continue
		}

		// Check if recent webhook was received (skip if yes)
		if h.hasRecentWebhook(ctx, userID, "slack", "message", 5*time.Minute) {
			h.logger.Debug("Skipping Slack sync (recent webhook)",
				slog.String("user_id", userID.String()),
			)
			continue
		}

		// TODO: Implement Slack API polling for missed messages
		// This would use the Slack Web API to fetch recent messages
		h.logger.Debug("Slack message polling not yet implemented",
			slog.String("user_id", userID.String()),
		)
	}

	h.logger.Info("Slack messages sync job complete",
		slog.Int("synced", syncCount),
		slog.Int("errors", errorCount),
	)

	return nil
}

// =============================================================================
// LINEAR ISSUES SYNC
// =============================================================================

// SyncLinearIssues syncs Linear issues for all connected users.
func (h *SyncJobHandler) SyncLinearIssues(ctx context.Context, payload map[string]interface{}) error {
	h.logger.Info("Starting Linear issues sync job")

	// Get all users with Linear OAuth tokens
	rows, err := h.pool.Query(ctx, `
		SELECT DISTINCT user_id FROM oauth_tokens
		WHERE provider = 'linear'
		  AND deleted_at IS NULL
		  AND expires_at > NOW()
	`)
	if err != nil {
		return fmt.Errorf("query linear users: %w", err)
	}
	defer rows.Close()

	syncCount := 0
	errorCount := 0

	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			h.logger.Error("Failed to scan user ID", slog.Any("error", err))
			continue
		}

		// Check if recent webhook was received (skip if yes)
		if h.hasRecentWebhook(ctx, userID, "linear", "issue", 2*time.Minute) {
			h.logger.Debug("Skipping Linear sync (recent webhook)",
				slog.String("user_id", userID.String()),
			)
			continue
		}

		// TODO: Implement Linear GraphQL API polling for issues
		// This would use Linear GraphQL API to fetch recent issue updates
		h.logger.Debug("Linear issue polling not yet implemented",
			slog.String("user_id", userID.String()),
		)
	}

	h.logger.Info("Linear issues sync job complete",
		slog.Int("synced", syncCount),
		slog.Int("errors", errorCount),
	)

	return nil
}

// =============================================================================
// HUBSPOT CONTACTS SYNC
// =============================================================================

// SyncHubSpotContacts syncs HubSpot contacts for all connected users.
func (h *SyncJobHandler) SyncHubSpotContacts(ctx context.Context, payload map[string]interface{}) error {
	h.logger.Info("Starting HubSpot contacts sync job")

	// Get all users with HubSpot OAuth tokens
	rows, err := h.pool.Query(ctx, `
		SELECT DISTINCT user_id FROM oauth_tokens
		WHERE provider = 'hubspot'
		  AND deleted_at IS NULL
		  AND expires_at > NOW()
	`)
	if err != nil {
		return fmt.Errorf("query hubspot users: %w", err)
	}
	defer rows.Close()

	syncCount := 0
	errorCount := 0

	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			h.logger.Error("Failed to scan user ID", slog.Any("error", err))
			continue
		}

		// Check if recent webhook was received (skip if yes)
		if h.hasRecentWebhook(ctx, userID, "hubspot", "contact", 10*time.Minute) {
			h.logger.Debug("Skipping HubSpot sync (recent webhook)",
				slog.String("user_id", userID.String()),
			)
			continue
		}

		// TODO: Implement HubSpot REST API polling for contacts
		// This would use HubSpot CRM API to fetch recent contact updates
		h.logger.Debug("HubSpot contact polling not yet implemented",
			slog.String("user_id", userID.String()),
		)
	}

	h.logger.Info("HubSpot contacts sync job complete",
		slog.Int("synced", syncCount),
		slog.Int("errors", errorCount),
	)

	return nil
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// hasRecentWebhook checks if a webhook was received recently for this user/provider/resource.
func (h *SyncJobHandler) hasRecentWebhook(ctx context.Context, userID uuid.UUID, provider, resourceType string, within time.Duration) bool {
	var lastEventAt *time.Time
	err := h.pool.QueryRow(ctx, `
		SELECT last_event_at
		FROM webhook_subscriptions
		WHERE user_id = $1
		  AND provider = $2
		  AND resource_type = $3
		  AND status = 'active'
	`, userID, provider, resourceType).Scan(&lastEventAt)

	if err != nil || lastEventAt == nil {
		return false
	}

	return time.Since(*lastEventAt) < within
}
