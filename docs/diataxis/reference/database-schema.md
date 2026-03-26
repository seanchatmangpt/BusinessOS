# Reference: Database Schema

**Information-oriented.** PostgreSQL schema lookup: tables, columns, types, constraints, relationships.

**Format:** Table → Columns → Types → Constraints → Indexes → Relationships

---

## Quick Decision Tree

**"Which table stores X?"**

| What | Table | Notes |
|-----|-------|-------|
| User account info | `user` (Better Auth managed) | External, not in schema.sql |
| Chat conversation | `conversations` | Thread of messages |
| Chat message | `messages` | Text + metadata |
| Project | `projects` | With status, priority, dates |
| Task/todo item | `tasks` | Subtasks via parent_task_id |
| Document/profile | `contexts` | PERSON, BUSINESS, PROJECT, DOCUMENT |
| Team member | `team_members` | Staff, skills, capacity |
| Client/company | `clients` | CRM customer |
| Workspace node | `nodes` | BUSINESS, PROJECT, LEARNING, OPERATIONAL |
| Artifact (code, doc) | `artifacts` | CODE, DOCUMENT, MARKDOWN, REACT, HTML, SVG |
| Project file | `generated_files` | Output from app generation |
| Calendar event | `calendar_events` | Meetings, reminders |
| Daily log entry | `daily_logs` | Journal, standup |
| OAuth token | `oauth_tokens` | Encrypted external credentials |
| Audit log | `audit_logs` | SOC2 A1 compliance |

---

## Core Tables

### user (Better Auth Managed)

**Purpose:** User accounts (managed externally by Better Auth).

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | VARCHAR(255) | PRIMARY KEY | Unique user ID |
| `name` | VARCHAR(255) | | Display name |
| `email` | VARCHAR(255) | UNIQUE | Email address |
| `email_verified` | BOOLEAN | DEFAULT FALSE | Email confirmed? |
| `image` | VARCHAR(500) | | Avatar URL |
| `created_at` | TIMESTAMPTZ | DEFAULT NOW() | Account creation |
| `updated_at` | TIMESTAMPTZ | DEFAULT NOW() | Last update |

**Indexes:**
- PRIMARY KEY: `id`
- UNIQUE: `email`

**Relations:**
- FK: contexts.user_id → user.id
- FK: conversations.user_id → user.id
- FK: projects.user_id → user.id
- FK: tasks.user_id → user.id
- FK: team_members.user_id → user.id

---

### conversations

**Purpose:** Chat conversation threads.

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Unique conversation ID |
| `user_id` | VARCHAR(255) | NOT NULL, FK | Owner of conversation |
| `title` | VARCHAR(255) | DEFAULT 'New Conversation' | Display name |
| `context_id` | UUID | FK | Linked document/context |
| `created_at` | TIMESTAMP | DEFAULT NOW() | Creation time |
| `updated_at` | TIMESTAMP | DEFAULT NOW() | Last message time |

**Indexes:**
- PRIMARY KEY: `id`
- INDEX: `user_id` (fast user lookups)

**Relations:**
- FK: messages.conversation_id → conversations.id (CASCADE)
- FK: contexts.id → conversations.context_id

---

### messages

**Purpose:** Individual chat messages within conversations.

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Unique message ID |
| `conversation_id` | UUID | NOT NULL, FK | Parent conversation |
| `role` | messagerole | NOT NULL | USER, ASSISTANT, SYSTEM |
| `content` | TEXT | NOT NULL | Message body |
| `created_at` | TIMESTAMP | DEFAULT NOW() | Send time |
| `message_metadata` | JSONB | | Additional data (embeddings, tool calls, etc.) |

**Columns: message_metadata**
- `tool_calls`: Array of function calls
- `citations`: References to artifacts/contexts
- `embedding_vector`: For semantic search

**Indexes:**
- PRIMARY KEY: `id`
- INDEX: `conversation_id` (fast message lookup)

**Relations:**
- FK: conversation_id → conversations.id (CASCADE)
- FK: message_id ← artifacts.message_id

---

### conversation_tags

