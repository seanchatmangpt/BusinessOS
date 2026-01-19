# Deep Research Agent - Cost Model Spreadsheet

**Date:** 2026-01-19
**Purpose:** Quick reference cost calculations at different scales

---

## 📊 Cost Calculator - Interactive Model

### Assumptions (Editable)

```
LLM Model (Primary): Claude 3.5 Sonnet
LLM Model (Optimized): Claude 3.5 Haiku
Search API: Tavily AI Pro
Database: PostgreSQL (Supabase)
Compute: GCP Cloud Run (us-central1)
```

---

## 1. Per-Task Cost Breakdown

### Base Configuration (Unoptimized)

| Component | Input Tokens | Output Tokens | Cost (Sonnet) |
|-----------|-------------|---------------|---------------|
| Planning | 500 | 300 | $0.0024 |
| Sub-query generation | 800 | 400 | $0.0036 |
| Search summarization (5x) | 15,000 | 2,000 | $0.0750 |
| Relevance scoring | 5,000 | 500 | $0.0225 |
| Report generation | 20,000 | 3,000 | $0.1050 |
| Citation formatting | 2,000 | 500 | $0.0090 |
| **Subtotal: LLM** | **43,300** | **6,700** | **$0.2175** |
| Search API (5 queries) | - | - | $0.0100 |
| Database write | - | - | $0.0001 |
| Compute (Cloud Run) | - | - | $0.0007 |
| **TOTAL PER TASK** | - | - | **$0.2283** |

---

### Optimized Configuration (Haiku for sub-tasks)

| Component | Input Tokens | Output Tokens | Model | Cost |
|-----------|-------------|---------------|-------|------|
| Planning | 500 | 300 | Haiku | $0.0002 |
| Sub-query generation | 800 | 400 | Haiku | $0.0004 |
| Search summarization (5x) | 15,000 | 2,000 | Haiku | $0.0063 |
| Relevance scoring | 5,000 | 500 | Haiku | $0.0019 |
| Report generation | 20,000 | 3,000 | **Sonnet** | $0.1050 |
| Citation formatting | 2,000 | 500 | Haiku | $0.0008 |
| **Subtotal: LLM** | **43,300** | **6,700** | - | **$0.1146** |
| Search API (5 queries) | - | - | Tavily | $0.0100 |
| Database write | - | - | PG | $0.0001 |
| Compute (Cloud Run) | - | - | GCP | $0.0007 |
| **TOTAL PER TASK** | - | - | - | **$0.1254** |

**Savings: 45% reduction** ($0.2283 → $0.1254)

---

### Aggressive Optimization (Caching + Token Reduction)

| Component | Base Cost | Cache Hit (40%) | Token Reduction (20%) | Final Cost |
|-----------|-----------|-----------------|----------------------|------------|
| LLM costs | $0.1146 | $0.0688 | $0.0550 | $0.0550 |
| Search API | $0.0100 | $0.0060 | $0.0060 | $0.0060 |
| Database | $0.0001 | $0.0001 | $0.0001 | $0.0001 |
| Compute | $0.0007 | $0.0007 | $0.0007 | $0.0007 |
| **TOTAL** | **$0.1254** | **$0.0756** | **$0.0618** | **$0.0618** |

**Total savings: 73% reduction** ($0.2283 → $0.0618)

**Still 6.2x above $0.01 target**

---

## 2. Monthly Cost at Scale

### Scenario A: Unoptimized Costs

| Scale | Tasks/Day | Tasks/Month | LLM | Search | DB | Compute | **TOTAL** | $/Task |
|-------|-----------|-------------|-----|--------|----|---------|-----------:|-------:|
| 1x | 10 | 300 | $65.25 | $3.00 | $0.03 | $0.22 | **$68.50** | $0.2283 |
| 10x | 100 | 3,000 | $652.50 | $30.00 | $0.25 | $2.22 | **$684.97** | $0.2283 |
| 100x | 1,000 | 30,000 | $6,525.00 | $300.00 | $2.56 | $22.21 | **$6,849.77** | $0.2283 |
| 1000x | 10,000 | 300,000 | $65,250.00 | $3,000.00 | $25.60 | $222.12 | **$68,497.72** | $0.2283 |

---

### Scenario B: Optimized (Haiku)

