package linkedin

import (
	"time"
)

// Contact represents a LinkedIn contact from CSV import.
type Contact struct {
	ID             int64      `json:"id"`
	LinkedInID     string     `json:"linkedin_id"`
	Name           string     `json:"name"`
	Title          string     `json:"title"`
	Company        string     `json:"company"`
	Industry       string     `json:"industry"`
	ConnectionDate *time.Time `json:"connection_date"`
	ICPScore       float64    `json:"icp_score"`
	ICPScoredAt    *time.Time `json:"icp_scored_at"`
	RawCSV         string     `json:"raw_csv"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// OutreachSequence represents a LinkedIn outreach workflow.
type OutreachSequence struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	TargetICPMinScore float64   `json:"target_icp_min_score"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// SequenceStep represents a single step within an outreach sequence.
type SequenceStep struct {
	ID              int64     `json:"id"`
	SequenceID      int64     `json:"sequence_id"`
	StepOrder       int       `json:"step_order"`
	MessageTemplate string    `json:"message_template"`
	DelayDays       int       `json:"delay_days"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// SequenceEnrollment tracks a contact's progress through an outreach sequence.
type SequenceEnrollment struct {
	ID          int64      `json:"id"`
	ContactID   int64      `json:"contact_id"`
	SequenceID  int64      `json:"sequence_id"`
	CurrentStep int        `json:"current_step"`
	EnrolledAt  time.Time  `json:"enrolled_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// LinkedInMessageQueue represents a message pending send via LinkedIn.
type LinkedInMessageQueue struct {
	ID          int64      `json:"id"`
	ContactID   int64      `json:"contact_id"`
	StepID      int64      `json:"step_id"`
	ScheduledAt time.Time  `json:"scheduled_at"`
	SentAt      *time.Time `json:"sent_at"`
	Status      string     `json:"status"` // "pending", "sent", "failed", "skipped"
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ICPScoreResult represents the scoring outcome for a contact.
type ICPScoreResult struct {
	ContactID int64              `json:"contact_id"`
	Name      string             `json:"name"`
	Title     string             `json:"title"`
	Company   string             `json:"company"`
	Score     float64            `json:"score"`
	Qualified bool               `json:"qualified"`
	Breakdown map[string]float64 `json:"breakdown"`
}

// ImportCSVRequest represents a LinkedIn CSV import request.
type ImportCSVRequest struct {
	CSVContent string `json:"csv_content" binding:"required"`
	Filename   string `json:"filename"`
}

// ImportCSVResponse represents the result of a CSV import.
type ImportCSVResponse struct {
	ContactsImported int      `json:"contacts_imported"`
	ContactsUpdated  int      `json:"contacts_updated"`
	ContactsFailed   int      `json:"contacts_failed"`
	Errors           []string `json:"errors,omitempty"`
}

// EnrollOutreachRequest represents a request to enroll contacts in a sequence.
type EnrollOutreachRequest struct {
	SequenceID  int64   `json:"sequence_id" binding:"required"`
	MinScore    float64 `json:"min_score" binding:"required"`
	TargetCount int     `json:"target_count"`
}

// EnrollOutreachResponse represents enrollment results.
type EnrollOutreachResponse struct {
	Enrolled int      `json:"enrolled"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors,omitempty"`
}

// ContactListResponse represents paginated contact list results.
type ContactListResponse struct {
	Contacts []Contact `json:"contacts"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	HasMore  bool      `json:"has_more"`
}
