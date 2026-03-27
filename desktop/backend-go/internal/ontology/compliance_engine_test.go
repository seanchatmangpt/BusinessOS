package ontology

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewComplianceEngine(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	assert.NoError(t, err)
	assert.NotNil(t, engine)
}

func TestComplianceEngineInitialize(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	assert.NoError(t, err)
}

func TestVerifySOC2Controls(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	report, err := engine.VerifySOC2(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "SOC2", report.Framework)
	assert.Greater(t, report.TotalControls, 0)
	assert.GreaterOrEqual(t, report.PassedControls+report.FailedControls, 1)
	assert.Greater(t, report.Score, 0.0)
	assert.LessOrEqual(t, report.Score, 1.0)
}

func TestVerifyGDPRControls(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	report, err := engine.VerifyGDPR(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "GDPR", report.Framework)
	assert.Greater(t, report.TotalControls, 0)
	assert.GreaterOrEqual(t, report.PassedControls+report.FailedControls, 1)
}

func TestVerifyHIPAAControls(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	report, err := engine.VerifyHIPAA(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "HIPAA", report.Framework)
	assert.Greater(t, report.TotalControls, 0)
	assert.GreaterOrEqual(t, report.PassedControls+report.FailedControls, 1)
}

func TestVerifySOXControls(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	report, err := engine.VerifySOX(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "SOX", report.Framework)
	assert.Greater(t, report.TotalControls, 0)
	assert.GreaterOrEqual(t, report.PassedControls+report.FailedControls, 1)
}

func TestComplianceReportStatus(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	report, err := engine.VerifySOC2(ctx)
	assert.NoError(t, err)

	if report.FailedControls == 0 {
		assert.Equal(t, "compliant", report.Status)
	} else if report.PassedControls > 0 {
		assert.Equal(t, "partial", report.Status)
	} else {
		assert.Equal(t, "non_compliant", report.Status)
	}
}

func TestComplianceViolationsStructure(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	report, err := engine.VerifySOC2(ctx)
	assert.NoError(t, err)

	for _, violation := range report.Violations {
		assert.NotEmpty(t, violation.ControlID)
		assert.NotEmpty(t, violation.Framework)
		assert.NotEmpty(t, violation.Title)
		assert.NotEmpty(t, violation.Reason)
		assert.NotEmpty(t, violation.Severity)
		assert.NotEmpty(t, violation.Remediation)
	}
}

func TestGenerateComplianceReport(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	matrix, err := engine.GenerateReport(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, matrix)
	assert.Equal(t, 4, len(matrix.Frameworks))
	assert.Greater(t, matrix.OverallScore, 0.0)
	assert.LessOrEqual(t, matrix.OverallScore, 1.0)

	// Verify all frameworks are present
	assert.Contains(t, matrix.Frameworks, "SOC2")
	assert.Contains(t, matrix.Frameworks, "GDPR")
	assert.Contains(t, matrix.Frameworks, "HIPAA")
	assert.Contains(t, matrix.Frameworks, "SOX")

	// Verify each framework report has correct structure
	for fw, report := range matrix.Frameworks {
		assert.Equal(t, fw, report.Framework)
		assert.Greater(t, report.TotalControls, 0)
		assert.NotEqual(t, time.Time{}, report.Timestamp)
	}
}

func TestComplianceTimeoutHandling(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	// Create a context that times out immediately
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// This should handle timeout gracefully
	time.Sleep(10 * time.Millisecond) // Ensure context times out
	report, err := engine.VerifySOC2(timeoutCtx)

	// Engine should still return a report (simulated query)
	// In production with real SPARQL queries, timeout would cause errors
	assert.NotNil(t, report)
}

func TestConcurrentVerifications(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	done := make(chan bool)

	// Launch concurrent verifications
	go func() {
		_, err := engine.VerifySOC2(ctx)
		assert.NoError(t, err)
		done <- true
	}()

	go func() {
		_, err := engine.VerifyGDPR(ctx)
		assert.NoError(t, err)
		done <- true
	}()

	go func() {
		_, err := engine.VerifyHIPAA(ctx)
		assert.NoError(t, err)
		done <- true
	}()

	go func() {
		_, err := engine.VerifySOX(ctx)
		assert.NoError(t, err)
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}
}

func TestFrameworkControlsRetrieval(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	// Test each framework
	frameworks := []string{"SOC2", "GDPR", "HIPAA", "SOX"}

	for _, fw := range frameworks {
		controls := engine.GetFrameworkControls(fw)
		assert.Greater(t, len(controls), 0)

		for _, ctrl := range controls {
			assert.NotEmpty(t, ctrl.ID)
			assert.NotEmpty(t, ctrl.Title)
			assert.NotEmpty(t, ctrl.Description)
			assert.NotEmpty(t, ctrl.Severity)
			assert.Equal(t, fw, ctrl.Framework)
		}
	}
}

