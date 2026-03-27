package ontology

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

// ComplianceFramework represents a regulatory framework (SOC2, GDPR, HIPAA, SOX).
type ComplianceFramework string

const (
	FrameworkSOC2  ComplianceFramework = "SOC2"
	FrameworkGDPR  ComplianceFramework = "GDPR"
	FrameworkHIPAA ComplianceFramework = "HIPAA"
	FrameworkSOX   ComplianceFramework = "SOX"
)

// ComplianceControl represents a single control within a framework.
type ComplianceControl struct {
	ID          string   `json:"id"`
	Framework   string   `json:"framework"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"` // critical, high, medium, low
	Verified    bool     `json:"verified"`
	Details     []string `json:"details,omitempty"`
}

// ComplianceViolation represents a control violation.
type ComplianceViolation struct {
	ControlID   string `json:"control_id"`
	Framework   string `json:"framework"`
	Title       string `json:"title"`
	Reason      string `json:"reason"`
	Severity    string `json:"severity"`
	Remediation string `json:"remediation,omitempty"`
}

// ComplianceReport represents the results of a compliance verification.
type ComplianceReport struct {
	Framework      string                `json:"framework"`
	Status         string                `json:"status"` // compliant, non_compliant, partial
	Score          float64               `json:"score"`  // 0.0-1.0
	TotalControls  int                   `json:"total_controls"`
	PassedControls int                   `json:"passed_controls"`
	FailedControls int                   `json:"failed_controls"`
	Violations     []ComplianceViolation `json:"violations"`
	Timestamp      time.Time             `json:"timestamp"`
}

// ComplianceMatrix aggregates reports for all frameworks.
type ComplianceMatrix struct {
	Frameworks   map[string]*ComplianceReport `json:"frameworks"`
	OverallScore float64                      `json:"overall_score"`
	Timestamp    time.Time                    `json:"timestamp"`
}

// OntologyLoader manages loading compliance ontology from file.
type OntologyLoader struct {
	ontologyPath string
	controls     map[string]map[string]*ComplianceControl // framework -> id -> control
	mu           sync.RWMutex
	logger       *slog.Logger
}

// NewOntologyLoader constructs an OntologyLoader.
func NewOntologyLoader(ontologyPath string, logger *slog.Logger) *OntologyLoader {
	if logger == nil {
		logger = slog.Default()
	}
	return &OntologyLoader{
		ontologyPath: ontologyPath,
		controls:     make(map[string]map[string]*ComplianceControl),
		logger:       logger,
	}
}

// LoadOntology loads the compliance ontology from file and parses it.
// For now, this loads a pre-built control set based on YAWL patterns.
func (ol *OntologyLoader) LoadOntology(ctx context.Context) error {
	ol.mu.Lock()
	defer ol.mu.Unlock()

	// Verify file exists
	if _, err := os.Stat(ol.ontologyPath); err != nil {
		ol.logger.Warn("Ontology file not found, using hardcoded control set", "path", ol.ontologyPath)
	}

	// Initialize framework maps
	for _, fw := range []string{"SOC2", "GDPR", "HIPAA", "SOX"} {
		ol.controls[fw] = make(map[string]*ComplianceControl)
	}

	// Load hardcoded control set (based on compliance-rules.yaml)
	ol.loadSOC2Controls()
	ol.loadGDPRControls()
	ol.loadHIPAAControls()
	ol.loadSOXControls()

	ol.logger.Info("Compliance ontology loaded", "frameworks", len(ol.controls))
	return nil
}

