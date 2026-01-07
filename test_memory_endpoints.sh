#!/bin/bash

BASE_URL="http://localhost:8001/api"
TOKEN="${1:-test-token}"
WORKSPACE_ID="550e8400-e29b-41d4-a716-446655440000"

echo "=================================="
echo "TESTING MEMORY ENDPOINTS"
echo "=================================="
echo ""

# Test 1: List workspace memories
echo "1. GET /workspaces/:id/memories - List workspace memories"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/workspaces/$WORKSPACE_ID/memories" \
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

# Test 2: List private memories
echo "2. GET /workspaces/:id/memories/private - List private memories"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/workspaces/$WORKSPACE_ID/memories/private" \
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

# Test 3: List accessible memories
echo "3. GET /workspaces/:id/memories/accessible - List all accessible"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/workspaces/$WORKSPACE_ID/memories/accessible" \
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

# Test 4: Create workspace memory
echo "4. POST /workspaces/:id/memories - Create workspace memory"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/workspaces/$WORKSPACE_ID/memories" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Memory",
    "summary": "Test memory for API testing",
    "content": "This is a test memory content",
    "memory_type": "general",
    "visibility": "workspace",
    "tags": ["test", "api"]
  }')
http_code=$(echo "$response" | tail -n1)

if [ "$http_code" = "201" ] || [ "$http_code" = "401" ] || [ "$http_code" = "403" ]; then
  echo "   Status: $http_code"
  echo "   ✓ Endpoint responding"
else
  echo "   Status: $http_code"
  echo "   ✗ Unexpected status"
fi
echo ""

# Test 5: Create private memory
echo "5. POST /workspaces/:id/memories - Create private memory"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/workspaces/$WORKSPACE_ID/memories" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Private Test Memory",
    "summary": "Private memory for testing",
    "content": "Private content",
    "memory_type": "general",
    "visibility": "private",
    "tags": ["private", "test"]
  }')
http_code=$(echo "$response" | tail -n1)

if [ "$http_code" = "201" ] || [ "$http_code" = "401" ] || [ "$http_code" = "403" ]; then
  echo "   Status: $http_code"
  echo "   ✓ Endpoint responding"
else
  echo "   Status: $http_code"
  echo "   ✗ Unexpected status"
fi
echo ""

echo "=================================="
echo "MEMORY ENDPOINTS: COMPLETE"
echo "=================================="
