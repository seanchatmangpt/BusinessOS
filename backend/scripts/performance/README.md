# Performance Testing Suite

**Version:** 1.0.0
**Last Updated:** 2026-01-26

Comprehensive k6-based performance testing infrastructure for BusinessOS.

## Quick Start

### 1. Install k6

```bash
./install_k6.sh
```

### 2. Start Backend

```bash
cd ../..
go run cmd/server/main.go
```

### 3. Set Environment Variables

```bash
export BASE_URL=http://localhost:8001
export AUTH_TOKEN=your_jwt_token_here
export WORKSPACE_ID=your_workspace_id_here
```

### 4. Run Tests

```bash
# OSA load test (17 minutes total)
k6 run load_test_osa.js

# Hybrid architecture test (16 minutes)
k6 run load_test_hybrid.js

# Spike test (4 minutes)
k6 run spike_test.js

# Endurance test (2 hours by default)
k6 run endurance_test.js
```

## Test Suites

### 1. OSA Load Test (`load_test_osa.js`)

Tests OSA integration endpoints under realistic load:

- **Endpoints:** `/api/osa/generate`, `/api/osa/status/:id`, `/api/osa/orchestrate`
- **Scenarios:** 100, 500, 1000 req/s
- **Duration:** 5 minutes each = 15 minutes total
- **Metrics:** P50/P95/P99 latency, error rate, success rate

**Expected Results:**
- P95 latency < 500ms (generate), < 200ms (status), < 3s (orchestrate)
- Error rate < 1%
- Success rate > 99.9%

### 2. Hybrid Architecture Test (`load_test_hybrid.js`)

Tests routing efficiency:

- **Distribution:** 70% direct path, 30% CoT path
- **Load:** Ramping 10 → 50 → 100 VUs
- **Duration:** 16 minutes total
- **Metrics:** Direct vs CoT latency, routing overhead

**Expected Results:**
- Direct path P95 < 200ms
- CoT path P95 < 3s
- Routing overhead P95 < 50ms

### 3. Spike Test (`spike_test.js`)

Tests resilience under sudden load:

- **Pattern:** 100 → 5000 → 100 req/s
- **Duration:** 4 minutes total
- **Metrics:** Circuit breaker behavior, recovery time

**Expected Results:**
- Circuit breaker activates during spike
- Error rate < 5% during spike
- Recovery time < 5s

### 4. Endurance Test (`endurance_test.js`)

Long-running stability test:

- **Load:** 500 req/s sustained
- **Duration:** 2 hours (configurable)
- **Metrics:** Memory leaks, performance degradation

**Expected Results:**
- Performance degradation < 20%
- No memory leaks
- Stable hourly metrics

## File Structure

```
scripts/performance/
├── README.md                    # This file
├── install_k6.sh                # Cross-platform k6 installation
├── load_test_osa.js             # OSA endpoints load test
├── load_test_hybrid.js          # Hybrid architecture test
├── spike_test.js                # Spike/resilience test
├── endurance_test.js            # Long-running stability test
└── analyze_results.sh           # Results comparison script
```

## Advanced Usage

### Export Results as JSON

```bash
k6 run --out json=results.json load_test_osa.js
```

### Generate HTML Report

```bash
k6 report results.json --out report.html
```

### Compare Against Baseline

```bash
# Run test
k6 run --out json=current.json load_test_osa.js

# Compare
./analyze_results.sh baseline/baseline-osa.json current.json
```

### Custom Duration (Endurance)

```bash
k6 run -e DURATION_HOURS=4 endurance_test.js
```

### Verbose Analysis

```bash
./analyze_results.sh baseline.json current.json --verbose
```

### Generate HTML Comparison

```bash
./analyze_results.sh baseline.json current.json --html
```

## CI/CD Integration

Performance tests run automatically via GitHub Actions:

- **Schedule:** Weekly (Sundays at 2 AM UTC)
- **Manual:** Via workflow dispatch
- **Alerts:** Creates issue if degradation > 20%

### Trigger Manual Run

```bash
gh workflow run performance-tests.yml
```

### Run Specific Test

```bash
gh workflow run performance-tests.yml \
  -f test_suite=osa
```

## Performance Targets

| Path | P50 | P95 | P99 | Success Rate |
|------|-----|-----|-----|--------------|
| **Direct** | < 50ms | < 100ms | < 200ms | > 99.9% |
| **CoT** | < 500ms | < 2s | < 3s | > 99.5% |

## Troubleshooting

### Connection Refused

```bash
# Check backend is running
curl http://localhost:8001/api/health

# Check port
netstat -an | grep 8001
```

### Authentication Failures

```bash
# Get fresh token
curl -X POST http://localhost:8001/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"testpassword123"}'

# Export token
export AUTH_TOKEN="eyJhbGci..."
```

### High Error Rates

Check backend logs:
```bash
tail -f ../../logs/server.log | grep ERROR
```

## Documentation

Full documentation: `../../docs/PERFORMANCE_TESTING.md`

## Support

- **Issues:** https://github.com/yourusername/BusinessOS/issues
- **Slack:** #performance-testing
- **k6 Docs:** https://k6.io/docs/

---

**Maintained by:** DevOps Team
**Review Schedule:** Quarterly
