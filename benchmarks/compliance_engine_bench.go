package benchmarks

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/rhl/businessos-backend/internal/ontology"
)

// BenchmarkComplianceSOC2Verification measures SOC2 compliance verification performance.
// Reports: operations/sec, mean/p95/p99 latency (ms)
// SLA: p95 < 1000ms
func BenchmarkComplianceSOC2Verification(b *testing.B) {
	loader := ontology.NewOntologyLoader("", slog.Default())
	ctx := context.Background()

	// Load ontology once
	if err := loader.LoadOntology(ctx); err != nil {
		b.Fatalf("Failed to load ontology: %v", err)
	}

	engine := ontology.NewComplianceEngine(loader, slog.Default())

	latencies := make([]float64, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		// Verify SOC2 framework
		_, err := engine.VerifyFramework(ctx, ontology.FrameworkSOC2)
		latencies[i] = float64(time.Since(start).Milliseconds())

		if err != nil {
			b.Logf("Error in SOC2 verification: %v", err)
		}
	}
	b.StopTimer()

	stats := calculateStats(latencies)
	b.Logf("ComplianceSOC2Verification Statistics (ops/sec, ms):")
	b.Logf("  Operations/sec: %.2f", float64(b.N)/b.Elapsed().Seconds())
	b.Logf("  Mean latency: %.2f ms", stats.mean)
	b.Logf("  P95 latency: %.2f ms", stats.p95)
	b.Logf("  P99 latency: %.2f ms", stats.p99)
	b.Logf("  Min latency: %.2f ms", stats.min)
	b.Logf("  Max latency: %.2f ms", stats.max)

	if stats.p95 > 1000.0 {
		b.Logf("WARNING: P95 latency (%.2f ms) exceeds SLA of 1000ms", stats.p95)
	}
}

// BenchmarkComplianceGDPRVerification measures GDPR compliance verification performance.
// Reports: operations/sec, mean/p95/p99 latency (ms)
// SLA: p95 < 1000ms
func BenchmarkComplianceGDPRVerification(b *testing.B) {
	loader := ontology.NewOntologyLoader("", slog.Default())
	ctx := context.Background()

	if err := loader.LoadOntology(ctx); err != nil {
		b.Fatalf("Failed to load ontology: %v", err)
	}

	engine := ontology.NewComplianceEngine(loader, slog.Default())

	latencies := make([]float64, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := engine.VerifyFramework(ctx, ontology.FrameworkGDPR)
		latencies[i] = float64(time.Since(start).Milliseconds())

		if err != nil {
			b.Logf("Error in GDPR verification: %v", err)
		}
	}
	b.StopTimer()

	stats := calculateStats(latencies)
	b.Logf("ComplianceGDPRVerification Statistics (ops/sec, ms):")
	b.Logf("  Operations/sec: %.2f", float64(b.N)/b.Elapsed().Seconds())
	b.Logf("  Mean latency: %.2f ms", stats.mean)
	b.Logf("  P95 latency: %.2f ms", stats.p95)
	b.Logf("  P99 latency: %.2f ms", stats.p99)
	b.Logf("  Min latency: %.2f ms", stats.min)
	b.Logf("  Max latency: %.2f ms", stats.max)

	if stats.p95 > 1000.0 {
		b.Logf("WARNING: P95 latency (%.2f ms) exceeds SLA of 1000ms", stats.p95)
	}
}

// BenchmarkComplianceHIPAAVerification measures HIPAA compliance verification performance.
// Reports: operations/sec, mean/p95/p99 latency (ms)
// SLA: p95 < 1000ms
func BenchmarkComplianceHIPAAVerification(b *testing.B) {
	loader := ontology.NewOntologyLoader("", slog.Default())
	ctx := context.Background()

	if err := loader.LoadOntology(ctx); err != nil {
		b.Fatalf("Failed to load ontology: %v", err)
	}

	engine := ontology.NewComplianceEngine(loader, slog.Default())

	latencies := make([]float64, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := engine.VerifyFramework(ctx, ontology.FrameworkHIPAA)
		latencies[i] = float64(time.Since(start).Milliseconds())

		if err != nil {
			b.Logf("Error in HIPAA verification: %v", err)
		}
	}
	b.StopTimer()

	stats := calculateStats(latencies)
	b.Logf("ComplianceHIPAAVerification Statistics (ops/sec, ms):")
	b.Logf("  Operations/sec: %.2f", float64(b.N)/b.Elapsed().Seconds())
	b.Logf("  Mean latency: %.2f ms", stats.mean)
	b.Logf("  P95 latency: %.2f ms", stats.p95)
	b.Logf("  P99 latency: %.2f ms", stats.p99)
	b.Logf("  Min latency: %.2f ms", stats.min)
	b.Logf("  Max latency: %.2f ms", stats.max)

	if stats.p95 > 1000.0 {
		b.Logf("WARNING: P95 latency (%.2f ms) exceeds SLA of 1000ms", stats.p95)
	}
}

