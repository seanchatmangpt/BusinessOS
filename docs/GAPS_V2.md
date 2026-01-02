# BusinessOS V2: Remaining Implementation Gaps

This document tracks the tasks from the [taks_v2.md](file:///c:/Users/Pichau/Desktop/BusinessOS-main-dev/docs/taks_v2.md) master plan that are NOT yet fully operational.

## 🔴 CRITICAL GAPS

### 1. Output Styles & Block System (Part 3)
- [ ] **Markdown-to-Block Translation**: Backend logic to convert LLM Markdown responses into the structured `Block` format (`paragraph`, `bullet_list`, `code`, `callout`) used by the frontend documents.
- [ ] **Context-Specific Style Selection**: Automatic switching of output styles based on `focus_mode` or `agent_type` (e.g., 'technical' for @coder, 'executive' for @analyst).
- [ ] **User Style Overrides**: Persistence for user-defined default styles in the `user_output_preferences` table.

### 2. Deep Context Integration (Part 4)
- [ ] **Conversation Summarization**: Background job/service to generate and update `conversation_summaries`, allowing the system to index and search past chat history semantically.
- [ ] **Voice Note Semantic Search**: Indexing voice note transcriptions with vector embeddings for inclusion in the `TreeSearchTool` results.
- [ ] **Node Hierarchy Inheritance**: Logic to automatically pull "Parent Node" context when a child node is selected.

### 3. Self-Learning & Behavior Patterns (Part 6)
- [ ] **Behavior Pattern Detection**: Service to analyze interaction history and detect patterns like preferred response length, common technical topics, or timezone-dependent behavior.
- [ ] **Explicit Learning Validation**: UI/UX flow for users to "Confirm" or "Reject" facts the AI thinks it has learned about them.
- [ ] **Automatic Context Injection**: System-wide middleware to inject "User Facts" (e.g., tech stack, company size) into every agent session without manual tool calls.

## 🟡 ENHANCEMENTS & REFINEMENT

### Context Management
- [ ] **Context Tree API Refinement**: Dedicated stats endpoints for the tree visualization (`/api/context-tree/stats`).
- [ ] **Token Budgeting (Priority-Based)**: Implementing strict LRU eviction across *all* context types (Document Chunks vs Memories vs Recent Chat).

### Developer Experience
- [ ] **Application Profiler Sync**: Auto-syncing `ApplicationProfiles` with local file changes or Git events.
- [ ] **Specialized Specialist Prompts**: Updating @coder, @analyst, and @researcher prompts to be as precise and recursive as the new Orchestrator V2.
