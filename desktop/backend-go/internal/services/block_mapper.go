package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// BlockMapperService converts markdown content to structured blocks
type BlockMapperService struct {
	db     *sql.DB
	logger *slog.Logger
}

// Block represents a structured content block
type Block struct {
	ID         string                 `json:"id"`
	Type       BlockType              `json:"type"`
	Content    string                 `json:"content"`
	RawContent string                 `json:"raw_content,omitempty"`
	Language   string                 `json:"language,omitempty"`
	Level      int                    `json:"level,omitempty"`
	Children   []*Block               `json:"children,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Hash       string                 `json:"hash"`
	StartLine  int                    `json:"start_line,omitempty"`
	EndLine    int                    `json:"end_line,omitempty"`
}

// BlockType represents the type of content block
type BlockType string

const (
	BlockTypeParagraph   BlockType = "paragraph"
	BlockTypeHeading     BlockType = "heading"
	BlockTypeCode        BlockType = "code"
	BlockTypeCodeInline  BlockType = "code_inline"
	BlockTypeBlockquote  BlockType = "blockquote"
	BlockTypeList        BlockType = "list"
	BlockTypeListItem    BlockType = "list_item"
	BlockTypeTable       BlockType = "table"
	BlockTypeTableRow    BlockType = "table_row"
	BlockTypeTableCell   BlockType = "table_cell"
	BlockTypeImage       BlockType = "image"
	BlockTypeLink        BlockType = "link"
	BlockTypeHR          BlockType = "horizontal_rule"
	BlockTypeHTML        BlockType = "html"
	BlockTypeThinking    BlockType = "thinking"
	BlockTypeArtifact    BlockType = "artifact"
	BlockTypeCallout     BlockType = "callout"
	BlockTypeMath        BlockType = "math"
	BlockTypeFrontmatter BlockType = "frontmatter"
	BlockTypeTask        BlockType = "task"
	BlockTypeFootnote    BlockType = "footnote"
)

// BlockDocument represents a parsed document with blocks
type BlockDocument struct {
	ID          string                 `json:"id"`
	SourceID    string                 `json:"source_id,omitempty"`
	Title       string                 `json:"title,omitempty"`
	Blocks      []*Block               `json:"blocks"`
	Outline     []*OutlineEntry        `json:"outline,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Hash        string                 `json:"hash"`
	TotalBlocks int                    `json:"total_blocks"`
	CreatedAt   time.Time              `json:"created_at"`
}

// OutlineEntry represents a heading in the document outline
type OutlineEntry struct {
	ID       string          `json:"id"`
	Level    int             `json:"level"`
	Title    string          `json:"title"`
	BlockID  string          `json:"block_id"`
	Children []*OutlineEntry `json:"children,omitempty"`
}

// BlockMapperOptions configures parsing behavior
type BlockMapperOptions struct {
	ExtractOutline     bool `json:"extract_outline"`
	PreserveRawContent bool `json:"preserve_raw_content"`
	ParseCodeBlocks    bool `json:"parse_code_blocks"`
	ParseTables        bool `json:"parse_tables"`
	ParseThinking      bool `json:"parse_thinking"`
	ParseArtifacts     bool `json:"parse_artifacts"`
	IncludeLineNumbers bool `json:"include_line_numbers"`
}

// DefaultBlockMapperOptions returns default parsing options
func DefaultBlockMapperOptions() *BlockMapperOptions {
	return &BlockMapperOptions{
		ExtractOutline:     true,
		PreserveRawContent: false,
		ParseCodeBlocks:    true,
		ParseTables:        true,
		ParseThinking:      true,
		ParseArtifacts:     true,
		IncludeLineNumbers: true,
	}
}

// NewBlockMapperService creates a new block mapper service
func NewBlockMapperService(db *sql.DB, logger *slog.Logger) *BlockMapperService {
	return &BlockMapperService{
		db:     db,
		logger: logger,
	}
}