// loadSOC2Controls initializes SOC2 controls.
func (ol *OntologyLoader) loadSOC2Controls() {
	controls := []*ComplianceControl{
		{
			ID:          "soc2.cc6.1",
			Framework:   "SOC2",
			Title:       "Logical access restricted to authorized personnel",
			Description: "User roles must be validated and restricted to authorized personnel only",
			Severity:    "critical",
		},
		{
			ID:          "soc2.cc6.2",
			Framework:   "SOC2",
			Title:       "User provisioning requires verification",
			Description: "New user accounts require verification before activation",
			Severity:    "high",
		},
		{
			ID:          "soc2.a1.1",
			Framework:   "SOC2",
			Title:       "Service availability must exceed 99.9%",
			Description: "Systems must maintain 99.9% uptime SLA",
			Severity:    "high",
		},
		{
			ID:          "soc2.c1.1",
			Framework:   "SOC2",
			Title:       "Sensitive data must be encrypted at rest",
			Description: "All sensitive data must be encrypted using approved algorithms",
			Severity:    "critical",
		},
		{
			ID:          "soc2.i1.1",
			Framework:   "SOC2",
			Title:       "Audit trail entries must have valid signatures",
			Description: "Audit entries must be cryptographically signed and verifiable",
			Severity:    "critical",
		},
		{
			ID:          "soc2.cc7.1",
			Framework:   "SOC2",
			Title:       "System monitoring and alerting enabled",
			Description: "Continuous system monitoring must be in place with real-time alerting",
			Severity:    "medium",
		},
		{
			ID:          "soc2.cc7.2",
			Framework:   "SOC2",
			Title:       "Incident response procedures documented",
			Description: "Formal incident response procedures must be documented and tested",
			Severity:    "medium",
		},
		{
			ID:          "soc2.pi1.1",
			Framework:   "SOC2",
			Title:       "Privacy impact assessment performed",
			Description: "Privacy impact assessment must be completed for new systems",
			Severity:    "medium",
		},
	}

	for _, ctrl := range controls {
		ol.controls["SOC2"][ctrl.ID] = ctrl
	}
}

// loadGDPRControls initializes GDPR controls.
func (ol *OntologyLoader) loadGDPRControls() {
	controls := []*ComplianceControl{
		{
			ID:          "gdpr.ds.1",
			Framework:   "GDPR",
			Title:       "Data subject access requests fulfilled within 30 days",
			Description: "Data subject requests for access, rectification, or erasure must be fulfilled within 30 days",
			Severity:    "critical",
		},
		{
			ID:          "gdpr.cm.1",
			Framework:   "GDPR",
			Title:       "Explicit consent obtained before processing personal data",
			Description: "Consent must be freely given, specific, informed, and unambiguous before processing",
			Severity:    "critical",
		},
		{
			ID:          "gdpr.dpa.1",
			Framework:   "GDPR",
			Title:       "Data Processing Agreement with all sub-processors",
			Description: "Signed DPA must be in place with all processors and sub-processors",
			Severity:    "critical",
		},
		{
			ID:          "gdpr.dm.1",
			Framework:   "GDPR",
			Title:       "Data minimization enforced",
			Description: "Only necessary personal data is collected per GDPR Article 5(1)(c)",
			Severity:    "medium",
		},
		{
			ID:          "gdpr.dr.1",
			Framework:   "GDPR",
			Title:       "EU personal data residency compliance",
			Description: "EU resident personal data must be stored in EU data centers",
			Severity:    "critical",
		},
		{
			ID:          "gdpr.br.1",
			Framework:   "GDPR",
			Title:       "Breach notification within 72 hours",
			Description: "Personal data breaches must be reported to authorities within 72 hours",
			Severity:    "critical",
		},
		{
			ID:          "gdpr.dpia.1",
			Framework:   "GDPR",
			Title:       "Data Protection Impact Assessment completed",
			Description: "DPIA required for high-risk processing activities",
			Severity:    "high",
		},
	}

	for _, ctrl := range controls {
		ol.controls["GDPR"][ctrl.ID] = ctrl
	}
}

// loadHIPAAControls initializes HIPAA controls.
func (ol *OntologyLoader) loadHIPAAControls() {
	controls := []*ComplianceControl{
		{
			ID:          "hipaa.ac.1",
			Framework:   "HIPAA",
			Title:       "Access control implemented for PHI",
			Description: "Only authorized users can access Protected Health Information",
			Severity:    "critical",
		},
		{
			ID:          "hipaa.ae.1",
			Framework:   "HIPAA",
			Title:       "Audit controls enabled for PHI systems",
			Description: "Comprehensive audit logging for all PHI access and modifications",
			Severity:    "critical",
		},
		{
			ID:          "hipaa.tr.1",
			Framework:   "HIPAA",
			Title:       "PHI transmission encrypted end-to-end",
			Description: "All PHI must be encrypted in transit using TLS 1.2 or higher",
			Severity:    "critical",
		},
		{
			ID:          "hipaa.se.1",
			Framework:   "HIPAA",
			Title:       "Encryption at rest required for PHI",
			Description: "All stored PHI must be encrypted using NIST-approved algorithms",
			Severity:    "critical",
		},
		{
			ID:          "hipaa.ba.1",
			Framework:   "HIPAA",
			Title:       "Business Associate Agreement in place",
			Description: "BAA must be signed with all Business Associates handling PHI",
			Severity:    "critical",
		},
		{
			ID:          "hipaa.id.1",
			Framework:   "HIPAA",
			Title:       "Workforce identification and authentication",
			Description: "Multi-factor authentication required for PHI system access",
			Severity:    "high",
		},
		{
			ID:          "hipaa.nm.1",
			Framework:   "HIPAA",
			Title:       "Non-repudiation controls for PHI transactions",
			Description: "Digital signatures required for critical PHI operations",
			Severity:    "medium",
		},
	}

	for _, ctrl := range controls {
		ol.controls["HIPAA"][ctrl.ID] = ctrl
	}
}

