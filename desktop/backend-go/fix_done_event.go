// +build ignore

package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

func main() {
	filePath := "internal/agents/orchestration.go"
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Normalize to LF for pattern matching
	contentStr := strings.ReplaceAll(string(content), "\r\n", "\n")
	hasCRLF := bytes.Contains(content, []byte("\r\n"))
	fmt.Println("File uses CRLF:", hasCRLF)

	// Fix 1: executeDelegation - add EventTypeDone before return when agent channel closes
	old1 := `		case event, ok := <-agentEvents:
			if !ok {
				execStep.Output = output.String()
				cot.UpdateStep(execStep.ID, execStep.Output, "completed")
				cot.FinalOutput = output.String()
				return
			}`
	new1 := `		case event, ok := <-agentEvents:
			if !ok {
				execStep.Output = output.String()
				cot.UpdateStep(execStep.ID, execStep.Output, "completed")
				cot.FinalOutput = output.String()
				events <- streaming.StreamEvent{Type: streaming.EventTypeDone}
				return
			}`

	if strings.Contains(contentStr, old1) {
		contentStr = strings.Replace(contentStr, old1, new1, 1)
		fmt.Println("Fix 1 applied: Added EventTypeDone to executeDelegation")
	} else {
		fmt.Println("Fix 1: Pattern not found or already applied")
	}

	// Fix 2: executeDirectly - same fix
	old2 := `		case event, ok := <-agentEvents:
			if !ok {
				step.Output = output.String()
				cot.UpdateStep(step.ID, step.Output, "completed")
				cot.FinalOutput = output.String()
				return`
	new2 := `		case event, ok := <-agentEvents:
			if !ok {
				step.Output = output.String()
				cot.UpdateStep(step.ID, step.Output, "completed")
				cot.FinalOutput = output.String()
				events <- streaming.StreamEvent{Type: streaming.EventTypeDone}
				return`

	if strings.Contains(contentStr, old2) {
		contentStr = strings.Replace(contentStr, old2, new2, 1)
		fmt.Println("Fix 2 applied: Added EventTypeDone to executeDirectly")
	} else {
		fmt.Println("Fix 2: Pattern not found or already applied")
	}

	// Convert back to CRLF if original used CRLF
	if hasCRLF {
		contentStr = strings.ReplaceAll(contentStr, "\n", "\r\n")
	}

	err = os.WriteFile(filePath, []byte(contentStr), 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}
	fmt.Println("Fixes applied successfully!")
}