// ParseMarkdown converts markdown content to a BlockDocument
func (s *BlockMapperService) ParseMarkdown(ctx context.Context, content string, opts *BlockMapperOptions) (*BlockDocument, error) {
	if opts == nil {
		opts = DefaultBlockMapperOptions()
	}

	doc := &BlockDocument{
		ID:        uuid.New().String(),
		Blocks:    make([]*Block, 0),
		Outline:   make([]*OutlineEntry, 0),
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	// Calculate document hash
	hash := sha256.Sum256([]byte(content))
	doc.Hash = hex.EncodeToString(hash[:])

	lines := strings.Split(content, "\n")
	lineNum := 0

	for lineNum < len(lines) {
		block, consumed := s.parseBlock(lines, lineNum, opts)
		if block != nil {
			if opts.IncludeLineNumbers {
				block.StartLine = lineNum + 1
				block.EndLine = lineNum + consumed
			}
			doc.Blocks = append(doc.Blocks, block)
		}
		if consumed == 0 {
			consumed = 1
		}
		lineNum += consumed
	}

	// Extract outline from headings
	if opts.ExtractOutline {
		doc.Outline = s.extractOutline(doc.Blocks)
	}

	// Extract title from first heading
	for _, b := range doc.Blocks {
		if b.Type == BlockTypeHeading && b.Level == 1 {
			doc.Title = b.Content
			break
		}
	}

	doc.TotalBlocks = len(doc.Blocks)

	return doc, nil
}

// parseBlock parses a single block starting at lineNum
func (s *BlockMapperService) parseBlock(lines []string, lineNum int, opts *BlockMapperOptions) (*Block, int) {
	if lineNum >= len(lines) {
		return nil, 0
	}

	line := lines[lineNum]
	trimmedLine := strings.TrimSpace(line)

	// Skip empty lines
	if trimmedLine == "" {
		return nil, 1
	}

	// Frontmatter (YAML)
	if lineNum == 0 && trimmedLine == "---" {
		return s.parseFrontmatter(lines, lineNum)
	}

	// Thinking tags
	if opts.ParseThinking && strings.HasPrefix(trimmedLine, "<thinking") {
		return s.parseThinking(lines, lineNum)
	}

	// Artifact tags
	if opts.ParseArtifacts && strings.HasPrefix(trimmedLine, "<artifact") {
		return s.parseArtifact(lines, lineNum)
	}

	// Code blocks
	if opts.ParseCodeBlocks && strings.HasPrefix(trimmedLine, "```") {
		return s.parseCodeBlock(lines, lineNum)
	}

	// Headings
	if strings.HasPrefix(trimmedLine, "#") {
		return s.parseHeading(line), 1
	}

	// Horizontal rule
	if s.isHorizontalRule(trimmedLine) {
		return &Block{
			ID:   uuid.New().String(),
			Type: BlockTypeHR,
			Hash: s.hashContent("---"),
		}, 1
	}

	// Blockquote
	if strings.HasPrefix(trimmedLine, ">") {
		return s.parseBlockquote(lines, lineNum)
	}

	// Lists
	if s.isListItem(trimmedLine) {
		return s.parseList(lines, lineNum)
	}

	// Tables
	if opts.ParseTables && strings.HasPrefix(trimmedLine, "|") {
		return s.parseTable(lines, lineNum)
	}

	// Math blocks
	if strings.HasPrefix(trimmedLine, "$$") {
		return s.parseMathBlock(lines, lineNum)
	}

	// Callouts (Obsidian-style)
	if strings.HasPrefix(trimmedLine, "> [!") {
		return s.parseCallout(lines, lineNum)
	}

	// HTML blocks
	if strings.HasPrefix(trimmedLine, "<") && !strings.HasPrefix(trimmedLine, "<thinking") && !strings.HasPrefix(trimmedLine, "<artifact") {
		return s.parseHTMLBlock(lines, lineNum)
	}

	// Default: paragraph
	return s.parseParagraph(lines, lineNum)
}

// parseFrontmatter parses YAML frontmatter
func (s *BlockMapperService) parseFrontmatter(lines []string, lineNum int) (*Block, int) {
	var content strings.Builder
	consumed := 1 // First ---

	for i := lineNum + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			consumed++
			break
		}
		content.WriteString(lines[i])
		content.WriteString("\n")
		consumed++
	}

	return &Block{
		ID:      uuid.New().String(),
		Type:    BlockTypeFrontmatter,
		Content: strings.TrimSpace(content.String()),
		Hash:    s.hashContent(content.String()),
	}, consumed
}

