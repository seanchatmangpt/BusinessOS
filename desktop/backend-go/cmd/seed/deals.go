package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ============================================================================
// Deal Seeding — Financial Instruments for Fortune 5 Compliance
// ============================================================================
// Seeds 10 realistic financial deals across domains: equity, fixed income,
// derivatives, commodities, structured products. Mix of statuses from draft
// to closed. Deterministic internal_reference (SEED-DEAL-NNN) for cleanup.
// ============================================================================

var dealIDs = []uuid.UUID{
	uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
	uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
	uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
	uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"),
	uuid.MustParse("550e8400-e29b-41d4-a716-446655440005"),
	uuid.MustParse("550e8400-e29b-41d4-a716-446655440006"),
	uuid.MustParse("550e8400-e29b-41d4-a716-446655440007"),
	uuid.MustParse("550e8400-e29b-41d4-a716-446655440008"),
	uuid.MustParse("550e8400-e29b-41d4-a716-446655440009"),
	uuid.MustParse("550e8400-e29b-41d4-a716-446655440010"),
}

type Deal struct {
	ID                uuid.UUID
	Name              string
	AmountCents       int64  // Always use cents to avoid floating-point issues
	Currency          string // ISO 4217 code
	Status            string // draft, proposed, negotiating, approved, executed, settled, closed, rejected
	Domain            string // equity, fixed_income, derivatives, commodities, fx, structured, other
	CreatedBy         string // user ID
	DealDate          *string
	SettlementDate    *string
	MaturityDate      *string
	Description       string
	RiskRating        *string // AAA, AA, A, BBB, BB, B, CCC, CC, C, D
	InternalReference string
	CreatedAt         time.Time
}

