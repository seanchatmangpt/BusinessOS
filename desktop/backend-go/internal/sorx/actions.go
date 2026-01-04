// Package sorx provides built-in action handlers for the skill execution engine.
package sorx

import (
	"context"
	"fmt"
)

func init() {
	// Register all built-in actions
	RegisterAction("gmail.list_messages", gmailListMessages)
	RegisterAction("gmail.send_email", gmailSendEmail)
	RegisterAction("gmail.search", gmailSearch)

	RegisterAction("google_calendar.list_events", googleCalendarListEvents)
	RegisterAction("google_calendar.create_event", googleCalendarCreateEvent)

	RegisterAction("hubspot.list_contacts", hubspotListContacts)
	RegisterAction("hubspot.create_contact", hubspotCreateContact)

	RegisterAction("linear.list_issues", linearListIssues)
	RegisterAction("linear.create_issue", linearCreateIssue)

	RegisterAction("slack.send_message", slackSendMessage)
	RegisterAction("slack.list_channels", slackListChannels)

	RegisterAction("notion.search", notionSearch)
	RegisterAction("notion.create_page", notionCreatePage)

	RegisterAction("ai.extract_actions", aiExtractActions)
	RegisterAction("ai.summarize", aiSummarize)
	RegisterAction("ai.classify", aiClassify)

	RegisterAction("transform.map_fields", transformMapFields)
	RegisterAction("transform.filter", transformFilter)

	RegisterAction("businessos.create_tasks", businessOSCreateTasks)
	RegisterAction("businessos.upsert_clients", businessOSUpsertClients)
	RegisterAction("businessos.create_daily_log", businessOSCreateDailyLog)
	RegisterAction("businessos.import_tasks", businessOSImportTasks)
	RegisterAction("businessos.create_nodes", businessOSCreateNodes)
	RegisterAction("businessos.list_pending_tasks", businessOSListPendingTasks)
	RegisterAction("businessos.get_client_summary", businessOSGetClientSummary)
	RegisterAction("businessos.get_pipeline_summary", businessOSGetPipelineSummary)
	RegisterAction("businessos.get_meeting_context", businessOSGetMeetingContext)

	RegisterAction("google_calendar.get_event", googleCalendarGetEvent)

	// Context gathering actions for command-based skills
	RegisterAction("businessos.gather_context", businessOSGatherContext)
}

func businessOSGatherContext(ctx context.Context, ac ActionContext) (interface{}, error) {
	sources, _ := ac.Params["sources"].([]interface{})

	// Placeholder - would gather context from specified sources
	// In production, this would:
	// 1. Query the database for documents, conversations, artifacts, etc.
	// 2. Build a context bundle with relevant information
	// 3. Return structured data for the next step

	sourceList := make([]string, 0)
	for _, s := range sources {
		if str, ok := s.(string); ok {
			sourceList = append(sourceList, str)
		}
	}

	return map[string]interface{}{
		"sources":   sourceList,
		"documents": []interface{}{},
		"conversations": []interface{}{},
		"artifacts": []interface{}{},
		"clients": []interface{}{},
		"projects": []interface{}{},
		"context_built": true,
	}, nil
}

// ============================================================================
// Gmail Actions
// ============================================================================

func gmailListMessages(ctx context.Context, ac ActionContext) (interface{}, error) {
	// In production, this would use the Gmail API with the user's credentials
	maxResults := 50
	if val, ok := ac.Params["max_results"].(float64); ok {
		maxResults = int(val)
	}

	// Placeholder - would call Gmail API
	return map[string]interface{}{
		"messages": []interface{}{},
		"count":    0,
		"max":      maxResults,
	}, nil
}

func gmailSendEmail(ctx context.Context, ac ActionContext) (interface{}, error) {
	to, _ := ac.Params["to"].(string)
	subject, _ := ac.Params["subject"].(string)
	body, _ := ac.Params["body"].(string)

	if to == "" || subject == "" {
		return nil, fmt.Errorf("to and subject are required")
	}

	// Placeholder - would call Gmail API
	return map[string]interface{}{
		"sent":    true,
		"to":      to,
		"subject": subject,
		"body":    body,
	}, nil
}