// parseThinking parses thinking tags
func (s *BlockMapperService) parseThinking(lines []string, lineNum int) (*Block, int) {
	var content strings.Builder
	consumed := 0
	inThinking := false

	for i := lineNum; i < len(lines); i++ {
		line := lines[i]
		consumed++

		if strings.Contains(line, "<thinking") {
			inThinking = true
			// Extract content after opening tag if on same line
			if idx := strings.Index(line, ">"); idx != -1 {
				content.WriteString(line[idx+1:])
				content.WriteString("\n")
			}
			continue
		}

		if strings.Contains(line, "</thinking>") {
			// Extract content before closing tag
			if idx := strings.Index(line, "</thinking>"); idx > 0 {
				content.WriteString(line[:idx])
			}
			break
		}

		if inThinking {
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	return &Block{
		ID:      uuid.New().String(),
		Type:    BlockTypeThinking,
		Content: strings.TrimSpace(content.String()),
		Hash:    s.hashContent(content.String()),
	}, consumed
}

// parseArtifact parses artifact tags
func (s *BlockMapperService) parseArtifact(lines []string, lineNum int) (*Block, int) {
	var content strings.Builder
	consumed := 0
	metadata := make(map[string]interface{})

	openingLine := lines[lineNum]

	// Extract attributes from opening tag
	if attrMatch := regexp.MustCompile(`identifier="([^"]+)"`).FindStringSubmatch(openingLine); len(attrMatch) > 1 {
		metadata["identifier"] = attrMatch[1]
	}
	if attrMatch := regexp.MustCompile(`type="([^"]+)"`).FindStringSubmatch(openingLine); len(attrMatch) > 1 {
		metadata["type"] = attrMatch[1]
	}
	if attrMatch := regexp.MustCompile(`title="([^"]+)"`).FindStringSubmatch(openingLine); len(attrMatch) > 1 {
		metadata["title"] = attrMatch[1]
	}
	if attrMatch := regexp.MustCompile(`language="([^"]+)"`).FindStringSubmatch(openingLine); len(attrMatch) > 1 {
		metadata["language"] = attrMatch[1]
	}

	for i := lineNum; i < len(lines); i++ {
		line := lines[i]
		consumed++

		if i == lineNum {
			// Check if content starts on same line
			if idx := strings.Index(line, ">"); idx != -1 && idx < len(line)-1 {
				content.WriteString(line[idx+1:])
				content.WriteString("\n")
			}
			continue
		}

		if strings.Contains(line, "</artifact>") {
			if idx := strings.Index(line, "</artifact>"); idx > 0 {
				content.WriteString(line[:idx])
			}
			break
		}

		content.WriteString(line)
		content.WriteString("\n")
	}

	return &Block{
		ID:       uuid.New().String(),
		Type:     BlockTypeArtifact,
		Content:  strings.TrimSpace(content.String()),
		Metadata: metadata,
		Hash:     s.hashContent(content.String()),
	}, consumed
}

// parseCodeBlock parses fenced code blocks
func (s *BlockMapperService) parseCodeBlock(lines []string, lineNum int) (*Block, int) {
	firstLine := strings.TrimSpace(lines[lineNum])
	language := strings.TrimPrefix(firstLine, "```")
	language = strings.TrimSpace(language)

	var content strings.Builder
	consumed := 1

	for i := lineNum + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "```" {
			consumed++
			break
		}
		content.WriteString(lines[i])
		content.WriteString("\n")
		consumed++
	}

	return &Block{
		ID:       uuid.New().String(),
		Type:     BlockTypeCode,
		Content:  strings.TrimSuffix(content.String(), "\n"),
		Language: language,
		Hash:     s.hashContent(content.String()),
	}, consumed
}

// parseHeading parses heading lines
func (s *BlockMapperService) parseHeading(line string) *Block {
	level := 0
	for _, c := range line {
		if c == '#' {
			level++
		} else {
			break
		}
	}

	content := strings.TrimSpace(strings.TrimLeft(line, "# "))

	return &Block{
		ID:      uuid.New().String(),
		Type:    BlockTypeHeading,
		Content: content,
		Level:   level,
		Hash:    s.hashContent(content),
	}
}

// parseBlockquote parses blockquote blocks
func (s *BlockMapperService) parseBlockquote(lines []string, lineNum int) (*Block, int) {
	var content strings.Builder
	consumed := 0

	for i := lineNum; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(trimmed, ">") && trimmed != "" {
			break
		}
		if trimmed == "" {
			consumed++
			break
		}

		// Remove > prefix
		lineContent := strings.TrimPrefix(trimmed, ">")
		lineContent = strings.TrimPrefix(lineContent, " ")
		content.WriteString(lineContent)
		content.WriteString("\n")
		consumed++
	}

	return &Block{
		ID:      uuid.New().String(),
		Type:    BlockTypeBlockquote,
		Content: strings.TrimSpace(content.String()),
		Hash:    s.hashContent(content.String()),
	}, consumed
}

