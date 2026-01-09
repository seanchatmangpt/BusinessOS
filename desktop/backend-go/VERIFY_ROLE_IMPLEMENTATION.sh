#!/bin/bash
# Verification Script for Role-Based Agent Behavior Implementation
# Run this to verify all components are in place

set -e

echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║   ROLE-BASED AGENT BEHAVIOR - IMPLEMENTATION VERIFICATION    ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""

SUCCESS=0
FAILURES=0

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

check_file() {
    if [ -f "$1" ]; then
        echo -e "${GREEN}✓${NC} $2"
        ((SUCCESS++))
    else
        echo -e "${RED}✗${NC} $2 - NOT FOUND: $1"
        ((FAILURES++))
    fi
}

check_directory() {
    if [ -d "$1" ]; then
        echo -e "${GREEN}✓${NC} $2"
        ((SUCCESS++))
    else
        echo -e "${RED}✗${NC} $2 - NOT FOUND: $1"
        ((FAILURES++))
    fi
}

check_pattern() {
    if grep -q "$2" "$1" 2>/dev/null; then
        echo -e "${GREEN}✓${NC} $3"
        ((SUCCESS++))
    else
        echo -e "${RED}✗${NC} $3 - Pattern not found in $1"
        ((FAILURES++))
    fi
}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. DATABASE SCHEMA"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

check_file "internal/database/migrations/026_workspaces_and_roles.sql" "Migration 026 exists"
check_pattern "internal/database/migrations/026_workspaces_and_roles.sql" "CREATE TABLE.*workspaces" "Workspaces table defined"
check_pattern "internal/database/migrations/026_workspaces_and_roles.sql" "CREATE TABLE.*workspace_roles" "Workspace roles table defined"
check_pattern "internal/database/migrations/026_workspaces_and_roles.sql" "CREATE TABLE.*workspace_members" "Workspace members table defined"
check_pattern "internal/database/migrations/026_workspaces_and_roles.sql" "CREATE TABLE.*role_permissions" "Role permissions table defined"
check_pattern "internal/database/migrations/026_workspaces_and_roles.sql" "seed_default_workspace_roles" "Default roles seed function defined"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2. SERVICES"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

check_file "internal/services/role_context.go" "RoleContextService exists"
check_pattern "internal/services/role_context.go" "type RoleContextService struct" "RoleContextService defined"
check_pattern "internal/services/role_context.go" "GetUserRoleContext" "GetUserRoleContext method exists"
check_pattern "internal/services/role_context.go" "HasPermission" "HasPermission method exists"
check_pattern "internal/services/role_context.go" "GetRoleContextPrompt" "GetRoleContextPrompt method exists"

check_file "internal/services/workspace_service.go" "WorkspaceService exists"
check_pattern "internal/services/workspace_service.go" "type WorkspaceService struct" "WorkspaceService defined"
check_pattern "internal/services/workspace_service.go" "CreateWorkspace" "CreateWorkspace method exists"
check_pattern "internal/services/workspace_service.go" "AddMember" "AddMember method exists"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3. MIDDLEWARE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

check_file "internal/middleware/permission_check.go" "Permission middleware exists"
check_pattern "internal/middleware/permission_check.go" "InjectRoleContext" "InjectRoleContext middleware exists"
check_pattern "internal/middleware/permission_check.go" "RequirePermission" "RequirePermission middleware exists"
check_pattern "internal/middleware/permission_check.go" "RequireWorkspaceOwner" "RequireWorkspaceOwner middleware exists"
check_pattern "internal/middleware/permission_check.go" "RequireWorkspaceAdmin" "RequireWorkspaceAdmin middleware exists"
check_pattern "internal/middleware/permission_check.go" "RequireHierarchyLevel" "RequireHierarchyLevel middleware exists"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4. HANDLERS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

check_file "internal/handlers/workspace_handlers.go" "Workspace handlers exist"
check_pattern "internal/handlers/workspace_handlers.go" "CreateWorkspace" "CreateWorkspace handler exists"
check_pattern "internal/handlers/workspace_handlers.go" "AddWorkspaceMember" "AddWorkspaceMember handler exists"
check_pattern "internal/handlers/workspace_handlers.go" "UpdateWorkspaceMemberRole" "UpdateWorkspaceMemberRole handler exists"
check_pattern "internal/handlers/workspace_handlers.go" "ListWorkspaceRoles" "ListWorkspaceRoles handler exists"

check_file "internal/handlers/handlers.go" "Main handlers file exists"
check_pattern "internal/handlers/handlers.go" "roleContextService.*RoleContextService" "roleContextService field exists"
check_pattern "internal/handlers/handlers.go" "SetRoleContextService" "SetRoleContextService method exists"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5. ROUTE REGISTRATION"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

check_pattern "internal/handlers/handlers.go" "/workspaces" "Workspace routes registered"
check_pattern "internal/handlers/handlers.go" "InjectRoleContext" "Role context injected in routes"
check_pattern "internal/handlers/handlers.go" "RequireWorkspaceOwner" "Owner permission check on routes"
check_pattern "internal/handlers/handlers.go" "RequireWorkspaceAdmin" "Admin permission check on routes"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "6. AGENT INTEGRATION"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

check_file "internal/agents/agent_v2.go" "Agent V2 interface exists"
check_pattern "internal/agents/agent_v2.go" "SetRoleContextPrompt" "SetRoleContextPrompt in interface"

check_file "internal/agents/base_agent_v2.go" "Base Agent V2 exists"
check_pattern "internal/agents/base_agent_v2.go" "roleContextPrompt" "roleContextPrompt field exists"

check_file "internal/handlers/chat_v2.go" "Chat V2 handler exists"
check_pattern "internal/handlers/chat_v2.go" "roleContextService.GetUserRoleContext" "Role context retrieval in chat"
check_pattern "internal/handlers/chat_v2.go" "SetRoleContextPrompt" "Role context injection in agents"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "7. SERVER INITIALIZATION"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

check_file "cmd/server/main.go" "Server main file exists"
check_pattern "cmd/server/main.go" "NewRoleContextService" "RoleContextService initialized"
check_pattern "cmd/server/main.go" "SetRoleContextService" "RoleContextService set in handlers"
check_pattern "cmd/server/main.go" "NewWorkspaceService" "WorkspaceService initialized"
check_pattern "cmd/server/main.go" "SetWorkspaceService" "WorkspaceService set in handlers"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "8. TESTS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

check_file "internal/services/role_context_test.go" "Role context tests exist"

echo ""
echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║                     VERIFICATION SUMMARY                      ║"
echo "╠═══════════════════════════════════════════════════════════════╣"
echo -e "║  ${GREEN}Successful Checks: $SUCCESS${NC}                                    "
echo -e "║  ${RED}Failed Checks:     $FAILURES${NC}                                    "
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""

if [ $FAILURES -eq 0 ]; then
    echo -e "${GREEN}✓ ALL CHECKS PASSED${NC}"
    echo ""
    echo -e "${YELLOW}Next Steps:${NC}"
    echo "1. Run database migration:"
    echo "   psql business_os < internal/database/migrations/026_workspaces_and_roles.sql"
    echo ""
    echo "2. Restart backend server"
    echo ""
    echo "3. Test the implementation (see ROLE_BASED_AGENT_DEPLOYMENT.md)"
    echo ""
    exit 0
else
    echo -e "${RED}✗ SOME CHECKS FAILED${NC}"
    echo ""
    echo "Please review the failed checks above and ensure all files are present."
    echo ""
    exit 1
fi