// BenchmarkComplianceSOXVerification measures SOX compliance verification performance.
// Reports: operations/sec, mean/p95/p99 latency (ms)
// SLA: p95 < 1000ms
func BenchmarkComplianceSOXVerification(b *testing.B) {
	loader := ontology.NewOntologyLoader("", slog.Default())
	ctx := context.Background()

	if err := loader.LoadOntology(ctx); err != nil {
		b.Fatalf("Failed to load ontology: %v", err)
	}

	engine := ontology.NewComplianceEngine(loader, slog.Default())

	latencies := make([]float64, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := engine.VerifyFramework(ctx, ontology.FrameworkSOX)
		latencies[i] = float64(time.Since(start).Milliseconds())

		if err != nil {
			b.Logf("Error in SOX verification: %v", err)
		}
	}
	b.StopTimer()

	stats := calculateStats(latencies)
	b.Logf("ComplianceSOXVerification Statistics (ops/sec, ms):")
	b.Logf("  Operations/sec: %.2f", float64(b.N)/b.Elapsed().Seconds())
	b.Logf("  Mean latency: %.2f ms", stats.mean)
	b.Logf("  P95 latency: %.2f ms", stats.p95)
	b.Logf("  P99 latency: %.2f ms", stats.p99)
	b.Logf("  Min latency: %.2f ms", stats.min)
	b.Logf("  Max latency: %.2f ms", stats.max)

	if stats.p95 > 1000.0 {
		b.Logf("WARNING: P95 latency (%.2f ms) exceeds SLA of 1000ms", stats.p95)
	}
}

// BenchmarkComplianceReportGeneration measures the performance of generating a full compliance report.
// Reports: operations/sec, mean/p95/p99 latency (ms)
// SLA: p95 < 2000ms (report aggregates all frameworks)
func BenchmarkComplianceReportGeneration(b *testing.B) {
	loader := ontology.NewOntologyLoader("", slog.Default())
	ctx := context.Background()

	if err := loader.LoadOntology(ctx); err != nil {
		b.Fatalf("Failed to load ontology: %v", err)
	}

	engine := ontology.NewComplianceEngine(loader, slog.Default())

	latencies := make([]float64, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := engine.GenerateReport(ctx)
		latencies[i] = float64(time.Since(start).Milliseconds())

		if err != nil {
			b.Logf("Error in report generation: %v", err)
		}
	}
	b.StopTimer()

	stats := calculateStats(latencies)
	b.Logf("ComplianceReportGeneration Statistics (ops/sec, ms):")
	b.Logf("  Operations/sec: %.2f", float64(b.N)/b.Elapsed().Seconds())
	b.Logf("  Mean latency: %.2f ms", stats.mean)
	b.Logf("  P95 latency: %.2f ms", stats.p95)
	b.Logf("  P99 latency: %.2f ms", stats.p99)
	b.Logf("  Min latency: %.2f ms", stats.min)
	b.Logf("  Max latency: %.2f ms", stats.max)

	if stats.p95 > 2000.0 {
		b.Logf("WARNING: P95 latency (%.2f ms) exceeds SLA of 2000ms", stats.p95)
	}
}