// parseList parses list blocks
func (s *BlockMapperService) parseList(lines []string, lineNum int) (*Block, int) {
	block := &Block{
		ID:       uuid.New().String(),
		Type:     BlockTypeList,
		Children: make([]*Block, 0),
	}

	consumed := 0
	isOrdered := s.isOrderedListItem(strings.TrimSpace(lines[lineNum]))
	block.Metadata = map[string]interface{}{"ordered": isOrdered}

	for i := lineNum; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			consumed++
			break
		}
		if !s.isListItem(trimmed) {
			break
		}

		// Parse list item
		itemContent := s.extractListItemContent(trimmed)
		isTask, taskComplete := s.isTaskItem(itemContent)

		item := &Block{
			ID:      uuid.New().String(),
			Type:    BlockTypeListItem,
			Content: itemContent,
			Hash:    s.hashContent(itemContent),
		}

		if isTask {
			item.Type = BlockTypeTask
			item.Metadata = map[string]interface{}{"completed": taskComplete}
		}

		block.Children = append(block.Children, item)
		consumed++
	}

	// Build content summary
	var contentBuilder strings.Builder
	for _, child := range block.Children {
		contentBuilder.WriteString(child.Content)
		contentBuilder.WriteString("\n")
	}
	block.Content = strings.TrimSpace(contentBuilder.String())
	block.Hash = s.hashContent(block.Content)

	return block, consumed
}

// parseTable parses markdown tables
func (s *BlockMapperService) parseTable(lines []string, lineNum int) (*Block, int) {
	block := &Block{
		ID:       uuid.New().String(),
		Type:     BlockTypeTable,
		Children: make([]*Block, 0),
	}

	consumed := 0
	var contentBuilder strings.Builder

	for i := lineNum; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || !strings.HasPrefix(trimmed, "|") {
			break
		}

		// Skip separator row
		if s.isTableSeparator(trimmed) {
			consumed++
			continue
		}

		// Parse table row
		cells := s.parseTableRow(trimmed)
		row := &Block{
			ID:       uuid.New().String(),
			Type:     BlockTypeTableRow,
			Children: make([]*Block, 0),
		}

		for _, cell := range cells {
			row.Children = append(row.Children, &Block{
				ID:      uuid.New().String(),
				Type:    BlockTypeTableCell,
				Content: cell,
				Hash:    s.hashContent(cell),
			})
		}

		block.Children = append(block.Children, row)
		contentBuilder.WriteString(trimmed)
		contentBuilder.WriteString("\n")
		consumed++
	}

	block.Content = strings.TrimSpace(contentBuilder.String())
	block.Hash = s.hashContent(block.Content)

	return block, consumed
}