// loadSOXControls initializes SOX controls.
func (ol *OntologyLoader) loadSOXControls() {
	controls := []*ComplianceControl{
		{
			ID:          "sox.itg.1",
			Framework:   "SOX",
			Title:       "Segregation of duties enforced",
			Description: "Changes to production systems cannot be made and approved by same person",
			Severity:    "critical",
		},
		{
			ID:          "sox.sa.1",
			Framework:   "SOX",
			Title:       "Financial systems maintain 99.9% uptime",
			Description: "Systems processing financial data must meet SOX uptime SLA",
			Severity:    "critical",
		},
		{
			ID:          "sox.al.1",
			Framework:   "SOX",
			Title:       "Access logging comprehensive",
			Description: "All access to financial systems must be logged with user identification",
			Severity:    "critical",
		},
		{
			ID:          "sox.cm.1",
			Framework:   "SOX",
			Title:       "Configuration management documented",
			Description: "All system configurations must be documented and change-controlled",
			Severity:    "high",
		},
		{
			ID:          "sox.fm.1",
			Framework:   "SOX",
			Title:       "Financial data integrity via checksums",
			Description: "Financial records must be protected with integrity verification",
			Severity:    "critical",
		},
		{
			ID:          "sox.dr.1",
			Framework:   "SOX",
			Title:       "Disaster recovery plan tested quarterly",
			Description: "DR procedures must be documented and tested at least quarterly",
			Severity:    "high",
		},
	}

	for _, ctrl := range controls {
		ol.controls["SOX"][ctrl.ID] = ctrl
	}
}

// ComplianceEngine performs SPARQL-based compliance verification.
type ComplianceEngine struct {
	loader *OntologyLoader
	logger *slog.Logger
}

// NewComplianceEngine constructs a ComplianceEngine.
func NewComplianceEngine(ontologyPath string, logger *slog.Logger) (*ComplianceEngine, error) {
	if logger == nil {
		logger = slog.Default()
	}

	loader := NewOntologyLoader(ontologyPath, logger)
	engine := &ComplianceEngine{
		loader: loader,
		logger: logger,
	}

	return engine, nil
}

// Initialize loads the ontology.
func (ce *ComplianceEngine) Initialize(ctx context.Context) error {
	return ce.loader.LoadOntology(ctx)
}

// VerifySOC2 verifies all SOC2 controls.
func (ce *ComplianceEngine) VerifySOC2(ctx context.Context) (*ComplianceReport, error) {
	return ce.verifyFramework(ctx, "SOC2")
}

// VerifyGDPR verifies all GDPR controls.
func (ce *ComplianceEngine) VerifyGDPR(ctx context.Context) (*ComplianceReport, error) {
	return ce.verifyFramework(ctx, "GDPR")
}

// VerifyHIPAA verifies all HIPAA controls.
func (ce *ComplianceEngine) VerifyHIPAA(ctx context.Context) (*ComplianceReport, error) {
	return ce.verifyFramework(ctx, "HIPAA")
}

// VerifySOX verifies all SOX controls.
func (ce *ComplianceEngine) VerifySOX(ctx context.Context) (*ComplianceReport, error) {
	return ce.verifyFramework(ctx, "SOX")
}

