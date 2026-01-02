# Exclusive Backend Development Tasks for Pedro - V2

This document contains **EXCLUSIVELY** the tasks assigned to Pedro in the "Memory, Context & Intelligence System" plan. All shared or Nick-specific tasks have been removed.

---

## 1. Intelligence & Memory Services
- [x] **Memory Service (`services/memory.go`)**:
    - Implement CRUD operations for episodic memories.
    - Build semantic search and retrieval logic using embeddings.
    - Develop auto-extraction logic for memories from conversations and voice notes.
    - Implement importance scoring and access tracking.
    - Implement User Facts management (preferences, facts, style).

## 2. Context Management & Tools
- [x] **Tree Search Tools**: Develop the following Go tools for AI agents:
    - `TreeSearchTool`: For searching titles, content, or semantic matching in the knowledge base.
    - `LoadContextTool`: For loading specific documents/memories into an agent's context.
    - `BrowseTreeTool`: For hierarchical navigation of the context tree.
- [x] **Context Service (`services/context.go`)**:
    - Build context building logic for agents based on project/node selection.
    - Manage context profiles and tree operation logic.
- [x] **Context Window Tracking (`services/context_tracker.go`)**:
    - Implement token usage monitoring per agent session.
    - Handle LRU-style (Least Recently Used) context eviction to stay within model limits.

## 3. Block System Integration
- [x] **Block Type Mapping (`services/block_mapper.go`)**:
    - Build the service to convert markdown/AI responses into the platform's specific block-based data structure (JSON).

## 4. Document & Files System
- [x] **Document Processor (`services/document_processor.go`)**:
    - Build the processing pipeline for uploaded files.
    - Implement text extraction from PDFs, Markdown, and Docx.
    - Develop chunking logic to split large documents for better semantic retrieval.
    - Build semantic search specifically for the document library.
- [x] **Context Profile Logic**: Implement the backend logic to link various items (documents, memories, artifacts) to profiles.

## 5. Intelligent Chat Features
- [x] **Chat History Intelligence**:
    - Implement conversation summarization logic.
    - Build topic extraction and decision tracking from past chats.
- [x] **Context Injection**: Update chat handlers to inject retrieved context (memories/docs) into the system prompt.
- [x] **Tree Response Logic**: Implement the backend logic to generate the hierarchical tree structure (JSON) used for the frontend visualization.

## 6. Self-Learning & Application Context
- [x] **Learning Service (`services/learning.go`)**:
    - Build logic to process user feedback and corrections.
    - Implement behavior pattern detection.
    - Maintain and update the user's personalization profile.
- [x] **App Profiler Service**: Develop the logic to auto-profile codebases (Phase 8), identifying components, modules, and tech stacks.

---

## Summary of Pedro's Exclusive Responsibilities
Pedro is primarily responsible for the **functional intelligence layer**:
- Logic for memory/context retrieval and storage.
- The "Load/Search" toolset for agents.
- Text processing and document intelligence.
- The mapping between AI text and UI Blocks.
- Learning and personalization logic.

---

## Implementation Status: 100% COMPLETE

All tasks have been fully implemented. Key files:

| Component | File | Status |
|-----------|------|--------|
| Memory Service | `services/memory.go`, `services/memory_extractor.go` | Complete |
| Tree Search Tools | `services/context.go` (lines 185-280) | Complete |
| Context Service | `services/context.go`, `services/project_context.go` | Complete |
| Context Tracker | `services/context_tracker.go` (LRU eviction) | Complete |
| Block Mapper | `services/block_mapper.go` | Complete |
| Document Processor | `services/document_processor.go` (PDF/DOCX/Markdown) | Complete |
| Chat Intelligence | `services/conversation_intelligence.go` | Complete |
| Context Injection | `services/project_context.go` | Complete |
| Learning Service | `services/learning.go` | Complete |
| App Profiler | `services/app_profiler.go` | Complete |

### Database Migrations
- `016_memories.sql` - Memory tables
- `017_context_system.sql` - Context tree structure
- `018_output_styles.sql` - Output formatting
- `019_documents.sql` / `019_documents_no_vector.sql` - Document storage
- `020_context_integration.sql` - Context integration
- `021_learning_system.sql` - Learning/feedback system
- `022_application_profiles.sql` - App profiling
- `023_pedro_tasks_schema_fix.sql` - Schema fixes

### API Handlers (all registered in handlers.go:614-644)
- `/api/documents/*` - Document CRUD and search
- `/api/learning/*` - Feedback and behavior patterns
- `/api/app-profiles/*` - Application profiling
- `/api/intelligence/*` - Conversation intelligence