// parseMathBlock parses $$ math blocks
func (s *BlockMapperService) parseMathBlock(lines []string, lineNum int) (*Block, int) {
	var content strings.Builder
	consumed := 1 // First $$

	for i := lineNum + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "$$" {
			consumed++
			break
		}
		content.WriteString(lines[i])
		content.WriteString("\n")
		consumed++
	}

	return &Block{
		ID:      uuid.New().String(),
		Type:    BlockTypeMath,
		Content: strings.TrimSpace(content.String()),
		Hash:    s.hashContent(content.String()),
	}, consumed
}

// parseCallout parses Obsidian-style callouts
func (s *BlockMapperService) parseCallout(lines []string, lineNum int) (*Block, int) {
	firstLine := lines[lineNum]

	// Extract callout type
	typeMatch := regexp.MustCompile(`>\s*\[!(\w+)\]`).FindStringSubmatch(firstLine)
	calloutType := "note"
	if len(typeMatch) > 1 {
		calloutType = typeMatch[1]
	}

	var content strings.Builder
	consumed := 0

	for i := lineNum; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(trimmed, ">") && trimmed != "" {
			break
		}
		if trimmed == "" {
			consumed++
			break
		}

		lineContent := strings.TrimPrefix(trimmed, ">")
		lineContent = strings.TrimPrefix(lineContent, " ")
		content.WriteString(lineContent)
		content.WriteString("\n")
		consumed++
	}

	return &Block{
		ID:      uuid.New().String(),
		Type:    BlockTypeCallout,
		Content: strings.TrimSpace(content.String()),
		Metadata: map[string]interface{}{
			"callout_type": calloutType,
		},
		Hash: s.hashContent(content.String()),
	}, consumed
}

// parseHTMLBlock parses raw HTML blocks
func (s *BlockMapperService) parseHTMLBlock(lines []string, lineNum int) (*Block, int) {
	var content strings.Builder
	consumed := 0

	// Find the tag name
	tagMatch := regexp.MustCompile(`<(\w+)`).FindStringSubmatch(lines[lineNum])
	if len(tagMatch) < 2 {
		return s.parseParagraph(lines, lineNum)
	}
	tagName := tagMatch[1]
	closingTag := fmt.Sprintf("</%s>", tagName)

	for i := lineNum; i < len(lines); i++ {
		content.WriteString(lines[i])
		content.WriteString("\n")
		consumed++

		if strings.Contains(lines[i], closingTag) {
			break
		}
	}

	return &Block{
		ID:      uuid.New().String(),
		Type:    BlockTypeHTML,
		Content: strings.TrimSpace(content.String()),
		Hash:    s.hashContent(content.String()),
	}, consumed
}

// parseParagraph parses regular paragraph text
func (s *BlockMapperService) parseParagraph(lines []string, lineNum int) (*Block, int) {
	var content strings.Builder
	consumed := 0

	for i := lineNum; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			consumed++
			break
		}

		// Check if next block type starts
		if strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "```") ||
			strings.HasPrefix(trimmed, ">") ||
			strings.HasPrefix(trimmed, "|") ||
			s.isListItem(trimmed) ||
			s.isHorizontalRule(trimmed) {
			break
		}

		content.WriteString(lines[i])
		content.WriteString("\n")
		consumed++
	}

	return &Block{
		ID:      uuid.New().String(),
		Type:    BlockTypeParagraph,
		Content: strings.TrimSpace(content.String()),
		Hash:    s.hashContent(content.String()),
	}, consumed
}

