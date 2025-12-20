package terminal

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Manager handles terminal session lifecycle
type Manager struct {
	sessions      map[string]*Session
	mu            sync.RWMutex
	maxSessions   int
	idleTimeout   time.Duration
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

// NewManager creates a new terminal session manager
func NewManager() *Manager {
	m := &Manager{
		sessions:    make(map[string]*Session),
		maxSessions: 100,
		idleTimeout: 30 * time.Minute,
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine
	m.startCleanup()

	return m
}

// CreateSession creates a new terminal session
func (m *Manager) CreateSession(userID string, cols, rows int, shell, workingDir string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check user session limit (max 5 per user)
	userSessions := m.getUserSessionCount(userID)
	if userSessions >= 5 {
		return nil, fmt.Errorf("session limit reached (max 5 per user)")
	}

	// Check global session limit
	if len(m.sessions) >= m.maxSessions {
		return nil, fmt.Errorf("maximum global session limit reached")
	}

	// Determine working directory
	if workingDir == "" {
		workingDir = getDefaultWorkingDir()
	}

	// Create session
	session := &Session{
		ID:           uuid.New().String(),
		UserID:       userID,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Cols:         cols,
		Rows:         rows,
		Shell:        shell,
		WorkingDir:   workingDir,
		Status:       StatusActive,
		Environment:  m.buildEnvironment(userID),
	}

	// Start PTY
	if err := startPTY(session); err != nil {
		return nil, fmt.Errorf("failed to start PTY: %w", err)
	}

	// Store session
	m.sessions[session.ID] = session

	return session, nil
}

// GetSession retrieves a session by ID
func (m *Manager) GetSession(sessionID string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	return session, nil
}

// GetUserSessions retrieves all sessions for a user
func (m *Manager) GetUserSessions(userID string) []SessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sessions []SessionInfo
	for _, session := range m.sessions {
		if session.UserID == userID {
			sessions = append(sessions, session.ToInfo())
		}
	}

	return sessions
}

// UpdateActivity updates the last activity time for a session
func (m *Manager) UpdateActivity(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, exists := m.sessions[sessionID]; exists {
		session.LastActivity = time.Now()
	}
}

// CloseSession closes a terminal session
func (m *Manager) CloseSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// Close PTY and kill process
	closePTY(session)

	// Update status
	session.Status = StatusClosed

	// Remove from active sessions
	delete(m.sessions, sessionID)

	return nil
}

// ResizeSession resizes the terminal
func (m *Manager) ResizeSession(sessionID string, cols, rows int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	session.Cols = cols
	session.Rows = rows

	return resizePTY(session, cols, rows)
}

// Shutdown closes all sessions and stops cleanup
func (m *Manager) Shutdown() {
	close(m.stopCleanup)
	if m.cleanupTicker != nil {
		m.cleanupTicker.Stop()
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Close all sessions
	for sessionID, session := range m.sessions {
		closePTY(session)
		delete(m.sessions, sessionID)
	}
}

// Helper functions

func (m *Manager) getUserSessionCount(userID string) int {
	count := 0
	for _, session := range m.sessions {
		if session.UserID == userID && session.Status == StatusActive {
			count++
		}
	}
	return count
}

func (m *Manager) buildEnvironment(userID string) map[string]string {
	env := make(map[string]string)
	env["TERM"] = "xterm-256color"
	env["LANG"] = "en_US.UTF-8"
	env["COLORTERM"] = "truecolor"
	return env
}

func (m *Manager) startCleanup() {
	m.cleanupTicker = time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-m.cleanupTicker.C:
				m.cleanupIdleSessions()
			case <-m.stopCleanup:
				return
			}
		}
	}()
}

func (m *Manager) cleanupIdleSessions() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for sessionID, session := range m.sessions {
		if now.Sub(session.LastActivity) > m.idleTimeout {
			closePTY(session)
			delete(m.sessions, sessionID)
		}
	}
}
