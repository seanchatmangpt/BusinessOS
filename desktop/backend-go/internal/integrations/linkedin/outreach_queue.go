package linkedin

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// OutreachQueueManager handles message queuing and rate limiting for LinkedIn outreach.
// Rate limit: max 5 messages per contact per day (key: linkedin:rate:{contact_id}:{date}).
type OutreachQueueManager struct {
	logger      *slog.Logger
	redisClient *redis.Client
	repo        *Repository
}

// NewOutreachQueueManager creates a new outreach queue manager.
func NewOutreachQueueManager(logger *slog.Logger, redisClient *redis.Client, repo *Repository) *OutreachQueueManager {
	return &OutreachQueueManager{
		logger:      logger,
		redisClient: redisClient,
		repo:        repo,
	}
}

// EnqueueMessage schedules a message for sending via LinkedIn.
// Returns error if rate limit exceeded.
func (oq *OutreachQueueManager) EnqueueMessage(ctx context.Context, contactID, stepID int64) error {
	// Check rate limit: max 5 messages per contact per day
	if exceeded, err := oq.checkRateLimit(ctx, contactID); err != nil {
		oq.logger.Error("Rate limit check failed", "contact_id", contactID, "error", err)
		return fmt.Errorf("rate limit check failed: %w", err)
	} else if exceeded {
		oq.logger.Warn("Rate limit exceeded", "contact_id", contactID)
		return fmt.Errorf("rate limit exceeded: max 5 messages per contact per day")
	}

	// Create queue entry
	scheduledAt := time.Now().Add(30 * time.Second) // Send in 30 seconds
	msg := &LinkedInMessageQueue{
		ContactID:   contactID,
		StepID:      stepID,
		ScheduledAt: scheduledAt,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := oq.repo.CreateMessage(msg); err != nil {
		oq.logger.Error("Message creation failed", "contact_id", contactID, "error", err)
		return fmt.Errorf("message creation failed: %w", err)
	}

	// Increment rate limit counter
	if err := oq.incrementRateLimit(ctx, contactID); err != nil {
		oq.logger.Error("Rate limit increment failed", "contact_id", contactID, "error", err)
		// Don't fail the request; message is already queued
	}

	oq.logger.Debug("Message enqueued",
		"contact_id", contactID,
		"step_id", stepID,
		"scheduled_at", scheduledAt,
	)

	return nil
}

// EnqueueBatch schedules multiple messages in bulk.
// Respects rate limits per contact.
func (oq *OutreachQueueManager) EnqueueBatch(ctx context.Context, contactIDs []int64, stepID int64) (int, []string) {
	var queued int
	var errors []string

	for _, contactID := range contactIDs {
		if err := oq.EnqueueMessage(ctx, contactID, stepID); err != nil {
			errors = append(errors, fmt.Sprintf("contact_id=%d: %v", contactID, err))
		} else {
			queued++
		}
	}

	oq.logger.Info("Batch enqueue completed",
		"queued", queued,
		"failed", len(errors),
	)

	return queued, errors
}

// checkRateLimit checks if a contact has exceeded the daily message limit (5/day).
// Returns true if limit exceeded, false otherwise.
func (oq *OutreachQueueManager) checkRateLimit(ctx context.Context, contactID int64) (bool, error) {
	now := time.Now()
	dateStr := now.Format("2006-01-02") // YYYY-MM-DD
	key := fmt.Sprintf("linkedin:rate:%d:%s", contactID, dateStr)

	val, err := oq.redisClient.Get(ctx, key).Int64()
	if err != nil && err != redis.Nil {
		return false, fmt.Errorf("redis get failed: %w", err)
	}

	// val is 0 if key doesn't exist (Nil error)
	return val >= 5, nil
}

// incrementRateLimit increments the per-contact daily message counter.
// Counter expires at end of day (TTL = seconds until midnight).
func (oq *OutreachQueueManager) incrementRateLimit(ctx context.Context, contactID int64) error {
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	key := fmt.Sprintf("linkedin:rate:%d:%s", contactID, dateStr)

	// Increment counter
	if err := oq.redisClient.Incr(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis incr failed: %w", err)
	}

	// Set TTL to midnight UTC
	midnight := now.AddDate(0, 0, 1).Truncate(24 * time.Hour)
	ttl := time.Until(midnight)

	if err := oq.redisClient.Expire(ctx, key, ttl).Err(); err != nil {
		oq.logger.Warn("Failed to set TTL on rate limit key", "key", key, "error", err)
		// Don't fail; the counter will expire eventually
	}

	return nil
}

// GetPendingMessages retrieves all messages pending send.
func (oq *OutreachQueueManager) GetPendingMessages(ctx context.Context) ([]*LinkedInMessageQueue, error) {
	return oq.repo.GetPendingMessages()
}

// MarkSent marks a message as successfully sent.
func (oq *OutreachQueueManager) MarkSent(ctx context.Context, messageID int64) error {
	return oq.repo.MarkMessageSent(messageID)
}

// MarkFailed marks a message as failed.
func (oq *OutreachQueueManager) MarkFailed(ctx context.Context, messageID int64, reason string) error {
	return oq.repo.MarkMessageFailed(messageID, reason)
}
