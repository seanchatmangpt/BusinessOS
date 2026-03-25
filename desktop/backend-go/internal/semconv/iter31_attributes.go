package semconv

import "go.opentelemetry.io/otel/attribute"

// Iter31 — MCP cache, A2A SLO, PM complexity, consensus threshold vote, healing rate limit, LLM distillation

// MCP Tool Cache
const MCPToolCacheHitKey = attribute.Key("mcp.tool.cache.hit")
const MCPToolCacheTTLMsKey = attribute.Key("mcp.tool.cache.ttl_ms")
const MCPToolCacheKeyKey = attribute.Key("mcp.tool.cache.key")

func MCPToolCacheHit(val bool) attribute.KeyValue     { return MCPToolCacheHitKey.Bool(val) }
func MCPToolCacheTTLMs(val int64) attribute.KeyValue  { return MCPToolCacheTTLMsKey.Int64(val) }
func MCPToolCacheKey(val string) attribute.KeyValue   { return MCPToolCacheKeyKey.String(val) }

// A2A SLO
const A2ASLOIDKey = attribute.Key("a2a.slo.id")
const A2ASLOTargetLatencyMsKey = attribute.Key("a2a.slo.target_latency_ms")
const A2ASLOComplianceRateKey = attribute.Key("a2a.slo.compliance_rate")
const A2ASLOBreachCountKey = attribute.Key("a2a.slo.breach_count")

func A2ASLOID(val string) attribute.KeyValue             { return A2ASLOIDKey.String(val) }
func A2ASLOTargetLatencyMs(val int64) attribute.KeyValue { return A2ASLOTargetLatencyMsKey.Int64(val) }
func A2ASLOComplianceRate(val float64) attribute.KeyValue { return A2ASLOComplianceRateKey.Float64(val) }
func A2ASLOBreachCount(val int64) attribute.KeyValue     { return A2ASLOBreachCountKey.Int64(val) }

// Process Mining Complexity
const ProcessMiningComplexityScoreKey = attribute.Key("process.mining.complexity.score")
const ProcessMiningComplexityMetricKey = attribute.Key("process.mining.complexity.metric")
const ProcessMiningComplexityVariantCountKey = attribute.Key("process.mining.complexity.variant_count")

func ProcessMiningComplexityScore(val float64) attribute.KeyValue  { return ProcessMiningComplexityScoreKey.Float64(val) }
func ProcessMiningComplexityMetric(val string) attribute.KeyValue  { return ProcessMiningComplexityMetricKey.String(val) }
func ProcessMiningComplexityVariantCount(val int64) attribute.KeyValue { return ProcessMiningComplexityVariantCountKey.Int64(val) }

// Consensus Threshold Vote
const ConsensusThresholdVoteTypeKey = attribute.Key("consensus.threshold.vote_type")
const ConsensusThresholdYeaCountKey = attribute.Key("consensus.threshold.yea_count")
const ConsensusThresholdNayCountKey = attribute.Key("consensus.threshold.nay_count")

func ConsensusThresholdVoteType(val string) attribute.KeyValue  { return ConsensusThresholdVoteTypeKey.String(val) }
func ConsensusThresholdYeaCount(val int64) attribute.KeyValue   { return ConsensusThresholdYeaCountKey.Int64(val) }
func ConsensusThresholdNayCount(val int64) attribute.KeyValue   { return ConsensusThresholdNayCountKey.Int64(val) }

// Healing Rate Limit
const HealingRateLimitRequestsPerSecKey = attribute.Key("healing.rate_limit.requests_per_sec")
const HealingRateLimitBurstSizeKey = attribute.Key("healing.rate_limit.burst_size")
const HealingRateLimitCurrentRateKey = attribute.Key("healing.rate_limit.current_rate")

func HealingRateLimitRequestsPerSec(val float64) attribute.KeyValue { return HealingRateLimitRequestsPerSecKey.Float64(val) }
func HealingRateLimitBurstSize(val int64) attribute.KeyValue        { return HealingRateLimitBurstSizeKey.Int64(val) }
func HealingRateLimitCurrentRate(val float64) attribute.KeyValue    { return HealingRateLimitCurrentRateKey.Float64(val) }

// LLM Distillation
const LLMDistillationTeacherModelKey = attribute.Key("llm.distillation.teacher_model")
const LLMDistillationStudentModelKey = attribute.Key("llm.distillation.student_model")
const LLMDistillationCompressionRatioKey = attribute.Key("llm.distillation.compression_ratio")
const LLMDistillationKLDivergenceKey = attribute.Key("llm.distillation.kl_divergence")

func LLMDistillationTeacherModel(val string) attribute.KeyValue      { return LLMDistillationTeacherModelKey.String(val) }
func LLMDistillationStudentModel(val string) attribute.KeyValue      { return LLMDistillationStudentModelKey.String(val) }
func LLMDistillationCompressionRatio(val float64) attribute.KeyValue { return LLMDistillationCompressionRatioKey.Float64(val) }
func LLMDistillationKLDivergence(val float64) attribute.KeyValue     { return LLMDistillationKLDivergenceKey.Float64(val) }
