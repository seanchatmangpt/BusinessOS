package prompt

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// varPattern matches {{variableName}} placeholders.
var varPattern = regexp.MustCompile(`\{\{(\w+)\}\}`)

// maxVersionHistory is the maximum number of historical versions retained per prompt.
const maxVersionHistory = 10

// PromptVersion holds a historical snapshot of a prompt.
type PromptVersion struct {
	Version   int
	Content   string
	UpdatedAt time.Time
}

// PromptTemplate is a fully parsed prompt loaded from disk.
type PromptTemplate struct {
	Name      string            // filename without extension
	Version   int
	Content   string
	Variables []string          // names extracted from {{varName}} patterns
	Sections  map[string]string // "## SectionName" → body text until next section
	UpdatedAt time.Time
}

// PromptService loads prompt files from disk and keeps them hot-reloaded via fsnotify.
type PromptService struct {
	promptsDir string
	mu         sync.RWMutex
	templates  map[string]*PromptTemplate
	versions   map[string][]PromptVersion // name → last N versions
	watcher    *fsnotify.Watcher
	logger     *slog.Logger
}

// NewPromptService creates the service, loads all .md and .txt files in promptsDir,
// and starts an fsnotify watcher for hot-reload. The caller must call Close() when done.
func NewPromptService(promptsDir string, logger *slog.Logger) (*PromptService, error) {
	if logger == nil {
		logger = slog.Default()
	}

	absDir, err := filepath.Abs(promptsDir)
	if err != nil {
		return nil, fmt.Errorf("prompt: resolve promptsDir: %w", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("prompt: create fsnotify watcher: %w", err)
	}

	svc := &PromptService{
		promptsDir: absDir,
		templates:  make(map[string]*PromptTemplate),
		versions:   make(map[string][]PromptVersion),
		watcher:    watcher,
		logger:     logger,
	}

	if err := svc.loadAll(); err != nil {
		_ = watcher.Close()
		return nil, fmt.Errorf("prompt: initial load: %w", err)
	}

	if err := watcher.Add(absDir); err != nil {
		_ = watcher.Close()
		return nil, fmt.Errorf("prompt: watch directory %q: %w", absDir, err)
	}

	go svc.watchLoop()

	logger.Info("prompt service started",
		"dir", absDir,
		"loaded", len(svc.templates),
	)

	return svc, nil
}

// Get returns the named prompt template (filename without extension).
// Returns an error wrapping a descriptive message when the name is not found.
func (s *PromptService) Get(ctx context.Context, name string) (*PromptTemplate, error) {
	s.mu.RLock()
	t, ok := s.templates[name]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("prompt: template %q not found", name)
	}
	// Return a shallow copy so callers cannot mutate internal state.
	cp := *t
	return &cp, nil
}

// Render substitutes {{variable}} placeholders in the named template.
// Returns an error if any variable referenced in the template is absent from vars.
func (s *PromptService) Render(ctx context.Context, name string, vars map[string]string) (string, error) {
	t, err := s.Get(ctx, name)
	if err != nil {
		return "", err
	}

	var missing []string
	result := varPattern.ReplaceAllStringFunc(t.Content, func(match string) string {
		sub := varPattern.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		key := sub[1]
		val, ok := vars[key]
		if !ok {
			missing = append(missing, key)
			return match
		}
		return val
	})

	if len(missing) > 0 {
		return "", fmt.Errorf("prompt: render %q: missing variables: %s", name, strings.Join(missing, ", "))
	}

	return result, nil
}

// GetSection returns the body text of a named section (## SectionName) within a template.
// Returns an error when either the template or the section is not found.
func (s *PromptService) GetSection(ctx context.Context, name, section string) (string, error) {
	t, err := s.Get(ctx, name)
	if err != nil {
		return "", err
	}

	body, ok := t.Sections[section]
	if !ok {
		return "", fmt.Errorf("prompt: template %q has no section %q", name, section)
	}

	return body, nil
}

// List returns the names of all currently loaded prompt templates.
func (s *PromptService) List(ctx context.Context) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.templates))
	for k := range s.templates {
		names = append(names, k)
	}
	return names
}

// GetVersionHistory returns up to maxVersionHistory past versions of the named prompt.
// The slice is ordered oldest-first. Returns an empty slice (not an error) when no
// history exists yet.
func (s *PromptService) GetVersionHistory(ctx context.Context, name string) ([]PromptVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.templates[name]; !ok {
		return nil, fmt.Errorf("prompt: template %q not found", name)
	}

	hist := s.versions[name]
	out := make([]PromptVersion, len(hist))
	copy(out, hist)
	return out, nil
}

