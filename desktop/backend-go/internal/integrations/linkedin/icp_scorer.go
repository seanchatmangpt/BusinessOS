package linkedin

import (
	"log/slog"
	"strings"
)

// ICPScorer implements ICP scoring: title (50%) + industry (30%) + company (20%).
type ICPScorer struct {
	logger *slog.Logger
}

// NewICPScorer creates a new ICP scorer.
func NewICPScorer(logger *slog.Logger) *ICPScorer {
	return &ICPScorer{logger: logger}
}

// ScoreContact evaluates a contact's fit based on title, industry, and company signals.
// Returns a score from 0.0 to 1.0.
func (s *ICPScorer) ScoreContact(contact *Contact) float64 {
	breakdown := map[string]float64{
		"title":    s.scoreTitle(contact.Title),
		"industry": s.scoreIndustry(contact.Industry),
		"company":  s.scoreCompany(contact.Company),
	}

	// Weighted average: title 50% + industry 30% + company 20%
	score := (breakdown["title"] * 0.50) + (breakdown["industry"] * 0.30) + (breakdown["company"] * 0.20)

	s.logger.Debug("ICP scored contact",
		"name", contact.Name,
		"title", contact.Title,
		"company", contact.Company,
		"title_score", breakdown["title"],
		"industry_score", breakdown["industry"],
		"company_score", breakdown["company"],
		"final_score", score,
	)

	return score
}

// scoreTitle evaluates job title keywords. Target: VP, Director, Manager, C-suite roles.
func (s *ICPScorer) scoreTitle(title string) float64 {
	if title == "" {
		return 0.0
	}

	title = strings.ToLower(title)

	// High-value keywords (score 1.0)
	highValue := []string{"ceo", "cto", "cfo", "coo", "vp ", "chief", "president", "founder"}
	for _, keyword := range highValue {
		if strings.Contains(title, keyword) {
			return 1.0
		}
	}

	// Medium-value keywords (score 0.7)
	mediumValue := []string{"director", "manager", "head of", "lead", "senior", "principal"}
	for _, keyword := range mediumValue {
		if strings.Contains(title, keyword) {
			return 0.7
		}
	}

	// Low-value keywords (score 0.4)
	lowValue := []string{"engineer", "analyst", "coordinator", "specialist", "consultant"}
	for _, keyword := range lowValue {
		if strings.Contains(title, keyword) {
			return 0.4
		}
	}

	return 0.2 // Default: neutral score
}

// scoreIndustry evaluates industry fit. Target: Tech, Finance, Enterprise, Healthcare.
func (s *ICPScorer) scoreIndustry(industry string) float64 {
	if industry == "" {
		return 0.0
	}

	industry = strings.ToLower(industry)

	// High-fit industries (score 1.0) — enterprise workflow/automation focus
	highFit := []string{"software", "technology", "fintech", "financial services", "banking",
		"insurance", "enterprise software", "saas", "cloud", "consulting", "healthcare",
		"manufacturing", "logistics", "energy", "government"}
	for _, ind := range highFit {
		if strings.Contains(industry, ind) {
			return 1.0
		}
	}

	// Medium-fit industries (score 0.6)
	mediumFit := []string{"retail", "e-commerce", "real estate", "automotive", "pharma",
		"telecommunications", "media", "entertainment", "hospitality"}
	for _, ind := range mediumFit {
		if strings.Contains(industry, ind) {
			return 0.6
		}
	}

	// Low-fit industries (score 0.3)
	lowFit := []string{"agriculture", "construction", "non-profit", "education"}
	for _, ind := range lowFit {
		if strings.Contains(industry, ind) {
			return 0.3
		}
	}

	return 0.4 // Default: neutral
}

// scoreCompany evaluates company size/reputation proxy via name signals.
// Look for Fortune 500 keywords or Series A/B/C signals.
func (s *ICPScorer) scoreCompany(company string) float64 {
	if company == "" {
		return 0.0
	}

	company = strings.ToLower(company)

	// Known Fortune 500 companies (score 1.0)
	fortune500 := []string{
		"apple", "microsoft", "google", "amazon", "oracle", "salesforce",
		"ibm", "intel", "cisco", "vmware", "jpmorgan", "bofa", "citi",
		"goldmansachs", "merrill", "morgan stanley", "hsbc", "ubs",
		"wells fargo", "american express", "capital one", "discover",
		"walmart", "costco", "target", "target", "home depot",
		"unilever", "procter gamble", "nestlé", "coca-cola", "pepsi",
		"johnson", "pfizer", "merck", "gilead", "moderna",
		"lockheed", "boeing", "northrop", "raytheon",
	}
	for _, f500 := range fortune500 {
		if strings.Contains(company, f500) {
			return 1.0
		}
	}

	// Mid-market / Series B+ (score 0.7)
	if strings.Contains(company, "inc") || strings.Contains(company, "corp") {
		return 0.7
	}

	// Early-stage / startup (score 0.4)
	if strings.Contains(company, "startup") || strings.Contains(company, "labs") {
		return 0.4
	}

	return 0.5 // Default: neutral
}