// extractOutline builds an outline from heading blocks
func (s *BlockMapperService) extractOutline(blocks []*Block) []*OutlineEntry {
	outline := make([]*OutlineEntry, 0)
	stack := make([]*OutlineEntry, 0)

	for _, b := range blocks {
		if b.Type != BlockTypeHeading {
			continue
		}

		entry := &OutlineEntry{
			ID:       uuid.New().String(),
			Level:    b.Level,
			Title:    b.Content,
			BlockID:  b.ID,
			Children: make([]*OutlineEntry, 0),
		}

		// Find parent
		for len(stack) > 0 && stack[len(stack)-1].Level >= b.Level {
			stack = stack[:len(stack)-1]
		}

		if len(stack) == 0 {
			outline = append(outline, entry)
		} else {
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, entry)
		}

		stack = append(stack, entry)
	}

	return outline
}

// Helper functions

func (s *BlockMapperService) isHorizontalRule(line string) bool {
	line = strings.TrimSpace(line)
	if len(line) < 3 {
		return false
	}
	return regexp.MustCompile(`^[-*_]{3,}$`).MatchString(line)
}

func (s *BlockMapperService) isListItem(line string) bool {
	return regexp.MustCompile(`^(\s*[-*+]|\s*\d+\.)\s`).MatchString(line)
}

func (s *BlockMapperService) isOrderedListItem(line string) bool {
	return regexp.MustCompile(`^\d+\.\s`).MatchString(line)
}

func (s *BlockMapperService) extractListItemContent(line string) string {
	// Remove list marker
	return regexp.MustCompile(`^(\s*[-*+]|\s*\d+\.)\s*`).ReplaceAllString(line, "")
}

func (s *BlockMapperService) isTaskItem(content string) (bool, bool) {
	if strings.HasPrefix(content, "[ ] ") {
		return true, false
	}
	if strings.HasPrefix(content, "[x] ") || strings.HasPrefix(content, "[X] ") {
		return true, true
	}
	return false, false
}

func (s *BlockMapperService) isTableSeparator(line string) bool {
	return regexp.MustCompile(`^\|[\s:-]+\|$`).MatchString(line)
}

func (s *BlockMapperService) parseTableRow(line string) []string {
	// Remove leading and trailing |
	line = strings.Trim(line, "|")
	cells := strings.Split(line, "|")

	for i, cell := range cells {
		cells[i] = strings.TrimSpace(cell)
	}

	return cells
}

func (s *BlockMapperService) hashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:8])
}

// BlocksToMarkdown converts blocks back to markdown
func (s *BlockMapperService) BlocksToMarkdown(blocks []*Block) string {
	var builder strings.Builder

	for _, block := range blocks {
		s.blockToMarkdown(&builder, block)
		builder.WriteString("\n")
	}

	return strings.TrimSpace(builder.String())
}

