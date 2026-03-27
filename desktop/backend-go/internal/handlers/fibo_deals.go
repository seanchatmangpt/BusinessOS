package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/utils"
)

// FIBODealsHandler handles HTTP endpoints for FIBO deal management.
type FIBODealsHandler struct {
	dealsService *services.FIBODealsService
	logger       *slog.Logger
}

// NewFIBODealsHandler creates a new FIBO deals handler.
func NewFIBODealsHandler(dealsService *services.FIBODealsService) *FIBODealsHandler {
	return &FIBODealsHandler{
		dealsService: dealsService,
		logger:       slog.Default(),
	}
}

// RegisterFIBODealsRoutes wires /api/deals routes.
func RegisterFIBODealsRoutes(api *gin.RouterGroup, h *FIBODealsHandler, auth gin.HandlerFunc) {
	deals := api.Group("/deals")
	deals.Use(auth)
	{
		// Create a new deal
		deals.POST("", h.CreateDeal)

		// Get deal by ID
		deals.GET("/:id", h.GetDeal)

		// List all deals with pagination
		deals.GET("", h.ListDeals)

		// Update an existing deal
		deals.PATCH("/:id", h.UpdateDeal)

		// Verify deal compliance
		deals.POST("/:id/verify-compliance", h.VerifyCompliance)
	}
}

// ============================================================================
// Request / Response Types
// ============================================================================

// createDealRequest is the request body for POST /api/deals
type createDealRequest struct {
	Name              string  `json:"name" binding:"required"`
	Amount            float64 `json:"amount" binding:"required,gt=0"`
	Currency          string  `json:"currency"`
	BuyerID           string  `json:"buyer_id" binding:"required"`
	SellerID          string  `json:"seller_id" binding:"required"`
	ExpectedCloseDate string  `json:"expected_close_date"`
	Probability       int     `json:"probability" binding:"min=0,max=100"`
	Stage             string  `json:"stage"`
}

// updateDealRequest is the request body for PATCH /api/deals/:id
type updateDealRequest struct {
	Name              *string  `json:"name,omitempty"`
	Amount            *float64 `json:"amount,omitempty"`
	Currency          *string  `json:"currency,omitempty"`
	Status            *string  `json:"status,omitempty"`
	ExpectedCloseDate *string  `json:"expected_close_date,omitempty"`
	Probability       *int     `json:"probability,omitempty"`
	Stage             *string  `json:"stage,omitempty"`
}

