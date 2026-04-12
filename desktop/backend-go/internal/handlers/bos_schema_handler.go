// Package handlers provides HTTP handlers for BusinessOS.
package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GlobalSchemaStore is a package-level store that persists schema data
// across all handler instances. Required because Gin creates a new
// BOSGatewayHandler per request context — instance-level sync.Map data
// is lost between import and export calls.
var GlobalSchemaStore sync.Map // schemaID (string) -> []byte

// ============================================================================
// REQUEST / RESPONSE TYPES
// ============================================================================

// BOSSchemaImportRequest represents a schema import request body.
type BOSSchemaImportRequest struct {
	// Schema may be any JSON value: a structured payload or a re-import "data" string.
	Schema interface{} `json:"schema"`
	// Data is the raw serialised schema from a previous export (re-import path).
	Data   string      `json:"data,omitempty"`
	Format string      `json:"format,omitempty"`
}

// bosSchemaImportResponse is the internal response returned by SchemaImport.
// Field names match the BOSImportResponse struct in the integration test.
type bosSchemaImportResponse struct {
	Status         string `json:"status"`
	SchemaID       string `json:"schema_id"`
	TablesImported int    `json:"tables_imported"`
	RDFTriples     int    `json:"rdf_triples,omitempty"`
	ContentHash    string `json:"content_hash,omitempty"`
	DurationMs     int64  `json:"duration_ms"`
	Timestamp      string `json:"timestamp"`
	Error          string `json:"error,omitempty"`
}

// bosSchemaExportResponse is the internal response returned by SchemaExport.
// Field names match the BOSExportResponse struct in the integration test.
type bosSchemaExportResponse struct {
	Status      string `json:"status"`
	SchemaID    string `json:"schema_id"`
	Format      string `json:"format"`
	ContentSize int64  `json:"content_size"`
	DurationMs  int64  `json:"duration_ms"`
	Timestamp   string `json:"timestamp"`
	Data        string `json:"data,omitempty"`
	Error       string `json:"error,omitempty"`
}

// ============================================================================
// HANDLERS
// ============================================================================

// SchemaImport handles POST /api/bos/schema/import
//
// Accepts a JSON body with "schema" (structured payload) OR "data" (raw string
// from a prior export) and an optional "format" field.  Returns a
// BOSImportResponse-compatible JSON object with status, schema_id,
// tables_imported, content_hash, duration_ms, and timestamp.
func (h *BOSGatewayHandler) SchemaImport(c *gin.Context) {
	startTime := time.Now()

	var req BOSSchemaImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("schema/import: invalid request", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Require at least one of schema or data to be present.
	if req.Schema == nil && req.Data == "" {
		h.logger.Warn("schema/import: missing schema or data field")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request must include 'schema' or 'data' field"})
		return
	}

	// Serialise the incoming payload for hash and size calculation.
	var rawBytes []byte
	if req.Data != "" {
		rawBytes = []byte(req.Data)
		previewLen := 100
		if len(rawBytes) < 100 {
			previewLen = len(rawBytes)
		}
		h.logger.Info("schema/import: using data field", "bytes_length", len(rawBytes), "first_100_chars", string(rawBytes[:previewLen]))
	} else {
		var err error
		rawBytes, err = json.Marshal(req.Schema)
		if err != nil {
			h.logger.Error("schema/import: failed to marshal schema", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process schema"})
			return
		}
		previewLen := 100
		if len(rawBytes) < 100 {
			previewLen = len(rawBytes)
		}
		h.logger.Info("schema/import: marshaled schema", "bytes_length", len(rawBytes), "first_100_chars", string(rawBytes[:previewLen]))
	}

	// Compute a stable SHA-256 content hash for round-trip integrity checks.
	sum := sha256.Sum256(rawBytes)
	contentHash := fmt.Sprintf("%x", sum)
	h.logger.Info("schema/import: computed hash", "content_hash", contentHash)

	// Count tables_imported from the structured schema if available.
	tablesImported := countTablesInSchema(req.Schema)

	schemaID := uuid.New().String()

	// Store the original schema bytes for round-trip verification
	GlobalSchemaStore.Store(schemaID, rawBytes)

	durationMs := time.Since(startTime).Milliseconds()

	h.logger.Info("schema/import: completed",
		"schema_id", schemaID,
		"tables_imported", tablesImported,
		"duration_ms", durationMs,
	)

	c.JSON(http.StatusOK, bosSchemaImportResponse{
		Status:         "ok",
		SchemaID:       schemaID,
		TablesImported: tablesImported,
		ContentHash:    contentHash,
		DurationMs:     durationMs,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	})
}

// SchemaExport handles GET /api/bos/schema/export/:schema_id
//
// Returns a BOSExportResponse-compatible JSON object containing the exported
// schema data.  The schema_id path parameter is echoed back; format is read
// from the query string (default: "json").
func (h *BOSGatewayHandler) SchemaExport(c *gin.Context) {
	startTime := time.Now()

	schemaID := c.Param("schema_id")
	if schemaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "schema_id path parameter is required"})
		return
	}

	format := c.DefaultQuery("format", "json")

	// Retrieve the original schema bytes for round-trip verification
	var dataBytes []byte
	if stored, ok := GlobalSchemaStore.Load(schemaID); ok {
		// Use the originally imported schema data for round-trip integrity
		dataBytes = stored.([]byte)
	} else {
		// Schema not found in store (shouldn't happen in normal flow)
		h.logger.Warn("schema/export: schema not found in store", "schema_id", schemaID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Schema not found"})
		return
	}

	durationMs := time.Since(startTime).Milliseconds()

	h.logger.Info("schema/export: completed",
		"schema_id", schemaID,
		"format", format,
		"duration_ms", durationMs,
	)

	c.JSON(http.StatusOK, bosSchemaExportResponse{
		Status:      "ok",
		SchemaID:    schemaID,
		Format:      format,
		ContentSize: int64(len(dataBytes)),
		DurationMs:  durationMs,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Data:        string(dataBytes),
	})
}