**Purpose:** Tags for conversations (organization, search).

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Unique tag ID |
| `conversation_id` | UUID | NOT NULL, FK | Parent conversation |
| `tag` | VARCHAR(100) | NOT NULL | Tag string |
| `created_at` | TIMESTAMP | DEFAULT NOW() | Creation time |

**Relations:**
- FK: conversation_id → conversations.id (CASCADE)

---

## Project Management

### projects

**Purpose:** Projects (containers for tasks, conversations, team).

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Unique project ID |
| `user_id` | VARCHAR(255) | NOT NULL | Owner |
| `name` | VARCHAR(255) | NOT NULL | Project name |
| `description` | TEXT | | Long description |
| `status` | projectstatus | DEFAULT 'ACTIVE' | ACTIVE, PAUSED, COMPLETED, ARCHIVED |
| `priority` | projectpriority | DEFAULT 'MEDIUM' | CRITICAL, HIGH, MEDIUM, LOW |
| `client_name` | VARCHAR(255) | | Client name (denormalized) |
| `client_id` | UUID | FK | Link to clients table |
| `project_type` | VARCHAR(100) | DEFAULT 'internal' | internal, client, research |
| `project_metadata` | JSONB | | Custom fields |
| `start_date` | DATE | | Project start |
| `due_date` | DATE | | Project deadline |
| `completed_at` | TIMESTAMP | | Completion time (if done) |
| `visibility` | VARCHAR(20) | DEFAULT 'private' | private, team, public |
| `owner_id` | VARCHAR(255) | | Alternative owner ref |
| `created_at` | TIMESTAMP | DEFAULT NOW() | Creation time |
| `updated_at` | TIMESTAMP | DEFAULT NOW() | Last update |

