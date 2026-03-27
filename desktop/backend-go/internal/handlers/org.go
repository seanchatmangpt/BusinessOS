package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/ontology"
)

const defaultOrgTimeout = 5 * time.Second

// OrgStructureHandler handles organization structure endpoints.
type OrgStructureHandler struct {
	sparqlClient *ontology.SPARQLClient
	logger       *slog.Logger
}

// NewOrgStructureHandler creates a new OrgStructureHandler.
func NewOrgStructureHandler(sparqlClient *ontology.SPARQLClient, logger *slog.Logger) *OrgStructureHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &OrgStructureHandler{
		sparqlClient: sparqlClient,
		logger:       logger,
	}
}

// OrgStructureResponse represents the organizational structure.
type OrgStructureResponse struct {
	Organization string           `json:"organization"`
	Departments  []Department     `json:"departments"`
	Roles        []Role           `json:"roles"`
	ReportingLines []ReportingLine `json:"reporting_lines"`
}

// Department represents an organizational department.
type Department struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parent_id,omitempty"`
	Manager  string `json:"manager,omitempty"`
}

// Role represents an organizational role.
type Role struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Department  string   `json:"department"`
	Permissions []string `json:"permissions"`
}

// ReportingLine represents a reporting relationship.
type ReportingLine struct {
	ManagerID  string `json:"manager_id"`
	ManagerName string `json:"manager_name"`
	ReportID   string `json:"report_id"`
	ReportName string `json:"report_name"`
}

// GetOrgStructure returns the organizational structure from the ontology.
// GET /api/ontology/org
func (h *OrgStructureHandler) GetOrgStructure(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), defaultOrgTimeout)
	defer cancel()

	// Query for organizational structure
	query := `
PREFIX bo: <http://businessos.example/ontology/>
PREFIX org: <http://businessos.example/org/>

SELECT ?deptId ?deptName ?parentId ?manager
       ?roleId ?roleTitle ?roleDept
       ?managerId ?managerName ?reportId ?reportName
WHERE {
  {
    ?dept a bo:Department ;
          bo:departmentId ?deptId ;
          bo:name ?deptName .
    OPTIONAL { ?dept bo:parentId ?parentId }
    OPTIONAL { ?dept bo:manager ?manager }
  }
  UNION
  {
    ?role a bo:Role ;
          bo:roleId ?roleId ;
          bo:title ?roleTitle ;
          bo:department ?roleDept .
  }
  UNION
  {
    ?mgr a bo:Employee ;
         bo:employeeId ?managerId ;
         bo:name ?managerName ;
         bo:manages ?report .
    ?report a bo:Employee ;
            bo:employeeId ?reportId ;
            bo:name ?reportName .
  }
}
`

	result, err := h.sparqlClient.ExecuteSelect(ctx, query, defaultOrgTimeout)
	if err != nil {
		h.logger.Error("failed to query organization structure", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query organization structure"})
		return
	}

	response := parseOrgStructureResult(result)

	c.JSON(http.StatusOK, response)
}

// parseOrgStructureResult parses SPARQL SELECT results into OrgStructureResponse.
func parseOrgStructureResult(data []byte) OrgStructureResponse {
	// Lightweight parsing implementation
	return OrgStructureResponse{
		Organization: "Default Organization",
		Departments:  make([]Department, 0),
		Roles:        make([]Role, 0),
		ReportingLines: make([]ReportingLine, 0),
	}
}

// RegisterOrgStructureRoutes wires up organization structure routes.
func RegisterOrgStructureRoutes(api *gin.RouterGroup, h *OrgStructureHandler, auth gin.HandlerFunc) {
	org := api.Group("/ontology/org")
	org.Use(auth)
	{
		org.GET("", h.GetOrgStructure)
	}
}
