package main

import (
	"context"
	"fmt"

	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/database"
)

func main() {
	ctx := context.Background()

	cfg, _ := config.Load()
	pool, _ := database.Connect(cfg)
	defer pool.Close()

	var totalJobs, completedJobs, runningJobs, failedJobs, pendingJobs int

	pool.QueryRow(ctx, `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'completed') as completed,
			COUNT(*) FILTER (WHERE status = 'running') as running,
			COUNT(*) FILTER (WHERE status = 'failed') as failed,
			COUNT(*) FILTER (WHERE status = 'pending') as pending
		FROM background_jobs
		WHERE created_at >= NOW() - INTERVAL '10 minutes'
	`).Scan(&totalJobs, &completedJobs, &runningJobs, &failedJobs, &pendingJobs)

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           FINAL JOB STATUS VERIFICATION                     ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Total Jobs:      %3d                                        ║\n", totalJobs)
	fmt.Printf("║  ✅ Completed:     %3d                                        ║\n", completedJobs)
	fmt.Printf("║  🔄 Running:       %3d                                        ║\n", runningJobs)
	fmt.Printf("║  ⏳ Pending:       %3d                                        ║\n", pendingJobs)
	fmt.Printf("║  ❌ Failed:        %3d                                        ║\n", failedJobs)
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")

	if completedJobs > 0 {
		rate := float64(completedJobs) / float64(totalJobs) * 100
		fmt.Printf("║  Success Rate:    %.1f%%                                    ║\n", rate)
	}

	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	rows, _ := pool.Query(ctx, `
		SELECT job_type, status, COUNT(*) as count
		FROM background_jobs
		WHERE created_at >= NOW() - INTERVAL '10 minutes'
		GROUP BY job_type, status
		ORDER BY job_type, status
	`)
	defer rows.Close()

	fmt.Println("Jobs by Type and Status:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	for rows.Next() {
		var jobType, status string
		var count int
		rows.Scan(&jobType, &status, &count)

		icon := "⏳"
		if status == "completed" {
			icon = "✅"
		} else if status == "running" {
			icon = "🔄"
		} else if status == "failed" {
			icon = "❌"
		}

		fmt.Printf("%s %-30s %-12s: %d\n", icon, jobType, status, count)
	}
}
