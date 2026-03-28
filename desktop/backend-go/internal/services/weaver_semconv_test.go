package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeaverSemconvEnabled_DefaultFalse(t *testing.T) {
	t.Setenv("WEAVER_SEMCONV_ENABLED", "")
	assert.False(t, WeaverSemconvEnabled())
}

func TestWeaverSemconvEnabled_Truthy(t *testing.T) {
	t.Setenv("WEAVER_SEMCONV_ENABLED", "true")
	assert.True(t, WeaverSemconvEnabled())
}

func TestWeaverSemconvRegistryPath_Default(t *testing.T) {
	t.Setenv("WEAVER_SEMCONV_REGISTRY", "")
	assert.Equal(t, "/semconv/model", WeaverSemconvRegistryPath())
}

func TestWeaverSemconvConfigured_RequiresDirectory(t *testing.T) {
	t.Setenv("WEAVER_SEMCONV_ENABLED", "true")
	dir := t.TempDir()
	t.Setenv("WEAVER_SEMCONV_REGISTRY", dir)
	assert.True(t, weaverSemconvConfigured())

	t.Setenv("WEAVER_SEMCONV_REGISTRY", "/nonexistent-weaver-semconv-path-xyz")
	assert.False(t, weaverSemconvConfigured())
}

func TestMCPService_GetAllTools_NoWeaverWhenDisabled(t *testing.T) {
	t.Setenv("WEAVER_SEMCONV_ENABLED", "")
	m := NewMCPService(nil, "", nil, nil, nil)
	tools := m.GetAllTools()
	require.NotEmpty(t, tools)
	for _, x := range tools {
		assert.False(t, len(x.Name) > len(weaverSemconvPrefix) && x.Name[:len(weaverSemconvPrefix)] == weaverSemconvPrefix,
			"unexpected semconv tool %q when weaver disabled", x.Name)
	}
}

func TestExecuteWeaverSemconvTool_InvalidPrefix(t *testing.T) {
	_, err := ExecuteWeaverSemconvTool(context.Background(), "wrong.prefix", nil)
	require.Error(t, err)
}