func gmailSearch(ctx context.Context, ac ActionContext) (interface{}, error) {
	query, _ := ac.Params["query"].(string)

	// Placeholder - would call Gmail API
	return map[string]interface{}{
		"query":    query,
		"messages": []interface{}{},
		"count":    0,
	}, nil
}

// ============================================================================
// Google Calendar Actions
// ============================================================================

func googleCalendarListEvents(ctx context.Context, ac ActionContext) (interface{}, error) {
	daysAhead := 7
	if val, ok := ac.Params["days_ahead"].(float64); ok {
		daysAhead = int(val)
	}

	// Placeholder - would call Google Calendar API
	return map[string]interface{}{
		"events":     []interface{}{},
		"count":      0,
		"days_ahead": daysAhead,
	}, nil
}

func googleCalendarCreateEvent(ctx context.Context, ac ActionContext) (interface{}, error) {
	title, _ := ac.Params["title"].(string)
	startTime, _ := ac.Params["start_time"].(string)

	if title == "" || startTime == "" {
		return nil, fmt.Errorf("title and start_time are required")
	}

	// Placeholder - would call Google Calendar API
	return map[string]interface{}{
		"created": true,
		"title":   title,
	}, nil
}

// ============================================================================
// HubSpot Actions
// ============================================================================

func hubspotListContacts(ctx context.Context, ac ActionContext) (interface{}, error) {
	// Placeholder - would call HubSpot API
	return map[string]interface{}{
		"contacts": []interface{}{},
		"count":    0,
	}, nil
}

func hubspotCreateContact(ctx context.Context, ac ActionContext) (interface{}, error) {
	email, _ := ac.Params["email"].(string)
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	// Placeholder - would call HubSpot API
	return map[string]interface{}{
		"created": true,
		"email":   email,
	}, nil
}

// ============================================================================
// Linear Actions
// ============================================================================

func linearListIssues(ctx context.Context, ac ActionContext) (interface{}, error) {
	// Placeholder - would call Linear API
	return map[string]interface{}{
		"issues": []interface{}{},
		"count":  0,
	}, nil
}

func linearCreateIssue(ctx context.Context, ac ActionContext) (interface{}, error) {
	title, _ := ac.Params["title"].(string)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	// Placeholder - would call Linear API
	return map[string]interface{}{
		"created": true,
		"title":   title,
	}, nil
}

// ============================================================================
// Slack Actions
// ============================================================================

func slackSendMessage(ctx context.Context, ac ActionContext) (interface{}, error) {
	channel, _ := ac.Params["channel"].(string)
	text, _ := ac.Params["text"].(string)

	if channel == "" || text == "" {
		return nil, fmt.Errorf("channel and text are required")
	}

	// Placeholder - would call Slack API
	return map[string]interface{}{
		"sent":    true,
		"channel": channel,
	}, nil
}

func slackListChannels(ctx context.Context, ac ActionContext) (interface{}, error) {
	// Placeholder - would call Slack API
	return map[string]interface{}{
		"channels": []interface{}{},
		"count":    0,
	}, nil
}

// ============================================================================
// Notion Actions
// ============================================================================

func notionSearch(ctx context.Context, ac ActionContext) (interface{}, error) {
	query, _ := ac.Params["query"].(string)

	// Placeholder - would call Notion API
	return map[string]interface{}{
		"query":   query,
		"results": []interface{}{},
		"count":   0,
	}, nil
}

func notionCreatePage(ctx context.Context, ac ActionContext) (interface{}, error) {
	title, _ := ac.Params["title"].(string)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	// Placeholder - would call Notion API
	return map[string]interface{}{
		"created": true,
		"title":   title,
	}, nil
}

// ============================================================================
// AI Actions
// ============================================================================

func aiExtractActions(ctx context.Context, ac ActionContext) (interface{}, error) {
	source, _ := ac.Params["source"].(string)

	// Placeholder - would call AI service
	return map[string]interface{}{
		"source":  source,
		"actions": []interface{}{},
		"count":   0,
	}, nil
}

func aiSummarize(ctx context.Context, ac ActionContext) (interface{}, error) {
	text, _ := ac.Params["text"].(string)

	// Placeholder - would call AI service
	return map[string]interface{}{
		"summary": fmt.Sprintf("Summary of: %s...", text[:min(50, len(text))]),
	}, nil
}