**Indexes:**
- PRIMARY KEY: `id`
- INDEX: `user_id` (user's projects)
- INDEX: `client_id` (client's projects)
- INDEX: `status` (filter by status)
- INDEX: `due_date` (sort by deadline)

**Relations:**
- FK: client_id → clients.id (SET NULL on delete)
- FK: ← tasks.project_id
- FK: ← project_conversations.project_id

---

### project_notes

**Purpose:** Notes attached to projects.

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Unique note ID |
| `project_id` | UUID | NOT NULL, FK | Parent project |
| `content` | TEXT | NOT NULL | Note text |
| `created_at` | TIMESTAMP | DEFAULT NOW() | Creation time |

**Relations:**
- FK: project_id → projects.id (CASCADE)

---

### project_conversations

**Purpose:** Many-to-many: conversations linked to projects.

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `project_id` | UUID | NOT NULL, FK | Project |
| `conversation_id` | UUID | NOT NULL, FK | Conversation |
| `linked_at` | TIMESTAMP | DEFAULT NOW() | When linked |

**Composite Key:**
- PRIMARY KEY: (project_id, conversation_id)

**Relations:**
- FK: project_id → projects.id (CASCADE)
- FK: conversation_id → conversations.id (CASCADE)

---

### project_statuses

**Purpose:** Custom status values for projects (Kanban board).

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Status ID |
| `project_id` | UUID | NOT NULL, FK | Parent project |
| `name` | VARCHAR(100) | NOT NULL | Status name (e.g., "In Review") |
| `color` | VARCHAR(7) | DEFAULT '#6B7280' | Hex color code |
| `position` | INT | DEFAULT 0 | Sort order |
| `is_done_state` | BOOLEAN | DEFAULT FALSE | Mark as "complete" status? |

**Relations:**
- FK: project_id → projects.id (CASCADE)

---

### tasks

**Purpose:** Individual tasks (work items, todos).

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Unique task ID |
| `user_id` | VARCHAR(255) | NOT NULL | Owner/creator |
| `title` | VARCHAR(500) | NOT NULL | Task name |
| `description` | TEXT | | Details |
| `status` | taskstatus | DEFAULT 'todo' | todo, in_progress, done, cancelled |
| `priority` | taskpriority | DEFAULT 'medium' | critical, high, medium, low |
| `due_date` | TIMESTAMP | | Deadline |
| `start_date` | TIMESTAMP | | Work start time |
| `completed_at` | TIMESTAMP | | Actual completion time |
| `project_id` | UUID | FK | Parent project |
| `assignee_id` | UUID | FK | Team member assigned |
| `parent_task_id` | UUID | FK | Parent task (for subtasks) |
| `custom_status_id` | UUID | | Custom status from project_statuses |
| `position` | INT | DEFAULT 0 | Sort order in list |
| `created_at` | TIMESTAMP | DEFAULT NOW() | Creation time |
| `updated_at` | TIMESTAMP | DEFAULT NOW() | Last update |

**Indexes:**
- PRIMARY KEY: `id`
- INDEX: `user_id` (user's tasks)
- INDEX: `parent_task_id` (subtasks)
- INDEX: `(user_id, position)` (sorted task list)

**Relations:**
- FK: project_id → projects.id (SET NULL)
- FK: assignee_id → team_members.id (SET NULL)
- FK: parent_task_id → tasks.id (CASCADE)

---

## Knowledge & Documents

### contexts

**Purpose:** Documents, profiles, knowledge entries (PERSON, BUSINESS, PROJECT, DOCUMENT).

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Unique context ID |
| `user_id` | VARCHAR(255) | NOT NULL | Owner |
| `name` | VARCHAR(255) | NOT NULL | Title |
| `type` | contexttype | DEFAULT 'CUSTOM' | PERSON, BUSINESS, PROJECT, DOCUMENT |
| `content` | TEXT | | Full content |
| `structured_data` | JSONB | | Structured fields (schema-flexible) |
| `system_prompt_template` | TEXT | | LLM system prompt for this context |
| `blocks` | JSONB | DEFAULT '[]' | Document blocks (rich text structure) |
| `cover_image` | VARCHAR(500) | | Cover image URL |
| `icon` | VARCHAR(50) | | Icon name or emoji |
| `parent_id` | UUID | FK | Folder/parent context |
| `is_template` | BOOLEAN | DEFAULT FALSE | Reusable template? |
| `is_archived` | BOOLEAN | DEFAULT FALSE | Archived? |
| `last_edited_at` | TIMESTAMP | | Last modification |
| `word_count` | INTEGER | DEFAULT 0 | For search/stats |
| `is_public` | BOOLEAN | DEFAULT FALSE | Public share? |
| `share_id` | VARCHAR(32) | UNIQUE | Public share token |
| `property_schema` | JSONB | DEFAULT '[]' | Schema for structured_data fields |
| `properties` | JSONB | DEFAULT '{}' | Instance properties |
| `client_id` | UUID | | Associated client |
| `created_at` | TIMESTAMP | DEFAULT NOW() | Creation time |
| `updated_at` | TIMESTAMP | DEFAULT NOW() | Last update |

**Indexes:**
- PRIMARY KEY: `id`
- INDEX: `user_id` (user's contexts)
- INDEX: `parent_id` (folder structure)
- INDEX: `is_archived` (filter archived)
- INDEX: `share_id` (public lookup)

**Relations:**
- FK: parent_id → contexts.id (SET NULL)
- FK: client_id → clients.id
- FK: ← conversations.context_id
- FK: ← artifacts.context_id
- FK: ← nodes.context_id

---

## Artifacts & Generated Content

### artifacts

**Purpose:** Generated artifacts (code, documents, markdown, React, HTML, SVG).

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Unique artifact ID |
| `user_id` | VARCHAR(255) | NOT NULL | Creator |
| `conversation_id` | UUID | FK | Source conversation |
| `message_id` | UUID | FK | Source message |
| `project_id` | UUID | FK | Associated project |
| `context_id` | UUID | FK | Associated document |
| `title` | VARCHAR(255) | NOT NULL | Artifact name |
| `type` | artifacttype | NOT NULL | CODE, DOCUMENT, MARKDOWN, REACT, HTML, SVG |
| `language` | VARCHAR(50) | | Programming language (if CODE) |
| `content` | TEXT | NOT NULL | Full content |
| `summary` | VARCHAR(500) | | Brief description |
| `version` | INTEGER | DEFAULT 1 | Version number |
| `created_at` | TIMESTAMP | DEFAULT NOW() | Creation time |
| `updated_at` | TIMESTAMP | DEFAULT NOW() | Last update |

**Indexes:**
- PRIMARY KEY: `id`
- INDEX: `user_id` (user's artifacts)
- INDEX: `conversation_id` (artifacts in conversation)

**Relations:**
- FK: conversation_id → conversations.id (SET NULL)
- FK: message_id → messages.id (SET NULL)
- FK: project_id → projects.id (SET NULL)
- FK: context_id → contexts.id (SET NULL)

---

### artifact_versions

**Purpose:** Version history for artifacts.

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Version ID |
| `artifact_id` | UUID | NOT NULL, FK | Parent artifact |
| `version` | INTEGER | NOT NULL | Version number |
| `content` | TEXT | NOT NULL | Content at this version |
| `created_at` | TIMESTAMP | DEFAULT NOW() | Version creation |

**Relations:**
- FK: artifact_id → artifacts.id (CASCADE)

---

## Organization & Team

### team_members

**Purpose:** Team members (staff, contractors, capacity tracking).

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Unique member ID |
| `user_id` | VARCHAR(255) | NOT NULL | Associated user (can be NULL for contractors) |
| `name` | VARCHAR(255) | NOT NULL | Display name |
| `email` | VARCHAR(255) | NOT NULL | Email address |
| `role` | VARCHAR(255) | NOT NULL | Engineer, Designer, Manager, etc. |
| `avatar_url` | TEXT | | Profile photo URL |
| `status` | memberstatus | DEFAULT 'AVAILABLE' | AVAILABLE, BUSY, OVERLOADED, OOO |
| `capacity` | INTEGER | DEFAULT 0 | Hours/week available |
| `manager_id` | UUID | FK | Reporting manager |
| `skills` | JSONB | | Array of skill names |
| `hourly_rate` | NUMERIC(10,2) | | For budget tracking |
| `share_calendar` | BOOLEAN | DEFAULT FALSE | Share calendar with team? |
| `calendar_user_id` | VARCHAR(255) | | Google/Microsoft calendar ID |
| `joined_at` | TIMESTAMP | DEFAULT NOW() | Onboarding date |
| `created_at` | TIMESTAMP | DEFAULT NOW() | Record creation |
| `updated_at` | TIMESTAMP | DEFAULT NOW() | Last update |

**Indexes:**
- PRIMARY KEY: `id`
- INDEX: `user_id` (lookup by user)

**Relations:**
- FK: manager_id → team_members.id (SET NULL)
- FK: ← tasks.assignee_id
- FK: ← team_member_activities.member_id

---

### team_member_activities

**Purpose:** Activity log for team members (onboarding, training, reviews).

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Activity ID |
| `member_id` | UUID | NOT NULL, FK | Team member |
| `activity_type` | VARCHAR(100) | NOT NULL | onboarding, training, review, etc. |
| `description` | TEXT | NOT NULL | Details |
| `created_at` | TIMESTAMP | DEFAULT NOW() | When it happened |

**Relations:**
- FK: member_id → team_members.id (CASCADE)

---

## Workspace Structure

### nodes

**Purpose:** Workspace hierarchy (BUSINESS, PROJECT, LEARNING, OPERATIONAL).

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Unique node ID |
| `user_id` | VARCHAR(255) | NOT NULL | Owner |
| `parent_id` | UUID | FK | Parent node (hierarchy) |
| `context_id` | UUID | FK | Linked knowledge context |
| `name` | VARCHAR(255) | NOT NULL | Node name |
| `type` | nodetype | NOT NULL | BUSINESS, PROJECT, LEARNING, OPERATIONAL |
| `health` | nodehealth | DEFAULT 'NOT_STARTED' | HEALTHY, NEEDS_ATTENTION, CRITICAL, NOT_STARTED |
| `purpose` | TEXT | | Why this node exists |
| `current_status` | TEXT | | Current state |
| `this_week_focus` | JSONB | | This week's priorities |
| `decision_queue` | JSONB | | Pending decisions |
| `delegation_ready` | JSONB | | Ready to delegate |
| `is_active` | BOOLEAN | DEFAULT FALSE | Currently active? |
| `is_archived` | BOOLEAN | DEFAULT FALSE | Archived? |
| `sort_order` | INTEGER | DEFAULT 0 | Display order |
| `created_at` | TIMESTAMP | DEFAULT NOW() | Creation time |
| `updated_at` | TIMESTAMP | DEFAULT NOW() | Last update |

**Indexes:**
- PRIMARY KEY: `id`
- INDEX: `user_id` (user's nodes)

**Relations:**
- FK: parent_id → nodes.id (SET NULL)
- FK: context_id → contexts.id
- FK: ← node_projects.node_id
- FK: ← node_contexts.node_id
- FK: ← node_conversations.node_id

---

### node_projects

**Purpose:** Many-to-many: projects linked to nodes.

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `node_id` | UUID | NOT NULL, FK | Workspace node |
| `project_id` | UUID | NOT NULL, FK | Project |
| `linked_at` | TIMESTAMP | DEFAULT NOW() | When linked |
| `linked_by` | VARCHAR(255) | | Who created link |

**Composite Key:**
- PRIMARY KEY: (node_id, project_id)

**Indexes:**
- INDEX: `node_id` (projects for node)
- INDEX: `project_id` (nodes for project)

---

### node_contexts

**Purpose:** Many-to-many: contexts linked to nodes.

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `node_id` | UUID | NOT NULL, FK | Workspace node |
| `context_id` | UUID | NOT NULL, FK | Context document |
| `linked_at` | TIMESTAMP | DEFAULT NOW() | When linked |
| `linked_by` | VARCHAR(255) | | Who created link |

**Composite Key:**
- PRIMARY KEY: (node_id, context_id)

**Indexes:**
- INDEX: `node_id`
- INDEX: `context_id`

---

### node_conversations

**Purpose:** Many-to-many: conversations linked to nodes.

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `node_id` | UUID | NOT NULL, FK | Workspace node |
| `conversation_id` | UUID | NOT NULL, FK | Conversation |
| `linked_at` | TIMESTAMP | DEFAULT NOW() | When linked |
| `linked_by` | VARCHAR(255) | | Who created link |

**Composite Key:**
- PRIMARY KEY: (node_id, conversation_id)

**Indexes:**
- INDEX: `node_id`
- INDEX: `conversation_id`

---

## CRM & Clients

### clients

**Purpose:** Customers, leads, prospects, companies.

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Unique client ID |
| `user_id` | VARCHAR(255) | NOT NULL | Client owner |
| `name` | VARCHAR(255) | NOT NULL | Client/company name |
| `type` | clienttype | NOT NULL | company, individual |
| `status` | clientstatus | NOT NULL | lead, prospect, active, inactive, churned |
| `email` | VARCHAR(255) | | Primary email |
| `phone` | VARCHAR(20) | | Phone number |
| `website` | VARCHAR(255) | | Website URL |
| `industry` | VARCHAR(100) | | Industry category |
| `company_size` | VARCHAR(50) | | Employees: 1-10, 11-50, 51-200, 200+ |
| `address` | TEXT | | Street address |
| `city` | VARCHAR(100) | | City |
| `state` | VARCHAR(100) | | State/province |
| `country` | VARCHAR(100) | | Country |
| `postal_code` | VARCHAR(20) | | ZIP/postal code |
| `notes` | TEXT | | Internal notes |
| `metadata` | JSONB | | Custom fields |
| `created_at` | TIMESTAMP | DEFAULT NOW() | First contact |
| `updated_at` | TIMESTAMP | DEFAULT NOW() | Last update |

**Indexes:**
- PRIMARY KEY: `id`
- INDEX: `user_id` (client owner)
- INDEX: `status` (filter by status)

**Relations:**
- FK: ← projects.client_id

---

## Calendar & Scheduling

### calendar_events

**Purpose:** Calendar meetings, reminders, deadlines.

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Event ID |
| `user_id` | VARCHAR(255) | NOT NULL | Owner |
| `title` | VARCHAR(255) | NOT NULL | Event name |
| `description` | TEXT | | Details |
| `start_time` | TIMESTAMPTZ | NOT NULL | Event start |
| `end_time` | TIMESTAMPTZ | NOT NULL | Event end |
| `location` | VARCHAR(255) | | Meeting location/URL |
| `organizer` | VARCHAR(255) | | Meeting organizer email |
| `attendees` | JSONB | | Array of attendee emails |
| `is_all_day` | BOOLEAN | DEFAULT FALSE | All-day event? |
| `is_recurring` | BOOLEAN | DEFAULT FALSE | Recurring event? |
| `recurrence_rule` | VARCHAR(500) | | iCalendar RRULE |
| `calendar_id` | VARCHAR(255) | | External calendar ID (Google, etc.) |
| `external_id` | VARCHAR(255) | | External event ID (for sync) |
| `created_at` | TIMESTAMP | DEFAULT NOW() | Creation time |
| `updated_at` | TIMESTAMP | DEFAULT NOW() | Last update |

**Indexes:**
- PRIMARY KEY: `id`
- INDEX: `user_id` (user's events)
- INDEX: `start_time` (timeline queries)

---

## Logging & Compliance

### audit_logs

**Purpose:** SOC2 A1 compliance — immutable audit trail of all system actions.

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Unique log ID |
| `user_id` | VARCHAR(255) | | Who performed action (can be system/null) |
| `action` | VARCHAR(255) | NOT NULL | CREATE, READ, UPDATE, DELETE, etc. |
| `resource_type` | VARCHAR(100) | NOT NULL | projects, tasks, contexts, users, etc. |
| `resource_id` | VARCHAR(255) | | ID of affected resource |
| `details` | JSONB | | Action details (before/after values) |
| `ip_address` | VARCHAR(45) | | IPv4 or IPv6 |
| `user_agent` | VARCHAR(500) | | Browser/client info |
| `status` | VARCHAR(20) | | success, failure |
| `error_message` | TEXT | | If status=failure |
| `timestamp` | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | When action occurred |

**Indexes:**
- PRIMARY KEY: `id`
- INDEX: `timestamp` (chronological lookup)
- INDEX: `user_id` (user's actions)
- INDEX: `resource_type` (filter by resource)

**Constraints:**
- IMMUTABLE: No UPDATE/DELETE allowed in production (append-only log)

---

### daily_logs

**Purpose:** Daily journal, standup entries, retrospectives.

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Entry ID |
| `user_id` | VARCHAR(255) | NOT NULL | Author |
| `date` | DATE | NOT NULL | Log date |
| `title` | VARCHAR(255) | | Optional title |
| `content` | TEXT | | Entry content |
| `mood` | VARCHAR(50) | | Mood rating (optional) |
| `achievements` | JSONB | | Array of wins |
| `challenges` | JSONB | | Array of blockers |
| `notes` | JSONB | | Additional notes |
| `created_at` | TIMESTAMP | DEFAULT NOW() | Creation time |
| `updated_at` | TIMESTAMP | DEFAULT NOW() | Last edit |

**Indexes:**
- PRIMARY KEY: `id`
- INDEX: `user_id` (user's logs)
- INDEX: `date` (timeline)

---

## Authentication & Integrations

### oauth_tokens

**Purpose:** External OAuth tokens (encrypted, for Google, Slack, etc.).

**Columns:**
| Column | Type | Constraints | Purpose |
|--------|------|-----------|---------|
| `id` | UUID | PRIMARY KEY | Token ID |
| `user_id` | VARCHAR(255) | NOT NULL | Token owner |
| `service` | VARCHAR(50) | NOT NULL | google, slack, notion, hubspot, etc. |
| `access_token` | TEXT | NOT NULL | Encrypted access token |
| `refresh_token` | TEXT | | Encrypted refresh token |
| `expires_at` | TIMESTAMP | | Token expiration |
| `scope` | VARCHAR(500) | | OAuth scopes granted |
| `created_at` | TIMESTAMP | DEFAULT NOW() | When obtained |
| `updated_at` | TIMESTAMP | DEFAULT NOW() | Last refresh |

**Indexes:**
- PRIMARY KEY: `id`
- INDEX: `user_id, service` (find user's token for service)

**Security:**
- `access_token` and `refresh_token` encrypted with `TOKEN_ENCRYPTION_KEY`
- Never log token values

---

## Additional Tables

### focus_items

**Purpose:** Weekly/daily focus priorities.

**Columns:**
- `id` UUID PRIMARY KEY
- `user_id` VARCHAR(255) NOT NULL
- `week_start` DATE NOT NULL
- `priority` INT (1-5, 1=highest)
- `description` TEXT
- `status` VARCHAR(20) (pending, done)
- `created_at` TIMESTAMP DEFAULT NOW()

---

### user_settings

**Purpose:** User preferences, notifications, display options.

**Columns:**
- `id` UUID PRIMARY KEY
- `user_id` VARCHAR(255) NOT NULL UNIQUE
- `theme` VARCHAR(20) (light, dark, auto)
- `notifications_enabled` BOOLEAN DEFAULT true
- `email_digest_frequency` VARCHAR(20) (daily, weekly, never)
- `timezone` VARCHAR(50)
- `language` VARCHAR(10)
- `metadata` JSONB
- `updated_at` TIMESTAMP DEFAULT NOW()

---

## Enum Types

### messagerole
```sql
CREATE TYPE messagerole AS ENUM ('USER', 'ASSISTANT', 'SYSTEM', 'user', 'assistant', 'system');
```

### artifacttype
```sql
CREATE TYPE artifacttype AS ENUM ('CODE', 'DOCUMENT', 'MARKDOWN', 'REACT', 'HTML', 'SVG');
```

### contexttype
```sql
CREATE TYPE contexttype AS ENUM ('PERSON', 'BUSINESS', 'PROJECT', 'CUSTOM', 'document', 'DOCUMENT');
```

### projectstatus
```sql
CREATE TYPE projectstatus AS ENUM ('ACTIVE', 'PAUSED', 'COMPLETED', 'ARCHIVED');
```

### projectpriority
```sql
CREATE TYPE projectpriority AS ENUM ('CRITICAL', 'HIGH', 'MEDIUM', 'LOW');
```

### nodetype
```sql
CREATE TYPE nodetype AS ENUM ('BUSINESS', 'PROJECT', 'LEARNING', 'OPERATIONAL');
```

### nodehealth
```sql
CREATE TYPE nodehealth AS ENUM ('HEALTHY', 'NEEDS_ATTENTION', 'CRITICAL', 'NOT_STARTED');
```

### taskstatus
```sql
CREATE TYPE taskstatus AS ENUM ('todo', 'in_progress', 'done', 'cancelled');
```

### taskpriority
```sql
CREATE TYPE taskpriority AS ENUM ('critical', 'high', 'medium', 'low');
```

### memberstatus
```sql
CREATE TYPE memberstatus AS ENUM ('AVAILABLE', 'BUSY', 'OVERLOADED', 'OOO');
```

### clienttype
```sql
CREATE TYPE clienttype AS ENUM ('company', 'individual');
```

### clientstatus
```sql
CREATE TYPE clientstatus AS ENUM ('lead', 'prospect', 'active', 'inactive', 'churned');
```

### interactiontype
```sql
CREATE TYPE interactiontype AS ENUM ('call', 'email', 'meeting', 'note');
```

---

## Relationships Summary

**User (center):**
- user → conversations, projects, tasks, contexts, team_members, clients, nodes, artifacts, audit_logs, daily_logs, calendar_events, oauth_tokens, user_settings

**Projects (hub):**
- projects ← tasks, project_conversations, node_projects, artifacts
- projects → clients, contexts (optional)

**Conversations (hub):**
- conversations → messages, artifacts, contexts (optional)
- conversations ← project_conversations, node_conversations

**Nodes (hierarchy):**
- nodes → projects (M-M via node_projects)
- nodes → contexts (M-M via node_contexts)
- nodes → conversations (M-M via node_conversations)
- nodes → parent nodes (self-referential)

**Contexts (knowledge):**
- contexts ← artifacts, conversations, nodes
- contexts → parent contexts (self-referential)

---

## See Also

- API endpoints reference
- Configuration options reference
- How-to: Query the database
- Migrations: `/migrations/` directory
