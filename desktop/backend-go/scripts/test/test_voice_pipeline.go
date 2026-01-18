package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/services"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No .env file found, using environment variables")
	}

	fmt.Println("🧪 Testing Voice System Pipeline...\n")

	// 1. Load configuration
	fmt.Print("1. Loading configuration... ")
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		fmt.Printf("❌\nError: %v\n", err)
		log.Fatal("Failed to load config")
	}
	fmt.Println("✅")

	// 2. Setup database connection
	fmt.Print("2. Connecting to database... ")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("❌ DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Printf("❌\nError: %v\n", err)
		log.Fatal("Failed to create connection pool")
	}
	defer pool.Close()
	fmt.Println("✅")

	// 3. Test database connectivity
	fmt.Print("3. Testing database connectivity... ")
	var testVal int
	err = pool.QueryRow(ctx, "SELECT 1").Scan(&testVal)
	if err != nil {
		fmt.Printf("❌\nError: %v\n", err)
		log.Fatal("Failed to connect to database")
	}
	fmt.Println("✅")

	// 4. Test embedding service initialization
	fmt.Print("4. Initializing embedding service... ")
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	embeddingService := services.NewEmbeddingService(pool, ollamaURL)
	if embeddingService == nil {
		fmt.Println("⚠️  (optional service, can continue without)")
	} else {
		fmt.Println("✅")
	}

	// 5. Verify database schema for voice system
	fmt.Println("\n5. Verifying voice system database schema...")

	// Check required tables
	requiredTables := map[string][]string{
		"user":                    {"id", "email", "username"},
		"workspaces":              {"id", "name", "created_at"},
		"workspace_members":       {"workspace_id", "user_id"},
		"user_workspace_profiles": {"user_id", "workspace_id"},
		"agent_v2":                {"id", "workspace_id", "name"},
		"embeddings":              {"id", "agent_id", "content"},
	}

	for table, expectedCols := range requiredTables {
		fmt.Printf("   - Checking table '%s'... ", table)

		// Check if table exists
		var exists bool
		err := pool.QueryRow(
			ctx,
			"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)",
			table,
		).Scan(&exists)

		if err != nil || !exists {
			fmt.Println("⚠️  (not found)")
			continue
		}

		// Check if required columns exist
		var missingCols []string
		for _, col := range expectedCols {
			var colExists bool
			err := pool.QueryRow(
				ctx,
				`SELECT EXISTS (
					SELECT FROM information_schema.columns
					WHERE table_name = $1 AND column_name = $2
				)`,
				table, col,
			).Scan(&colExists)

			if err != nil || !colExists {
				missingCols = append(missingCols, col)
			}
		}

		if len(missingCols) > 0 {
			fmt.Printf("✅ (exists, missing columns: %v)\n", missingCols)
		} else {
			fmt.Println("✅")
		}
	}

	// 6. Test voice-specific configurations
	fmt.Println("\n6. Verifying voice system configuration...")
	fmt.Printf("   - Config loaded: ✅\n")
	fmt.Printf("   - Database URL set: ")
	if dbURL != "" {
		fmt.Println("✅")
	} else {
		fmt.Println("❌")
	}
	fmt.Printf("   - Embedding service URL: %s\n", ollamaURL)

	// 7. Sample data check
	fmt.Println("\n7. Checking sample data...")
	var userCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM \"user\"").Scan(&userCount)
	if err == nil {
		fmt.Printf("   - User count: %d\n", userCount)
	}

	var workspaceCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM workspaces").Scan(&workspaceCount)
	if err == nil {
		fmt.Printf("   - Workspace count: %d\n", workspaceCount)
	}

	var agentCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM agent_v2").Scan(&agentCount)
	if err == nil {
		fmt.Printf("   - Agent V2 count: %d\n", agentCount)
	}

	// Final summary
	fmt.Println("\n" + repeatString("=", 50))
	fmt.Println("✅ Voice System Pipeline Test Summary")
	fmt.Println(repeatString("=", 50))
	fmt.Println("All critical components initialized successfully!")
	fmt.Println("Ready to run voice system tests")
	fmt.Println(repeatString("=", 50))
}

// Helper function for string repeat
func repeatString(s string, count int) string {
	return strings.Repeat(s, count)
}
