#!/bin/bash

BASE_URL="http://localhost:8001/api"
TOKEN="${1:-test-token}"

echo "=================================="
echo "TESTING WORKSPACE ENDPOINTS"
echo "=================================="
echo ""

# Test 1: List workspaces
echo "1. GET /workspaces - List all workspaces"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/workspaces" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" = "200" ] || [ "$http_code" = "401" ]; then
  echo "   Status: $http_code"
  echo "   Response: $(echo $body | head -c 100)..."
  echo "   ✓ Endpoint responding"
else
  echo "   Status: $http_code"
  echo "   ✗ Unexpected status"
fi
echo ""

# Test 2: Get specific workspace
echo "2. GET /workspaces/:id - Get workspace details"
WORKSPACE_ID="550e8400-e29b-41d4-a716-446655440000"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/workspaces/$WORKSPACE_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" = "200" ] || [ "$http_code" = "404" ] || [ "$http_code" = "401" ]; then
  echo "   Status: $http_code"
  echo "   ✓ Endpoint responding"
else
  echo "   Status: $http_code"
  echo "   ✗ Unexpected status"
fi
echo ""

# Test 3: Create workspace
echo "3. POST /workspaces - Create new workspace"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/workspaces" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Workspace",
    "slug": "test-workspace",
    "description": "Test workspace for API testing"
  }')
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" = "201" ] || [ "$http_code" = "401" ] || [ "$http_code" = "409" ]; then
  echo "   Status: $http_code"
  echo "   ✓ Endpoint responding"
else
  echo "   Status: $http_code"
  echo "   ✗ Unexpected status"
fi
echo ""

# Test 4: List workspace members
echo "4. GET /workspaces/:id/members - List members"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/workspaces/$WORKSPACE_ID/members" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json")
http_code=$(echo "$response" | tail -n1)

if [ "$http_code" = "200" ] || [ "$http_code" = "404" ] || [ "$http_code" = "401" ]; then
  echo "   Status: $http_code"
  echo "   ✓ Endpoint responding"
else
  echo "   Status: $http_code"
  echo "   ✗ Unexpected status"
fi
echo ""

# Test 5: List workspace roles
echo "5. GET /workspaces/:id/roles - List roles"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/workspaces/$WORKSPACE_ID/roles" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json")
http_code=$(echo "$response" | tail -n1)

if [ "$http_code" = "200" ] || [ "$http_code" = "404" ] || [ "$http_code" = "401" ]; then
  echo "   Status: $http_code"
  echo "   ✓ Endpoint responding"
else
  echo "   Status: $http_code"
  echo "   ✗ Unexpected status"
fi
echo ""

echo "=================================="
echo "WORKSPACE ENDPOINTS: COMPLETE"
echo "=================================="
