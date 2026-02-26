package e2b

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ---- Domain types -----------------------------------------------------------

// ExecutionConfig controls sandbox lifecycle behaviour.
type ExecutionConfig struct {
	// Timeout is the HTTP-level deadline for a single sandbox operation.
	Timeout time.Duration

	// MaxRetries is the number of build-test-fix iterations in ExecuteWithFixLoop.
	MaxRetries int

	// RetryDelay is a fixed pause between fix-loop iterations (not used for
	// transient HTTP errors; those follow the RetryStrategy backoff).
	RetryDelay time.Duration

	// KeepSandbox instructs the bridge to leave the sandbox running after
	// execution so callers can inspect its state.
	KeepSandbox bool
}

// DefaultExecutionConfig returns sensible defaults.
func DefaultExecutionConfig() *ExecutionConfig {
	return &ExecutionConfig{
		Timeout:     10 * time.Minute,
		MaxRetries:  3,
		RetryDelay:  3 * time.Second,
		KeepSandbox: false,
	}
}

// ExecutionResult holds the outcome of a single sandbox run.
type ExecutionResult struct {
	SandboxID string         `json:"sandbox_id"`
	Success   bool           `json:"success"`
	Phase     ExecutionPhase `json:"phase"`
	Error     string         `json:"error,omitempty"`
	Stdout    string         `json:"stdout,omitempty"`
	Stderr    string         `json:"stderr,omitempty"`
}

// IsSuccess reports whether the sandbox run completed without errors.
func (r *ExecutionResult) IsSuccess() bool {
	return r != nil && r.Success && r.Error == ""
}

// HasError reports whether the result contains an error description.
func (r *ExecutionResult) HasError() bool {
	return r != nil && r.Error != ""
}

// GetErrorDetails returns a human-readable string summarising the error.
func (r *ExecutionResult) GetErrorDetails() string {
	if r == nil {
		return ""
	}
	if r.Error != "" && r.Stderr != "" {
		return fmt.Sprintf("[%s] %s\nSTDERR: %s", r.Phase, r.Error, r.Stderr)
	}
	if r.Error != "" {
		return fmt.Sprintf("[%s] %s", r.Phase, r.Error)
	}
	return ""
}

// ExecutionSummary accumulates results across all fix-loop iterations.
type ExecutionSummary struct {
	TotalAttempts  int
	SuccessfulRuns int
	FixesApplied   int
	AllResults     []*ExecutionResult
	LastResult     *ExecutionResult
	FinalSandboxID string
	ErrorsSummary  []string
	FilesUpdated   []string
	TotalDuration  time.Duration
}

// SandboxStatus holds the runtime state returned by the bridge.
type SandboxStatus struct {
	SandboxID string `json:"sandbox_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at,omitempty"`
}

// CleanupResult is the bridge's response to a sandbox deletion request.
type CleanupResult struct {
	SandboxID string `json:"sandbox_id"`
	Deleted   bool   `json:"deleted"`
}

// UpdateFilesResult is the bridge's response to a file-update request.
type UpdateFilesResult struct {
	SandboxID    string   `json:"sandbox_id"`
	UpdatedFiles []string `json:"updated_files"`
}

// wire types for HTTP requests and error responses.
type testExecutionRequest struct {
	Path        string `json:"path"`
	IterationID string `json:"iteration_id,omitempty"`
	KeepSandbox bool   `json:"keep_sandbox"`
}

