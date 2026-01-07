package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// TEST SETUP HELPERS
// =============================================================================

// setupTestHandlers creates test handlers with dependencies
func setupTestHandlers(pool *pgxpool.Pool) *Handlers {
	workspaceService := services.NewWorkspaceService(pool)
	memoryHierarchyService := services.NewMemoryHierarchyService(pool)

	return &Handlers{
		workspaceService:           workspaceService,
		memoryHierarchyService:     memoryHierarchyService,
		roleContextService:         services.NewRoleContextService(pool),
		// Add other services as needed
	}
}

// setupTestRouter creates a test router with middleware
func setupTestRouter(handlers *Handlers, user *middleware.User) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock auth middleware that sets the test user
	router.Use(func(c *gin.Context) {
		if user != nil {
			middleware.SetCurrentUser(c, user)
		}
		c.Next()
	})

	// Register routes
	api := router.Group("/api")
	{
		// Workspace routes
		api.POST("/workspaces", handlers.CreateWorkspace)
		api.GET("/workspaces", handlers.ListWorkspaces)
		api.GET("/workspaces/:id", handlers.GetWorkspace)
		api.PUT("/workspaces/:id", handlers.UpdateWorkspace)
		api.DELETE("/workspaces/:id", handlers.DeleteWorkspace)

		// Member routes
		api.GET("/workspaces/:id/members", handlers.ListWorkspaceMembers)
		api.POST("/workspaces/:id/members/invite", handlers.AddWorkspaceMember)
		api.PUT("/workspaces/:id/members/:userId", handlers.UpdateWorkspaceMemberRole)
		api.DELETE("/workspaces/:id/members/:userId", handlers.RemoveWorkspaceMember)

		// Role routes
		api.GET("/workspaces/:id/roles", handlers.ListWorkspaceRoles)
		api.GET("/workspaces/:id/profile", handlers.GetWorkspaceProfile)
		api.GET("/workspaces/:id/role-context", handlers.GetUserRoleContext)

		// Memory routes
		api.GET("/workspaces/:id/memories", handlers.GetWorkspaceMemories)
	}

	return router
}

// createTestUser creates a test user
func createTestUser(id string) *middleware.User {
	return &middleware.User{
		ID:    id,
		Email: id + "@test.com",
	}
}

