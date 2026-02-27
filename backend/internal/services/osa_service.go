package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	osa "github.com/Miosa-osa/sdk-go"
)

// OSAService wraps the sdk-go Client and exposes BusinessOS-aware helpers.
// This is the real implementation; the stub OSASyncService in stubs.go remains
// unchanged for onboarding_service.go compatibility.
type OSAService struct {
	client osa.Client
}

// NewOSAService creates an OSAService backed by an initialised sdk-go Client.
func NewOSAService(client osa.Client) *OSAService {
	return &OSAService{client: client}
}

// Client returns the underlying sdk-go Client.
// Handlers that need direct access to methods not wrapped by OSAService
// (e.g. Orchestrate, Stream, Classify) can use this accessor.
func (s *OSAService) Client() osa.Client {
	return s.client
}

// SyncUserWithOSA sends a user-scoped orchestration request to OSA.
// It maps the BusinessOS userID to the sdk-go OrchestrateRequest.UserID field.
func (s *OSAService) SyncUserWithOSA(ctx context.Context, userID uuid.UUID, input string) (*osa.AgentResponse, error) {
	if input == "" {
		return nil, fmt.Errorf("osa: input must not be empty")
	}
	req := osa.OrchestrateRequest{
		Input:  input,
		UserID: userID.String(),
	}
	resp, err := s.client.Orchestrate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("osa: SyncUserWithOSA: %w", err)
	}
	return resp, nil
}

// SyncWorkspaceWithOSA sends a workspace-scoped orchestration request to OSA.
// It maps the BusinessOS workspaceID to the sdk-go OrchestrateRequest.WorkspaceID field.
func (s *OSAService) SyncWorkspaceWithOSA(ctx context.Context, workspaceID uuid.UUID, input string) (*osa.AgentResponse, error) {
	if input == "" {
		return nil, fmt.Errorf("osa: input must not be empty")
	}
	req := osa.OrchestrateRequest{
		Input:       input,
		WorkspaceID: workspaceID.String(),
	}
	resp, err := s.client.Orchestrate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("osa: SyncWorkspaceWithOSA: %w", err)
	}
	return resp, nil
}

// Health delegates to the underlying sdk-go client.Health().
func (s *OSAService) Health(ctx context.Context) (*osa.HealthStatus, error) {
	status, err := s.client.Health(ctx)
	if err != nil {
		return nil, fmt.Errorf("osa: Health: %w", err)
	}
	return status, nil
}

// ListSkills delegates to the underlying sdk-go client.ListSkills().
func (s *OSAService) ListSkills(ctx context.Context) ([]osa.SkillDefinition, error) {
	skills, err := s.client.ListSkills(ctx)
	if err != nil {
		return nil, fmt.Errorf("osa: ListSkills: %w", err)
	}
	return skills, nil
}