// dealResponse is the JSON response for deal endpoints.
type dealResponse struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Amount            float64 `json:"amount"`
	Currency          string  `json:"currency"`
	Status            string  `json:"status"`
	BuyerID           string  `json:"buyer_id"`
	SellerID          string  `json:"seller_id"`
	ExpectedCloseDate string  `json:"expected_close_date"`
	Probability       int     `json:"probability"`
	Stage             string  `json:"stage"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	RDFTripleCount    int     `json:"rdf_triple_count"`
	ComplianceStatus  string  `json:"compliance_status"`
	KYCVerified       bool    `json:"kyc_verified"`
	AMLScreening      string  `json:"aml_screening"`
}

// ============================================================================
// Handlers
// ============================================================================

// CreateDeal handles POST /api/deals
// Creates a new financial deal with FIBO ontology integration.
func (h *FIBODealsHandler) CreateDeal(c *gin.Context) {
	var req createDealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, h.logger, err)
		return
	}

	deal := &services.Deal{
		Name:        req.Name,
		Amount:      req.Amount,
		Currency:    req.Currency,
		BuyerID:     req.BuyerID,
		SellerID:    req.SellerID,
		Probability: req.Probability,
		Stage:       req.Stage,
	}

	// Parse expected close date if provided
	if req.ExpectedCloseDate != "" {
		closeDate, err := parseDateTime(req.ExpectedCloseDate)
		if err != nil {
			utils.RespondBadRequest(c, h.logger, fmt.Sprintf("invalid expected_close_date: %v", err))
			return
		}
		deal.ExpectedCloseDate = closeDate
	}

	// Create deal via service
	result, err := h.dealsService.CreateDeal(c.Request.Context(), deal)
	if err != nil {
		h.logger.Error("failed to create deal", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Deal creation failed: %v", err),
		})
		return
	}

	// Return created deal
	c.JSON(http.StatusCreated, dealToResponse(result))
}

// GetDeal handles GET /api/deals/:id
// Retrieves a single deal by ID from FIBO ontology.
func (h *FIBODealsHandler) GetDeal(c *gin.Context) {
	dealID := c.Param("id")
	if dealID == "" {
		utils.RespondBadRequest(c, h.logger, "deal_id parameter required")
		return
	}

	result, err := h.dealsService.GetDeal(c.Request.Context(), dealID)
	if err != nil {
		h.logger.Error("failed to get deal", "deal_id", dealID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Deal not found: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, dealToResponse(result))
}

// ListDeals handles GET /api/deals
// Lists all deals with optional pagination.
func (h *FIBODealsHandler) ListDeals(c *gin.Context) {
	// Parse pagination parameters
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	results, err := h.dealsService.ListDeals(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("failed to list deals", "limit", limit, "offset", offset, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Deal listing failed: %v", err),
		})
		return
	}

	// Convert to response format
	responses := make([]dealResponse, len(results))
	for i, deal := range results {
		responses[i] = dealToResponse(deal)
	}

	c.JSON(http.StatusOK, gin.H{
		"deals":  responses,
		"count":  len(results),
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateDeal handles PATCH /api/deals/:id
// Updates an existing deal's fields in FIBO ontology.
func (h *FIBODealsHandler) UpdateDeal(c *gin.Context) {
	dealID := c.Param("id")
	if dealID == "" {
		utils.RespondBadRequest(c, h.logger, "deal_id parameter required")
		return
	}

	var req updateDealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, h.logger, err)
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["dealName"] = *req.Name
	}
	if req.Amount != nil {
		updates["dealAmount"] = *req.Amount
	}
	if req.Currency != nil {
		updates["dealCurrency"] = *req.Currency
	}
	if req.Status != nil {
		updates["dealStatus"] = *req.Status
	}
	if req.ExpectedCloseDate != nil {
		updates["expectedCloseDate"] = *req.ExpectedCloseDate
	}
	if req.Probability != nil {
		updates["dealProbability"] = *req.Probability
	}
	if req.Stage != nil {
		updates["currentStage"] = *req.Stage
	}

	result, err := h.dealsService.UpdateDeal(c.Request.Context(), dealID, updates)
	if err != nil {
		h.logger.Error("failed to update deal", "deal_id", dealID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Deal update failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, dealToResponse(result))
}

// VerifyCompliance handles POST /api/deals/:id/verify-compliance
// Checks deal compliance against KYC, AML, and SOX requirements.
func (h *FIBODealsHandler) VerifyCompliance(c *gin.Context) {
	dealID := c.Param("id")
	if dealID == "" {
		utils.RespondBadRequest(c, h.logger, "deal_id parameter required")
		return
	}

	result, err := h.dealsService.VerifyCompliance(c.Request.Context(), dealID)
	if err != nil {
		h.logger.Error("failed to verify compliance", "deal_id", dealID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Compliance verification failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deal_id":    dealID,
		"compliance": result,
	})
}

// ============================================================================
// Utilities
// ============================================================================

// dealToResponse converts a Deal to a dealResponse for JSON serialization.
func dealToResponse(deal *services.Deal) dealResponse {
	return dealResponse{
		ID:                deal.ID,
		Name:              deal.Name,
		Amount:            deal.Amount,
		Currency:          deal.Currency,
		Status:            deal.Status,
		BuyerID:           deal.BuyerID,
		SellerID:          deal.SellerID,
		ExpectedCloseDate: deal.ExpectedCloseDate.Format("2006-01-02T15:04:05Z"),
		Probability:       deal.Probability,
		Stage:             deal.Stage,
		CreatedAt:         deal.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         deal.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		RDFTripleCount:    deal.RDFTripleCount,
		ComplianceStatus:  deal.ComplianceStatus,
		KYCVerified:       deal.KYCVerified,
		AMLScreening:      deal.AMLScreening,
	}
}

// parseDateTime parses ISO 8601 datetime strings.
func parseDateTime(s string) (time.Time, error) {
	// Try RFC3339 format
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	// Try date-only format
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid datetime format")
}