func (s *BlockMapperService) blockToMarkdown(builder *strings.Builder, block *Block) {
	switch block.Type {
	case BlockTypeParagraph:
		builder.WriteString(block.Content)
		builder.WriteString("\n")

	case BlockTypeHeading:
		builder.WriteString(strings.Repeat("#", block.Level))
		builder.WriteString(" ")
		builder.WriteString(block.Content)
		builder.WriteString("\n")

	case BlockTypeCode:
		builder.WriteString("```")
		builder.WriteString(block.Language)
		builder.WriteString("\n")
		builder.WriteString(block.Content)
		builder.WriteString("\n```\n")

	case BlockTypeBlockquote:
		lines := strings.Split(block.Content, "\n")
		for _, line := range lines {
			builder.WriteString("> ")
			builder.WriteString(line)
			builder.WriteString("\n")
		}

	case BlockTypeList:
		for i, child := range block.Children {
			if block.Metadata["ordered"] == true {
				builder.WriteString(fmt.Sprintf("%d. ", i+1))
			} else {
				builder.WriteString("- ")
			}
			builder.WriteString(child.Content)
			builder.WriteString("\n")
		}

	case BlockTypeTable:
		for i, row := range block.Children {
			builder.WriteString("|")
			for _, cell := range row.Children {
				builder.WriteString(" ")
				builder.WriteString(cell.Content)
				builder.WriteString(" |")
			}
			builder.WriteString("\n")

			// Add separator after header
			if i == 0 {
				builder.WriteString("|")
				for range row.Children {
					builder.WriteString(" --- |")
				}
				builder.WriteString("\n")
			}
		}

	case BlockTypeHR:
		builder.WriteString("---\n")

	case BlockTypeThinking:
		builder.WriteString("<thinking>\n")
		builder.WriteString(block.Content)
		builder.WriteString("\n</thinking>\n")

	case BlockTypeArtifact:
		builder.WriteString("<artifact")
		if id, ok := block.Metadata["identifier"].(string); ok {
			builder.WriteString(fmt.Sprintf(" identifier=\"%s\"", id))
		}
		if t, ok := block.Metadata["type"].(string); ok {
			builder.WriteString(fmt.Sprintf(" type=\"%s\"", t))
		}
		if title, ok := block.Metadata["title"].(string); ok {
			builder.WriteString(fmt.Sprintf(" title=\"%s\"", title))
		}
		builder.WriteString(">\n")
		builder.WriteString(block.Content)
		builder.WriteString("\n</artifact>\n")

	case BlockTypeMath:
		builder.WriteString("$$\n")
		builder.WriteString(block.Content)
		builder.WriteString("\n$$\n")

	case BlockTypeCallout:
		calloutType := "note"
		if t, ok := block.Metadata["callout_type"].(string); ok {
			calloutType = t
		}
		lines := strings.Split(block.Content, "\n")
		for i, line := range lines {
			if i == 0 {
				builder.WriteString(fmt.Sprintf("> [!%s] %s\n", calloutType, line))
			} else {
				builder.WriteString("> ")
				builder.WriteString(line)
				builder.WriteString("\n")
			}
		}

	default:
		builder.WriteString(block.Content)
		builder.WriteString("\n")
	}
}

// SaveBlockDocument persists a block document to the database
func (s *BlockMapperService) SaveBlockDocument(ctx context.Context, doc *BlockDocument) error {
	blocksJSON, err := json.Marshal(doc.Blocks)
	if err != nil {
		return err
	}
	outlineJSON, err := json.Marshal(doc.Outline)
	if err != nil {
		return err
	}
	metadataJSON, err := json.Marshal(doc.Metadata)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO block_documents (id, source_id, title, blocks, outline, metadata, hash, total_blocks, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (id) DO UPDATE SET
		    blocks = EXCLUDED.blocks,
		    outline = EXCLUDED.outline,
		    metadata = EXCLUDED.metadata,
		    hash = EXCLUDED.hash,
		    total_blocks = EXCLUDED.total_blocks`,
		doc.ID, doc.SourceID, doc.Title, blocksJSON, outlineJSON, metadataJSON,
		doc.Hash, doc.TotalBlocks, doc.CreatedAt)

	return err
}

// GetBlockDocument retrieves a block document by ID
func (s *BlockMapperService) GetBlockDocument(ctx context.Context, id string) (*BlockDocument, error) {
	var doc BlockDocument
	var blocksJSON, outlineJSON, metadataJSON []byte

	err := s.db.QueryRowContext(ctx,
		`SELECT id, source_id, title, blocks, outline, metadata, hash, total_blocks, created_at
		 FROM block_documents WHERE id = $1`,
		id).Scan(&doc.ID, &doc.SourceID, &doc.Title, &blocksJSON, &outlineJSON, &metadataJSON,
		&doc.Hash, &doc.TotalBlocks, &doc.CreatedAt)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(blocksJSON, &doc.Blocks)
	json.Unmarshal(outlineJSON, &doc.Outline)
	json.Unmarshal(metadataJSON, &doc.Metadata)

	return &doc, nil
}
