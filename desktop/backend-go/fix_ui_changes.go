// +build ignore

package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Fix 1: Add web search count variable and thinking event in chat_v2.go
	fixChatV2()

	// Fix 2: Remove emojis from orchestration.go
	fixOrchestration()
}

func fixChatV2() {
	filePath := "internal/handlers/chat_v2.go"
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading chat_v2.go:", err)
		return
	}

	contentStr := strings.ReplaceAll(string(content), "\r\n", "\n")
	hasCRLF := bytes.Contains(content, []byte("\r\n"))
	fmt.Println("chat_v2.go uses CRLF:", hasCRLF)

	// Fix 1a: Add searchResultCount variable after searchContextText
	old1 := `var searchContextText string`
	new1 := `var searchContextText string
	var searchResultCount int`

	if strings.Contains(contentStr, old1) && !strings.Contains(contentStr, "searchResultCount int") {
		contentStr = strings.Replace(contentStr, old1, new1, 1)
		fmt.Println("Fix 1a applied: Added searchResultCount variable")
	} else {
		fmt.Println("Fix 1a: Pattern not found or already applied")
	}

	// Fix 1b: Set searchResultCount when web search returns results
	old2 := `if len(focusCtx.SearchContext) > 0 {
				searchContextText = focusService.FormatContextForPrompt(focusCtx)
				log.Printf("[ChatV2] Web search returned %d results for focus mode", len(focusCtx.SearchContext))
			}`
	new2 := `if len(focusCtx.SearchContext) > 0 {
				searchContextText = focusService.FormatContextForPrompt(focusCtx)
				searchResultCount = len(focusCtx.SearchContext)
				log.Printf("[ChatV2] Web search returned %d results for focus mode", searchResultCount)
			}`

	if strings.Contains(contentStr, old2) {
		contentStr = strings.Replace(contentStr, old2, new2, 1)
		fmt.Println("Fix 1b applied: Set searchResultCount")
	} else {
		fmt.Println("Fix 1b: Pattern not found or already applied")
	}

	// Fix 1c: Add web search thinking event after initial thinking event
	old3 := `writeSSEEvent(w, streaming.StreamEvent{
				Type: streaming.EventTypeThinking,
				Data: streaming.ThinkingStep{
					Step:      "analyzing",
					Content:   "Processing your request...",
					Agent:     string(agentType),
					Completed: false,
				},
			})
		}`
	new3 := `writeSSEEvent(w, streaming.StreamEvent{
				Type: streaming.EventTypeThinking,
				Data: streaming.ThinkingStep{
					Step:      "analyzing",
					Content:   "Processing your request...",
					Agent:     string(agentType),
					Completed: false,
				},
			})

			// Send web search notification if search was performed
			if searchResultCount > 0 {
				writeSSEEvent(w, streaming.StreamEvent{
					Type: streaming.EventTypeThinking,
					Data: streaming.ThinkingStep{
						Step:      "web_search",
						Content:   fmt.Sprintf("Web search completed: %d sources found", searchResultCount),
						Agent:     string(agentType),
						Completed: true,
					},
				})
			}
		}`

	if strings.Contains(contentStr, old3) {
		contentStr = strings.Replace(contentStr, old3, new3, 1)
		fmt.Println("Fix 1c applied: Added web search thinking event")
	} else {
		fmt.Println("Fix 1c: Pattern not found or already applied")
	}

	if hasCRLF {
		contentStr = strings.ReplaceAll(contentStr, "\n", "\r\n")
	}

	err = os.WriteFile(filePath, []byte(contentStr), 0644)
	if err != nil {
		fmt.Println("Error writing chat_v2.go:", err)
		return
	}
	fmt.Println("chat_v2.go fixes applied successfully!")
}

func fixOrchestration() {
	filePath := "internal/agents/orchestration.go"
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading orchestration.go:", err)
		return
	}

	contentStr := strings.ReplaceAll(string(content), "\r\n", "\n")
	hasCRLF := bytes.Contains(content, []byte("\r\n"))
	fmt.Println("orchestration.go uses CRLF:", hasCRLF)

	// Remove emojis
	replacements := []struct {
		old string
		new string
	}{
		{`"🧠 *Analyzing request...*\n\n"`, `"Analyzing request...\n\n"`},
		{`fmt.Sprintf("📋 *Plan: %s*\n\n", plan.Strategy)`, `fmt.Sprintf("Plan: %s\n\n", plan.Strategy)`},
		{`fmt.Sprintf("🤖 **%s Agent**\n\n", targetAgent)`, `fmt.Sprintf("%s Agent\n\n", targetAgent)`},
		{`fmt.Sprintf("🔄 **Multi-Agent Execution** (%d agents)\n\n", len(plan.Steps))`, `fmt.Sprintf("Multi-Agent Execution (%d agents)\n\n", len(plan.Steps))`},
		{`"### 🧠 Chain of Thought Summary\n\n"`, `"### Chain of Thought Summary\n\n"`},
	}

	for i, r := range replacements {
		if strings.Contains(contentStr, r.old) {
			contentStr = strings.Replace(contentStr, r.old, r.new, 1)
			fmt.Printf("Emoji fix %d applied: removed emoji\n", i+1)
		} else {
			fmt.Printf("Emoji fix %d: Pattern not found or already applied\n", i+1)
		}
	}

	if hasCRLF {
		contentStr = strings.ReplaceAll(contentStr, "\n", "\r\n")
	}

	err = os.WriteFile(filePath, []byte(contentStr), 0644)
	if err != nil {
		fmt.Println("Error writing orchestration.go:", err)
		return
	}
	fmt.Println("orchestration.go fixes applied successfully!")
}
