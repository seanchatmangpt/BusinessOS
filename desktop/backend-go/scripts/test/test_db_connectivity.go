package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No .env file found, using environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("❌ DATABASE_URL not set")
	}

	fmt.Println("🧪 Testing Database Connectivity...\n")

	// Connect
	fmt.Print("1. Connecting to database... ")
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		fmt.Printf("❌\nError: %v\n", err)
		log.Fatal("Failed to create connection pool")
	}
	defer pool.Close()
	fmt.Println("✅")

	// Test basic connectivity
	fmt.Print("2. Testing basic query... ")
	var count int
	err = pool.QueryRow(context.Background(), "SELECT 1").Scan(&count)
	if err != nil {
		fmt.Printf("❌\nError: %v\n", err)
		log.Fatal("Failed to execute query")
	}
	fmt.Println("✅")

	// Test user table
	fmt.Print("3. Checking user table... ")
	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM \"user\"").Scan(&count)
	if err != nil {
		fmt.Printf("❌\nError: %v\n", err)
		log.Fatal("Failed to query user table")
	}
	fmt.Printf("✅ (found %d users)\n", count)

	// Test voice-specific tables
	fmt.Println("\n4. Checking voice-related tables...")
	tables := []string{
		"workspace_members",
		"user_workspace_profiles",
		"workspaces",
		"agent_v2",
		"embeddings",
	}

	for _, table := range tables {
		fmt.Printf("   - Checking '%s'... ", table)
		var exists bool
		err = pool.QueryRow(
			context.Background(),
			"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)",
			table,
		).Scan(&exists)

		if err != nil {
			fmt.Printf("❌ (error: %v)\n", err)
			continue
		}

		if exists {
			// Get row count
			var rowCount int
			err = pool.QueryRow(
				context.Background(),
				fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", table),
			).Scan(&rowCount)
			if err != nil {
				fmt.Printf("✅ (exists, %d rows)\n", rowCount)
			} else {
				fmt.Printf("✅ (exists, %d rows)\n", rowCount)
			}
		} else {
			fmt.Println("⚠️  (not found)")
		}
	}

	// Test connection pool stats
	fmt.Println("\n5. Connection Pool Stats:")
	stats := pool.Stat()
	fmt.Printf("   - Acquired connections: %d\n", stats.AcquiredConns())
	fmt.Printf("   - Idle connections: %d\n", stats.IdleConns())
	fmt.Printf("   - Total connections: %d\n", stats.TotalConns())

	fmt.Println("\n✅ Database connectivity test completed successfully!")
}
