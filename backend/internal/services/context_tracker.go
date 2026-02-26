package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ContextTrackerService monitors token usage and manages context window eviction
type ContextTrackerService struct {
	db     *sql.DB
	logger *slog.Logger
	mu     sync.RWMutex

	// In-memory LRU cache for active conversations
	activeContexts map[string]*ConversationContext
	maxContexts    int

	// Token budget defaults
	defaultMaxTokens    int
	defaultReserveRatio float64
}

// ConversationContext tracks token usage for a conversation
type ConversationContext struct {
	ConversationID string                 `json:"conversation_id"`
	UserID         string                 `json:"user_id"`
	TotalTokens    int                    `json:"total_tokens"`
	MaxTokens      int                    `json:"max_tokens"`
	ReserveTokens  int                    `json:"reserve_tokens"`
	Blocks         []*ContextBlock        `json:"blocks"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      time.Time              `json:"created_at"`
	LastAccessedAt time.Time              `json:"last_accessed_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// ContextBlock represents a block of content in the context window
type ContextBlock struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // system, user, assistant, memory, document
	Content     string                 `json:"content"`
	TokenCount  int                    `json:"token_count"`
	Priority    int                    `json:"priority"` // Higher = more important
	Pinned      bool                   `json:"pinned"`   // Never evict pinned blocks
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	AccessCount int                    `json:"access_count"`
}

// TokenUsageStats provides statistics about token usage
type TokenUsageStats struct {
	ConversationID   string  `json:"conversation_id"`
	TotalTokens      int     `json:"total_tokens"`
	MaxTokens        int     `json:"max_tokens"`
	UsagePercentage  float64 `json:"usage_percentage"`
	AvailableTokens  int     `json:"available_tokens"`
	ReserveTokens    int     `json:"reserve_tokens"`
	BlockCount       int     `json:"block_count"`
	PinnedTokens     int     `json:"pinned_tokens"`
	EvictableTokens  int     `json:"evictable_tokens"`
	OldestBlockAge   string  `json:"oldest_block_age"`
	EstimatedTurns   int     `json:"estimated_turns"`
}

// EvictionResult contains results of a context eviction
type EvictionResult struct {
	EvictedBlocks  []string `json:"evicted_blocks"`
	TokensFreed    int      `json:"tokens_freed"`
	BlocksRemoved  int      `json:"blocks_removed"`
	NewTotalTokens int      `json:"new_total_tokens"`
}

// ContextTrackerConfig configuration for the service
type ContextTrackerConfig struct {
	MaxActiveContexts   int
	DefaultMaxTokens    int
	DefaultReserveRatio float64
}

// NewContextTrackerService creates a new context tracker service
func NewContextTrackerService(db *sql.DB, logger *slog.Logger, config *ContextTrackerConfig) *ContextTrackerService {
	if config == nil {
		config = &ContextTrackerConfig{
			MaxActiveContexts:   100,
			DefaultMaxTokens:    128000, // Claude's context window
			DefaultReserveRatio: 0.2,    // Reserve 20% for response
		}
	}

	return &ContextTrackerService{
		db:                  db,
		logger:              logger,
		activeContexts:      make(map[string]*ConversationContext),
		maxContexts:         config.MaxActiveContexts,
		defaultMaxTokens:    config.DefaultMaxTokens,
		defaultReserveRatio: config.DefaultReserveRatio,
	}
}

// GetOrCreateContext gets or creates a context for a conversation
func (s *ContextTrackerService) GetOrCreateContext(ctx context.Context, conversationID, userID string) (*ConversationContext, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check in-memory cache first
	if cc, exists := s.activeContexts[conversationID]; exists {
		cc.LastAccessedAt = time.Now()
		return cc, nil
	}

	// Try to load from database
	cc, err := s.loadContextFromDB(ctx, conversationID)
	if err == nil && cc != nil {
		cc.LastAccessedAt = time.Now()
		s.activeContexts[conversationID] = cc
		s.evictLRUIfNeeded()
		return cc, nil
	}

	// Create new context
	reserveTokens := int(float64(s.defaultMaxTokens) * s.defaultReserveRatio)
	cc = &ConversationContext{
		ConversationID: conversationID,
		UserID:         userID,
		TotalTokens:    0,
		MaxTokens:      s.defaultMaxTokens,
		ReserveTokens:  reserveTokens,
		Blocks:         make([]*ContextBlock, 0),
		Metadata:       make(map[string]interface{}),
		CreatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
		UpdatedAt:      time.Now(),
	}

	s.activeContexts[conversationID] = cc
	s.evictLRUIfNeeded()

	// Persist to database
	if err := s.saveContextToDB(ctx, cc); err != nil {
		s.logger.Warn("failed to save context to DB", "error", err)
	}

	return cc, nil
}