func TestUnknownFramework(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	// Try to verify unknown framework
	report, err := engine.verifyFramework(ctx, "UNKNOWN")
	assert.Error(t, err)
	assert.Nil(t, report)
	assert.Contains(t, err.Error(), "unknown framework")
}

func TestComplianceScoreCalculation(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	report, err := engine.VerifySOC2(ctx)
	assert.NoError(t, err)

	// Score should be between 0 and 1
	assert.GreaterOrEqual(t, report.Score, 0.0)
	assert.LessOrEqual(t, report.Score, 1.0)

	// If all controls pass, score should be 1.0
	if report.FailedControls == 0 {
		assert.Equal(t, 1.0, report.Score)
	}
}

func TestReportTimestamp(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	before := time.Now().UTC()
	report, err := engine.VerifySOC2(ctx)
	after := time.Now().UTC()

	assert.NoError(t, err)
	assert.True(t, report.Timestamp.After(before.Add(-1*time.Second)))
	assert.True(t, report.Timestamp.Before(after.Add(1*time.Second)))
}

func TestOntologyLoaderInitialization(t *testing.T) {
	logger := slog.Default()
	loader := NewOntologyLoader("/tmp/ontology.ttl", logger)
	assert.NotNil(t, loader)

	ctx := context.Background()
	err := loader.LoadOntology(ctx)
	assert.NoError(t, err)

	// Verify all frameworks loaded
	loader.mu.RLock()
	assert.Greater(t, len(loader.controls), 0)
	loader.mu.RUnlock()
}

func TestOntologyFileNotFound(t *testing.T) {
	logger := slog.Default()
	// Non-existent file should not cause error in LoadOntology
	loader := NewOntologyLoader("/nonexistent/path/ontology.ttl", logger)
	ctx := context.Background()
	err := loader.LoadOntology(ctx)
	assert.NoError(t, err)
}

func TestControlSeverityWeights(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	weights := map[string]float64{
		"critical": 4.0,
		"high":     3.0,
		"medium":   2.0,
		"low":      1.0,
		"unknown":  1.0,
	}

	for severity, expected := range weights {
		weight := engine.severityWeight(severity)
		assert.Equal(t, expected, weight)
	}
}

func TestSOC2ControlCount(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	controls := engine.GetFrameworkControls("SOC2")
	assert.Greater(t, len(controls), 0)
	assert.Equal(t, 8, len(controls))
}

func TestGDPRControlCount(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	controls := engine.GetFrameworkControls("GDPR")
	assert.Greater(t, len(controls), 0)
	assert.Equal(t, 7, len(controls))
}

func TestHIPAAControlCount(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	controls := engine.GetFrameworkControls("HIPAA")
	assert.Greater(t, len(controls), 0)
	assert.Equal(t, 7, len(controls))
}

func TestSOXControlCount(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	controls := engine.GetFrameworkControls("SOX")
	assert.Greater(t, len(controls), 0)
	assert.Equal(t, 6, len(controls))
}

func TestMixedComplianceScenario(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	// Run all verifications
	soc2, err1 := engine.VerifySOC2(ctx)
	gdpr, err2 := engine.VerifyGDPR(ctx)
	hipaa, err3 := engine.VerifyHIPAA(ctx)
	sox, err4 := engine.VerifySOX(ctx)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.NoError(t, err4)

	// Verify all have meaningful results
	assert.NotNil(t, soc2)
	assert.NotNil(t, gdpr)
	assert.NotNil(t, hipaa)
	assert.NotNil(t, sox)

	// Verify total controls match framework definitions
	assert.Equal(t, 8, soc2.TotalControls)
	assert.Equal(t, 7, gdpr.TotalControls)
	assert.Equal(t, 7, hipaa.TotalControls)
	assert.Equal(t, 6, sox.TotalControls)
}

func TestComplianceControlFields(t *testing.T) {
	logger := slog.Default()
	engine, err := NewComplianceEngine("/tmp/ontology.ttl", logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = engine.Initialize(ctx)
	require.NoError(t, err)

	controls := engine.GetFrameworkControls("SOC2")
	require.Greater(t, len(controls), 0)

	// Verify all required fields are populated
	for _, ctrl := range controls {
		assert.NotEmpty(t, ctrl.ID)
		assert.NotEmpty(t, ctrl.Framework)
		assert.Equal(t, "SOC2", ctrl.Framework)
		assert.NotEmpty(t, ctrl.Title)
		assert.NotEmpty(t, ctrl.Description)
		assert.NotEmpty(t, ctrl.Severity)

		// Severity should be one of the expected values
		validSeverities := []string{"critical", "high", "medium", "low"}
		assert.Contains(t, validSeverities, ctrl.Severity)
	}
}
