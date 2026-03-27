package yawlv6_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rhl/businessos-backend/internal/integrations/yawlv6"
)

func TestHealth_WhenEngineUnreachable(t *testing.T) {
	// Point to a port nobody is listening on so Health() must return an error.
	t.Setenv("YAWLV6_URL", "http://localhost:19999")
	c := yawlv6.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := c.Health(ctx)
	assert.Error(t, err, "Health should return error when engine is unreachable")
}

func TestBuildSequenceSpec_ContainsAllTasks(t *testing.T) {
	xml := yawlv6.BuildSequenceSpec([]string{"step_a", "step_b", "step_c"})

	assert.Contains(t, xml, `id="step_a"`)
	assert.Contains(t, xml, `id="step_b"`)
	assert.Contains(t, xml, `id="step_c"`)
	assert.Contains(t, xml, "specificationSet")
	assert.Contains(t, xml, `join code="xor"`)
}

func TestBuildSequenceSpec_EmptyTasks_ReturnsEmpty(t *testing.T) {
	xml := yawlv6.BuildSequenceSpec([]string{})
	assert.Empty(t, xml)
}

func TestBuildSequenceSpec_ChainOrder(t *testing.T) {
	xml := yawlv6.BuildSequenceSpec([]string{"a", "b"})

	// "a" must flow into "b", "b" must flow into OutputCondition.
	aIdx := strings.Index(xml, `id="a"`)
	bIdx := strings.Index(xml, `id="b"`)
	require.True(t, aIdx >= 0 && bIdx >= 0)

	assert.Contains(t, xml, `nextElementRef id="b"`)
	assert.Contains(t, xml, `nextElementRef id="OutputCondition"`)
}

func TestBuildParallelSplitSpec_HasAndSplit(t *testing.T) {
	xml := yawlv6.BuildParallelSplitSpec("start", []string{"branch_a", "branch_b"})

	assert.Contains(t, xml, `split code="and"`)
	assert.Contains(t, xml, `id="branch_a"`)
	assert.Contains(t, xml, `id="branch_b"`)
	require.True(t, strings.Contains(xml, "specificationSet"))
	assert.Contains(t, xml, "OSA_ParallelSplit")
}

func TestBuildParallelSplitSpec_TriggerFlowsToBothBranches(t *testing.T) {
	xml := yawlv6.BuildParallelSplitSpec("trigger", []string{"left", "right"})

	assert.Contains(t, xml, `nextElementRef id="left"`)
	assert.Contains(t, xml, `nextElementRef id="right"`)
}

func TestConformanceResult_JSONRoundTrip(t *testing.T) {
	raw := []byte(`{"fitness":0.95,"violations":["v1"],"is_sound":true}`)

	var result yawlv6.ConformanceResult
	err := json.Unmarshal(raw, &result)
	require.NoError(t, err)

	assert.Equal(t, 0.95, result.Fitness)
	assert.Equal(t, []string{"v1"}, result.Violations)
	assert.True(t, result.IsSound)
}
