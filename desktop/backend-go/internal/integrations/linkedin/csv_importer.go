package linkedin

import (
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"
)

// CSVImporter parses LinkedIn CSV export and imports contacts.
// Expects: firstName, lastName, email, currentTitle, companyName, industry.
type CSVImporter struct {
	logger *slog.Logger
	repo   *Repository
}

// NewCSVImporter creates a new CSV importer.
func NewCSVImporter(logger *slog.Logger, repo *Repository) *CSVImporter {
	return &CSVImporter{
		logger: logger,
		repo:   repo,
	}
}

// ImportCSV parses a CSV buffer and upserts contacts into the database.
// Returns counts of imported, updated, and failed records.
func (ci *CSVImporter) ImportCSV(csvContent string) (int, int, int, []string) {
	reader := csv.NewReader(strings.NewReader(csvContent))
	reader.FieldsPerRecord = -1 // Allow variable field counts

	var imported, updated, failed int
	var errors []string

	// Read header row
	headers, err := reader.Read()
	if err != nil {
		ci.logger.Error("CSV header read failed", "error", err)
		errors = append(errors, fmt.Sprintf("Failed to read CSV header: %v", err))
		return 0, 0, 0, errors
	}

	// Map column indices
	columnMap := mapCSVColumns(headers)
	if len(columnMap) == 0 {
		ci.logger.Error("CSV column mapping failed: no recognizable columns")
		errors = append(errors, "CSV missing required columns: firstName, lastName, email, currentTitle, companyName")
		return 0, 0, 0, errors
	}

	// Read data rows
	lineNum := 2 // Start at 2 (after header)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			ci.logger.Warn("CSV row read error", "line", lineNum, "error", err)
			errors = append(errors, fmt.Sprintf("Line %d: %v", lineNum, err))
			failed++
			lineNum++
			continue
		}

		// Extract fields
		firstName := getCSVField(record, columnMap, "firstName")
		lastName := getCSVField(record, columnMap, "lastName")
		email := getCSVField(record, columnMap, "email")
		title := getCSVField(record, columnMap, "currentTitle")
		company := getCSVField(record, columnMap, "companyName")
		industry := getCSVField(record, columnMap, "industry")

		// Validate required fields
		if firstName == "" || lastName == "" || email == "" {
			ci.logger.Warn("CSV row missing required fields", "line", lineNum, "email", email)
			errors = append(errors, fmt.Sprintf("Line %d: missing firstName, lastName, or email", lineNum))
			failed++
			lineNum++
			continue
		}

		name := strings.TrimSpace(firstName + " " + lastName)

		// Check if contact already exists by email
		existing, err := ci.repo.GetContactByEmail(email)
		isUpdate := existing != nil && err == nil

		contact := &Contact{
			LinkedInID:     generateLinkedInID(email),
			Name:           name,
			Title:          title,
			Company:        company,
			Industry:       industry,
			ConnectionDate: ptrTime(time.Now()),
			RawCSV:         strings.Join(record, ","),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if isUpdate {
			// Update existing
			existing.Name = name
			existing.Title = title
			existing.Company = company
			existing.Industry = industry
			existing.UpdatedAt = time.Now()

			if err := ci.repo.UpdateContact(existing); err != nil {
				ci.logger.Error("Contact update failed", "email", email, "error", err)
				errors = append(errors, fmt.Sprintf("Line %d: update failed: %v", lineNum, err))
				failed++
			} else {
				updated++
			}
		} else {
			// Insert new
			if err := ci.repo.CreateContact(contact); err != nil {
				ci.logger.Error("Contact creation failed", "email", email, "error", err)
				errors = append(errors, fmt.Sprintf("Line %d: insert failed: %v", lineNum, err))
				failed++
			} else {
				imported++
			}
		}

		lineNum++
	}

	ci.logger.Info("CSV import completed",
		"imported", imported,
		"updated", updated,
		"failed", failed,
	)

	return imported, updated, failed, errors
}

// mapCSVColumns maps header row to column indices.
// Returns a map of column name -> index.
func mapCSVColumns(headers []string) map[string]int {
	columnMap := make(map[string]int)

	for i, header := range headers {
		headerLower := strings.ToLower(strings.TrimSpace(header))

		// Map variations of common LinkedIn column names
		switch {
		case strings.Contains(headerLower, "first") && strings.Contains(headerLower, "name"):
			columnMap["firstName"] = i
		case strings.Contains(headerLower, "last") && strings.Contains(headerLower, "name"):
			columnMap["lastName"] = i
		case strings.Contains(headerLower, "email"):
			columnMap["email"] = i
		case strings.Contains(headerLower, "title") && strings.Contains(headerLower, "current"):
			columnMap["currentTitle"] = i
		case strings.Contains(headerLower, "company"):
			columnMap["companyName"] = i
		case strings.Contains(headerLower, "industry"):
			columnMap["industry"] = i
		}
	}

	return columnMap
}

// getCSVField safely retrieves a field from a record by column name.
func getCSVField(record []string, columnMap map[string]int, fieldName string) string {
	if idx, ok := columnMap[fieldName]; ok && idx < len(record) {
		return strings.TrimSpace(record[idx])
	}
	return ""
}

// generateLinkedInID creates a synthetic LinkedIn ID from email.
// In production, this would come from the LinkedIn API.
func generateLinkedInID(email string) string {
	// Simple: hash email to a numeric ID
	return fmt.Sprintf("li_%s_%d", strings.Split(email, "@")[0], len(email)*12345)
}

// ptrTime returns a pointer to a time.Time.
func ptrTime(t time.Time) *time.Time {
	return &t
}