// AddBlock adds a new block to the context
func (s *ContextTrackerService) AddBlock(ctx context.Context, conversationID string, block *ContextBlock) (*EvictionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cc, exists := s.activeContexts[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation context not found: %s", conversationID)
	}

	if block.ID == "" {
		block.ID = uuid.New().String()
	}
	block.CreatedAt = time.Now()
	block.AccessCount = 1

	// Check if we need to evict
	availableTokens := cc.MaxTokens - cc.ReserveTokens - cc.TotalTokens
	var evictionResult *EvictionResult

	if block.TokenCount > availableTokens {
		// Need to evict
		tokensNeeded := block.TokenCount - availableTokens + 1000 // Buffer
		evictionResult = s.evictBlocks(cc, tokensNeeded)
	}

	// Add the block
	cc.Blocks = append(cc.Blocks, block)
	cc.TotalTokens += block.TokenCount
	cc.UpdatedAt = time.Now()
	cc.LastAccessedAt = time.Now()

	// Persist
	if err := s.saveContextToDB(ctx, cc); err != nil {
		s.logger.Warn("failed to save context after adding block", "error", err)
	}

	return evictionResult, nil
}

// evictBlocks removes lowest priority blocks until tokensNeeded are freed
func (s *ContextTrackerService) evictBlocks(cc *ConversationContext, tokensNeeded int) *EvictionResult {
	result := &EvictionResult{
		EvictedBlocks: make([]string, 0),
		TokensFreed:   0,
		BlocksRemoved: 0,
	}

	// Sort blocks by eviction priority (lowest priority first, then oldest)
	// Skip pinned blocks
	type evictionCandidate struct {
		index    int
		block    *ContextBlock
		priority int
	}

	candidates := make([]evictionCandidate, 0)
	for i, b := range cc.Blocks {
		if !b.Pinned {
			// Lower priority score = evict first
			// Consider: priority, recency, access count
			priorityScore := b.Priority*1000 + b.AccessCount
			candidates = append(candidates, evictionCandidate{
				index:    i,
				block:    b,
				priority: priorityScore,
			})
		}
	}

	// Sort by priority (ascending - lowest first)
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[i].priority > candidates[j].priority {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	// Evict until we have enough tokens
	evictedIndices := make(map[int]bool)
	for _, candidate := range candidates {
		if result.TokensFreed >= tokensNeeded {
			break
		}

		result.EvictedBlocks = append(result.EvictedBlocks, candidate.block.ID)
		result.TokensFreed += candidate.block.TokenCount
		result.BlocksRemoved++
		evictedIndices[candidate.index] = true
	}

	// Remove evicted blocks
	newBlocks := make([]*ContextBlock, 0, len(cc.Blocks)-result.BlocksRemoved)
	for i, b := range cc.Blocks {
		if !evictedIndices[i] {
			newBlocks = append(newBlocks, b)
		}
	}
	cc.Blocks = newBlocks
	cc.TotalTokens -= result.TokensFreed
	result.NewTotalTokens = cc.TotalTokens

	s.logger.Info("evicted blocks from context",
		"conversation_id", cc.ConversationID,
		"blocks_removed", result.BlocksRemoved,
		"tokens_freed", result.TokensFreed)

	return result
}

// GetTokenUsageStats returns statistics about token usage
func (s *ContextTrackerService) GetTokenUsageStats(ctx context.Context, conversationID string) (*TokenUsageStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cc, exists := s.activeContexts[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation context not found: %s", conversationID)
	}

	availableTokens := cc.MaxTokens - cc.ReserveTokens - cc.TotalTokens
	usagePercentage := float64(cc.TotalTokens) / float64(cc.MaxTokens-cc.ReserveTokens) * 100

	pinnedTokens := 0
	evictableTokens := 0
	var oldestBlock *ContextBlock

	for _, b := range cc.Blocks {
		if b.Pinned {
			pinnedTokens += b.TokenCount
		} else {
			evictableTokens += b.TokenCount
		}
		if oldestBlock == nil || b.CreatedAt.Before(oldestBlock.CreatedAt) {
			oldestBlock = b
		}
	}

	oldestAge := "N/A"
	if oldestBlock != nil {
		oldestAge = time.Since(oldestBlock.CreatedAt).Round(time.Second).String()
	}

	// Estimate remaining turns based on average token usage
	estimatedTurns := 0
	if len(cc.Blocks) > 0 {
		avgTokensPerTurn := cc.TotalTokens / len(cc.Blocks)
		if avgTokensPerTurn > 0 {
			estimatedTurns = availableTokens / avgTokensPerTurn
		}
	}

	return &TokenUsageStats{
		ConversationID:  conversationID,
		TotalTokens:     cc.TotalTokens,
		MaxTokens:       cc.MaxTokens,
		UsagePercentage: usagePercentage,
		AvailableTokens: availableTokens,
		ReserveTokens:   cc.ReserveTokens,
		BlockCount:      len(cc.Blocks),
		PinnedTokens:    pinnedTokens,
		EvictableTokens: evictableTokens,
		OldestBlockAge:  oldestAge,
		EstimatedTurns:  estimatedTurns,
	}, nil
}

// PinBlock marks a block as pinned (never evict)
func (s *ContextTrackerService) PinBlock(ctx context.Context, conversationID, blockID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cc, exists := s.activeContexts[conversationID]
	if !exists {
		return fmt.Errorf("conversation context not found: %s", conversationID)
	}

	for _, b := range cc.Blocks {
		if b.ID == blockID {
			b.Pinned = true
			cc.UpdatedAt = time.Now()
			return s.saveContextToDB(ctx, cc)
		}
	}

	return fmt.Errorf("block not found: %s", blockID)
}

// UnpinBlock removes the pinned status from a block
func (s *ContextTrackerService) UnpinBlock(ctx context.Context, conversationID, blockID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cc, exists := s.activeContexts[conversationID]
	if !exists {
		return fmt.Errorf("conversation context not found: %s", conversationID)
	}

	for _, b := range cc.Blocks {
		if b.ID == blockID {
			b.Pinned = false
			cc.UpdatedAt = time.Now()
			return s.saveContextToDB(ctx, cc)
		}
	}

	return fmt.Errorf("block not found: %s", blockID)
}

// SetBlockPriority updates the priority of a block
func (s *ContextTrackerService) SetBlockPriority(ctx context.Context, conversationID, blockID string, priority int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cc, exists := s.activeContexts[conversationID]
	if !exists {
		return fmt.Errorf("conversation context not found: %s", conversationID)
	}

	for _, b := range cc.Blocks {
		if b.ID == blockID {
			b.Priority = priority
			cc.UpdatedAt = time.Now()
			return s.saveContextToDB(ctx, cc)
		}
	}

	return fmt.Errorf("block not found: %s", blockID)
}

// RemoveBlock removes a specific block from the context
func (s *ContextTrackerService) RemoveBlock(ctx context.Context, conversationID, blockID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cc, exists := s.activeContexts[conversationID]
	if !exists {
		return fmt.Errorf("conversation context not found: %s", conversationID)
	}

	newBlocks := make([]*ContextBlock, 0, len(cc.Blocks))
	var removedTokens int
	found := false

	for _, b := range cc.Blocks {
		if b.ID == blockID {
			removedTokens = b.TokenCount
			found = true
		} else {
			newBlocks = append(newBlocks, b)
		}
	}

	if !found {
		return fmt.Errorf("block not found: %s", blockID)
	}

	cc.Blocks = newBlocks
	cc.TotalTokens -= removedTokens
	cc.UpdatedAt = time.Now()

	return s.saveContextToDB(ctx, cc)
}

// GetBlocks returns all blocks in a context
func (s *ContextTrackerService) GetBlocks(ctx context.Context, conversationID string) ([]*ContextBlock, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cc, exists := s.activeContexts[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation context not found: %s", conversationID)
	}

	// Return a copy to avoid race conditions
	blocks := make([]*ContextBlock, len(cc.Blocks))
	copy(blocks, cc.Blocks)

	return blocks, nil
}

// GetBlocksByType returns blocks of a specific type
func (s *ContextTrackerService) GetBlocksByType(ctx context.Context, conversationID, blockType string) ([]*ContextBlock, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cc, exists := s.activeContexts[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation context not found: %s", conversationID)
	}

	blocks := make([]*ContextBlock, 0)
	for _, b := range cc.Blocks {
		if b.Type == blockType {
			blocks = append(blocks, b)
		}
	}

	return blocks, nil
}

// SetMaxTokens updates the max tokens for a conversation
func (s *ContextTrackerService) SetMaxTokens(ctx context.Context, conversationID string, maxTokens int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cc, exists := s.activeContexts[conversationID]
	if !exists {
		return fmt.Errorf("conversation context not found: %s", conversationID)
	}

	cc.MaxTokens = maxTokens
	cc.ReserveTokens = int(float64(maxTokens) * s.defaultReserveRatio)
	cc.UpdatedAt = time.Now()

	return s.saveContextToDB(ctx, cc)
}

// ClearContext removes all non-pinned blocks from a context
func (s *ContextTrackerService) ClearContext(ctx context.Context, conversationID string) (*EvictionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cc, exists := s.activeContexts[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation context not found: %s", conversationID)
	}

	result := &EvictionResult{
		EvictedBlocks: make([]string, 0),
		TokensFreed:   0,
		BlocksRemoved: 0,
	}

	newBlocks := make([]*ContextBlock, 0)
	for _, b := range cc.Blocks {
		if b.Pinned {
			newBlocks = append(newBlocks, b)
		} else {
			result.EvictedBlocks = append(result.EvictedBlocks, b.ID)
			result.TokensFreed += b.TokenCount
			result.BlocksRemoved++
		}
	}

	cc.Blocks = newBlocks
	cc.TotalTokens -= result.TokensFreed
	result.NewTotalTokens = cc.TotalTokens
	cc.UpdatedAt = time.Now()

	if err := s.saveContextToDB(ctx, cc); err != nil {
		return result, err
	}

	return result, nil
}

// DeleteContext removes a context entirely
func (s *ContextTrackerService) DeleteContext(ctx context.Context, conversationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.activeContexts, conversationID)

	_, err := s.db.ExecContext(ctx,
		`DELETE FROM context_windows WHERE conversation_id = $1`,
		conversationID)

	return err
}

