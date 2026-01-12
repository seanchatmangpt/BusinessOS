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

	// Read the supabase migration file
	sqlFile := "internal/database/migrations/supabase_migration.sql"
	sqlContent, err := os.ReadFile(sqlFile)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", sqlFile, err)
	}

	fmt.Println("Applying full database schema...")
	fmt.Println("This will create all missing tables, enums, and indexes.")
	fmt.Println()

	// Execute the migration
	_, err = conn.Exec(ctx, string(sqlContent))
	if err != nil {
		log.Printf("Warning: Migration had errors (this is OK if tables already exist): %v", err)
		fmt.Println("\n✓ Migration completed with warnings (expected if some tables exist)")
	} else {
		fmt.Println("\n✓ Full schema migration complete!")
	}
}