// Reload force-reloads a single prompt from disk by name (filename without extension).
// It searches for a .md file first, then a .txt file.
func (s *PromptService) Reload(ctx context.Context, name string) error {
	for _, ext := range []string{".md", ".txt"} {
		path := filepath.Join(s.promptsDir, name+ext)
		if _, err := os.Stat(path); err == nil {
			return s.reloadFile(path, name)
		}
	}
	return fmt.Errorf("prompt: no file found for %q in %s", name, s.promptsDir)
}

// Close stops the fsnotify watcher. After Close, no further hot-reloads will occur.
func (s *PromptService) Close() error {
	if err := s.watcher.Close(); err != nil {
		return fmt.Errorf("prompt: close watcher: %w", err)
	}
	return nil
}

// ---- internal helpers ----

// loadAll reads every .md and .txt file in promptsDir into memory.
func (s *PromptService) loadAll() error {
	entries, err := os.ReadDir(s.promptsDir)
	if err != nil {
		return fmt.Errorf("read directory %q: %w", s.promptsDir, err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext != ".md" && ext != ".txt" {
			continue
		}
		name := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		path := filepath.Join(s.promptsDir, e.Name())
		if err := s.reloadFile(path, name); err != nil {
			return fmt.Errorf("load %s: %w", e.Name(), err)
		}
	}

	return nil
}

// reloadFile reads a single file from disk, parses it, and updates the in-memory map.
// If a previous version exists it is pushed into the history ring before replacement.
func (s *PromptService) reloadFile(path, name string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %q: %w", path, err)
	}

	content := string(data)
	now := time.Now()

	t := &PromptTemplate{
		Name:      name,
		Content:   content,
		Variables: extractVariables(content),
		Sections:  extractSections(content),
		UpdatedAt: now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Archive the old version before overwriting.
	if old, exists := s.templates[name]; exists {
		t.Version = old.Version + 1
		s.pushVersion(name, old)
	} else {
		t.Version = 1
	}

	s.templates[name] = t
	return nil
}

// pushVersion appends old to the history for name, keeping at most maxVersionHistory entries.
func (s *PromptService) pushVersion(name string, old *PromptTemplate) {
	hist := s.versions[name]
	hist = append(hist, PromptVersion{
		Version:   old.Version,
		Content:   old.Content,
		UpdatedAt: old.UpdatedAt,
	})
	if len(hist) > maxVersionHistory {
		hist = hist[len(hist)-maxVersionHistory:]
	}
	s.versions[name] = hist
}

// watchLoop runs in a goroutine and processes fsnotify events.
func (s *PromptService) watchLoop() {
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				ext := strings.ToLower(filepath.Ext(event.Name))
				if ext != ".md" && ext != ".txt" {
					continue
				}
				name := nameFromPath(event.Name)
				if err := s.reloadFile(event.Name, name); err != nil {
					s.logger.Error("hot-reload failed", "file", event.Name, "error", err)
				} else {
					s.logger.Info("prompt hot-reloaded", "name", name)
				}
			}
		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			s.logger.Error("fsnotify watcher error", "error", err)
		}
	}
}

// nameFromPath strips the directory and extension, returning just the base name.
func nameFromPath(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// extractVariables returns a deduplicated, ordered list of variable names found in
// content using the {{varName}} pattern.
func extractVariables(content string) []string {
	matches := varPattern.FindAllStringSubmatch(content, -1)
	seen := make(map[string]struct{})
	var out []string
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		key := m[1]
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}

// extractSections parses content on "\n## " boundaries and returns a map of
// section-name → trimmed body text. The preamble before the first "## " header
// is ignored (or can be captured under the empty string key if needed; here we skip it).
func extractSections(content string) map[string]string {
	sections := make(map[string]string)

	// Normalise line endings.
	content = strings.ReplaceAll(content, "\r\n", "\n")

	// Split on lines that begin with "## ".
	lines := strings.Split(content, "\n")

	var currentSection string
	var buf strings.Builder
	inSection := false

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			// Save the previous section body if we were inside one.
			if inSection {
				sections[currentSection] = strings.TrimSpace(buf.String())
				buf.Reset()
			}
			currentSection = strings.TrimPrefix(line, "## ")
			currentSection = strings.TrimSpace(currentSection)
			inSection = true
			continue
		}
		if inSection {
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}

	// Flush the last section.
	if inSection {
		sections[currentSection] = strings.TrimSpace(buf.String())
	}

	return sections
}
