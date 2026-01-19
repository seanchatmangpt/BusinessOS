// Package jobs provides initialization for sync job scheduling.
package jobs

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rhl/businessos-backend/internal/services"
)

// InitSyncJobs initializes and schedules all sync jobs.
func InitSyncJobs(
	ctx context.Context,
	pool *pgxpool.Pool,
	syncService *services.SyncService,
	webhookSubService *services.WebhookSubscriptionService,
	scheduler *services.JobScheduler,
	logger *slog.Logger,
) error {
	// Register job handlers with the background jobs service
	// Note: Job handlers are executed by the worker pool when jobs are dequeued
	// Handler registration happens in the worker initialization (see GetJobHandler function below)

	logger.Info("Registering sync job handlers")

	// Schedule recurring sync jobs using cron expressions
	// These jobs create background jobs that will be executed by workers

	jobs := []struct {
		jobType     string
		cron        string
		name        string
		description string
	}{
		{
			jobType:     JobTypeSyncGoogleCalendar,
			cron:        "*/5 * * * *", // Every 5 minutes
			name:        "Google Calendar Sync",
			description: "Polls Google Calendar for event updates",
		},
		{
			jobType:     JobTypeSyncSlackMessages,
			cron:        "*/2 * * * *", // Every 2 minutes
			name:        "Slack Messages Sync",
			description: "Polls Slack for missed message events",
		},
		{
			jobType:     JobTypeSyncLinearIssues,
			cron:        "* * * * *", // Every 1 minute
			name:        "Linear Issues Sync",
			description: "Polls Linear for fast task updates",
		},
		{
			jobType:     JobTypeSyncHubSpotContacts,
			cron:        "*/5 * * * *", // Every 5 minutes
			name:        "HubSpot Contacts Sync",
			description: "Polls HubSpot for contact and engagement data",
		},
		{
			jobType:     JobTypeSyncNotionPages,
			cron:        "*/2 * * * *", // Every 2 minutes
			name:        "Notion Pages Sync",
			description: "Polls Notion for page updates",
		},
		{
			jobType:     JobTypeSyncClickUpTasks,
			cron:        "*/2 * * * *", // Every 2 minutes
			name:        "ClickUp Tasks Sync",
			description: "Polls ClickUp for task updates",
		},
		{
			jobType:     JobTypeSyncAirtableRecords,
			cron:        "* * * * *", // Every 1 minute
			name:        "Airtable Records Sync",
			description: "Polls Airtable for record changes",
		},
		{
			jobType:     JobTypeSyncFathomMeetings,
			cron:        "*/15 * * * *", // Every 15 minutes
			name:        "Fathom Meetings Sync",
			description: "Polls Fathom for meeting recordings",
		},
		{
			jobType:     JobTypeSyncMicrosoftCalendar,
			cron:        "*/5 * * * *", // Every 5 minutes
			name:        "Microsoft Calendar Sync",
			description: "Polls Microsoft Calendar for event updates",
		},
	}

	// Create scheduled jobs (skip if already exist)
	for _, job := range jobs {
		// Check if job already exists
		existing, err := scheduler.ListScheduledJobs(ctx, true)
		if err != nil {
			logger.Error("Failed to list scheduled jobs", slog.Any("error", err))
			continue
		}

		exists := false
		for _, existingJob := range existing {
			if existingJob.JobType == job.jobType {
				exists = true
				logger.Debug("Scheduled job already exists, skipping",
					slog.String("job_type", job.jobType),
				)
				break
			}
		}

		if !exists {
			// Create the scheduled job
			scheduledJob, err := scheduler.CreateScheduledJob(ctx, services.CreateScheduledJobRequest{
				JobType:        job.jobType,
				Payload:        map[string]interface{}{},
				CronExpression: job.cron,
				Timezone:       "UTC",
				Name:           &job.name,
				Description:    &job.description,
			})
			if err != nil {
				logger.Error("Failed to create scheduled job",
					slog.String("job_type", job.jobType),
					slog.Any("error", err),
				)
				continue
			}

			logger.Info("Scheduled job created",
				slog.String("job_type", job.jobType),
				slog.String("cron", job.cron),
				slog.String("job_id", scheduledJob.ID.String()),
			)
		}
	}

	logger.Info("Sync job handlers registered and scheduled")

	// Note: Job execution handlers need to be registered in the worker
	// This is typically done in cmd/server/main.go or a worker initialization file
	// Example:
	//   worker.RegisterHandler(JobTypeSyncGoogleCalendar, handler.SyncGoogleCalendar)
	//   worker.RegisterHandler(JobTypeSyncSlackMessages, handler.SyncSlackMessages)
	//   etc.

	return nil
}

// GetJobHandler returns a sync job handler instance.
// This is used to register handlers with the worker pool.
func GetJobHandler(
	pool *pgxpool.Pool,
	syncService *services.SyncService,
	webhookSubService *services.WebhookSubscriptionService,
	logger *slog.Logger,
) *SyncJobHandler {
	return NewSyncJobHandler(pool, syncService, webhookSubService, logger)
}