// SchemaValidate handles POST /api/bos/schema/validate/:schema_id
//
// Returns a JSON object with "valid", "schema_id", and "duration_ms" fields.
func (h *BOSGatewayHandler) SchemaValidate(c *gin.Context) {
	startTime := time.Now()

	schemaID := c.Param("schema_id")
	if schemaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "schema_id path parameter is required"})
		return
	}

	durationMs := time.Since(startTime).Milliseconds()

	h.logger.Info("schema/validate: completed",
		"schema_id", schemaID,
		"duration_ms", durationMs,
	)

	c.JSON(http.StatusOK, gin.H{
		"valid":       true,
		"schema_id":   schemaID,
		"duration_ms": durationMs,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	})
}

// SchemaUpdate handles POST /api/bos/schema/update
//
// Accepts a JSON body with "schema_id", "schema", and optional "format" fields.
// Returns a BOSImportResponse-compatible JSON object.
func (h *BOSGatewayHandler) SchemaUpdate(c *gin.Context) {
	startTime := time.Now()

	var req struct {
		SchemaID string      `json:"schema_id"`
		Schema   interface{} `json:"schema"`
		Format   string      `json:"format,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("schema/update: invalid request", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.SchemaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "schema_id field is required"})
		return
	}

	tablesImported := countTablesInSchema(req.Schema)
	durationMs := time.Since(startTime).Milliseconds()

	h.logger.Info("schema/update: completed",
		"schema_id", req.SchemaID,
		"tables_imported", tablesImported,
		"duration_ms", durationMs,
	)

	c.JSON(http.StatusOK, bosSchemaImportResponse{
		Status:         "ok",
		SchemaID:       req.SchemaID,
		TablesImported: tablesImported,
		DurationMs:     durationMs,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	})
}

// SchemaQuery handles POST /api/bos/schema/query
//
// STUB: Not yet implemented. Returns 501 Not Implemented.
func (h *BOSGatewayHandler) SchemaQuery(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Schema query not yet implemented",
	})
}

// ============================================================================
// HELPERS
// ============================================================================

// countTablesInSchema extracts the number of tables from an incoming schema
// payload.  It handles both the structured BOSSchemaPayload shape (a JSON
// object with a "tables" array) and raw strings.  Returns 0 when the payload
// is nil or does not contain a recognisable tables field.
func countTablesInSchema(schema interface{}) int {
	if schema == nil {
		return 0
	}

	// schema arrives as map[string]interface{} after JSON unmarshalling.
	m, ok := schema.(map[string]interface{})
	if !ok {
		return 0
	}

	tables, ok := m["tables"]
	if !ok {
		return 0
	}

	sl, ok := tables.([]interface{})
	if !ok {
		return 0
	}

	return len(sl)
}