func seedDeals(ctx context.Context, pool *pgxpool.Pool, userID string) {
	// Time offsets for realistic distribution
	now := time.Now()

	dateStr := func(offset time.Duration) *string {
		d := now.Add(offset).Format("2006-01-02")
		return &d
	}

	deals := []Deal{
		// Equity deals (3)
		{
			dealIDs[0],
			"TechCorp Series B Round",
			5000000000, // $50M in cents
			"USD",
			"executed",
			"equity",
			userID,
			dateStr(-25 * 24 * time.Hour),
			dateStr(-15 * 24 * time.Hour),
			nil,
			"Venture capital Series B funding for SaaS platform",
			ptrString("BB"),
			"SEED-DEAL-001",
			now.Add(-25 * 24 * time.Hour),
		},
		{
			dealIDs[1],
			"Healthcare Merger Acquisition",
			12500000000, // $125M in cents
			"USD",
			"negotiating",
			"equity",
			userID,
			dateStr(-20 * 24 * time.Hour),
			nil,
			nil,
			"Strategic acquisition of regional healthcare provider",
			ptrString("A"),
			"SEED-DEAL-002",
			now.Add(-20 * 24 * time.Hour),
		},
		{
			dealIDs[2],
			"GreenEnergy IPO Preparation",
			20000000000, // 200M EUR in cents
			"EUR",
			"approved",
			"equity",
			userID,
			dateStr(-15 * 24 * time.Hour),
			dateStr(-5 * 24 * time.Hour),
			nil,
			"Initial Public Offering preparation for renewable energy company",
			ptrString("AA"),
			"SEED-DEAL-003",
			now.Add(-15 * 24 * time.Hour),
		},
		// Fixed Income deals (3)
		{
			dealIDs[3],
			"Corporate Bond Issuance 5Y",
			7500000000, // 75M GBP in cents
			"GBP",
			"settled",
			"fixed_income",
			userID,
			dateStr(-30 * 24 * time.Hour),
			dateStr(-28 * 24 * time.Hour),
			dateStr(4 * 365 * 24 * time.Hour), // 5-year maturity
			"Senior unsecured 5-year corporate bond at 3.5%",
			ptrString("BBB"),
			"SEED-DEAL-004",
			now.Add(-30 * 24 * time.Hour),
		},
		{
			dealIDs[4],
			"Government Debt Restructuring",
			30000000000, // 300M USD in cents
			"USD",
			"executed",
			"fixed_income",
			userID,
			dateStr(-18 * 24 * time.Hour),
			dateStr(-12 * 24 * time.Hour),
			dateStr(6 * 365 * 24 * time.Hour), // 10-year maturity
			"Sovereign debt restructuring with 10-year maturity",
			ptrString("AA"),
			"SEED-DEAL-005",
			now.Add(-18 * 24 * time.Hour),
		},
		{
			dealIDs[5],
			"Municipal Bond Portfolio",
			5000000000, // 50M USD in cents
			"USD",
			"draft",
			"fixed_income",
			userID,
			nil,
			nil,
			nil,
			"Tax-exempt municipal bonds for infrastructure",
			ptrString("A"),
			"SEED-DEAL-006",
			now.Add(-5 * 24 * time.Hour),
		},
		// Derivatives (2)
		{
			dealIDs[6],
			"Currency Swap EUR/GBP",
			15000000000, // 150M EUR in cents
			"EUR",
			"executed",
			"derivatives",
			userID,
			dateStr(-10 * 24 * time.Hour),
			dateStr(-8 * 24 * time.Hour),
			dateStr(365 * 24 * time.Hour), // 1-year maturity
			"Interest rate and currency swap hedge",
			ptrString("A"),
			"SEED-DEAL-007",
			now.Add(-10 * 24 * time.Hour),
		},
		{
			dealIDs[7],
			"Equity Index Options Strategy",
			2500000000, // 25M USD in cents
			"USD",
			"proposed",
			"derivatives",
			userID,
			dateStr(-3 * 24 * time.Hour),
			nil,
			dateStr(91 * 24 * time.Hour), // 3-month maturity
			"S&P 500 structured note with leverage",
			ptrString("CCC"),
			"SEED-DEAL-008",
			now.Add(-3 * 24 * time.Hour),
		},
		// Commodities + Other (2)
		{
			dealIDs[8],
			"Crude Oil Futures Contract",
			8000000000, // 80M USD in cents
			"USD",
			"executed",
			"commodities",
			userID,
			dateStr(-8 * 24 * time.Hour),
			dateStr(-6 * 24 * time.Hour),
			dateStr(273 * 24 * time.Hour), // ~9-month maturity
			"WTI crude oil hedging for Q2-Q4 2026",
			ptrString("BB"),
			"SEED-DEAL-009",
			now.Add(-8 * 24 * time.Hour),
		},
		{
			dealIDs[9],
			"Real Estate Development Partnership",
			11000000000, // 110M USD in cents
			"USD",
			"closed",
			"other",
			userID,
			dateStr(-120 * 24 * time.Hour),
			dateStr(-100 * 24 * time.Hour),
			nil,
			"Commercial real estate syndication - completed",
			ptrString("B"),
			"SEED-DEAL-010",
			now.Add(-120 * 24 * time.Hour),
		},
	}

	for i, d := range deals {
		_, err := pool.Exec(ctx, `
			INSERT INTO deals (
				id, name, amount_cents, currency, status, domain,
				created_by, deal_date, settlement_date, maturity_date,
				description, risk_rating, internal_reference, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			ON CONFLICT (id) DO NOTHING`,
			d.ID, d.Name, d.AmountCents, d.Currency, d.Status, d.Domain,
			d.CreatedBy, d.DealDate, d.SettlementDate, d.MaturityDate,
			d.Description, d.RiskRating, d.InternalReference, d.CreatedAt,
		)
		if err != nil {
			log.Printf("  deal %d (%s): %v", i+1, d.Name, err)
		} else {
			fmt.Printf("  + %s (%s, %s %s)\n", d.Name, d.Status, formatAmount(d.AmountCents), d.Currency)
		}
	}

	fmt.Printf("Seeded %d deals\n", len(deals))
}

func ptrString(s string) *string {
	return &s
}

func formatAmount(cents int64) string {
	whole := cents / 100
	frac := cents % 100
	return fmt.Sprintf("$%d.%02d", whole, frac)
}
