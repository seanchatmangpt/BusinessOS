package benchmarks

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rhl/businessos-backend/internal/services"
)

// BenchmarkCreateDeal measures the performance of deal creation.
// Reports: operations/sec, mean/p95/p99 latency (ms)
// SLA: p95 < 500ms
func BenchmarkCreateDeal(b *testing.B) {
	service := services.NewFIBODealsService("http://localhost:8890")
	ctx := context.Background()

	latencies := make([]float64, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		deal := &services.Deal{
			Name:      fmt.Sprintf("Deal_%d", i),
			Amount:    100000.0 + float64(i),
			Currency:  "USD",
			BuyerID:   fmt.Sprintf("buyer_%d", i%10),
			SellerID:  fmt.Sprintf("seller_%d", i%10),
			Stage:     "prospecting",
			Probability: 50,
		}

		_, err := service.CreateDeal(ctx, deal)
		latencies[i] = float64(time.Since(start).Milliseconds())

		if err != nil {
			b.Logf("Error in CreateDeal: %v", err)
		}
	}
	b.StopTimer()

	// Calculate statistics
	stats := calculateStats(latencies)
	b.Logf("CreateDeal Statistics (ops/sec, ms):")
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

// BenchmarkGetDeal measures the performance of retrieving a deal by ID.
// Reports: operations/sec, mean/p95/p99 latency (ms)
// SLA: p95 < 500ms
func BenchmarkGetDeal(b *testing.B) {
	service := services.NewFIBODealsService("http://localhost:8890")
	ctx := context.Background()

	// Pre-create a deal to retrieve
	deal := &services.Deal{
		ID:        "bench_deal_001",
		Name:      "Benchmark Deal",
		Amount:    500000.0,
		Currency:  "USD",
		BuyerID:   "buyer_bench",
		SellerID:  "seller_bench",
		Stage:     "negotiation",
		Probability: 75,
	}
	service.CreateDeal(ctx, deal)

	latencies := make([]float64, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := service.GetDeal(ctx, "bench_deal_001")
		latencies[i] = float64(time.Since(start).Milliseconds())

		if err != nil {
			b.Logf("Error in GetDeal: %v", err)
		}
	}
	b.StopTimer()

	stats := calculateStats(latencies)
	b.Logf("GetDeal Statistics (ops/sec, ms):")
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

// BenchmarkListDeals measures the performance of listing deals with pagination.
// Reports: operations/sec, mean/p95/p99 latency (ms)
// SLA: p95 < 500ms
func BenchmarkListDeals(b *testing.B) {
	service := services.NewFIBODealsService("http://localhost:8890")
	ctx := context.Background()

	latencies := make([]float64, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := service.ListDeals(ctx, 50, 0)
		latencies[i] = float64(time.Since(start).Milliseconds())

		if err != nil {
			b.Logf("Error in ListDeals: %v", err)
		}
	}
	b.StopTimer()

	stats := calculateStats(latencies)
	b.Logf("ListDeals Statistics (ops/sec, ms):")
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

// BenchmarkListDealsLargePage measures listing with larger page size (500 items).
// Reports: operations/sec, mean/p95/p99 latency (ms)
// SLA: p95 < 1000ms
func BenchmarkListDealsLargePage(b *testing.B) {
	service := services.NewFIBODealsService("http://localhost:8890")
	ctx := context.Background()

	latencies := make([]float64, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := service.ListDeals(ctx, 500, 0)
		latencies[i] = float64(time.Since(start).Milliseconds())

		if err != nil {
			b.Logf("Error in ListDealsLargePage: %v", err)
		}
	}
	b.StopTimer()

	stats := calculateStats(latencies)
	b.Logf("ListDealsLargePage Statistics (ops/sec, ms):")
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

// BenchmarkUpdateDeal measures the performance of updating a deal.
// Reports: operations/sec, mean/p95/p99 latency (ms)
// SLA: p95 < 500ms
func BenchmarkUpdateDeal(b *testing.B) {
	service := services.NewFIBODealsService("http://localhost:8890")
	ctx := context.Background()

	// Pre-create a deal
	deal := &services.Deal{
		ID:        "bench_update_001",
		Name:      "Update Benchmark",
		Amount:    250000.0,
		Currency:  "USD",
		BuyerID:   "buyer_update",
		SellerID:  "seller_update",
		Stage:     "proposal",
		Probability: 60,
	}
	service.CreateDeal(ctx, deal)

	latencies := make([]float64, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		updates := map[string]interface{}{
			"dealAmount": 250000.0 + float64(i),
			"currentStage": "proposal",
		}

		_, err := service.UpdateDeal(ctx, "bench_update_001", updates)
		latencies[i] = float64(time.Since(start).Milliseconds())

		if err != nil {
			b.Logf("Error in UpdateDeal: %v", err)
		}
	}
	b.StopTimer()

	stats := calculateStats(latencies)
	b.Logf("UpdateDeal Statistics (ops/sec, ms):")
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

// BenchmarkVerifyDealCompliance measures the performance of compliance verification.
// Reports: operations/sec, mean/p95/p99 latency (ms)
// SLA: p95 < 1000ms (compliance checks are more expensive)
func BenchmarkVerifyDealCompliance(b *testing.B) {
	service := services.NewFIBODealsService("http://localhost:8890")
	ctx := context.Background()

	// Pre-create a deal for compliance verification
	deal := &services.Deal{
		ID:        "bench_compliance_001",
		Name:      "Compliance Benchmark Deal",
		Amount:    1000000.0,
		Currency:  "USD",
		BuyerID:   "buyer_comp",
		SellerID:  "seller_comp",
		Stage:     "due_diligence",
		Probability: 80,
	}
	service.CreateDeal(ctx, deal)

	latencies := make([]float64, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := service.VerifyCompliance(ctx, "bench_compliance_001")
		latencies[i] = float64(time.Since(start).Milliseconds())

		if err != nil {
			b.Logf("Error in VerifyCompliance: %v", err)
		}
	}
	b.StopTimer()

	stats := calculateStats(latencies)
	b.Logf("VerifyDealCompliance Statistics (ops/sec, ms):")
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

// BenchmarkDealLifecycle measures a complete deal workflow: create -> get -> update -> verify.
// Reports: operations/sec, mean/p95/p99 latency (ms)
// SLA: p95 < 3000ms (aggregate of all operations)
func BenchmarkDealLifecycle(b *testing.B) {
	service := services.NewFIBODealsService("http://localhost:8890")
	ctx := context.Background()

	latencies := make([]float64, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		// Create deal
		deal := &services.Deal{
			ID:        fmt.Sprintf("bench_lifecycle_%d", i),
			Name:      fmt.Sprintf("Lifecycle Deal %d", i),
			Amount:    500000.0,
			Currency:  "USD",
			BuyerID:   "buyer_lifecycle",
			SellerID:  "seller_lifecycle",
			Stage:     "prospecting",
			Probability: 50,
		}
		_, err := service.CreateDeal(ctx, deal)
		if err != nil {
			b.Logf("Error in CreateDeal: %v", err)
			continue
		}

		// Get deal
		_, err = service.GetDeal(ctx, deal.ID)
		if err != nil {
			b.Logf("Error in GetDeal: %v", err)
			continue
		}

		// Update deal
		updates := map[string]interface{}{
			"currentStage": "negotiation",
			"dealProbability": 75,
		}
		_, err = service.UpdateDeal(ctx, deal.ID, updates)
		if err != nil {
			b.Logf("Error in UpdateDeal: %v", err)
			continue
		}

		// Verify compliance
		_, err = service.VerifyCompliance(ctx, deal.ID)
		if err != nil {
			b.Logf("Error in VerifyCompliance: %v", err)
			continue
		}

		latencies[i] = float64(time.Since(start).Milliseconds())
	}
	b.StopTimer()

	stats := calculateStats(latencies)
	b.Logf("DealLifecycle Statistics (ops/sec, ms):")
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

// BenchmarkListDealsPagination measures pagination performance at different offsets.
// Reports: operations/sec, mean/p95/p99 latency (ms)
// SLA: p95 < 500ms
func BenchmarkListDealsPagination(b *testing.B) {
	service := services.NewFIBODealsService("http://localhost:8890")
	ctx := context.Background()

	latencies := make([]float64, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		// Vary offset to test pagination at different positions
		offset := (i % 10) * 50

		_, err := service.ListDeals(ctx, 50, offset)
		latencies[i] = float64(time.Since(start).Milliseconds())

		if err != nil {
			b.Logf("Error in ListDealsPagination: %v", err)
		}
	}
	b.StopTimer()

	stats := calculateStats(latencies)
	b.Logf("ListDealsPagination Statistics (ops/sec, ms):")
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

// ============================================================================
// Statistics Helper
// ============================================================================

type LatencyStats struct {
	mean float64
	p95  float64
	p99  float64
	min  float64
	max  float64
}

func calculateStats(latencies []float64) LatencyStats {
	if len(latencies) == 0 {
		return LatencyStats{}
	}

	// Sort latencies for percentile calculation
	sorted := make([]float64, len(latencies))
	copy(sorted, latencies)
	quickSort(sorted)

	// Calculate mean
	sum := 0.0
	for _, l := range latencies {
		sum += l
	}
	mean := sum / float64(len(latencies))

	// Calculate percentiles
	p95Idx := (95 * len(sorted)) / 100
	if p95Idx >= len(sorted) {
		p95Idx = len(sorted) - 1
	}
	p95 := sorted[p95Idx]

	p99Idx := (99 * len(sorted)) / 100
	if p99Idx >= len(sorted) {
		p99Idx = len(sorted) - 1
	}
	p99 := sorted[p99Idx]

	return LatencyStats{
		mean: mean,
		p95:  p95,
		p99:  p99,
		min:  sorted[0],
		max:  sorted[len(sorted)-1],
	}
}

func quickSort(arr []float64) {
	if len(arr) <= 1 {
		return
	}
	quickSortHelper(arr, 0, len(arr)-1)
}

func quickSortHelper(arr []float64, low, high int) {
	if low < high {
		pi := partition(arr, low, high)
		quickSortHelper(arr, low, pi-1)
		quickSortHelper(arr, pi+1, high)
	}
}

func partition(arr []float64, low, high int) int {
	pivot := arr[high]
	i := low - 1

	for j := low; j < high; j++ {
		if arr[j] < pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}
