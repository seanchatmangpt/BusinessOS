package prompt_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rhl/businessos-backend/internal/prompt"
)

// newTestLogger returns a slog.Logger that discards output during tests.
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// writeFile writes content to a file inside dir with the given filename.
func writeFile(t *testing.T, dir, filename, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644))
}

func TestNewPromptService(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "greeting.md", "Hello, {{name}}!")
	writeFile(t, dir, "farewell.txt", "Goodbye, {{name}}!")
	writeFile(t, dir, "ignored.json", `{"not": "a prompt"}`)

	svc, err := prompt.NewPromptService(dir, newTestLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	names := svc.List(context.Background())
	sort.Strings(names)

	assert.Equal(t, []string{"farewell", "greeting"}, names,
		"only .md and .txt files should be loaded; .json must be ignored")
}

func TestRender(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "hello.md", "Hi {{first}} {{last}}, welcome to {{place}}!")

	svc, err := prompt.NewPromptService(dir, newTestLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	got, err := svc.Render(context.Background(), "hello", map[string]string{
		"first": "John",
		"last":  "Doe",
		"place": "BusinessOS",
	})
	require.NoError(t, err)
	assert.Equal(t, "Hi John Doe, welcome to BusinessOS!", got)
}

func TestRender_MissingVar(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "tmpl.md", "Hello {{name}}, your score is {{score}}.")

	svc, err := prompt.NewPromptService(dir, newTestLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	_, err = svc.Render(context.Background(), "tmpl", map[string]string{
		"name": "Alice",
		// "score" is intentionally omitted
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "score", "error message should name the missing variable")
}

func TestGetSection(t *testing.T) {
	content := `# Document Title

## Introduction
This is the intro paragraph.
It spans multiple lines.

## Instructions
Step 1: do this.
Step 2: do that.

## Footer
End of document.
`
	dir := t.TempDir()
	writeFile(t, dir, "doc.md", content)

	svc, err := prompt.NewPromptService(dir, newTestLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	intro, err := svc.GetSection(context.Background(), "doc", "Introduction")
	require.NoError(t, err)
	assert.Contains(t, intro, "intro paragraph")
	assert.Contains(t, intro, "multiple lines")

	instructions, err := svc.GetSection(context.Background(), "doc", "Instructions")
	require.NoError(t, err)
	assert.Contains(t, instructions, "Step 1")
	assert.Contains(t, instructions, "Step 2")

	footer, err := svc.GetSection(context.Background(), "doc", "Footer")
	require.NoError(t, err)
	assert.Equal(t, "End of document.", footer)
}

func TestGetSection_NotFound(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "simple.md", "## Only Section\nSome content.")

	svc, err := prompt.NewPromptService(dir, newTestLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	_, err = svc.GetSection(context.Background(), "simple", "Nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Nonexistent")
}

func TestGetVersionHistory_InitiallyEmpty(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "v.md", "version one content")

	svc, err := prompt.NewPromptService(dir, newTestLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	hist, err := svc.GetVersionHistory(context.Background(), "v")
	require.NoError(t, err)
	assert.Empty(t, hist, "no history should exist immediately after initial load")
}

func TestGetVersionHistory_AfterReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "v.md")
	require.NoError(t, os.WriteFile(path, []byte("content v1"), 0o644))

	svc, err := prompt.NewPromptService(dir, newTestLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	// Force a second load to create a version history entry.
	require.NoError(t, os.WriteFile(path, []byte("content v2"), 0o644))
	require.NoError(t, svc.Reload(context.Background(), "v"))

	hist, err := svc.GetVersionHistory(context.Background(), "v")
	require.NoError(t, err)
	// At least 1 history entry must exist (fsnotify may also fire, creating extras).
	require.NotEmpty(t, hist, "expected at least one version history entry")
	// The first entry must be v1 content (oldest).
	assert.Equal(t, "content v1", hist[0].Content)
	assert.Equal(t, 1, hist[0].Version)
}

func TestGetVersionHistory_CappedAt10(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "v.md")
	require.NoError(t, os.WriteFile(path, []byte("init"), 0o644))

	svc, err := prompt.NewPromptService(dir, newTestLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	// Reload 15 times — history must be capped at 10.
	for i := 0; i < 15; i++ {
		require.NoError(t, os.WriteFile(path, []byte("content"), 0o644))
		require.NoError(t, svc.Reload(context.Background(), "v"))
	}

	hist, err := svc.GetVersionHistory(context.Background(), "v")
	require.NoError(t, err)
	assert.LessOrEqual(t, len(hist), 10, "version history must not exceed 10 entries")
}

func TestList(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "alpha.md", "a")
	writeFile(t, dir, "beta.txt", "b")
	writeFile(t, dir, "gamma.md", "c")

	svc, err := prompt.NewPromptService(dir, newTestLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	names := svc.List(context.Background())
	sort.Strings(names)

	assert.True(t, slices.Contains(names, "alpha"))
	assert.True(t, slices.Contains(names, "beta"))
	assert.True(t, slices.Contains(names, "gamma"))
	assert.Len(t, names, 3)
}

func TestGet_NotFound(t *testing.T) {
	dir := t.TempDir()

	svc, err := prompt.NewPromptService(dir, newTestLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	_, err = svc.Get(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestVariableExtraction(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "vars.md", "{{a}} and {{b}} and {{a}} again, {{c}}")

	svc, err := prompt.NewPromptService(dir, newTestLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	tmpl, err := svc.Get(context.Background(), "vars")
	require.NoError(t, err)

	// Deduplicated: "a" appears twice in content but should appear once in Variables.
	vars := tmpl.Variables
	sort.Strings(vars)
	assert.Equal(t, []string{"a", "b", "c"}, vars)
}

func TestHotReload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping hot-reload test in short mode")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "live.md")
	require.NoError(t, os.WriteFile(path, []byte("original"), 0o644))

	svc, err := prompt.NewPromptService(dir, newTestLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	// Write a new version; the watcher should pick it up within 2 seconds.
	require.NoError(t, os.WriteFile(path, []byte("updated"), 0o644))

	var updated bool
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		tmpl, err := svc.Get(context.Background(), "live")
		if err == nil && tmpl.Content == "updated" {
			updated = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	assert.True(t, updated, "hot-reload should update the template within 2 seconds")
}

func TestReload_FileNotFound(t *testing.T) {
	dir := t.TempDir()

	svc, err := prompt.NewPromptService(dir, newTestLogger())
	require.NoError(t, err)
	t.Cleanup(func() { _ = svc.Close() })

	err = svc.Reload(context.Background(), "ghost")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ghost")
}
