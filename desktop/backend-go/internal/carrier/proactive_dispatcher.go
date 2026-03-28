package carrier

import (
	"log/slog"

	"github.com/rhl/businessos-backend/internal/services"
)

// SSEBroadcaster is the interface that ProactiveDispatcher uses to fan out
// events to connected SSE clients. *services.SSEBroadcaster satisfies this
// interface; a mock can be injected in tests.
type SSEBroadcaster interface {
	SendToAll(event services.SSEEvent)
}

// ProactiveDispatcher handles decision-request and proactive-signal commands
// from the CARRIER queue by broadcasting them to all connected SSE clients.
//
// It is separate from ProactiveConsumer so that the broadcast logic can be
// tested without a live RabbitMQ connection.
type ProactiveDispatcher struct {
	broadcaster SSEBroadcaster
	logger      *slog.Logger
}

// NewProactiveDispatcher creates a ProactiveDispatcher. broadcaster must not
// be nil; it is the SSEBroadcaster that fans out events to connected clients.
func NewProactiveDispatcher(broadcaster SSEBroadcaster) *ProactiveDispatcher {
	return &ProactiveDispatcher{
		broadcaster: broadcaster,
		logger:      slog.Default(),
	}
}

// NewTestableProactiveDispatcher creates a ProactiveDispatcher with an injected
// broadcaster. Intended for use in unit tests.
func NewTestableProactiveDispatcher(broadcaster SSEBroadcaster) *ProactiveDispatcher {
	return &ProactiveDispatcher{
		broadcaster: broadcaster,
		logger:      slog.Default(),
	}
}

// HandleRequestDecision processes a request_decision command by broadcasting
// an SSE event of type "request_decision" to all connected clients.
//
// Armstrong rule: no recover(); if broadcaster panics that is information about
// a real bug and must surface.
func (d *ProactiveDispatcher) HandleRequestDecision(cmd ActionCommand) {
	question, _ := cmd.Params["question"].(string)
	options, _ := cmd.Params["options"].([]interface{})
	deadline, _ := cmd.Params["deadline"].(string)
	context, _ := cmd.Params["context"].(string)

	optionStrs := make([]string, 0, len(options))
	for _, o := range options {
		if s, ok := o.(string); ok {
			optionStrs = append(optionStrs, s)
		}
	}

	d.logger.Info("proactive dispatcher: broadcasting decision request",
		"correlation_id", cmd.CorrelationID,
		"execution_id", cmd.ExecutionID,
		"step_id", cmd.StepID,
		"os_instance_id", cmd.OSInstanceID,
		"question", question,
		"options", optionStrs,
		"deadline", deadline,
		"context", context,
	)

	d.broadcaster.SendToAll(services.SSEEvent{
		Type: "request_decision",
		Data: map[string]interface{}{
			"correlation_id": cmd.CorrelationID,
			"execution_id":   cmd.ExecutionID,
			"step_id":        cmd.StepID,
			"os_instance_id": cmd.OSInstanceID,
			"question":       question,
			"options":        optionStrs,
			"deadline":       deadline,
			"context":        context,
		},
	})
}

// HandleProactiveSignal processes a proactive_signal command by broadcasting
// an SSE event of type "proactive_signal" to all connected clients.
//
// Armstrong rule: no recover(); panics surface immediately.
func (d *ProactiveDispatcher) HandleProactiveSignal(cmd ActionCommand) {
	signalType, _ := cmd.Params["signal_type"].(string)
	severity, _ := cmd.Params["severity"].(string)
	message, _ := cmd.Params["message"].(string)
	metric, _ := cmd.Params["metric"].(string)
	value, _ := cmd.Params["value"]
	threshold, _ := cmd.Params["threshold"]

	d.logger.Info("proactive dispatcher: broadcasting proactive signal",
		"correlation_id", cmd.CorrelationID,
		"execution_id", cmd.ExecutionID,
		"os_instance_id", cmd.OSInstanceID,
		"signal_type", signalType,
		"severity", severity,
		"message", message,
		"metric", metric,
		"value", value,
		"threshold", threshold,
	)

	d.broadcaster.SendToAll(services.SSEEvent{
		Type: "proactive_signal",
		Data: map[string]interface{}{
			"correlation_id": cmd.CorrelationID,
			"execution_id":   cmd.ExecutionID,
			"os_instance_id": cmd.OSInstanceID,
			"signal_type":    signalType,
			"severity":       severity,
			"message":        message,
			"metric":         metric,
			"value":          value,
			"threshold":      threshold,
		},
	})
}