type updateFilesRequest struct {
	SandboxID string            `json:"sandbox_id"`
	Files     map[string]string `json:"files"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// ---- Client -----------------------------------------------------------------

// ClientConfig holds the parameters required to construct a Client.
type ClientConfig struct {
	// BaseURL is the URL of the E2B bridge service (e.g. "http://localhost:8080").
	BaseURL string

	// APIKey is the E2B API key. If empty the E2B_API_KEY environment variable
	// is used.
	APIKey string

	// TenantID tags every request with a tenant identifier for isolation.
	TenantID string

	// Execution overrides; nil means DefaultExecutionConfig() is used.
	Execution *ExecutionConfig

	// Logger is used for structured output. If nil slog.Default() is used.
	Logger *slog.Logger
}

// Client is the primary E2B bridge client. It is safe for concurrent use.
//
// Create with NewClient or NewClientWithConfig; never embed or copy by value.
type Client struct {
	baseURL         string
	apiKey          string
	tenantID        string
	httpClient      *http.Client
	config          *ExecutionConfig
	retryStrategies map[ErrorType]*RetryStrategy
	logger          *slog.Logger
}

// NewClient constructs a Client using defaults and the E2B_API_KEY environment
// variable.
func NewClient(ctx context.Context, baseURL string) (*Client, error) {
	return NewClientWithConfig(ctx, ClientConfig{BaseURL: baseURL})
}

// NewClientWithConfig constructs a Client with explicit configuration. The
// context is reserved for future use (e.g. acquiring initial auth tokens).
func NewClientWithConfig(_ context.Context, cfg ClientConfig) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("e2b client: BaseURL must not be empty")
	}

	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("E2B_API_KEY")
	}

	exec := cfg.Execution
	if exec == nil {
		exec = DefaultExecutionConfig()
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	c := &Client{
		baseURL:  cfg.BaseURL,
		apiKey:   apiKey,
		tenantID: cfg.TenantID,
		httpClient: &http.Client{
			Timeout: exec.Timeout,
		},
		config:          exec,
		retryStrategies: DefaultRetryStrategies(),
		logger:          logger,
	}
	return c, nil
}

// SetRetryStrategy overrides the retry strategy for a specific error type.
func (c *Client) SetRetryStrategy(errorType ErrorType, strategy *RetryStrategy) {
	if c.retryStrategies == nil {
		c.retryStrategies = make(map[ErrorType]*RetryStrategy)
	}
	c.retryStrategies[errorType] = strategy
}

// GetConfig returns a copy of the current execution configuration.
func (c *Client) GetConfig() ExecutionConfig {
	return *c.config
}

// ---- Sandbox lifecycle methods ----------------------------------------------

// TestExecution runs a single sandbox execution against projectPath.
func (c *Client) TestExecution(ctx context.Context, projectPath string) (*ExecutionResult, error) {
	return c.testExecutionWithID(ctx, projectPath, "")
}

// testExecutionWithID is the internal implementation that accepts an optional
// iteration ID.
func (c *Client) testExecutionWithID(ctx context.Context, projectPath, iterationID string) (*ExecutionResult, error) {
	req := testExecutionRequest{
		Path:        projectPath,
		IterationID: iterationID,
		KeepSandbox: c.config.KeepSandbox,
	}

	var result ExecutionResult
	if err := c.makeRequest(ctx, http.MethodPost, "/test-execution", req, &result); err != nil {
		return nil, fmt.Errorf("test execution: %w", err)
	}
	return &result, nil
}

// UpdateSandboxFiles replaces or creates files inside a running sandbox.
func (c *Client) UpdateSandboxFiles(ctx context.Context, sandboxID string, files map[string]string) (*UpdateFilesResult, error) {
	if sandboxID == "" {
		return nil, NewValidationError("sandbox ID must not be empty")
	}
	if len(files) == 0 {
		return nil, NewValidationError("files map must not be empty")
	}

	req := updateFilesRequest{
		SandboxID: sandboxID,
		Files:     files,
	}

	var result UpdateFilesResult
	if err := c.makeRequest(ctx, http.MethodPost, "/update-sandbox-files", req, &result); err != nil {
		return nil, fmt.Errorf("update sandbox files: %w", err)
	}
	return &result, nil
}

// GetSandboxStatus queries the running status of a sandbox.
func (c *Client) GetSandboxStatus(ctx context.Context, sandboxID string) (*SandboxStatus, error) {
	if sandboxID == "" {
		return nil, NewValidationError("sandbox ID must not be empty")
	}

	var result SandboxStatus
	endpoint := fmt.Sprintf("/sandbox-status/%s", sandboxID)
	if err := c.makeRequest(ctx, http.MethodGet, endpoint, nil, &result); err != nil {
		return nil, fmt.Errorf("get sandbox status: %w", err)
	}
	return &result, nil
}

// CleanupSandbox destroys a sandbox and releases its resources.
func (c *Client) CleanupSandbox(ctx context.Context, sandboxID string) (*CleanupResult, error) {
	if sandboxID == "" {
		return nil, NewValidationError("sandbox ID must not be empty")
	}

	var result CleanupResult
	endpoint := fmt.Sprintf("/sandbox/%s", sandboxID)
	if err := c.makeRequest(ctx, http.MethodDelete, endpoint, nil, &result); err != nil {
		return nil, fmt.Errorf("cleanup sandbox: %w", err)
	}
	return &result, nil
}

// ---- High-level execution methods ------------------------------------------

// ExecuteWithRetry runs a sandbox execution test, retrying on transient errors
// up to MaxRetries times with intelligent backoff.
func (c *Client) ExecuteWithRetry(ctx context.Context, projectPath string) (*ExecutionResult, error) {
	var (
		lastResult *ExecutionResult
		lastErr    error
	)

	for attempt := 1; attempt <= c.config.MaxRetries; attempt++ {
		c.logger.DebugContext(ctx, "sandbox execution attempt",
			"attempt", attempt,
			"max_retries", c.config.MaxRetries,
			"project_path", projectPath,
			"tenant_id", c.tenantID,
		)

		result, err := c.TestExecution(ctx, projectPath)
		if err != nil {
			e2bErr := ClassifyError(err, PhaseSetup, "")
			lastErr = e2bErr

			shouldRetry, delay := ShouldRetry(e2bErr, attempt, c.retryStrategies)
			if !shouldRetry || attempt >= c.config.MaxRetries {
				return nil, fmt.Errorf("execution failed after %d attempt(s): %w", attempt, e2bErr)
			}

			c.logger.WarnContext(ctx, "retryable error, waiting before retry",
				"attempt", attempt,
				"delay", delay,
				"error", e2bErr,
			)
			if err := sleepContext(ctx, delay); err != nil {
				return nil, err
			}
			continue
		}

		lastResult = result

		if result.IsSuccess() {
			c.logger.InfoContext(ctx, "sandbox execution succeeded", "attempt", attempt)
			return result, nil
		}

		execErr := NewExecutionError(result.Phase, result.Error, result.SandboxID)
		lastErr = execErr

		shouldRetry, delay := ShouldRetry(execErr, attempt, c.retryStrategies)
		if !shouldRetry || attempt >= c.config.MaxRetries {
			break
		}

		c.logger.WarnContext(ctx, "sandbox execution failed, retrying",
			"attempt", attempt,
			"phase", result.Phase,
			"delay", delay,
		)
		if err := sleepContext(ctx, delay); err != nil {
			return nil, err
		}
	}

	return lastResult, fmt.Errorf("execution failed after %d attempt(s): %w", c.config.MaxRetries, lastErr)
}

// FixerFunc is called when execution fails to generate file patches that should
// be applied before the next attempt. It receives the failing ExecutionResult
// and must return a map of sandbox-relative file paths to their new contents.
// Returning (nil, nil) or (empty map, nil) skips the update step.
type FixerFunc func(ctx context.Context, result *ExecutionResult) (map[string]string, error)

// ExecuteWithFixLoop runs the build-test-fix retry loop. On each failure it
// calls fixer (if non-nil) to generate file patches, applies them to the
// sandbox, then retries. The loop runs up to MaxRetries times.
func (c *Client) ExecuteWithFixLoop(ctx context.Context, projectPath string, fixer FixerFunc) (*ExecutionSummary, error) {
	summary := &ExecutionSummary{
		AllResults:    make([]*ExecutionResult, 0),
		ErrorsSummary: make([]string, 0),
		FilesUpdated:  make([]string, 0),
	}

	start := time.Now()
	defer func() {
		summary.TotalDuration = time.Since(start)
	}()

	var currentSandboxID string

	for attempt := 1; attempt <= c.config.MaxRetries; attempt++ {
		summary.TotalAttempts = attempt

		result, err := c.TestExecution(ctx, projectPath)
		if err != nil {
			return summary, fmt.Errorf("execution test on attempt %d: %w", attempt, err)
		}

		summary.AllResults = append(summary.AllResults, result)
		summary.LastResult = result
		currentSandboxID = result.SandboxID

		if result.IsSuccess() {
			summary.SuccessfulRuns = 1
			summary.FinalSandboxID = currentSandboxID
			c.logger.InfoContext(ctx, "fix-loop succeeded",
				"attempt", attempt,
				"sandbox_id", currentSandboxID,
				"tenant_id", c.tenantID,
			)
			return summary, nil
		}

		if result.HasError() {
			summary.ErrorsSummary = append(summary.ErrorsSummary, result.GetErrorDetails())
		}

		if attempt == c.config.MaxRetries || fixer == nil {
			break
		}

		fixes, err := fixer(ctx, result)
		if err != nil {
			c.logger.WarnContext(ctx, "fixer returned error, continuing to next attempt",
				"attempt", attempt,
				"error", err,
			)
			if sleepErr := sleepContext(ctx, c.config.RetryDelay); sleepErr != nil {
				return summary, sleepErr
			}
			continue
		}

		if len(fixes) == 0 {
			c.logger.DebugContext(ctx, "fixer produced no patches, continuing", "attempt", attempt)
			if sleepErr := sleepContext(ctx, c.config.RetryDelay); sleepErr != nil {
				return summary, sleepErr
			}
			continue
		}

		if currentSandboxID != "" {
			updateResult, updateErr := c.UpdateSandboxFiles(ctx, currentSandboxID, fixes)
			if updateErr != nil {
				c.logger.WarnContext(ctx, "failed to apply fixes to sandbox, will retry with fresh upload",
					"attempt", attempt,
					"sandbox_id", currentSandboxID,
					"error", updateErr,
				)
			} else {
				summary.FixesApplied++
				summary.FilesUpdated = append(summary.FilesUpdated, updateResult.UpdatedFiles...)
				c.logger.InfoContext(ctx, "applied fixes to sandbox",
					"attempt", attempt,
					"files_count", len(fixes),
					"sandbox_id", currentSandboxID,
				)
			}
		}

		if sleepErr := sleepContext(ctx, c.config.RetryDelay); sleepErr != nil {
			return summary, sleepErr
		}
	}

	summary.FinalSandboxID = currentSandboxID
	return summary, fmt.Errorf("execution failed after %d attempt(s)", c.config.MaxRetries)
}

// ---- Health checking --------------------------------------------------------

// CheckHealth verifies the bridge is reachable and responding. It returns nil
// on success.
func (c *Client) CheckHealth(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("create health check request: %w", err)
	}
	c.applyHeaders(req, false)

	c.logger.DebugContext(ctx, "checking e2b bridge health", "url", c.baseURL+"/health")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("e2b bridge unreachable at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("e2b bridge unhealthy: HTTP %d", resp.StatusCode)
	}

	c.logger.DebugContext(ctx, "e2b bridge is healthy")
	return nil
}

// CheckHealthWithRetry checks bridge health with exponential backoff.
func (c *Client) CheckHealthWithRetry(ctx context.Context, maxRetries int) error {
	backoff := time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := c.CheckHealth(ctx); err == nil {
			if attempt > 1 {
				c.logger.InfoContext(ctx, "e2b bridge became available", "attempt", attempt)
			}
			return nil
		} else {
			c.logger.WarnContext(ctx, "e2b health check failed",
				"attempt", attempt,
				"max_retries", maxRetries,
				"error", err,
			)

			if attempt == maxRetries {
				return fmt.Errorf("e2b bridge unavailable after %d attempts: %w", maxRetries, err)
			}
		}

		if err := sleepContext(ctx, backoff); err != nil {
			return err
		}
		backoff *= 2
	}

	return fmt.Errorf("e2b bridge unavailable after %d attempts", maxRetries)
}

// ---- SDK-style direct execution (no bridge) ---------------------------------

// SDKConfig holds the options for the SDK-backed executor.
type SDKConfig struct {
	// APIKey is the E2B API key. If empty E2B_API_KEY is used.
	APIKey string

	// TenantID tags sandbox operations for isolation tracking.
	TenantID string

	// Execution overrides; nil means DefaultExecutionConfig() is used.
	Execution *ExecutionConfig

	// Logger is used for structured output. If nil slog.Default() is used.
	Logger *slog.Logger
}

// SDKExecutor runs code directly in E2B sandboxes without an HTTP bridge. It
// uses the local filesystem as the source and uploads files itself.
//
// The interface is intentionally kept minimal; complex orchestration belongs in
// the service layer.
type SDKExecutor struct {
	apiKey   string
	tenantID string
	config   *ExecutionConfig
	logger   *slog.Logger
}

// NewSDKExecutor constructs an SDKExecutor. apiKey may be empty; if so
// E2B_API_KEY must be set in the environment.
func NewSDKExecutor(ctx context.Context, cfg SDKConfig) (*SDKExecutor, error) {
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("E2B_API_KEY")
	}

	exec := cfg.Execution
	if exec == nil {
		exec = DefaultExecutionConfig()
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	if apiKey == "" {
		logger.WarnContext(ctx, "E2B_API_KEY not set; sandbox creation will fail")
	}

	return &SDKExecutor{
		apiKey:   apiKey,
		tenantID: cfg.TenantID,
		config:   exec,
		logger:   logger,
	}, nil
}

// UploadFiles reads all non-ignored files under projectPath and returns a map
// of sandbox-relative paths to their contents. The caller is responsible for
// writing the returned map into a sandbox.
//
// Ignored paths: directories, hidden files, node_modules.
func UploadFiles(projectPath string) (map[string]string, error) {
	files := make(map[string]string)

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		if strings.Contains(path, "node_modules") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		rel, err := filepath.Rel(projectPath, path)
		if err != nil {
			return fmt.Errorf("relative path for %s: %w", path, err)
		}

		files[filepath.Join("/home/user/project", rel)] = string(content)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("upload files walk: %w", err)
	}
	return files, nil
}

// DetectPackageManager returns the install and test commands appropriate for
// the project at projectPath, based on the presence of well-known manifest
// files.
func DetectPackageManager(projectPath string) (installCmd, testCmd string, ok bool) {
	checks := []struct {
		file    string
		install string
		test    string
	}{
		{"package.json", "npm install", "npm test"},
		{"go.mod", "go mod download", "go test ./..."},
		{"requirements.txt", "pip install -r requirements.txt", "pytest"},
	}

	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(projectPath, c.file)); err == nil {
			return c.install, c.test, true
		}
	}
	return "", "", false
}

// ParseTestSuccess uses heuristics to decide whether command output indicates
// a successful test run.
func ParseTestSuccess(output string) bool {
	lower := strings.ToLower(output)

	failureKeywords := []string{"fail", "error", "✗", "✘", "failed", "failure"}
	for _, kw := range failureKeywords {
		if strings.Contains(lower, kw) {
			return false
		}
	}

	successKeywords := []string{"all tests passed", "tests passed", "ok", "pass", "✓", "✔"}
	for _, kw := range successKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}

	return false
}

// ---- HTTP transport helpers -------------------------------------------------

// makeRequest encodes the request body (if any), sends the HTTP request, and
// decodes the response into result. A single retry pass is used for the
// underlying transport call when the error strategy allows it.
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body, result interface{}) error {
	url := c.baseURL + endpoint

	var reqBody io.Reader
	var rawJSON []byte

	if body != nil {
		var err error
		rawJSON, err = json.Marshal(body)
		if err != nil {
			return NewValidationError(fmt.Sprintf("marshal request body: %v", err))
		}
		reqBody = bytes.NewReader(rawJSON)
	}

	const maxAttempts = 2 // internal transport-level retry
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if body != nil && attempt > 1 {
			reqBody = bytes.NewReader(rawJSON)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return NewValidationError(fmt.Sprintf("create request: %v", err))
		}

		c.applyHeaders(req, body != nil)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			e2bErr := ClassifyError(err, PhaseSetup, "")
			lastErr = e2bErr

			if attempt < maxAttempts {
				if ok, delay := ShouldRetry(e2bErr, attempt, c.retryStrategies); ok {
					c.logger.WarnContext(ctx, "transport error, retrying",
						"attempt", attempt,
						"delay", delay,
						"error", e2bErr,
					)
					if sleepErr := sleepContext(ctx, delay); sleepErr != nil {
						return sleepErr
					}
					continue
				}
			}
			return e2bErr
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return NewE2BError(ErrorTypeNetwork, "READ_ERROR",
				fmt.Sprintf("read response body: %v", readErr), true)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			e2bErr := c.classifyHTTPError(resp.StatusCode, respBody)
			lastErr = e2bErr

			if attempt < maxAttempts {
				if ok, delay := ShouldRetry(e2bErr, attempt, c.retryStrategies); ok {
					c.logger.WarnContext(ctx, "HTTP error, retrying",
						"status", resp.StatusCode,
						"attempt", attempt,
						"delay", delay,
					)
					if sleepErr := sleepContext(ctx, delay); sleepErr != nil {
						return sleepErr
					}
					continue
				}
			}
			return e2bErr
		}

		if result != nil {
			if unmarshalErr := json.Unmarshal(respBody, result); unmarshalErr != nil {
				return NewE2BError(ErrorTypeService, "PARSE_ERROR",
					fmt.Sprintf("unmarshal response: %v", unmarshalErr), false)
			}
		}
		return nil
	}

	return lastErr
}

// applyHeaders sets all standard headers on req.
func (c *Client) applyHeaders(req *http.Request, hasBody bool) {
	if hasBody {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", "BusinessOS-E2B-Client/1.0")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("X-API-Key", c.apiKey)
	}
	if c.tenantID != "" {
		req.Header.Set("X-Tenant-ID", c.tenantID)
	}
}

// classifyHTTPError converts an HTTP error status code and response body into a
// typed *E2BError.
func (c *Client) classifyHTTPError(statusCode int, respBody []byte) *E2BError {
	var errResp errorResponse
	if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
		switch statusCode {
		case http.StatusBadRequest:
			return NewE2BError(ErrorTypeValidation, "BAD_REQUEST", errResp.Error, false)
		case http.StatusNotFound:
			return NewE2BError(ErrorTypeValidation, "NOT_FOUND", errResp.Error, false)
		case http.StatusTooManyRequests:
			return NewE2BError(ErrorTypeRateLimit, "RATE_LIMITED", errResp.Error, true)
		case http.StatusInternalServerError:
			return NewE2BError(ErrorTypeService, "INTERNAL_ERROR", errResp.Error, true)
		case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return NewE2BError(ErrorTypeService, "SERVICE_UNAVAILABLE", errResp.Error, true)
		default:
			return NewE2BError(ErrorTypeService, "HTTP_ERROR", errResp.Error, statusCode >= 500)
		}
	}

	msg := fmt.Sprintf("HTTP %d: %s", statusCode, string(respBody))
	switch {
	case statusCode >= 500:
		return NewE2BError(ErrorTypeService, "SERVER_ERROR", msg, true)
	case statusCode == http.StatusTooManyRequests:
		return NewE2BError(ErrorTypeRateLimit, "RATE_LIMITED", msg, true)
	case statusCode >= 400:
		return NewE2BError(ErrorTypeValidation, "CLIENT_ERROR", msg, false)
	default:
		return NewE2BError(ErrorTypeUnknown, "HTTP_ERROR", msg, false)
	}
}

// sleepContext blocks for d or until ctx is cancelled, whichever comes first.
func sleepContext(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}
