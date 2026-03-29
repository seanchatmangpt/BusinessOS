package canopy

// IntelligencePayload is the JSON body sent to Canopy POST /api/v1/bos/intelligence.
type IntelligencePayload struct {
	// HealthSummary is a [0,1] float summarising overall system health.
	HealthSummary float64 `json:"health_summary"`
	// ConformanceScore is a [0,1] float from the last conformance check.
	ConformanceScore float64 `json:"conformance_score"`
	// TopRisk is the single highest-priority risk string.
	TopRisk string `json:"top_risk"`
	// ConwayViolations is the count of structural Conway violations detected.
	ConwayViolations int `json:"conway_violations"`
	// CaseCount is the number of active cases synced in this L0 run.
	CaseCount int `json:"case_count"`
	// HandoffCount is the number of handoff events synced in this L0 run.
	HandoffCount int `json:"handoff_count"`
	// Source identifies the originating system.
	Source string `json:"source"`
}
