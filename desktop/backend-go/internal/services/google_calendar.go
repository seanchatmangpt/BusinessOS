package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// GoogleCalendarService handles Google Calendar API operations
type GoogleCalendarService struct {
	pool   *pgxpool.Pool
	config *oauth2.Config
}

// NewGoogleCalendarService creates a new Google Calendar service
func NewGoogleCalendarService(pool *pgxpool.Pool) *GoogleCalendarService {
	cfg := config.AppConfig
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURI,
		Scopes: []string{
			calendar.CalendarReadonlyScope,
			calendar.CalendarEventsScope,
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleCalendarService{
		pool:   pool,
		config: oauthConfig,
	}
}

// GetAuthURL returns the Google OAuth URL for user authorization
func (s *GoogleCalendarService) GetAuthURL(state string) string {
	return s.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// ExchangeCode exchanges an authorization code for tokens
func (s *GoogleCalendarService) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.config.Exchange(ctx, code)
}

// SaveToken saves OAuth tokens to the database
func (s *GoogleCalendarService) SaveToken(ctx context.Context, userID string, token *oauth2.Token, email string) error {
	queries := sqlc.New(s.pool)

	scopes := []string{}
	if token.Extra("scope") != nil {
		scopes = append(scopes, token.Extra("scope").(string))
	}

	_, err := queries.CreateGoogleOAuthToken(ctx, sqlc.CreateGoogleOAuthTokenParams{
		UserID:       userID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    &token.TokenType,
		Expiry:       pgtype.Timestamptz{Time: token.Expiry, Valid: true},
		Scopes:       scopes,
		GoogleEmail:  &email,
	})

	return err
}

// UpdateToken updates existing OAuth tokens
func (s *GoogleCalendarService) UpdateToken(ctx context.Context, userID string, token *oauth2.Token) error {
	queries := sqlc.New(s.pool)

	_, err := queries.UpdateGoogleOAuthToken(ctx, sqlc.UpdateGoogleOAuthTokenParams{
		UserID:       userID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       pgtype.Timestamptz{Time: token.Expiry, Valid: true},
	})

	return err
}

// GetToken retrieves OAuth tokens from the database
func (s *GoogleCalendarService) GetToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	queries := sqlc.New(s.pool)

	dbToken, err := queries.GetGoogleOAuthToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  dbToken.AccessToken,
		RefreshToken: dbToken.RefreshToken,
		TokenType:    *dbToken.TokenType,
		Expiry:       dbToken.Expiry.Time,
	}, nil
}

// GetCalendarService creates a Google Calendar API service for a user
func (s *GoogleCalendarService) GetCalendarService(ctx context.Context, userID string) (*calendar.Service, error) {
	token, err := s.GetToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// Create token source that auto-refreshes
	tokenSource := s.config.TokenSource(ctx, token)

	// Check if token was refreshed
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Save refreshed token if it changed
	if newToken.AccessToken != token.AccessToken {
		if err := s.UpdateToken(ctx, userID, newToken); err != nil {
			// Log but don't fail
			fmt.Printf("Warning: failed to save refreshed token: %v\n", err)
		}
	}

	// Create calendar service
	srv, err := calendar.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	return srv, nil
}

// FetchEvents fetches events from Google Calendar
func (s *GoogleCalendarService) FetchEvents(ctx context.Context, userID string, timeMin, timeMax time.Time) ([]*calendar.Event, error) {
	srv, err := s.GetCalendarService(ctx, userID)
	if err != nil {
		return nil, err
	}

	events, err := srv.Events.List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(timeMin.Format(time.RFC3339)).
		TimeMax(timeMax.Format(time.RFC3339)).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}

	return events.Items, nil
}

// SyncEvents syncs events from Google Calendar to the database
func (s *GoogleCalendarService) SyncEvents(ctx context.Context, userID string, timeMin, timeMax time.Time) error {
	events, err := s.FetchEvents(ctx, userID, timeMin, timeMax)
	if err != nil {
		return err
	}

	queries := sqlc.New(s.pool)

	for _, event := range events {
		startTime, endTime := parseEventTimes(event)
		allDay := event.Start.Date != ""

		attendeesJSON, _ := json.Marshal(event.Attendees)

		_, err := queries.UpsertCalendarEvent(ctx, sqlc.UpsertCalendarEventParams{
			UserID:        userID,
			GoogleEventID: &event.Id,
			CalendarID:    stringPtr("primary"),
			Title:         &event.Summary,
			Description:   &event.Description,
			StartTime:     pgtype.Timestamptz{Time: startTime, Valid: true},
			EndTime:       pgtype.Timestamptz{Time: endTime, Valid: true},
			AllDay:        &allDay,
			Location:      &event.Location,
			Attendees:     attendeesJSON,
			Status:        &event.Status,
			Visibility:    &event.Visibility,
			HtmlLink:      &event.HtmlLink,
			Source:        stringPtr("google"),
		})
		if err != nil {
			fmt.Printf("Warning: failed to upsert event %s: %v\n", event.Id, err)
		}
	}

	return nil
}

// CreateEvent creates an event in Google Calendar
func (s *GoogleCalendarService) CreateEvent(ctx context.Context, userID string, event *calendar.Event) (*calendar.Event, error) {
	srv, err := s.GetCalendarService(ctx, userID)
	if err != nil {
		return nil, err
	}

	created, err := srv.Events.Insert("primary", event).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return created, nil
}

// UpdateEvent updates an event in Google Calendar
func (s *GoogleCalendarService) UpdateEvent(ctx context.Context, userID, eventID string, event *calendar.Event) (*calendar.Event, error) {
	srv, err := s.GetCalendarService(ctx, userID)
	if err != nil {
		return nil, err
	}

	updated, err := srv.Events.Update("primary", eventID, event).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return updated, nil
}

// DeleteEvent deletes an event from Google Calendar
func (s *GoogleCalendarService) DeleteEvent(ctx context.Context, userID, eventID string) error {
	srv, err := s.GetCalendarService(ctx, userID)
	if err != nil {
		return err
	}

	if err := srv.Events.Delete("primary", eventID).Do(); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	return nil
}

// DeleteToken removes OAuth tokens for a user
func (s *GoogleCalendarService) DeleteToken(ctx context.Context, userID string) error {
	queries := sqlc.New(s.pool)
	return queries.DeleteGoogleOAuthToken(ctx, userID)
}

// GetConnectionStatus checks if a user has connected their Google account
func (s *GoogleCalendarService) GetConnectionStatus(ctx context.Context, userID string) (*sqlc.GetGoogleOAuthStatusRow, error) {
	queries := sqlc.New(s.pool)
	status, err := queries.GetGoogleOAuthStatus(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// Helper functions

func parseEventTimes(event *calendar.Event) (start, end time.Time) {
	if event.Start.DateTime != "" {
		start, _ = time.Parse(time.RFC3339, event.Start.DateTime)
	} else if event.Start.Date != "" {
		start, _ = time.Parse("2006-01-02", event.Start.Date)
	}

	if event.End.DateTime != "" {
		end, _ = time.Parse(time.RFC3339, event.End.DateTime)
	} else if event.End.Date != "" {
		end, _ = time.Parse("2006-01-02", event.End.Date)
	}

	return start, end
}

func stringPtr(s string) *string {
	return &s
}
