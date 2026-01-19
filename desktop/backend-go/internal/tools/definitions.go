package tools

// ToolDefinition represents a Groq function calling tool
type ToolDefinition struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// GetAllTools returns all available tools for voice agent
func GetAllTools() []ToolDefinition {
	return []ToolDefinition{
		// Navigation tool
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "navigate_to_module",
				Description: "Open a module in the BusinessOS interface when the user asks to navigate somewhere or open something",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"module": map[string]interface{}{
							"type": "string",
							"enum": []string{
								"dashboard", "chat", "tasks", "projects",
								"team", "clients", "terminal", "settings",
								"pages", "knowledge-v2", "nodes", "daily",
								"notifications", "voice-notes", "help",
								"agents", "integrations", "usage", "app-store",
								"trash", "tables", "crm", "communication",
								"profile",
							},
							"description": "The module to open. Common requests: 'tasks' for tasks, 'projects' for projects, 'dashboard' for home, 'chat' for conversations, 'team' for team management, 'clients' for client management.",
						},
					},
					"required": []string{"module"},
				},
			},
		},

		// Task creation tool
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "create_task",
				Description: "Create a new task when user asks to add a task, reminder, or todo item",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title": map[string]interface{}{
							"type":        "string",
							"description": "The task title or description",
						},
						"due_date": map[string]interface{}{
							"type":        "string",
							"description": "Optional due date in YYYY-MM-DD format",
						},
						"priority": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"low", "medium", "high", "urgent"},
							"description": "Task priority level",
						},
					},
					"required": []string{"title"},
				},
			},
		},

		// Task listing tool
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "list_tasks",
				Description: "Get the user's tasks, optionally filtered by status",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"status": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"all", "active", "completed"},
							"description": "Filter tasks by status",
						},
						"limit": map[string]interface{}{
							"type":        "number",
							"description": "Maximum number of tasks to return (default 10)",
						},
					},
					"required": []string{},
				},
			},
		},

		// Project creation tool
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "create_project",
				Description: "Create a new project when user asks to start a new project",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "The project name",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Optional project description",
						},
					},
					"required": []string{"name"},
				},
			},
		},

		// Context search tool
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "search_context",
				Description: "Search the user's knowledge base, notes, and documents",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "The search query",
						},
					},
					"required": []string{"query"},
				},
			},
		},
	}
}
