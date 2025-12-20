package terminal

import (
	"os"
	"os/exec"
	"time"
)

// SessionStatus represents the state of a terminal session
type SessionStatus string

const (
	StatusActive SessionStatus = "active"
	StatusIdle   SessionStatus = "idle"
	StatusClosed SessionStatus = "closed"
)

// Session represents an active terminal session
type Session struct {
	ID           string            `json:"id"`
	UserID       string            `json:"user_id"`
	CreatedAt    time.Time         `json:"created_at"`
	LastActivity time.Time         `json:"last_activity"`
	Cols         int               `json:"cols"`
	Rows         int               `json:"rows"`
	Shell        string            `json:"shell"`
	WorkingDir   string            `json:"working_dir"`
	Environment  map[string]string `json:"-"`
	PTY          *os.File          `json:"-"`
	Cmd          *exec.Cmd         `json:"-"`
	Status       SessionStatus     `json:"status"`
}

// MessageType represents WebSocket message types
type MessageType string

const (
	MsgTypeInput     MessageType = "input"
	MsgTypeOutput    MessageType = "output"
	MsgTypeResize    MessageType = "resize"
	MsgTypeHeartbeat MessageType = "heartbeat"
	MsgTypeError     MessageType = "error"
	MsgTypeStatus    MessageType = "status"
)

// TerminalMessage represents a WebSocket message
type TerminalMessage struct {
	Type      MessageType            `json:"type"`
	SessionID string                 `json:"session_id,omitempty"`
	Data      string                 `json:"data,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ResizeData for terminal resize events
type ResizeData struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

// SessionInfo is the public session information
type SessionInfo struct {
	ID           string        `json:"id"`
	CreatedAt    time.Time     `json:"created_at"`
	LastActivity time.Time     `json:"last_activity"`
	Cols         int           `json:"cols"`
	Rows         int           `json:"rows"`
	Shell        string        `json:"shell"`
	WorkingDir   string        `json:"working_dir"`
	Status       SessionStatus `json:"status"`
}

// ToInfo converts a Session to public SessionInfo
func (s *Session) ToInfo() SessionInfo {
	return SessionInfo{
		ID:           s.ID,
		CreatedAt:    s.CreatedAt,
		LastActivity: s.LastActivity,
		Cols:         s.Cols,
		Rows:         s.Rows,
		Shell:        s.Shell,
		WorkingDir:   s.WorkingDir,
		Status:       s.Status,
	}
}