// BenchmarkComplianceFrameworkLookup measures the performance of looking up a single control.
// Reports: operations/sec, mean/p95/p99 latency (μs)
// SLA: p95 < 10ms (single control lookup should be very fast)
func BenchmarkComplianceFrameworkLookup(b *testing.B) {
	loader := ontology.NewOntologyLoader("", slog.Default())
	ctx := context.Background()

	if err := loader.LoadOntology(ctx); err != nil {
		b.Fatalf("Failed to load ontology: %v", err)
	}

	latencies := make([]float64, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		// Look up a specific control
		control := loader.GetControl("SOC2", "soc2.cc6.1")
		if control == nil {
			b.Logf("Warning: Control soc2.cc6.1 not found")
		}

		latencies[i] = float64(time.Since(start).Microseconds())
	}
	b.StopTimer()

	stats := calculateStats(latencies)
	b.Logf("ComplianceFrameworkLookup Statistics (ops/sec, μs):")
	b.Logf("  Operations/sec: %.2f", float64(b.N)/b.Elapsed().Seconds())
	b.Logf("  Mean latency: %.2f μs", stats.mean)
	b.Logf("  P95 latency: %.2f μs", stats.p95)
	b.Logf("  P99 latency: %.2f μs", stats.p99)
	b.Logf("  Min latency: %.2f μs", stats.min)
	b.Logf("  Max latency: %.2f μs", stats.max)

	if stats.p95 > 10000.0 {
		b.Logf("WARNING: P95 latency (%.2f μs) exceeds SLA of 10000μs (10ms)", stats.p95)
	}
}

// BenchmarkComplianceMultiFrameworkVerification measures verifying multiple frameworks sequentially.
// Reports: operations/sec, mean/p95/p99 latency (ms)
// SLA: p95 < 3000ms
func BenchmarkComplianceMultiFrameworkVerification(b *testing.B) {
	loader := ontology.NewOntologyLoader("", slog.Default())
	ctx := context.Background()

	if err := loader.LoadOntology(ctx); err != nil {
		b.Fatalf("Failed to load ontology: %v", err)
	}

	engine := ontology.NewComplianceEngine(loader, slog.Default())

	latencies := make([]float64, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		// Verify all frameworks sequentially
		frameworks := []ontology.ComplianceFramework{
			ontology.FrameworkSOC2,
			ontology.FrameworkGDPR,
			ontology.FrameworkHIPAA,
			ontology.FrameworkSOX,
		}

		for _, fw := range frameworks {
			_, err := engine.VerifyFramework(ctx, fw)
			if err != nil {
				b.Logf("Error verifying framework %s: %v", fw, err)
			}
		}

		latencies[i] = float64(time.Since(start).Milliseconds())
	}
	b.StopTimer()

	stats := calculateStats(latencies)
	b.Logf("ComplianceMultiFrameworkVerification Statistics (ops/sec, ms):")
	b.Logf("  Operations/sec: %.2f", float64(b.N)/b.Elapsed().Seconds())
	b.Logf("  Mean latency: %.2f ms", stats.mean)
	b.Logf("  P95 latency: %.2f ms", stats.p95)
	b.Logf("  P99 latency: %.2f ms", stats.p99)
	b.Logf("  Min latency: %.2f ms", stats.min)
	b.Logf("  Max latency: %.2f ms", stats.max)

	if stats.p95 > 3000.0 {
		b.Logf("WARNING: P95 latency (%.2f ms) exceeds SLA of 3000ms", stats.p95)
	}
}

// BenchmarkComplianceOntologyLoad measures the performance of loading the compliance ontology.
// Reports: operations/sec, mean/p95/p99 latency (ms)
// SLA: p95 < 500ms (ontology load is typically done on startup)
func BenchmarkComplianceOntologyLoad(b *testing.B) {
	ctx := context.Background()

	latencies := make([]float64, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		loader := ontology.NewOntologyLoader("", slog.Default())
		err := loader.LoadOntology(ctx)
		latencies[i] = float64(time.Since(start).Milliseconds())

		if err != nil {
			b.Logf("Error loading ontology: %v", err)
		}
	}
	b.StopTimer()

	stats := calculateStats(latencies)
	b.Logf("ComplianceOntologyLoad Statistics (ops/sec, ms):")
	b.Logf("  Operations/sec: %.2f", float64(b.N)/b.Elapsed().Seconds())
	b.Logf("  Mean latency: %.2f ms", stats.mean)
	b.Logf("  P95 latency: %.2f ms", stats.p95)
	b.Logf("  P99 latency: %.2f ms", stats.p99)
	b.Logf("  Min latency: %.2f ms", stats.min)
	b.Logf("  Max latency: %.2f ms", stats.max)

	if stats.p95 > 500.0 {
		b.Logf("WARNING: P95 latency (%.2f ms) exceeds SLA of 500ms", stats.p95)
	}
}