// makeRequest makes an HTTP request and returns the response
func makeRequest(t *testing.T, router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// =============================================================================
// WORKSPACE CRUD HANDLER TESTS
// =============================================================================

// TestCreateWorkspaceHandler tests POST /api/workspaces
func TestCreateWorkspaceHandler(t *testing.T) {
	t.Skip("Requires database - implement with testcontainers or run with live DB")

	var pool *pgxpool.Pool // Replace with actual pool
	handlers := setupTestHandlers(pool)
	user := createTestUser("test-owner-" + uuid.New().String()[:8])
	router := setupTestRouter(handlers, user)

	// Test: Create workspace successfully
	t.Run("Create workspace successfully", func(t *testing.T) {
		reqBody := services.CreateWorkspaceRequest{
			Name:     "Test Workspace",
			Slug:     "test-workspace",
			PlanType: "professional",
		}

		w := makeRequest(t, router, "POST", "/api/workspaces", reqBody)
		assert.Equal(t, http.StatusCreated, w.Code)

		var workspace services.Workspace
		err := json.Unmarshal(w.Body.Bytes(), &workspace)
		require.NoError(t, err)
		assert.Equal(t, "Test Workspace", workspace.Name)
		assert.Equal(t, user.ID, workspace.OwnerID)

		// Cleanup
		handlers.workspaceService.DeleteWorkspace(context.Background(), workspace.ID, user.ID)
	})

	// Test: Reject unauthenticated request
	t.Run("Reject unauthenticated request", func(t *testing.T) {
		routerNoAuth := setupTestRouter(handlers, nil)
		reqBody := services.CreateWorkspaceRequest{
			Name:     "Test",
			PlanType: "free",
		}

		w := makeRequest(t, routerNoAuth, "POST", "/api/workspaces", reqBody)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test: Reject invalid request body
	t.Run("Reject invalid request body", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "", // Empty name
		}

		w := makeRequest(t, router, "POST", "/api/workspaces", reqBody)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// TestListWorkspacesHandler tests GET /api/workspaces
func TestListWorkspacesHandler(t *testing.T) {
	t.Skip("Requires database - implement with testcontainers or run with live DB")

	var pool *pgxpool.Pool // Replace with actual pool
	handlers := setupTestHandlers(pool)
	user := createTestUser("test-user-" + uuid.New().String()[:8])
	router := setupTestRouter(handlers, user)
	ctx := context.Background()

	// Create test workspaces
	ws1, _ := handlers.workspaceService.CreateWorkspace(ctx, services.CreateWorkspaceRequest{Name: "Workspace 1", PlanType: "free"}, user.ID)
	defer handlers.workspaceService.DeleteWorkspace(ctx, ws1.ID, user.ID)

	ws2, _ := handlers.workspaceService.CreateWorkspace(ctx, services.CreateWorkspaceRequest{Name: "Workspace 2", PlanType: "free"}, user.ID)
	defer handlers.workspaceService.DeleteWorkspace(ctx, ws2.ID, user.ID)

	// Test: List user's workspaces
	t.Run("List user workspaces", func(t *testing.T) {
		w := makeRequest(t, router, "GET", "/api/workspaces", nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string][]services.Workspace
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		workspaces := response["workspaces"]
		assert.GreaterOrEqual(t, len(workspaces), 2)

		// Find our workspaces
		found := 0
		for _, ws := range workspaces {
			if ws.ID == ws1.ID || ws.ID == ws2.ID {
				found++
			}
		}
		assert.Equal(t, 2, found, "Should find both created workspaces")
	})
}

// TestGetWorkspaceHandler tests GET /api/workspaces/:id
func TestGetWorkspaceHandler(t *testing.T) {
	t.Skip("Requires database - implement with testcontainers or run with live DB")

	var pool *pgxpool.Pool // Replace with actual pool
	handlers := setupTestHandlers(pool)
	user := createTestUser("test-user-" + uuid.New().String()[:8])
	router := setupTestRouter(handlers, user)
	ctx := context.Background()

	workspace, _ := handlers.workspaceService.CreateWorkspace(ctx, services.CreateWorkspaceRequest{Name: "Test", PlanType: "free"}, user.ID)
	defer handlers.workspaceService.DeleteWorkspace(ctx, workspace.ID, user.ID)

	// Test: Get workspace as member
	t.Run("Get workspace as member", func(t *testing.T) {
		w := makeRequest(t, router, "GET", "/api/workspaces/"+workspace.ID.String(), nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var ws services.Workspace
		err := json.Unmarshal(w.Body.Bytes(), &ws)
		require.NoError(t, err)
		assert.Equal(t, workspace.ID, ws.ID)
	})

	// Test: Reject non-member
	t.Run("Reject non-member", func(t *testing.T) {
		otherUser := createTestUser("other-user")
		otherRouter := setupTestRouter(handlers, otherUser)

		w := makeRequest(t, otherRouter, "GET", "/api/workspaces/"+workspace.ID.String(), nil)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// TestUpdateWorkspaceHandler tests PUT /api/workspaces/:id
func TestUpdateWorkspaceHandler(t *testing.T) {
	t.Skip("Requires database - implement with testcontainers or run with live DB")

	var pool *pgxpool.Pool // Replace with actual pool
	handlers := setupTestHandlers(pool)
	owner := createTestUser("test-owner-" + uuid.New().String()[:8])
	member := createTestUser("test-member-" + uuid.New().String()[:8])
	ctx := context.Background()

	workspace, _ := handlers.workspaceService.CreateWorkspace(ctx, services.CreateWorkspaceRequest{Name: "Test", PlanType: "free"}, owner.ID)
	defer handlers.workspaceService.DeleteWorkspace(ctx, workspace.ID, owner.ID)

	// Add member
	handlers.workspaceService.AddMember(ctx, workspace.ID, services.AddMemberRequest{UserID: member.ID, Role: "member"}, owner.ID)

	// Test: Owner can update
	t.Run("Owner can update workspace", func(t *testing.T) {
		ownerRouter := setupTestRouter(handlers, owner)
		newName := "Updated Name"
		reqBody := services.UpdateWorkspaceRequest{
			Name: &newName,
		}

		w := makeRequest(t, ownerRouter, "PUT", "/api/workspaces/"+workspace.ID.String(), reqBody)
		assert.Equal(t, http.StatusOK, w.Code)

		var ws services.Workspace
		json.Unmarshal(w.Body.Bytes(), &ws)
		assert.Equal(t, "Updated Name", ws.Name)
	})

	// Test: Member cannot update
	t.Run("Member cannot update workspace", func(t *testing.T) {
		memberRouter := setupTestRouter(handlers, member)
		newName := "Hacked Name"
		reqBody := services.UpdateWorkspaceRequest{
			Name: &newName,
		}

		w := makeRequest(t, memberRouter, "PUT", "/api/workspaces/"+workspace.ID.String(), reqBody)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// TestDeleteWorkspaceHandler tests DELETE /api/workspaces/:id
func TestDeleteWorkspaceHandler(t *testing.T) {
	t.Skip("Requires database - implement with testcontainers or run with live DB")

	var pool *pgxpool.Pool // Replace with actual pool
	handlers := setupTestHandlers(pool)
	owner := createTestUser("test-owner-" + uuid.New().String()[:8])
	admin := createTestUser("test-admin-" + uuid.New().String()[:8])
	ctx := context.Background()

	// Test: Owner can delete
	t.Run("Owner can delete workspace", func(t *testing.T) {
		workspace, _ := handlers.workspaceService.CreateWorkspace(ctx, services.CreateWorkspaceRequest{Name: "Test", PlanType: "free"}, owner.ID)
		ownerRouter := setupTestRouter(handlers, owner)

		w := makeRequest(t, ownerRouter, "DELETE", "/api/workspaces/"+workspace.ID.String(), nil)
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify deletion
		_, err := handlers.workspaceService.GetWorkspace(ctx, workspace.ID)
		assert.Error(t, err)
	})

	// Test: Non-owner cannot delete
	t.Run("Non-owner cannot delete workspace", func(t *testing.T) {
		workspace, _ := handlers.workspaceService.CreateWorkspace(ctx, services.CreateWorkspaceRequest{Name: "Test", PlanType: "free"}, owner.ID)
		defer handlers.workspaceService.DeleteWorkspace(ctx, workspace.ID, owner.ID)

		// Add admin
		handlers.workspaceService.AddMember(ctx, workspace.ID, services.AddMemberRequest{UserID: admin.ID, Role: "admin"}, owner.ID)

		adminRouter := setupTestRouter(handlers, admin)
		w := makeRequest(t, adminRouter, "DELETE", "/api/workspaces/"+workspace.ID.String(), nil)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// =============================================================================
// MEMBER MANAGEMENT HANDLER TESTS
// =============================================================================

// TestAddMemberHandler tests POST /api/workspaces/:id/members/invite
func TestAddMemberHandler(t *testing.T) {
	t.Skip("Requires database - implement with testcontainers or run with live DB")

	var pool *pgxpool.Pool // Replace with actual pool
	handlers := setupTestHandlers(pool)
	owner := createTestUser("test-owner-" + uuid.New().String()[:8])
	member := createTestUser("test-member-" + uuid.New().String()[:8])
	newUser := createTestUser("test-new-" + uuid.New().String()[:8])
	ctx := context.Background()

	workspace, _ := handlers.workspaceService.CreateWorkspace(ctx, services.CreateWorkspaceRequest{Name: "Test", PlanType: "free"}, owner.ID)
	defer handlers.workspaceService.DeleteWorkspace(ctx, workspace.ID, owner.ID)

	// Add member
	handlers.workspaceService.AddMember(ctx, workspace.ID, services.AddMemberRequest{UserID: member.ID, Role: "member"}, owner.ID)

	// Test: Owner can invite
	t.Run("Owner can invite members", func(t *testing.T) {
		ownerRouter := setupTestRouter(handlers, owner)
		reqBody := services.AddMemberRequest{
			UserID: newUser.ID,
			Role:   "viewer",
		}

		w := makeRequest(t, ownerRouter, "POST", "/api/workspaces/"+workspace.ID.String()+"/members/invite", reqBody)
		assert.Equal(t, http.StatusCreated, w.Code)

		var newMember services.WorkspaceMember
		json.Unmarshal(w.Body.Bytes(), &newMember)
		assert.Equal(t, newUser.ID, newMember.UserID)
		assert.Equal(t, "viewer", newMember.Role)
	})

	// Test: Member cannot invite
	t.Run("Member cannot invite", func(t *testing.T) {
		memberRouter := setupTestRouter(handlers, member)
		reqBody := services.AddMemberRequest{
			UserID: "another-user",
			Role:   "member",
		}

		w := makeRequest(t, memberRouter, "POST", "/api/workspaces/"+workspace.ID.String()+"/members/invite", reqBody)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// TestUpdateMemberRoleHandler tests PUT /api/workspaces/:id/members/:userId
func TestUpdateMemberRoleHandler(t *testing.T) {
	t.Skip("Requires database - implement with testcontainers or run with live DB")

	var pool *pgxpool.Pool // Replace with actual pool
	handlers := setupTestHandlers(pool)
	owner := createTestUser("test-owner-" + uuid.New().String()[:8])
	member := createTestUser("test-member-" + uuid.New().String()[:8])
	ctx := context.Background()

	workspace, _ := handlers.workspaceService.CreateWorkspace(ctx, services.CreateWorkspaceRequest{Name: "Test", PlanType: "free"}, owner.ID)
	defer handlers.workspaceService.DeleteWorkspace(ctx, workspace.ID, owner.ID)

	// Add member
	handlers.workspaceService.AddMember(ctx, workspace.ID, services.AddMemberRequest{UserID: member.ID, Role: "member"}, owner.ID)

	// Test: Owner can update role
	t.Run("Owner can update member role", func(t *testing.T) {
		ownerRouter := setupTestRouter(handlers, owner)
		reqBody := map[string]string{
			"role": "admin",
		}

		w := makeRequest(t, ownerRouter, "PUT", "/api/workspaces/"+workspace.ID.String()+"/members/"+member.ID, reqBody)
		assert.Equal(t, http.StatusOK, w.Code)

		var updatedMember services.WorkspaceMember
		json.Unmarshal(w.Body.Bytes(), &updatedMember)
		assert.Equal(t, "admin", updatedMember.Role)
	})

	// Test: Cannot change owner role
	t.Run("Cannot change owner role", func(t *testing.T) {
		ownerRouter := setupTestRouter(handlers, owner)
		reqBody := map[string]string{
			"role": "member",
		}

		w := makeRequest(t, ownerRouter, "PUT", "/api/workspaces/"+workspace.ID.String()+"/members/"+owner.ID, reqBody)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// TestRemoveMemberHandler tests DELETE /api/workspaces/:id/members/:userId
func TestRemoveMemberHandler(t *testing.T) {
	t.Skip("Requires database - implement with testcontainers or run with live DB")

	var pool *pgxpool.Pool // Replace with actual pool
	handlers := setupTestHandlers(pool)
	owner := createTestUser("test-owner-" + uuid.New().String()[:8])
	member := createTestUser("test-member-" + uuid.New().String()[:8])
	ctx := context.Background()

	workspace, _ := handlers.workspaceService.CreateWorkspace(ctx, services.CreateWorkspaceRequest{Name: "Test", PlanType: "free"}, owner.ID)
	defer handlers.workspaceService.DeleteWorkspace(ctx, workspace.ID, owner.ID)

	// Add member
	handlers.workspaceService.AddMember(ctx, workspace.ID, services.AddMemberRequest{UserID: member.ID, Role: "member"}, owner.ID)

	// Test: Owner can remove member
	t.Run("Owner can remove member", func(t *testing.T) {
		ownerRouter := setupTestRouter(handlers, owner)

		w := makeRequest(t, ownerRouter, "DELETE", "/api/workspaces/"+workspace.ID.String()+"/members/"+member.ID, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify removal
		_, err := handlers.workspaceService.GetUserRole(ctx, workspace.ID, member.ID)
		assert.Error(t, err)
	})

	// Test: Cannot remove owner
	t.Run("Cannot remove owner", func(t *testing.T) {
		ownerRouter := setupTestRouter(handlers, owner)

		w := makeRequest(t, ownerRouter, "DELETE", "/api/workspaces/"+workspace.ID.String()+"/members/"+owner.ID, nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// =============================================================================
// ROLE CONTEXT HANDLER TESTS
// =============================================================================

// TestGetUserRoleContextHandler tests GET /api/workspaces/:id/role-context
func TestGetUserRoleContextHandler(t *testing.T) {
	t.Skip("Requires database - implement with testcontainers or run with live DB")

	var pool *pgxpool.Pool // Replace with actual pool
	handlers := setupTestHandlers(pool)
	owner := createTestUser("test-owner-" + uuid.New().String()[:8])
	router := setupTestRouter(handlers, owner)
	ctx := context.Background()

	workspace, _ := handlers.workspaceService.CreateWorkspace(ctx, services.CreateWorkspaceRequest{Name: "Test", PlanType: "free"}, owner.ID)
	defer handlers.workspaceService.DeleteWorkspace(ctx, workspace.ID, owner.ID)

	// Test: Get role context
	t.Run("Get user role context", func(t *testing.T) {
		w := makeRequest(t, router, "GET", "/api/workspaces/"+workspace.ID.String()+"/role-context", nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var roleContext map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &roleContext)
		require.NoError(t, err)

		assert.Equal(t, owner.ID, roleContext["user_id"])
		assert.Equal(t, "owner", roleContext["role_name"])
		assert.Equal(t, "Owner", roleContext["role_display_name"])
		assert.Equal(t, float64(1), roleContext["hierarchy_level"])
		assert.NotNil(t, roleContext["permissions"])
	})
}

// TestListWorkspaceRolesHandler tests GET /api/workspaces/:id/roles
func TestListWorkspaceRolesHandler(t *testing.T) {
	t.Skip("Requires database - implement with testcontainers or run with live DB")

	var pool *pgxpool.Pool // Replace with actual pool
	handlers := setupTestHandlers(pool)
	owner := createTestUser("test-owner-" + uuid.New().String()[:8])
	router := setupTestRouter(handlers, owner)
	ctx := context.Background()

	workspace, _ := handlers.workspaceService.CreateWorkspace(ctx, services.CreateWorkspaceRequest{Name: "Test", PlanType: "free"}, owner.ID)
	defer handlers.workspaceService.DeleteWorkspace(ctx, workspace.ID, owner.ID)

	// Test: List roles
	t.Run("List workspace roles", func(t *testing.T) {
		w := makeRequest(t, router, "GET", "/api/workspaces/"+workspace.ID.String()+"/roles", nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string][]services.WorkspaceRole
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		roles := response["roles"]
		assert.Len(t, roles, 6, "Should have 6 default roles")

		// Verify role names
		roleNames := make(map[string]bool)
		for _, role := range roles {
			roleNames[role.Name] = true
		}
		assert.True(t, roleNames["owner"])
		assert.True(t, roleNames["admin"])
		assert.True(t, roleNames["manager"])
		assert.True(t, roleNames["member"])
		assert.True(t, roleNames["viewer"])
		assert.True(t, roleNames["guest"])
	})
}

// =============================================================================
// MEMORY HIERARCHY HANDLER TESTS
// =============================================================================

// TestWorkspaceMemoriesHandler tests workspace memory access control
func TestWorkspaceMemoriesHandler(t *testing.T) {
	t.Skip("Requires database and memory handlers - implement with testcontainers")

	// TODO: This test would verify:
	// - Workspace memories visible to all members
	// - Private memories only visible to owner
	// - Shared memories visible to specified users
	// - Memory access control enforcement

	// Example test structure:
	// 1. Create workspace with owner
	// 2. Create workspace memory (visible to all)
	// 3. Create private memory (owner only)
	// 4. Create shared memory (specific users)
	// 5. Add member to workspace
	// 6. Verify member can see workspace memory
	// 7. Verify member cannot see private memory
	// 8. Verify member can see shared memory if included
}
