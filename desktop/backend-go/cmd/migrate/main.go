package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close(ctx)

	// Add share_calendar column
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'team_members' AND column_name = 'share_calendar'
			) THEN
				ALTER TABLE team_members ADD COLUMN share_calendar BOOLEAN DEFAULT FALSE;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding share_calendar: %v", err)
	} else {
		fmt.Println("✓ share_calendar column OK")
	}

	// Add calendar_user_id column
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'team_members' AND column_name = 'calendar_user_id'
			) THEN
				ALTER TABLE team_members ADD COLUMN calendar_user_id VARCHAR(255);
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding calendar_user_id: %v", err)
	} else {
		fmt.Println("✓ calendar_user_id column OK")
	}

	// Add client_id column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'client_id'
			) THEN
				ALTER TABLE projects ADD COLUMN client_id UUID REFERENCES clients(id) ON DELETE SET NULL;
				CREATE INDEX IF NOT EXISTS idx_projects_client ON projects(client_id);
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding client_id: %v", err)
	} else {
		fmt.Println("✓ client_id column OK")
	}

	// Add start_date column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'start_date'
			) THEN
				ALTER TABLE projects ADD COLUMN start_date TIMESTAMPTZ;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding start_date: %v", err)
	} else {
		fmt.Println("✓ start_date column OK")
	}

	// Add end_date column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'end_date'
			) THEN
				ALTER TABLE projects ADD COLUMN end_date TIMESTAMPTZ;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding end_date: %v", err)
	} else {
		fmt.Println("✓ end_date column OK")
	}

	// Add budget column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'budget'
			) THEN
				ALTER TABLE projects ADD COLUMN budget DECIMAL(12,2);
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding budget: %v", err)
	} else {
		fmt.Println("✓ budget column OK")
	}

	// Add due_date column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'due_date'
			) THEN
				ALTER TABLE projects ADD COLUMN due_date TIMESTAMPTZ;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding due_date: %v", err)
	} else {
		fmt.Println("✓ due_date column OK")
	}

	// Add completed_at column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'completed_at'
			) THEN
				ALTER TABLE projects ADD COLUMN completed_at TIMESTAMPTZ;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding completed_at: %v", err)
	} else {
		fmt.Println("✓ completed_at column OK")
	}

	// Add visibility column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'visibility'
			) THEN
				ALTER TABLE projects ADD COLUMN visibility VARCHAR(20) DEFAULT 'private';
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding visibility: %v", err)
	} else {
		fmt.Println("✓ visibility column OK")
	}

	// Add tags column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'tags'
			) THEN
				ALTER TABLE projects ADD COLUMN tags TEXT[];
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding tags: %v", err)
	} else {
		fmt.Println("✓ tags column OK")
	}

	// Add archived column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'archived'
			) THEN
				ALTER TABLE projects ADD COLUMN archived BOOLEAN DEFAULT FALSE;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding archived: %v", err)
	} else {
		fmt.Println("✓ archived column OK")
	}

	// Add progress column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'progress'
			) THEN
				ALTER TABLE projects ADD COLUMN progress INTEGER DEFAULT 0;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding progress: %v", err)
	} else {
		fmt.Println("✓ progress column OK")
	}

	// Add owner_id column to projects (no FK since users table may not exist)
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'owner_id'
			) THEN
				ALTER TABLE projects ADD COLUMN owner_id VARCHAR(255);
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding owner_id: %v", err)
	} else {
		fmt.Println("✓ owner_id column OK")
	}

	// ===== CHAIN OF THOUGHT (COT) THINKING SYSTEM =====

	// Create thinkingtype enum
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'thinkingtype') THEN
				CREATE TYPE thinkingtype AS ENUM ('analysis', 'planning', 'reflection', 'tool_use', 'reasoning', 'evaluation');
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning creating thinkingtype enum: %v", err)
	} else {
		fmt.Println("✓ thinkingtype enum OK")
	}

	// Create reasoning_templates table (must be created before thinking_traces due to FK)
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS reasoning_templates (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			system_prompt TEXT,
			thinking_instruction TEXT,
			output_format VARCHAR(50) DEFAULT 'streaming',
			show_thinking BOOLEAN DEFAULT true,
			save_thinking BOOLEAN DEFAULT true,
			max_thinking_tokens INT DEFAULT 4096,
			times_used INT DEFAULT 0,
			is_default BOOLEAN DEFAULT false,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_reasoning_templates_user ON reasoning_templates(user_id);
	`)
	if err != nil {
		log.Printf("Warning creating reasoning_templates table: %v", err)
	} else {
		fmt.Println("✓ reasoning_templates table OK")
	}

	// Create thinking_traces table
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS thinking_traces (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id VARCHAR(255) NOT NULL,
			conversation_id UUID REFERENCES conversations(id) ON DELETE CASCADE,
			message_id UUID REFERENCES messages(id) ON DELETE CASCADE,
			thinking_content TEXT NOT NULL,
			thinking_type thinkingtype DEFAULT 'reasoning',
			step_number INT DEFAULT 1,
			started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			completed_at TIMESTAMP WITH TIME ZONE,
			duration_ms INT,
			thinking_tokens INT DEFAULT 0,
			model_used VARCHAR(100),
			reasoning_template_id UUID REFERENCES reasoning_templates(id) ON DELETE SET NULL,
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_thinking_traces_user ON thinking_traces(user_id);
		CREATE INDEX IF NOT EXISTS idx_thinking_traces_conversation ON thinking_traces(conversation_id);
		CREATE INDEX IF NOT EXISTS idx_thinking_traces_message ON thinking_traces(message_id);
	`)
	if err != nil {
		log.Printf("Warning creating thinking_traces table: %v", err)
	} else {
		fmt.Println("✓ thinking_traces table OK")
	}

	// Add thinking settings columns to user_settings
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'user_settings' AND column_name = 'thinking_enabled'
			) THEN
				ALTER TABLE user_settings ADD COLUMN thinking_enabled BOOLEAN DEFAULT false;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding thinking_enabled: %v", err)
	} else {
		fmt.Println("✓ thinking_enabled column OK")
	}

	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'user_settings' AND column_name = 'thinking_show_in_ui'
			) THEN
				ALTER TABLE user_settings ADD COLUMN thinking_show_in_ui BOOLEAN DEFAULT true;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding thinking_show_in_ui: %v", err)
	} else {
		fmt.Println("✓ thinking_show_in_ui column OK")
	}

	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'user_settings' AND column_name = 'thinking_save_traces'
			) THEN
				ALTER TABLE user_settings ADD COLUMN thinking_save_traces BOOLEAN DEFAULT true;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding thinking_save_traces: %v", err)
	} else {
		fmt.Println("✓ thinking_save_traces column OK")
	}

	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'user_settings' AND column_name = 'thinking_default_template_id'
			) THEN
				ALTER TABLE user_settings ADD COLUMN thinking_default_template_id UUID REFERENCES reasoning_templates(id) ON DELETE SET NULL;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding thinking_default_template_id: %v", err)
	} else {
		fmt.Println("✓ thinking_default_template_id column OK")
	}

	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'user_settings' AND column_name = 'thinking_max_tokens'
			) THEN
				ALTER TABLE user_settings ADD COLUMN thinking_max_tokens INT DEFAULT 4096;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding thinking_max_tokens: %v", err)
	} else {
		fmt.Println("✓ thinking_max_tokens column OK")
	}

	// ===== CUSTOM AGENTS SYSTEM (Migration 009) =====

	// Create custom_agents table
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS custom_agents (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id VARCHAR(255) NOT NULL,
			name VARCHAR(50) NOT NULL,
			display_name VARCHAR(100) NOT NULL,
			description TEXT,
			avatar VARCHAR(50),
			system_prompt TEXT NOT NULL,
			model_preference VARCHAR(100),
			temperature DECIMAL(3,2) DEFAULT 0.7,
			max_tokens INTEGER DEFAULT 4096,
			capabilities TEXT[] DEFAULT '{}',
			tools_enabled TEXT[] DEFAULT '{}',
			context_sources TEXT[] DEFAULT '{}',
			thinking_enabled BOOLEAN DEFAULT FALSE,
			streaming_enabled BOOLEAN DEFAULT TRUE,
			category VARCHAR(50) DEFAULT 'general',
			is_public BOOLEAN DEFAULT FALSE,
			is_active BOOLEAN DEFAULT TRUE,
			times_used INTEGER DEFAULT 0,
			last_used_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(user_id, name)
		);
		CREATE INDEX IF NOT EXISTS idx_custom_agents_user_id ON custom_agents(user_id);
		CREATE INDEX IF NOT EXISTS idx_custom_agents_name ON custom_agents(user_id, name);
		CREATE INDEX IF NOT EXISTS idx_custom_agents_category ON custom_agents(category);
		CREATE INDEX IF NOT EXISTS idx_custom_agents_active ON custom_agents(user_id, is_active);
	`)
	if err != nil {
		log.Printf("Warning creating custom_agents table: %v", err)
	} else {
		fmt.Println("✓ custom_agents table OK")
	}

	// Create agent_presets table
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS agent_presets (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(50) NOT NULL UNIQUE,
			display_name VARCHAR(100) NOT NULL,
			description TEXT,
			avatar VARCHAR(50),
			system_prompt TEXT NOT NULL,
			model_preference VARCHAR(100),
			temperature DECIMAL(3,2) DEFAULT 0.7,
			max_tokens INTEGER DEFAULT 4096,
			capabilities TEXT[] DEFAULT '{}',
			tools_enabled TEXT[] DEFAULT '{}',
			context_sources TEXT[] DEFAULT '{}',
			thinking_enabled BOOLEAN DEFAULT FALSE,
			category VARCHAR(50) DEFAULT 'general',
			times_copied INTEGER DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`)
	if err != nil {
		log.Printf("Warning creating agent_presets table: %v", err)
	} else {
		fmt.Println("✓ agent_presets table OK")
	}

	// Insert default agent presets
	_, err = conn.Exec(ctx, `
		INSERT INTO agent_presets (name, display_name, description, avatar, system_prompt, category, capabilities, tools_enabled, thinking_enabled)
		VALUES
			('code-reviewer', 'Code Reviewer', 'Reviews code for bugs, best practices, and improvements', 'magnifying-glass',
			 'You are an expert code reviewer. Analyze code for:
1. **Bugs & Errors**: Identify potential bugs, edge cases, and runtime errors
2. **Best Practices**: Check adherence to coding standards and conventions
3. **Performance**: Spot inefficiencies and suggest optimizations
4. **Security**: Flag potential security vulnerabilities
5. **Maintainability**: Assess code readability and suggest improvements

Provide specific, actionable feedback with code examples when suggesting changes.',
			 'coding', ARRAY['code_review', 'analysis'], ARRAY['read_file', 'search_code'], TRUE),

			('technical-writer', 'Technical Writer', 'Creates clear documentation and technical content', 'pencil',
			 'You are an expert technical writer. Create clear, well-structured documentation that:
1. Uses simple, precise language
2. Includes relevant code examples
3. Follows standard documentation patterns
4. Anticipates reader questions
5. Provides both quick-start guides and detailed references

Adapt your writing style to the audience - from beginner-friendly tutorials to expert reference docs.',
			 'writing', ARRAY['documentation', 'writing'], ARRAY[]::TEXT[], FALSE),

			('data-analyst', 'Data Analyst', 'Analyzes data and creates insights', 'chart',
			 'You are an expert data analyst. When analyzing data:
1. Start with exploratory analysis to understand the data
2. Identify key patterns, trends, and anomalies
3. Use appropriate statistical methods
4. Create clear visualizations (describe them in detail)
5. Provide actionable insights and recommendations

Be precise with numbers and transparent about limitations or assumptions.',
			 'analysis', ARRAY['data_analysis', 'visualization'], ARRAY[]::TEXT[], TRUE),

			('business-strategist', 'Business Strategist', 'Provides strategic business advice and analysis', 'briefcase',
			 'You are a senior business strategist. Provide strategic advice by:
1. Understanding the business context and objectives
2. Analyzing market conditions and competition
3. Identifying opportunities and risks
4. Developing actionable recommendations
5. Considering implementation feasibility

Use frameworks like SWOT, Porter''s Five Forces, and business model canvas when appropriate.',
			 'business', ARRAY['strategy', 'analysis', 'planning'], ARRAY[]::TEXT[], TRUE),

			('creative-writer', 'Creative Writer', 'Helps with creative writing and content creation', 'sparkles',
			 'You are a talented creative writer. Help with:
1. Generating creative ideas and concepts
2. Writing engaging narratives and copy
3. Developing compelling characters and stories
4. Crafting persuasive marketing content
5. Editing and improving existing content

Match the desired tone, style, and voice. Be creative while staying on-brand.',
			 'writing', ARRAY['creative_writing', 'content_creation'], ARRAY[]::TEXT[], FALSE)
		ON CONFLICT (name) DO NOTHING;
	`)
	if err != nil {
		log.Printf("Warning inserting agent presets: %v", err)
	} else {
		fmt.Println("✓ agent_presets data OK")
	}

	fmt.Println("Migration complete!")
}