| Scale | Tasks/Day | Tasks/Month | LLM | Search | DB | Compute | **TOTAL** | $/Task |
|-------|-----------|-------------|-----|--------|----|---------|-----------:|-------:|
| 1x | 10 | 300 | $34.38 | $3.00 | $0.03 | $0.22 | **$37.63** | $0.1254 |
| 10x | 100 | 3,000 | $343.80 | $30.00 | $0.25 | $2.22 | **$376.27** | $0.1254 |
| 100x | 1,000 | 30,000 | $3,438.00 | $300.00 | $2.56 | $22.21 | **$3,762.77** | $0.1254 |
| 1000x | 10,000 | 300,000 | $34,380.00 | $3,000.00 | $25.60 | $222.12 | **$37,627.72** | $0.1254 |

**Savings vs Unoptimized:** 45%

---

### Scenario C: Aggressive Optimization

| Scale | Tasks/Day | Tasks/Month | LLM | Search | DB | Compute | **TOTAL** | $/Task |
|-------|-----------|-------------|-----|--------|----|---------|-----------:|-------:|
| 1x | 10 | 300 | $16.50 | $1.80 | $0.03 | $0.22 | **$18.55** | $0.0618 |
| 10x | 100 | 3,000 | $165.00 | $18.00 | $0.25 | $2.22 | **$185.47** | $0.0618 |
| 100x | 1,000 | 30,000 | $1,650.00 | $180.00 | $2.56 | $22.21 | **$1,854.77** | $0.0618 |
| 1000x | 10,000 | 300,000 | $16,500.00 | $1,800.00 | $25.60 | $222.12 | **$18,547.72** | $0.0618 |

**Savings vs Unoptimized:** 73%

---

### Scenario D: Ultra-Optimized (Future State)

**Assumptions:**
- Self-hosted Mixtral 8x7B (90% LLM cost reduction)
- Vector cache (70% hit rate)
- Enterprise search API pricing (50% reduction)
- Optimized compute (batch processing)

| Scale | Tasks/Day | Tasks/Month | LLM | Search | DB | Compute | **TOTAL** | $/Task |
|-------|-----------|-------------|-----|--------|----|---------|-----------:|-------:|
| 1x | 10 | 300 | $1.65 | $0.90 | $0.03 | $0.15 | **$2.73** | $0.0091 |
| 10x | 100 | 3,000 | $16.50 | $9.00 | $0.25 | $1.50 | **$27.25** | $0.0091 |
| 100x | 1,000 | 30,000 | $165.00 | $90.00 | $2.56 | $15.00 | **$272.56** | $0.0091 |
| 1000x | 10,000 | 300,000 | $1,650.00 | $900.00 | $25.60 | $150.00 | **$2,725.60** | $0.0091 |

**Achieves $0.01/task target** ✅

---

## 3. Revenue Model Scenarios

### Pricing Tier Assumptions

| Tier | Price/Month | Tasks Included | Effective $/Task |
|------|-------------|----------------|------------------|
| Free | $0 | 10 | $0.00 |
| Pro | $15 | 150 | $0.10 |
| Business | $49 | 1,000 | $0.049 |
| Enterprise | Custom | Unlimited | Negotiated |

---

### Revenue Projection - 1,000 Users

**User Distribution:**
- Free: 800 users (80%) → 8,000 tasks/month
- Pro: 150 users (15%) → 22,500 tasks/month
- Business: 50 users (5%) → 50,000 tasks/month
- **Total: 80,500 tasks/month**

**Revenue:**
- Free: $0
- Pro: 150 × $15 = $2,250
- Business: 50 × $49 = $2,450
- **Total Revenue: $4,700/month**

**Costs (Aggressive Optimization):**
- Cost/task: $0.0618
- Total tasks: 80,500
- **Total Cost: $4,975/month**

**Gross Margin: -6%** ❌ (BREAK-EVEN NEEDED)

---

### Revenue Projection - 1,000 Users (Optimized Pricing)

**User Distribution:**
- Free: 800 users (80%) → 8,000 tasks/month
- Pro: 150 users (15%) × $19 → 22,500 tasks/month
- Business: 50 users (5%) × $59 → 50,000 tasks/month

**Revenue:**
- Free: $0
- Pro: 150 × $19 = $2,850
- Business: 50 × $59 = $2,950
- **Total Revenue: $5,800/month**

**Costs:**
- Total Cost: $4,975/month (same as above)

**Gross Margin: 14%** ⚠️ (MARGINAL)

---

### Revenue Projection - 10,000 Users (Scale)

**User Distribution:**
- Free: 8,000 users (80%) → 80,000 tasks/month
- Pro: 1,500 users (15%) → 225,000 tasks/month
- Business: 500 users (5%) → 500,000 tasks/month
- **Total: 805,000 tasks/month**

**Revenue:**
- Free: $0
- Pro: 1,500 × $19 = $28,500
- Business: 500 × $59 = $29,500
- **Total Revenue: $58,000/month**

