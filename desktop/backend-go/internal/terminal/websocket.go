package terminal

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now (CORS is handled elsewhere)
		return true
	},
}

// WebSocketHandler handles terminal WebSocket connections
type WebSocketHandler struct {
	manager *Manager
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(manager *Manager) *WebSocketHandler {
	return &WebSocketHandler{
		manager: manager,
	}
}

// HandleConnection handles a WebSocket terminal connection
func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request, userID string) {
	log.Printf("[Terminal] HandleConnection starting for user: %s", userID)

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Terminal] Failed to upgrade WebSocket connection: %v", err)
		return
	}
	log.Printf("[Terminal] WebSocket upgraded successfully")
	defer conn.Close()

	// Parse query parameters for terminal configuration
	query := r.URL.Query()
	cols := parseIntParam(query.Get("cols"), 80)
	rows := parseIntParam(query.Get("rows"), 24)
	shell := query.Get("shell")
	if shell == "" {
		shell = "zsh" // Default to zsh on macOS
	}
	workingDir := query.Get("cwd")
	log.Printf("[Terminal] Config: cols=%d, rows=%d, shell=%s, cwd=%s", cols, rows, shell, workingDir)

	// Create terminal session
	log.Printf("[Terminal] Creating session...")
	session, err := h.manager.CreateSession(userID, cols, rows, shell, workingDir)
	if err != nil {
		log.Printf("[Terminal] CreateSession error: %v", err)
		h.sendError(conn, err.Error())
		return
	}
	log.Printf("[Terminal] Session created: %s", session.ID)
	defer h.manager.CloseSession(session.ID)

	// Send connected status with session ID
	log.Printf("[Terminal] Session created: %s for user %s", session.ID, userID)
	log.Printf("[Terminal] Sending connected status...")
	h.sendStatus(conn, "connected", map[string]interface{}{
		"session_id": session.ID,
		"cols":       cols,
		"rows":       rows,
		"shell":      shell,
	})
	log.Printf("[Terminal] Status message sent")

	// Send welcome banner via WebSocket (not PTY!)
	log.Printf("[Terminal] Sending welcome banner via WebSocket...")
	banner := GetWelcomeBanner()
	h.sendOutput(conn, session.ID, banner)
	log.Printf("[Terminal] Banner sent")

	// Start bi-directional streaming
	errChan := make(chan error, 2)

	// Read from WebSocket, write to PTY (user input)
	go h.handleInput(conn, session, errChan)

	// Read from PTY, write to WebSocket (terminal output)
	go h.handleOutput(conn, session, errChan)

	// Wait for error or disconnect
	err = <-errChan
	if err != nil && err != io.EOF {
		log.Printf("Terminal session %s ended: %v", session.ID, err)
	}
}

// handleInput reads from WebSocket and writes to PTY
func (h *WebSocketHandler) handleInput(conn *websocket.Conn, session *Session, errChan chan error) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			errChan <- err
			return
		}

		var msg TerminalMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			// Treat as raw input if not JSON
			_, writeErr := WritePTY(session, message)
			if writeErr != nil {
				errChan <- writeErr
				return
			}
			continue
		}

		switch msg.Type {
		case MsgTypeInput:
			// Write user input to PTY
			_, err := WritePTY(session, []byte(msg.Data))
			if err != nil {
				errChan <- err
				return
			}
			h.manager.UpdateActivity(session.ID)

		case MsgTypeResize:
			// Handle terminal resize
			var resizeData ResizeData
			if err := json.Unmarshal([]byte(msg.Data), &resizeData); err == nil {
				h.manager.ResizeSession(session.ID, resizeData.Cols, resizeData.Rows)
			}

		case MsgTypeHeartbeat:
			// Update last activity
			h.manager.UpdateActivity(session.ID)
		}
	}
}

// handleOutput reads from PTY and writes to WebSocket
func (h *WebSocketHandler) handleOutput(conn *websocket.Conn, session *Session, errChan chan error) {
	buffer := make([]byte, 4096)

	for {
		n, err := ReadPTY(session, buffer)
		if err != nil {
			errChan <- err
			return
		}

		if n > 0 {
			// Send output to client
			msg := TerminalMessage{
				Type:      MsgTypeOutput,
				SessionID: session.ID,
				Data:      string(buffer[:n]),
			}

			msgBytes, _ := json.Marshal(msg)
			if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
				errChan <- err
				return
			}
		}
	}
}

// sendError sends an error message to the client
func (h *WebSocketHandler) sendError(conn *websocket.Conn, message string) {
	msg := TerminalMessage{
		Type: MsgTypeError,
		Data: message,
	}
	msgBytes, _ := json.Marshal(msg)
	conn.WriteMessage(websocket.TextMessage, msgBytes)
}

// sendOutput sends output data to the client
func (h *WebSocketHandler) sendOutput(conn *websocket.Conn, sessionID string, data string) {
	msg := TerminalMessage{
		Type:      MsgTypeOutput,
		SessionID: sessionID,
		Data:      data,
	}
	msgBytes, _ := json.Marshal(msg)
	if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
		log.Printf("[Terminal] ERROR writing output message: %v", err)
	}
}

// sendStatus sends a status message to the client
func (h *WebSocketHandler) sendStatus(conn *websocket.Conn, status string, metadata map[string]interface{}) {
	msg := TerminalMessage{
		Type:     MsgTypeStatus,
		Data:     status,
		Metadata: metadata,
	}
	msgBytes, _ := json.Marshal(msg)
	log.Printf("[Terminal] Sending status message: %s", string(msgBytes))
	if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
		log.Printf("[Terminal] ERROR writing status message: %v", err)
	}
}

// parseIntParam parses an integer from string with default value
func parseIntParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}

// SetPongHandler sets up WebSocket keep-alive
func SetupKeepAlive(conn *websocket.Conn) {
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
}