func aiClassify(ctx context.Context, ac ActionContext) (interface{}, error) {
	text, _ := ac.Params["text"].(string)
	categories, _ := ac.Params["categories"].([]interface{})

	// Placeholder - would call AI service
	return map[string]interface{}{
		"text":       text,
		"categories": categories,
		"result":     "unknown",
	}, nil
}

// ============================================================================
// Transform Actions
// ============================================================================

func transformMapFields(ctx context.Context, ac ActionContext) (interface{}, error) {
	mapping, _ := ac.Params["mapping"].(string)

	// Placeholder - would transform data based on mapping config
	return map[string]interface{}{
		"mapping":     mapping,
		"transformed": []interface{}{},
	}, nil
}

func transformFilter(ctx context.Context, ac ActionContext) (interface{}, error) {
	condition, _ := ac.Params["condition"].(string)

	// Placeholder - would filter data based on condition
	return map[string]interface{}{
		"condition": condition,
		"filtered":  []interface{}{},
	}, nil
}

// ============================================================================
// BusinessOS Actions
// ============================================================================

func businessOSCreateTasks(ctx context.Context, ac ActionContext) (interface{}, error) {
	from, _ := ac.Params["from"].(string)

	// Get items from previous step results
	items := ac.Execution.StepResults[from]

	// Placeholder - would create tasks in BusinessOS
	return map[string]interface{}{
		"created": 0,
		"source":  from,
		"items":   items,
	}, nil
}

func businessOSUpsertClients(ctx context.Context, ac ActionContext) (interface{}, error) {
	// Placeholder - would upsert clients in BusinessOS
	return map[string]interface{}{
		"upserted": 0,
	}, nil
}

func businessOSCreateDailyLog(ctx context.Context, ac ActionContext) (interface{}, error) {
	// Placeholder - would create daily log entries
	return map[string]interface{}{
		"created": 0,
	}, nil
}

func businessOSImportTasks(ctx context.Context, ac ActionContext) (interface{}, error) {
	// Get decision result if any
	decisionResult := ac.Execution.Context["decision_result"]

	// Placeholder - would import tasks based on decision
	return map[string]interface{}{
		"imported": 0,
		"decision": decisionResult,
	}, nil
}

func businessOSCreateNodes(ctx context.Context, ac ActionContext) (interface{}, error) {
	nodeType, _ := ac.Params["type"].(string)
	source, _ := ac.Params["source"].(string)

	// Placeholder - would create knowledge nodes
	return map[string]interface{}{
		"created": 0,
		"type":    nodeType,
		"source":  source,
	}, nil
}

func businessOSListPendingTasks(ctx context.Context, ac ActionContext) (interface{}, error) {
	// Placeholder - would list pending tasks from BusinessOS
	return map[string]interface{}{
		"tasks": []interface{}{},
		"count": 0,
	}, nil
}

func businessOSGetClientSummary(ctx context.Context, ac ActionContext) (interface{}, error) {
	clientID, _ := ac.Params["client_id"].(string)

	// Placeholder - would get client summary from BusinessOS
	return map[string]interface{}{
		"client_id": clientID,
		"summary":   "Client data placeholder",
	}, nil
}

func businessOSGetPipelineSummary(ctx context.Context, ac ActionContext) (interface{}, error) {
	// Placeholder - would get pipeline summary from BusinessOS
	return map[string]interface{}{
		"pipeline": []interface{}{},
		"total_value": 0,
		"stages": map[string]interface{}{},
	}, nil
}

func businessOSGetMeetingContext(ctx context.Context, ac ActionContext) (interface{}, error) {
	// Placeholder - would gather meeting context from BusinessOS
	return map[string]interface{}{
		"attendees":       []interface{}{},
		"previous_notes":  []interface{}{},
		"related_clients": []interface{}{},
	}, nil
}

func googleCalendarGetEvent(ctx context.Context, ac ActionContext) (interface{}, error) {
	eventID, _ := ac.Params["event_id"].(string)

	// Placeholder - would get specific calendar event
	return map[string]interface{}{
		"event_id": eventID,
		"title":    "",
		"start":    "",
		"end":      "",
		"attendees": []interface{}{},
	}, nil
}

// Helper
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