**Costs (Aggressive Optimization):**
- Cost/task: $0.0618
- Total tasks: 805,000
- **Total Cost: $49,749/month**

**Gross Margin: 14%** ⚠️

**Costs (Ultra-Optimized):**
- Cost/task: $0.0091
- Total tasks: 805,000
- **Total Cost: $7,326/month**

**Gross Margin: 87%** ✅ (TARGET)

---

## 4. Break-Even Analysis

### Question: At what point do we break even?

**Assumptions:**
- Cost/task: $0.0618 (aggressive optimization)
- Pricing: Pro $19/month for 150 tasks
- Conversion rate: 15% to Pro

### Break-Even Calculation

| Total Users | Pro Users (15%) | Revenue | Tasks/Month | Cost | Gross Margin |
|-------------|-----------------|---------|-------------|------|--------------|
| 500 | 75 | $1,425 | 40,250 | $2,487 | **-75%** ❌ |
| 1,000 | 150 | $2,850 | 80,500 | $4,975 | **-75%** ❌ |
| 5,000 | 750 | $14,250 | 402,500 | $24,875 | **-75%** ❌ |
| 10,000 | 1,500 | $28,500 | 805,000 | $49,749 | **-75%** ❌ |

**CRITICAL FINDING:** Revenue scales linearly, but costs scale with usage (not revenue).

**Problem:** Free tier users consume 80% of costs but generate 0% of revenue.

---

### Revised Break-Even (Free Tier Reduced to 5 tasks/month)

**User Distribution:**
- Free: 8,000 users → 40,000 tasks/month (5 tasks each)
- Pro: 1,500 users → 225,000 tasks/month (150 tasks each)
- Business: 500 users → 500,000 tasks/month (1,000 tasks each)
- **Total: 765,000 tasks/month**

**Revenue:** $58,000/month (same)

**Costs:**
- Cost/task: $0.0618
- Total tasks: 765,000
- **Total Cost: $47,277/month**

**Gross Margin: 18%** ⚠️ (IMPROVED)

---

### Break-Even with Ultra-Optimization

**Costs:**
- Cost/task: $0.0091
- Total tasks: 765,000
- **Total Cost: $6,962/month**

**Gross Margin: 88%** ✅ (ACHIEVABLE AT SCALE)

---

## 5. Sensitivity Analysis

### Variable: Cost per Task

| Cost/Task | Monthly Cost (100k tasks) | Revenue (1k Pro users @ $19) | Gross Margin |
|-----------|---------------------------|------------------------------|--------------|
| $0.2283 | $22,830 | $19,000 | **-20%** ❌ |
| $0.1254 | $12,540 | $19,000 | **34%** ⚠️ |
| $0.0618 | $6,180 | $19,000 | **67%** ✅ |
| $0.0091 | $910 | $19,000 | **95%** ✅✅ |
| $0.0050 | $500 | $19,000 | **97%** ✅✅ |
| **$0.0010** | **$100** | **$19,000** | **99%** 🎯 |

**Target:** Achieve $0.01/task or lower for 70%+ margin.

---

### Variable: Pro Tier Pricing

| Pro Price | Revenue (1,500 users) | Cost (225k tasks @ $0.0618) | Gross Margin |
|-----------|-----------------------|-----------------------------|--------------|
| $9 | $13,500 | $13,905 | **-3%** ❌ |
| $12 | $18,000 | $13,905 | **23%** ⚠️ |
| $15 | $22,500 | $13,905 | **38%** ⚠️ |
| $19 | $28,500 | $13,905 | **51%** ✅ |
| $25 | $37,500 | $13,905 | **63%** ✅ |
| $29 | $43,500 | $13,905 | **68%** ✅ |

**Sweet spot:** $19-25/month for 50-60% margin at current costs.

---

### Variable: Conversion Rate

| Conversion to Pro | Pro Users (10k total) | Revenue @ $19 | Cost (tasks @ $0.0618) | Gross Margin |
|-------------------|-----------------------|---------------|------------------------|--------------|
| 5% | 500 | $9,500 | $47,277 | **-398%** ❌ |
| 10% | 1,000 | $19,000 | $47,277 | **-149%** ❌ |
| 15% | 1,500 | $28,500 | $47,277 | **-66%** ❌ |
| 20% | 2,000 | $38,000 | $47,277 | **-24%** ❌ |
| 30% | 3,000 | $57,000 | $47,277 | **17%** ⚠️ |
| 40% | 4,000 | $76,000 | $47,277 | **38%** ✅ |

**CRITICAL:** Need 30-40% conversion OR lower costs to break even.

---

## 6. Path to $0.01/Task

### Optimization Roadmap