// IncrementBlockAccess increments the access count for a block
func (s *ContextTrackerService) IncrementBlockAccess(ctx context.Context, conversationID, blockID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cc, exists := s.activeContexts[conversationID]
	if !exists {
		return fmt.Errorf("conversation context not found: %s", conversationID)
	}

	for _, b := range cc.Blocks {
		if b.ID == blockID {
			b.AccessCount++
			return nil
		}
	}

	return fmt.Errorf("block not found: %s", blockID)
}

// evictLRUIfNeeded removes least recently used contexts if cache is full
func (s *ContextTrackerService) evictLRUIfNeeded() {
	if len(s.activeContexts) <= s.maxContexts {
		return
	}

	// Find LRU context
	var lruKey string
	var lruTime time.Time

	for key, cc := range s.activeContexts {
		if lruKey == "" || cc.LastAccessedAt.Before(lruTime) {
			lruKey = key
			lruTime = cc.LastAccessedAt
		}
	}

	if lruKey != "" {
		s.logger.Info("evicting LRU context from cache", "conversation_id", lruKey)
		delete(s.activeContexts, lruKey)
	}
}

// loadContextFromDB loads a context from the database
func (s *ContextTrackerService) loadContextFromDB(ctx context.Context, conversationID string) (*ConversationContext, error) {
	var cc ConversationContext
	var blocksJSON, metadataJSON []byte

	err := s.db.QueryRowContext(ctx,
		`SELECT conversation_id, user_id, total_tokens, max_tokens, reserve_tokens,
		        blocks, metadata, created_at, last_accessed_at, updated_at
		 FROM context_windows
		 WHERE conversation_id = $1`,
		conversationID).Scan(
		&cc.ConversationID, &cc.UserID, &cc.TotalTokens, &cc.MaxTokens, &cc.ReserveTokens,
		&blocksJSON, &metadataJSON, &cc.CreatedAt, &cc.LastAccessedAt, &cc.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(blocksJSON, &cc.Blocks); err != nil {
		cc.Blocks = make([]*ContextBlock, 0)
	}
	if err := json.Unmarshal(metadataJSON, &cc.Metadata); err != nil {
		cc.Metadata = make(map[string]interface{})
	}

	return &cc, nil
}

// saveContextToDB persists a context to the database
func (s *ContextTrackerService) saveContextToDB(ctx context.Context, cc *ConversationContext) error {
	blocksJSON, err := json.Marshal(cc.Blocks)
	if err != nil {
		return err
	}
	metadataJSON, err := json.Marshal(cc.Metadata)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO context_windows (conversation_id, user_id, total_tokens, max_tokens, reserve_tokens,
		                              blocks, metadata, created_at, last_accessed_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (conversation_id) DO UPDATE SET
		    total_tokens = EXCLUDED.total_tokens,
		    max_tokens = EXCLUDED.max_tokens,
		    reserve_tokens = EXCLUDED.reserve_tokens,
		    blocks = EXCLUDED.blocks,
		    metadata = EXCLUDED.metadata,
		    last_accessed_at = EXCLUDED.last_accessed_at,
		    updated_at = EXCLUDED.updated_at`,
		cc.ConversationID, cc.UserID, cc.TotalTokens, cc.MaxTokens, cc.ReserveTokens,
		blocksJSON, metadataJSON, cc.CreatedAt, cc.LastAccessedAt, cc.UpdatedAt)

	return err
}

// GetActiveContextCount returns the number of active contexts in memory
func (s *ContextTrackerService) GetActiveContextCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.activeContexts)
}

// EstimateTokens estimates the token count for a string (approximation)
func EstimateTokens(text string) int {
	// Rough estimation: ~4 characters per token for English
	// This is a simplification; actual tokenization depends on the model
	return len(text) / 4
}

// CompactContext optimizes the context by merging similar blocks
func (s *ContextTrackerService) CompactContext(ctx context.Context, conversationID string) (*EvictionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cc, exists := s.activeContexts[conversationID]
	if !exists {
		return nil, fmt.Errorf("conversation context not found: %s", conversationID)
	}

	result := &EvictionResult{
		EvictedBlocks: make([]string, 0),
		TokensFreed:   0,
		BlocksRemoved: 0,
	}

	// Group blocks by type
	blocksByType := make(map[string][]*ContextBlock)
	for _, b := range cc.Blocks {
		blocksByType[b.Type] = append(blocksByType[b.Type], b)
	}

	// For now, just remove duplicate content
	seenContent := make(map[string]bool)
	newBlocks := make([]*ContextBlock, 0)

	for _, b := range cc.Blocks {
		if b.Pinned {
			newBlocks = append(newBlocks, b)
			continue
		}

		contentHash := fmt.Sprintf("%s:%s", b.Type, b.Content[:min(100, len(b.Content))])
		if seenContent[contentHash] {
			result.EvictedBlocks = append(result.EvictedBlocks, b.ID)
			result.TokensFreed += b.TokenCount
			result.BlocksRemoved++
		} else {
			seenContent[contentHash] = true
			newBlocks = append(newBlocks, b)
		}
	}

	cc.Blocks = newBlocks
	cc.TotalTokens -= result.TokensFreed
	result.NewTotalTokens = cc.TotalTokens
	cc.UpdatedAt = time.Now()

	if err := s.saveContextToDB(ctx, cc); err != nil {
		return result, err
	}

	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
