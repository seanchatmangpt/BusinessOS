package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	internalAuth "github.com/rhl/businessos-backend/internal/auth"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/container"
	"github.com/rhl/businessos-backend/internal/database"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/handlers"
	"github.com/rhl/businessos-backend/internal/middleware"
	redisClient "github.com/rhl/businessos-backend/internal/redis"
	"github.com/rhl/businessos-backend/internal/security"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/terminal"
	"github.com/rhl/businessos-backend/internal/workers"
)

// Note: middleware package now provides SessionCache for Redis-backed session validation

func main() {
	// Load .env file first (before config.Load) so os.Getenv works for all services
	// This is optional - in production, env vars are set directly
	if err := godotenv.Load(); err != nil {
		log.Printf("Note: No .env file found (this is fine in production)")
	}

	// Create background context for the application
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// ===== SECURITY INITIALIZATION =====
	// Validate production security configuration
	if err := security.ValidateAndFail(
		cfg.Environment,
		cfg.SecretKey,
		cfg.TokenEncryptionKey,
		cfg.RedisKeyHMACSecret,
	); err != nil {
		log.Fatalf("SECURITY ERROR: %v", err)
	}

	// Initialize token encryption (for OAuth tokens in database)
	if cfg.TokenEncryptionKey != "" {
		if err := security.InitGlobalEncryption(cfg.TokenEncryptionKey); err != nil {
			log.Fatalf("Failed to initialize token encryption: %v", err)
		}
		log.Printf("Token encryption initialized (AES-256-GCM)")
	} else {
		// In development, warn about plaintext storage
		warnings := security.WarnDevelopmentInsecure(cfg.TokenEncryptionKey, cfg.RedisKeyHMACSecret)
		for _, w := range warnings {
			log.Printf("WARNING: %s", w)
		}
	}

	// Generate unique instance ID for pub/sub (avoid message echo)
	instanceID := uuid.New().String()[:8]
	log.Printf("Server instance ID: %s", instanceID)

	// Connect to database (optional in dev)
	var pool *pgxpool.Pool
	dbConnected := false
	var dbErr error
	if cfg.DatabaseRequired {
		pool, err = database.Connect(cfg)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		dbConnected = true
		defer database.Close()
	} else {
		log.Printf("DATABASE_REQUIRED=false: starting without database (degraded mode)")
		p, err := database.Connect(cfg)
		if err != nil {
			dbErr = err
			log.Printf("Database unavailable (continuing): %v", err)
		} else {
			pool = p
			dbConnected = true
			defer database.Close()
		}
	}

	// Create a database/sql wrapper for services that use stdlib APIs.
	var sqlDB *sql.DB
	if dbConnected && pool != nil {
		sqlDB = stdlib.OpenDBFromPool(pool)
		defer sqlDB.Close()
	}

	// Sync YAML templates to database on startup (optional)
	if dbConnected && pool != nil && os.Getenv("SYNC_TEMPLATES_ON_STARTUP") == "true" {
		slog.Info("syncing YAML templates to database")
		templatesDir := os.Getenv("TEMPLATES_DIR")
		if templatesDir == "" {
			templatesDir = "internal/prompts/templates/osa"
		}

		syncService := services.NewTemplateSyncService(pool, slog.Default(), templatesDir)
		syncCtx, cancelSync := context.WithTimeout(ctx, 30*time.Second)
		defer cancelSync()

		if result, err := syncService.SyncTemplates(syncCtx); err != nil {
			slog.Warn("template sync failed", "error", err)
		} else {
			slog.Info("template sync completed",
				"inserted", result.Inserted,
				"updated", result.Updated,
				"errors", len(result.Errors),
			)
		}
	}

	// Connect to Redis (optional - for session storage and horizontal scaling)
	redisConnected := false
	var sessionCache *middleware.SessionCache
	var terminalPubSub *terminal.TerminalPubSub
	if cfg.RedisURL != "" {
		// Configure Redis with security settings
		redisCfg := redisClient.DefaultConfig()
		redisCfg.URL = cfg.RedisURL
		redisCfg.Password = cfg.RedisPassword
		redisCfg.TLSEnabled = cfg.RedisTLSEnabled

		// In development, allow insecure TLS for self-signed certs
		// NEVER do this in production
		if !cfg.IsProduction() && cfg.RedisTLSEnabled {
			redisCfg.TLSInsecure = true
		}

		if err := redisClient.Connect(ctx, redisCfg); err != nil {
			log.Printf("Warning: Redis unavailable: %v", err)
			log.Printf("Sessions will use direct DB auth (not optimal for horizontal scaling)")
		} else {
			redisConnected = true
			log.Printf("Redis connected successfully")
			defer redisClient.Close()

			// Create session cache for auth middleware with secure HMAC key
			sessionCacheConfig := &middleware.SessionCacheConfig{
				KeyPrefix:  "auth_session:",
				TTL:        15 * time.Minute,
				HMACSecret: cfg.RedisKeyHMACSecret, // Load from environment (CRITICAL for production)
			}
			var err error
			sessionCache, err = middleware.NewSessionCache(redisClient.Client(), sessionCacheConfig)
			if err != nil {
				log.Printf("Warning: Session cache initialization failed: %v", err)
				log.Printf("Sessions will use direct DB auth (not optimal for horizontal scaling)")
				sessionCache = nil
			} else {
				log.Printf("Session cache enabled (TTL=15m, HMAC-secured keys)")
			}

			// Create terminal pub/sub for horizontal scaling
			terminalPubSub = terminal.NewTerminalPubSub(redisClient.Client(), instanceID)
			log.Printf("Terminal pub/sub enabled (instance=%s)", instanceID)
		}
	}

	// Initialize container manager (optional - for Docker-based terminal isolation)
	var containerMonitor *container.ContainerMonitor
	containerMgr, err := container.NewContainerManager(ctx, "businessos-workspace:latest")
	if err != nil {
		log.Printf("Warning: Container manager unavailable: %v", err)
		log.Printf("Terminal will use local PTY mode")
		containerMgr = nil
	} else {
		log.Printf("Container manager initialized successfully")

		// Initialize and start container monitor for cleanup and health checks
		containerMonitor = container.NewContainerMonitor(containerMgr, nil)
		if err := containerMonitor.StartMonitoring(ctx); err != nil {
			log.Printf("Warning: Container monitor failed to start: %v", err)
		} else {
			log.Printf("Container monitor started (cleanup=%v, idle_timeout=%v)",
				container.DefaultMonitorConfig().CleanupInterval,
				container.DefaultMonitorConfig().IdleTimeout)
		}
	}

	// Create gin router
	router := gin.Default()

	// Limit request body to 10MB to prevent DoS via oversized payloads
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20) // 10MB max
		c.Next()
	})

	// Apply middleware
	router.Use(middleware.CORSMiddleware(cfg))

	// Apply security headers middleware (OWASP A05: Security Misconfiguration)
	// Sets X-Frame-Options, X-Content-Type-Options, HSTS, CSP, and other protective headers
	router.Use(middleware.SecurityHeaders(cfg))
	log.Printf("Security headers enabled (X-Frame-Options, CSP, HSTS, etc.)")

	// Apply CSRF protection (double submit cookie pattern)
	// Protects against CSRF attacks by validating token in cookie matches token in header
	// Works in combination with SameSite=Strict cookies for defense-in-depth
	csrfConfig := middleware.DefaultCSRFConfig()
	// In development (HTTP), disable Secure flag so browsers accept the cookie
	if !cfg.IsProduction() {
		csrfConfig.CookieSecure = false
		csrfConfig.CookieDomain = os.Getenv("COOKIE_DOMAIN")
		log.Printf("CSRF cookie Secure flag disabled for development (HTTP), domain set to: %s", csrfConfig.CookieDomain)
	}
	csrfConfig.Skipper = func(c *gin.Context) bool {
		path := c.Request.URL.Path
		// Skip CSRF for webhooks (third-party services can't set custom headers)
		if strings.HasPrefix(path, "/webhooks/") || strings.HasPrefix(path, "/api/webhooks/") || strings.HasPrefix(path, "/api/v1/webhooks/") {
			return true
		}
		// Skip CSRF for health checks (monitoring tools)
		if path == "/health" || path == "/ready" || path == "/health/detailed" {
			return true
		}
		return false
	}
	router.Use(middleware.CSRF(csrfConfig))
	log.Printf("CSRF protection enabled (double submit cookie pattern; excluded: webhooks, health)")

	// Apply global rate limiting (100 req/sec per IP, 200 req/sec per authenticated user)
	globalRateLimiter := middleware.GetGlobalHTTPRateLimiter()
	router.Use(middleware.RateLimitMiddleware(globalRateLimiter))
	log.Printf("Rate limiting enabled (100 req/s per IP, 200 req/s per user)")

	// Health check (no auth required)
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Business OS API", "version": "1.0.0", "instance_id": instanceID})
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Readiness check (includes dependencies)
	router.GET("/ready", func(c *gin.Context) {
		dbStatus := "disconnected"
		if dbConnected {
			dbStatus = "connected"
		} else if !cfg.DatabaseRequired {
			dbStatus = "disabled"
		}
		status := gin.H{
			"status":      "ready",
			"instance_id": instanceID,
			"database":    dbStatus,
			"redis":       "disconnected",
			"containers":  "unavailable",
		}

		// Check Redis
		if redisConnected && redisClient.IsConnected(c.Request.Context()) {
			status["redis"] = "connected"
		}

		// Check container manager
		if containerMgr != nil {
			status["containers"] = "available"
		}

		c.JSON(200, status)
	})

	// Detailed health check for monitoring
	router.GET("/health/detailed", func(c *gin.Context) {
		health := gin.H{
			"status":      "healthy",
			"instance_id": instanceID,
			"components":  gin.H{},
		}

		components := health["components"].(gin.H)

		// Database health
		dbComponent := gin.H{}
		if dbConnected {
			dbComponent["status"] = "connected"
		} else if !cfg.DatabaseRequired {
			dbComponent["status"] = "disabled"
			if dbErr != nil {
				dbComponent["error"] = dbErr.Error()
			}
		} else {
			dbComponent["status"] = "disconnected"
			if dbErr != nil {
				dbComponent["error"] = dbErr.Error()
			}
		}
		components["database"] = dbComponent

		// Redis health
		if redisConnected {
			redisHealth, err := redisClient.HealthCheck(c.Request.Context())
			if err != nil {
				components["redis"] = gin.H{"status": "error", "error": err.Error()}
			} else {
				components["redis"] = gin.H{
					"status":     "connected",
					"latency_ms": redisHealth.Latency.Milliseconds(),
					"pool_stats": redisHealth.PoolStats,
				}
			}
		} else {
			components["redis"] = gin.H{"status": "not_configured"}
		}

		// Container manager health
		if containerMgr != nil {
			components["containers"] = gin.H{"status": "available"}
		} else {
			components["containers"] = gin.H{"status": "unavailable"}
		}

		c.JSON(200, health)
	})

	// Serve uploaded files (profile photos, backgrounds) - no auth needed
	router.Static("/uploads", "./uploads")

	// ==================================================
	// API Versioning Setup
	// ==================================================
	// Initialize versioning configuration
	versioningConfig := middleware.DefaultVersioningConfig()
	v1Config := versioningConfig.Versions["v1"]

	// V1 API routes (current stable version)
	apiv1 := router.Group("/api/v1")
	apiv1.Use(middleware.DeprecationHeaders(v1Config))

	// Backward compatibility: /api/* -> /api/v1/* (deprecated)
	// This allows existing clients to work while we transition
	api := router.Group("/api")
	api.Use(middleware.VersionRedirect("v1", false))

	// If DB isn't available, only expose basic health endpoints.
	if !dbConnected || pool == nil {
		api.GET("/status", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":            "degraded",
				"database":          "unavailable",
				"database_required": cfg.DatabaseRequired,
			})
		})

		log.Printf("Backend running in degraded mode (no database): only /, /health, /ready, /health/detailed, /api/status, and /uploads are available")
		// Start server
		go func() {
			log.Printf("Server starting on port %s", cfg.ServerPort)
			if err := router.Run(":" + cfg.ServerPort); err != nil {
				log.Fatalf("Failed to start server: %v", err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")
		if containerMonitor != nil {
			log.Println("Stopping container monitor...")
			if err := containerMonitor.StopMonitoring(); err != nil {
				log.Printf("Warning: Error stopping container monitor: %v", err)
			}
		}
		if containerMgr != nil {
			log.Println("Shutting down container manager...")
			containerMgr.Shutdown()
		}
		database.Close()
		log.Println("Server stopped")
		return
	}

	// Initialize embedding service for RAG (semantic search)
	var embeddingService *services.EmbeddingService
	var contextBuilder *services.ContextBuilder
	var tieredContextService *services.TieredContextService
	embeddingService = services.NewEmbeddingService(pool, cfg.OllamaLocalURL)
	if embeddingService.HealthCheck(ctx) {
		contextBuilder = services.NewContextBuilder(pool, embeddingService)
		summarizer := services.NewSummarizerService(pool, cfg)
		tieredContextService = services.NewTieredContextService(pool, embeddingService, summarizer)
		log.Printf("Embedding service initialized (model=nomic-embed-text, dimensions=768)")
		log.Printf("Tiered context service enabled (scoped RAG, Level 1/2/3 context)")
	} else {
		log.Printf("Warning: Embedding service unavailable (Ollama not running or nomic-embed-text model not pulled)")
		log.Printf("RAG features will be disabled. Run: ollama pull nomic-embed-text")
		embeddingService = nil
	}

	// Initialize notification service with SSE broadcaster
	sseBroadcaster := services.NewSSEBroadcaster()
	notificationService := services.NewNotificationService(pool, sseBroadcaster)
	log.Printf("Notification service initialized (SSE real-time enabled)")

	// Start notification batch worker
	batchWorker := workers.NewBatchWorker(pool, notificationService.Dispatcher())
	go batchWorker.Start(ctx)
	log.Printf("Notification batch worker started (interval: 10s)")

	// Initialize Web Push service (optional - requires VAPID keys)
	var webPushService *services.WebPushService
	if cfg.VAPIDPublicKey != "" && cfg.VAPIDPrivateKey != "" {
		webPushService = services.NewWebPushService(pool, &services.WebPushConfig{
			VAPIDPublicKey:  cfg.VAPIDPublicKey,
			VAPIDPrivateKey: cfg.VAPIDPrivateKey,
			VAPIDContact:    cfg.VAPIDContact,
		})
		log.Printf("Web Push service initialized (VAPID keys configured)")
	} else {
		log.Printf("Web Push service disabled (VAPID keys not configured)")
		log.Printf("To enable: Generate keys with `npx web-push generate-vapid-keys` and set VAPID_PUBLIC_KEY, VAPID_PRIVATE_KEY")
	}

	// Initialize AI services (Memory, Context & Intelligence System)
	var documentProcessor *services.DocumentProcessor
	var learningService *services.LearningService
	var memoryService *services.MemoryService
	var autoLearningTriggers *services.AutoLearningTriggers
	var conversationIntelligence *services.ConversationIntelligenceService
	var memoryExtractor *services.MemoryExtractorService

	// Document Processor - requires embedding service for semantic search
	if embeddingService != nil {
		documentProcessor = services.NewDocumentProcessor(pool, embeddingService, "./uploads/documents")
		log.Printf("Document processor initialized (chunking + semantic search)")
	}

	// Learning Service - always available
	learningService = services.NewLearningService(pool)
	log.Printf("Learning service initialized (feedback + personalization)")

	// Memory Service - requires embedding service
	if embeddingService != nil {
		memoryService = services.NewMemoryService(pool, embeddingService)
		log.Printf("Memory service initialized (memory persistence)")
	}

	// Auto-Learning Triggers - requires learning, memory, and embedding services
	if learningService != nil && memoryService != nil && embeddingService != nil {
		autoLearningTriggers = services.NewAutoLearningTriggers(learningService, memoryService, embeddingService)
		log.Printf("Auto-learning triggers initialized (automatic pattern extraction)")
	}

	// Prompt Personalizer - requires pool, learning, memory, and embedding services
	var promptPersonalizer *services.PromptPersonalizer
	if pool != nil && learningService != nil && memoryService != nil && embeddingService != nil {
		promptPersonalizer = services.NewPromptPersonalizer(pool, learningService, memoryService, embeddingService)
		log.Printf("Prompt personalizer initialized (user-specific context injection)")
	}

	// ============================================================
	// Advanced RAG Services
	// ============================================================

	// Hybrid Search Service - requires pool and embedding service
	var hybridSearchService *services.HybridSearchService
	if pool != nil && embeddingService != nil {
		hybridSearchService = services.NewHybridSearchService(pool, embeddingService)
		log.Printf("Hybrid search service initialized (semantic + keyword with RRF)")
	}

	// Re-Ranker Service - requires pool and embedding service
	var rerankerService *services.ReRankerService
	if pool != nil && embeddingService != nil {
		rerankerService = services.NewReRankerService(pool, embeddingService)
		log.Printf("Re-ranker service initialized (multi-signal relevance scoring)")
	}

	// Agentic RAG Service - requires all RAG components
	var agenticRAGService *services.AgenticRAGService
	if pool != nil && hybridSearchService != nil && rerankerService != nil && embeddingService != nil && learningService != nil {
		agenticRAGService = services.NewAgenticRAGService(pool, hybridSearchService, rerankerService, embeddingService, learningService)
		log.Printf("Agentic RAG service initialized (intelligent adaptive retrieval)")
	}

	// ============================================================
	// Day 3: Performance Optimization (Caching + Query Expansion)
	// ============================================================

	// Embedding Cache Service - dedicated Redis cache for embeddings (new)
	var embeddingCache *services.EmbeddingCacheService
	var embeddingCacheAdapter *services.EmbeddingCacheAdapter
	if redisConnected && redisClient.Client() != nil {
		embeddingCacheConfig := services.DefaultEmbeddingCacheConfig()
		embeddingCache = services.NewEmbeddingCacheService(redisClient.Client(), pool, embeddingCacheConfig)
		embeddingCacheAdapter = services.NewEmbeddingCacheAdapter(embeddingCache)

		if embeddingCache.IsEnabled() {
			log.Printf("Embedding cache service initialized (24h text, 48h images)")

			// Connect embedding cache to embedding service
			if embeddingService != nil {
				embeddingService.SetEmbeddingCache(embeddingCacheAdapter)
				log.Printf("Embedding service now using dedicated embedding cache")
			}
		} else {
			log.Printf("Embedding cache disabled (Redis unavailable)")
		}
	}

	// RAG Cache Service - requires Redis for caching queries and embeddings (legacy)
	var ragCache *services.RAGCacheService
	if redisConnected && redisClient.Client() != nil {
		cacheConfig := services.DefaultRAGCacheConfig()
		ragCache = services.NewRAGCacheService(redisClient.Client(), cacheConfig)
		log.Printf("RAG cache service initialized (15min queries, 24hr embeddings - legacy)")

		// Connect legacy cache to embedding service as fallback
		if embeddingService != nil {
			embeddingService.SetCache(ragCache)
			log.Printf("Embedding service legacy cache enabled (fallback)")
		}

		// Connect cache to agentic RAG service for query result caching
		if agenticRAGService != nil {
			agenticRAGService.SetCache(ragCache)
			log.Printf("Agentic RAG cache enabled")
		}
	} else {
		log.Printf("RAG cache disabled (Redis not available)")
	}

	// Query Expansion Service - requires no dependencies, optional LLM integration
	var queryExpansion *services.QueryExpansionService
	queryExpansion = services.NewQueryExpansionService(nil) // nil = no LLM rewriting (can add later)
	log.Printf("Query expansion service initialized (60+ synonym mappings)")

	// Connect query expansion to agentic RAG for enhanced retrieval
	if queryExpansion != nil && agenticRAGService != nil {
		agenticRAGService.SetQueryExpansion(queryExpansion)
		log.Printf("Agentic RAG query expansion enabled")
	}

	// Image Embedding Service for Multi-modal Search (CLIP integration)
	var imageEmbeddingService *services.ImageEmbeddingService
	var multiModalSearchService *services.MultiModalSearchService

	// Check for CLIP provider configuration
	clipProvider := os.Getenv("CLIP_PROVIDER") // "openai", "replicate", "local"
	if clipProvider == "" {
		clipProvider = "local" // Default to local if not specified
	}

	imageEmbedConfig := services.ImageEmbeddingConfig{
		Provider:     clipProvider,
		APIKey:       os.Getenv("CLIP_API_KEY"),
		ModelName:    "clip-vit-base-patch32",
		Dimensions:   512,
		LocalBaseURL: os.Getenv("CLIP_LOCAL_URL"), // e.g., "http://localhost:8000"
	}

	// Only initialize if provider is configured or local server is available
	if clipProvider == "local" && imageEmbedConfig.LocalBaseURL == "" {
		log.Printf("Image embedding service disabled (CLIP_LOCAL_URL not set)")
		log.Printf("To enable: Set CLIP_LOCAL_URL=http://localhost:8000 and run CLIP server")
	} else if (clipProvider == "openai" || clipProvider == "replicate") && imageEmbedConfig.APIKey == "" {
		log.Printf("Image embedding service disabled (CLIP_API_KEY not set for %s)", clipProvider)
	} else {
		imageEmbeddingService = services.NewImageEmbeddingService(pool, imageEmbedConfig)
		log.Printf("Image embedding service initialized (provider=%s, model=%s)", clipProvider, imageEmbedConfig.ModelName)

		// Connect embedding cache to image embedding service
		if embeddingCacheAdapter != nil {
			imageEmbeddingService.SetEmbeddingCache(embeddingCacheAdapter)
			log.Printf("Image embedding service cache enabled (48h TTL)")
		}

		// Multi-modal Search Service - combines text + image search
		if hybridSearchService != nil && rerankerService != nil && embeddingService != nil {
			multiModalSearchService = services.NewMultiModalSearchService(
				pool,
				hybridSearchService,
				rerankerService,
				imageEmbeddingService,
				embeddingService,
			)
			log.Printf("Multi-modal search service initialized (text + image + cross-modal)")
			log.Printf("Feature 7 (Multi-modal Embeddings) complete: SearchWithImage ready!")
		}
	}

	// ============================================================

	// Conversation Intelligence - requires embedding service
	if embeddingService != nil {
		conversationIntelligence = services.NewConversationIntelligenceService(pool, embeddingService)
		log.Printf("Conversation intelligence initialized (analysis + summarization)")
	}

	// Memory Extractor - requires embedding service
	if embeddingService != nil {
		memoryExtractor = services.NewMemoryExtractorService(pool, embeddingService)

		// Wire LLM service for enhanced memory extraction (using Groq for speed/cost)
		if cfg.GroqAPIKey != "" {
			groqLLM := services.NewGroqService(cfg, "llama-3.1-8b-instant") // Fast model for extraction
			if groqLLM.HealthCheck(ctx) {
				memoryExtractor.SetLLMService(groqLLM)
				log.Printf("Memory extractor initialized with LLM-enhanced extraction (Groq llama-3.1-8b-instant)")
			} else {
				log.Printf("Memory extractor initialized (regex-only, Groq unavailable)")
			}
		} else {
			log.Printf("Memory extractor initialized (regex-only, no Groq API key)")
		}
	}

	// Initialize handlers with container manager, session cache, terminal pub/sub, embedding services, and notification service
	h := handlers.NewHandlers(pool, cfg, containerMgr, sessionCache, terminalPubSub, embeddingService, contextBuilder, tieredContextService, notificationService)

	// Set optional services
	if webPushService != nil {
		h.SetWebPushService(webPushService)
	}

	// Initialize Email service (Resend)
	emailService := services.NewEmailService()
	if emailService.IsEnabled() {
		h.SetEmailService(emailService)
		log.Printf("Email service initialized (Resend configured)")
	} else {
		log.Printf("Email service disabled (RESEND_API_KEY not set)")
	}

	// Initialize Comment service
	queries := sqlc.New(pool)
	commentService := services.NewCommentService(pool, queries, notificationService)
	h.SetCommentService(commentService)
	log.Printf("Comment service initialized (mentions & notifications enabled)")

	// Set AI services
	h.SetAIServices(documentProcessor, learningService, autoLearningTriggers, promptPersonalizer, conversationIntelligence, memoryExtractor)
	log.Printf("AI services registered (documents, learning, personalizer, intelligence, memory)")

	// Set RAG services
	h.SetRAGServices(hybridSearchService, rerankerService, agenticRAGService, memoryService)
	log.Printf("RAG services registered (hybrid search, re-ranker, agentic RAG, memory)")

	// Set Multi-modal Search services (Feature 7 - Multi-modal Embeddings)
	if multiModalSearchService != nil && imageEmbeddingService != nil {
		h.SetMultiModalServices(multiModalSearchService, imageEmbeddingService)
		log.Printf("Multi-modal services registered (image embeddings, text+image search)")
	}

	// Initialize Voice services (3D Desktop - Whisper + ElevenLabs)
	whisperService := services.NewWhisperService()
	elevenLabsService := services.NewElevenLabsService()
	h.SetVoiceServices(whisperService, elevenLabsService)
	if whisperService.IsAvailable() {
		log.Printf("Whisper service initialized (local speech-to-text)")
	} else {
		log.Printf("Whisper service not fully configured (model/binary not found)")
	}
	if elevenLabsService.IsConfigured() {
		log.Printf("ElevenLabs service initialized (OSA voice enabled)")
	} else {
		log.Printf("ElevenLabs service not configured (API key/voice ID not set)")
	}

	// Set Workspace service (Feature 1 - Team/Collaboration)
	workspaceService := services.NewWorkspaceService(pool)
	h.SetWorkspaceService(workspaceService)
	log.Printf("Workspace service registered (workspaces, members, roles)")

	// Set Workspace Version service (Feature 1 - Versioning & Snapshots)
	workspaceVersionService := services.NewWorkspaceVersionService(pool, slog.Default())
	h.SetWorkspaceVersionService(workspaceVersionService)
	log.Printf("Workspace version service registered (snapshots, restore)")

	// Set Role Context service (Feature 1 - Permission system)
	roleContextService := services.NewRoleContextService(pool)
	h.SetRoleContextService(roleContextService)
	log.Printf("Role context service registered (permission checks, hierarchy)")

	// Set Memory Hierarchy service (Q1 - Workspace Memory Management)
	memoryHierarchyService := services.NewMemoryHierarchyService(pool)
	h.SetMemoryHierarchyService(memoryHierarchyService)
	log.Printf("Memory hierarchy service registered (workspace memories)")

	// Set Workspace Invite service (Feature 1 - Email Invitations)
	inviteService := services.NewWorkspaceInviteService(pool)
	h.SetInviteService(inviteService)
	log.Printf("Workspace invite service registered (email invitations)")

	// Set Workspace Audit service (Feature 1 - Audit Logging)
	auditService := services.NewWorkspaceAuditService(pool)
	h.SetAuditService(auditService)
	log.Printf("Workspace audit service registered (audit logging)")

	// Set Project Access service (Feature 1 - Project-level Access Control)
	projectAccessService := services.NewProjectAccessService(pool)
	h.SetProjectAccessService(projectAccessService)
	log.Printf("Project access service registered (project member management)")

	// Initialize Skills Loader (Agent Skills System)
	skillsConfigPath := "./skills/skills.yaml"
	skillsLoader := services.NewSkillsLoader(skillsConfigPath)
	if err := skillsLoader.LoadConfig(); err != nil {
		log.Printf("Warning: Skills loader failed to initialize: %v", err)
		log.Printf("Agent skills system will be disabled")
	} else {
		h.SetSkillsLoader(skillsLoader)
		log.Printf("Skills loader initialized (%d skills loaded)", len(skillsLoader.GetEnabledSkills()))
	}

	// Optional background job: keep conversation_summaries fresh for context + semantic search.
	if conversationIntelligence != nil && cfg.ConversationSummaryJobEnabled {
		interval := time.Duration(cfg.ConversationSummaryJobIntervalMinutes) * time.Minute
		if interval <= 0 {
			interval = 30 * time.Minute
		}
		batchSize := cfg.ConversationSummaryJobBatchSize
		if batchSize <= 0 {
			batchSize = 25
		}
		maxMessages := cfg.ConversationSummaryJobMaxMessages
		if maxMessages <= 0 {
			maxMessages = 200
		}

		go func() {
			t := time.NewTicker(interval)
			defer t.Stop()
			log.Printf("Conversation summary job enabled (interval=%s batch=%d maxMessages=%d)", interval, batchSize, maxMessages)
			for {
				select {
				case <-t.C:
					count, err := conversationIntelligence.BackfillStaleSummaries(ctx, batchSize, maxMessages, false)
					if err != nil {
						log.Printf("Conversation summary job error: %v", err)
						continue
					}
					if count > 0 {
						log.Printf("Conversation summary job updated %d conversations", count)
					}
				}
			}
		}()
	}

	// Optional background job: detect behavior patterns and store them as user_facts for explicit confirmation.
	if learningService != nil && cfg.BehaviorPatternsJobEnabled {
		interval := time.Duration(cfg.BehaviorPatternsJobIntervalMinutes) * time.Minute
		if interval <= 0 {
			interval = 60 * time.Minute
		}
		batchSize := cfg.BehaviorPatternsJobUserBatchSize
		if batchSize <= 0 {
			batchSize = 50
		}

		go func() {
			t := time.NewTicker(interval)
			defer t.Stop()
			log.Printf("Behavior patterns job enabled (interval=%s userBatch=%d)", interval, batchSize)
			for {
				select {
				case <-t.C:
					usersProcessed, factsUpserted, err := learningService.BackfillRecentUsersBehaviorPatterns(ctx, batchSize)
					if err != nil {
						log.Printf("Behavior patterns job error: %v", err)
						continue
					}
					if usersProcessed > 0 || factsUpserted > 0 {
						log.Printf("Behavior patterns job processed %d users, upserted %d pattern facts", usersProcessed, factsUpserted)
					}
				}
			}
		}()
	}

	// Initialize Background Jobs System
	var jobsHandler *handlers.BackgroundJobsHandler
	var jobWorkers []*services.JobWorker
	var jobScheduler *services.JobScheduler

	if dbConnected && pool != nil {
		slog.Info("Initializing background jobs system...")

		// Create handler (includes service and scheduler)
		jobsHandler = handlers.NewBackgroundJobsHandler(pool)

		// Get service and scheduler instances
		jobsService := jobsHandler.GetService()
		jobScheduler = jobsHandler.GetScheduler()

		// Create and configure workers (3 workers)
		for i := 1; i <= 3; i++ {
			workerID := fmt.Sprintf("worker-%d", i)
			worker := services.NewJobWorker(jobsService, workerID, 5*time.Second)

			// Register example job handlers
			worker.RegisterHandler("email_send", services.ExampleEmailSendHandler)
			worker.RegisterHandler("report_generate", services.ExampleReportGenerateHandler)
			worker.RegisterHandler("sync_calendar", services.ExampleSyncCalendarHandler)

			// Register all custom production handlers
			worker.RegisterHandler("user_onboarding", handlers.UserOnboardingHandler)
			worker.RegisterHandler("workspace_export", handlers.WorkspaceExportHandler)
			worker.RegisterHandler("analytics_aggregation", handlers.AnalyticsAggregationHandler)
			worker.RegisterHandler("notification_batch", handlers.NotificationBatchHandler)
			worker.RegisterHandler("data_cleanup", handlers.DataCleanupHandler)
			worker.RegisterHandler("integration_sync", handlers.IntegrationSyncHandler)
			worker.RegisterHandler("backup", handlers.BackupHandler)

			jobWorkers = append(jobWorkers, worker)

			// Start worker
			if err := worker.Start(ctx); err != nil {
				slog.Error("Failed to start worker", "worker_id", workerID, "error", err)
			} else {
				slog.Info("Worker started", "worker_id", workerID)
			}
		}

		// Start scheduler
		if err := jobScheduler.Start(ctx); err != nil {
			slog.Error("Failed to start scheduler", "error", err)
		} else {
			slog.Info("Job scheduler started")
		}
	}

	// Register routes on BOTH /api (deprecated) and /api/v1 (current)
	// This ensures backward compatibility during transition period
	h.RegisterRoutes(api)   // Deprecated path with warning headers
	h.RegisterRoutes(apiv1) // Current versioned path

	// ─── Three-Tier Auth Routes ────────────────────────────────────────────
	// Resolve the active auth mode from config and wire the new auth handlers.
	{
		activeAuthMode := internalAuth.ParseAuthMode(cfg.AuthMode)
		log.Printf("Auth mode: %s", activeAuthMode)

		// In single-user mode: ensure the owner user and session exist.
		var singleSession *internalAuth.SingleUserSession
		if activeAuthMode == internalAuth.AuthModeSingle {
			var ssErr error
			singleSession, ssErr = internalAuth.EnsureSingleUser(ctx, pool)
			if ssErr != nil {
				log.Fatalf("Failed to initialise single-user session: %v", ssErr)
			}
			log.Printf("Single-user mode: permanent owner session ready")
		}

		// Public auth endpoints (no auth middleware required).
		authSetupHandler := handlers.NewAuthSetupHandler(pool, cfg, activeAuthMode)
		publicAuth := router.Group("/api/auth")
		publicAuth.GET("/mode", authSetupHandler.GetAuthMode)
		publicAuth.POST("/setup", authSetupHandler.CompleteSetup)
		publicAuth.GET("/invites/:token", authSetupHandler.ValidateInvite)

		// Same routes on v1.
		publicAuthV1 := router.Group("/api/v1/auth")
		publicAuthV1.GET("/mode", authSetupHandler.GetAuthMode)
		publicAuthV1.POST("/setup", authSetupHandler.CompleteSetup)
		publicAuthV1.GET("/invites/:token", authSetupHandler.ValidateInvite)

		// Protected invite creation endpoint.
		if activeAuthMode != internalAuth.AuthModeSingle {
			authMw := middleware.CachedAuthMiddleware(pool, sessionCache)
			protectedAuth := router.Group("/api/auth", authMw)
			protectedAuth.POST("/invites", authSetupHandler.CreateInvite)
			protectedAuthV1 := router.Group("/api/v1/auth", authMw)
			protectedAuthV1.POST("/invites", authSetupHandler.CreateInvite)
		}

		// GitHub OAuth routes (registered only when credentials are configured).
		githubHandler := handlers.NewGitHubAuthHandler(pool, cfg, sessionCache)
		if githubHandler != nil {
			router.GET("/api/v1/auth/github", githubHandler.InitiateGitHubLogin)
			router.GET("/api/v1/auth/github/callback", githubHandler.HandleGitHubCallback)
			router.GET("/api/auth/github", githubHandler.InitiateGitHubLogin)
			router.GET("/api/auth/github/callback", githubHandler.HandleGitHubCallback)
			log.Printf("GitHub OAuth routes registered (/api/v1/auth/github)")
		} else {
			log.Printf("GitHub OAuth disabled (GITHUB_CLIENT_ID not configured)")
		}

		// Log the mode-aware middleware status.
		_ = singleSession // Used inside auth.ModeMiddleware if needed
		log.Printf("Auth middleware: mode=%s requires_login=%v", activeAuthMode, activeAuthMode.RequiresLogin())
	}

	// Register background jobs routes (if handler available)
	if jobsHandler != nil {
		jobsHandler.RegisterRoutes(api)   // Deprecated
		jobsHandler.RegisterRoutes(apiv1) // Current
		slog.Info("Background jobs routes registered (/api/v1/jobs)")
	}

	// Start server
	srv := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Stop batch worker
	log.Println("Stopping notification batch worker...")
	batchWorker.Stop()

	// Stop container monitor first (if available)
	if containerMonitor != nil {
		log.Println("Stopping container monitor...")
		if err := containerMonitor.StopMonitoring(); err != nil {
			log.Printf("Warning: Error stopping container monitor: %v", err)
		}
	}

	// Shutdown container manager (if available)
	if containerMgr != nil {
		log.Println("Shutting down container manager...")
		containerMgr.Shutdown()
	}

	// Stop background jobs system
	if jobScheduler != nil {
		log.Println("Stopping job scheduler...")
		if err := jobScheduler.Stop(); err != nil {
			log.Printf("Warning: Error stopping scheduler: %v", err)
		}
	}

	for i, worker := range jobWorkers {
		if worker != nil && worker.IsRunning() {
			log.Printf("Stopping worker %d...", i+1)
			if err := worker.Stop(); err != nil {
				log.Printf("Warning: Error stopping worker %d: %v", i+1, err)
			}
		}
	}

	// Release stuck jobs (cleanup)
	if jobsHandler != nil {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if count, err := jobsHandler.GetService().ReleaseStuckJobs(cleanupCtx); err == nil && count > 0 {
			log.Printf("Released %d stuck jobs", count)
		}
	}

	// Close database connection
	database.Close()
	log.Println("Server stopped")
}