| Stage | Actions | Cost/Task | % Reduction | Timeline |
|-------|---------|-----------|-------------|----------|
| **Stage 0** | Baseline (all Sonnet) | $0.2283 | 0% | Current |
| **Stage 1** | Use Haiku for sub-tasks | $0.1254 | 45% | Week 1 |
| **Stage 2** | Implement caching (40% hit rate) | $0.0752 | 40% | Week 2-3 |
| **Stage 3** | Token reduction (20%) | $0.0602 | 20% | Week 4-6 |
| **Stage 4** | Search deduplication | $0.0482 | 20% | Month 2 |
| **Stage 5** | Vector cache (70% hit rate) | $0.0241 | 50% | Month 3 |
| **Stage 6** | Self-hosted model (Mixtral) | $0.0072 | 70% | Month 4-6 |
| **TARGET** | **All optimizations** | **< $0.0100** | **96%** | **6 months** |

**Conclusion:** $0.01/task is achievable with 6-month optimization roadmap.

---

## 7. Quick Decision Matrix

### Should we build this feature?

| Factor | Score (1-10) | Weight | Weighted |
|--------|--------------|--------|----------|
| **User value** (time savings) | 10 | 30% | 3.0 |
| **Competitive advantage** | 9 | 20% | 1.8 |
| **Technical feasibility** | 8 | 15% | 1.2 |
| **Cost viability** | 6 | 20% | 1.2 |
| **Revenue potential** | 7 | 15% | 1.05 |
| **TOTAL SCORE** | - | 100% | **8.25/10** ✅ |

**Verdict: PROCEED** (strong value prop, clear optimization path)

---

## 8. Cost Control Mechanisms

### Rate Limiting Strategy

| User Tier | Limit | Per | Soft Cap | Hard Cap |
|-----------|-------|-----|----------|----------|
| Free | 10 tasks | month | 10 | 10 |
| Free | 2 tasks | hour | 3 | 5 |
| Pro | 150 tasks | month | 200 | 250 |
| Pro | 10 tasks | hour | 15 | 20 |
| Business | 1,000 tasks | month | 1,500 | 2,000 |
| Business | 50 tasks | hour | 100 | 150 |

**Soft cap:** Warning message
**Hard cap:** Upgrade prompt

---

### Budget Alerts

| Threshold | Daily Spend | Action |
|-----------|-------------|--------|
| 50% budget | $50 | Email notification |
| 80% budget | $80 | Slack alert |
| 100% budget | $100 | Auto-disable new tasks |
| 120% budget | $120 | Emergency shutdown |

**Daily budget:** $100/day = $3,000/month

---

## 9. Monitoring Metrics

### Cost Metrics (Track Daily)

```
- LLM tokens per task (input/output breakdown)
- Search API calls per task
- Average cost per task
- Cache hit rate
- Daily spend vs budget
- Cost by user tier
- Cost by time of day
```

### Revenue Metrics (Track Weekly)

```
- New Pro conversions
- Churn rate by tier
- ARPU (Average Revenue Per User)
- LTV (Lifetime Value)
- CAC (Customer Acquisition Cost)
- Gross margin by tier
```

### Quality Metrics (Track Per Task)

```
- Time to completion
- Source count
- Citation count
- User rating (1-5 stars)
- Retry rate
- Error rate
```

---

## 10. Conclusion

### Summary Table

| Metric | Current | Target | Gap | Achievable? |
|--------|---------|--------|-----|-------------|
| Cost/task | $0.2283 | $0.0100 | 22.8x | ⚠️ Requires optimization |
| Time to complete | Unknown | < 3 min | TBD | ✅ Likely |
| Gross margin | -75% | 70% | 145% | ⚠️ At scale only |
| Break-even users | 10,000+ | 1,000 | 10x | ⚠️ Requires optimization |

### Recommendation: BUILD IT

**Why:**
1. Strong user value proposition (250x-370x ROI)
2. Clear path to cost optimization (6-month roadmap)
3. Competitive differentiation (integrated research)
4. Revenue potential at scale (87%+ margin possible)

**How:**
1. Start with MVP + aggressive cost monitoring
2. Launch Free tier only (10 tasks/month cap)
3. Iterate on optimization (target $0.06/task in Month 1)
4. Launch Pro tier when cost < $0.03/task
5. Scale when cost < $0.01/task

**When:**
- MVP: Week 1-2 (basic implementation)
- Optimization: Week 3-8 (cost reduction)
- Monetization: Month 3+ (Pro tier launch)
- Scale: Month 6+ (ultra-optimized costs)

---

**Last Updated:** 2026-01-19
**Next Review:** After MVP launch (collect real usage data)