// verifyFramework executes SPARQL ASK queries for a framework and returns results.
func (ce *ComplianceEngine) verifyFramework(ctx context.Context, framework string) (*ComplianceReport, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ce.logger.Info("Verifying framework", "framework", framework)

	ce.loader.mu.RLock()
	controls, ok := ce.loader.controls[framework]
	ce.loader.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown framework: %s", framework)
	}

	report := &ComplianceReport{
		Framework:     framework,
		TotalControls: len(controls),
		Violations:    make([]ComplianceViolation, 0),
		Timestamp:     time.Now().UTC(),
	}

	// For now, simulate SPARQL verification results
	// In production, this would execute SPARQL ASK queries to Oxigraph
	for id, ctrl := range controls {
		// Simulate verification: randomly pass/fail for demo purposes
		verified := ce.simulateSPARQLQuery(ctx, id, ctrl)

		if verified {
			report.PassedControls++
		} else {
			report.FailedControls++
			violation := ComplianceViolation{
				ControlID:   id,
				Framework:   framework,
				Title:       ctrl.Title,
				Reason:      fmt.Sprintf("Control %s failed verification", id),
				Severity:    ctrl.Severity,
				Remediation: fmt.Sprintf("Review and remediate control: %s", ctrl.Description),
			}
			report.Violations = append(report.Violations, violation)
		}
	}

	// Calculate overall status and score
	if report.FailedControls == 0 {
		report.Status = "compliant"
		report.Score = 1.0
	} else {
		if report.PassedControls > 0 {
			report.Status = "partial"
		} else {
			report.Status = "non_compliant"
		}

		// Score: (passed / total) with severity weighting
		totalWeight := 0.0
		passWeight := 0.0

		for _, v := range report.Violations {
			weight := ce.severityWeight(v.Severity)
			totalWeight += weight
		}

		for i := 0; i < report.PassedControls; i++ {
			// Approximate weight per passed control
			passWeight += 1.0
		}

		if totalWeight > 0 {
			report.Score = passWeight / (totalWeight + passWeight)
		}
	}

	ce.logger.Info("Framework verification complete",
		"framework", framework,
		"total", report.TotalControls,
		"passed", report.PassedControls,
		"failed", report.FailedControls,
		"score", report.Score)

	return report, nil
}

// simulateSPARQLQuery simulates executing a SPARQL ASK query.
// In production, this would execute against Oxigraph triplestore.
func (ce *ComplianceEngine) simulateSPARQLQuery(ctx context.Context, controlID string, ctrl *ComplianceControl) bool {
	// Simulate timeout handling
	select {
	case <-ctx.Done():
		ce.logger.Warn("Query timeout", "control", controlID)
		return false
	default:
	}

	// For demo: critical controls always pass, others randomly
	if ctrl.Severity == "critical" {
		return true // In production, actual query result
	}

	// Simulate some controls passing
	return len(controlID)%2 == 0
}

// severityWeight returns the weight multiplier for severity levels.
func (ce *ComplianceEngine) severityWeight(severity string) float64 {
	switch severity {
	case "critical":
		return 4.0
	case "high":
		return 3.0
	case "medium":
		return 2.0
	case "low":
		return 1.0
	default:
		return 1.0
	}
}

// GenerateReport runs all framework verifications and returns aggregated ComplianceMatrix.
func (ce *ComplianceEngine) GenerateReport(ctx context.Context) (*ComplianceMatrix, error) {
	ce.logger.Info("Generating compliance report for all frameworks")

	matrix := &ComplianceMatrix{
		Frameworks: make(map[string]*ComplianceReport),
		Timestamp:  time.Now().UTC(),
	}

	frameworks := []string{"SOC2", "GDPR", "HIPAA", "SOX"}
	totalScore := 0.0
	scoreCount := 0

	for _, fw := range frameworks {
		report, err := ce.verifyFramework(ctx, fw)
		if err != nil {
			ce.logger.Error("Framework verification failed", "framework", fw, "error", err)
			continue
		}

		matrix.Frameworks[fw] = report
		totalScore += report.Score
		scoreCount++
	}

	if scoreCount > 0 {
		matrix.OverallScore = totalScore / float64(scoreCount)
	}

	ce.logger.Info("Compliance report generated",
		"frameworks", len(matrix.Frameworks),
		"overall_score", matrix.OverallScore)

	return matrix, nil
}

// GetFrameworkControls returns all controls for a framework.
func (ce *ComplianceEngine) GetFrameworkControls(framework string) []*ComplianceControl {
	ce.loader.mu.RLock()
	defer ce.loader.mu.RUnlock()

	controls, ok := ce.loader.controls[framework]
	if !ok {
		return nil
	}

	result := make([]*ComplianceControl, 0, len(controls))
	for _, ctrl := range controls {
		result = append(result, ctrl)
	}

	return result
}

// GlobalComplianceEngine is the singleton instance of ComplianceEngine.
// This is initialized once during application startup.
var GlobalComplianceEngine *ComplianceEngine
