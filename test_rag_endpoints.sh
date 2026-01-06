#!/bin/bash

BASE_URL="http://localhost:8001/api"
TOKEN="${1:-test-token}"

echo "=================================="
echo "TESTING RAG ENDPOINTS"
echo "=================================="
echo ""

# Test 1: Hybrid Search
echo "1. POST /rag/search/hybrid - Hybrid search"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/rag/search/hybrid" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "authentication best practices",
    "semantic_weight": 0.7,
    "keyword_weight": 0.3,
    "max_results": 5
  }')
http_code=$(echo "$response" | tail -n1)

if [ "$http_code" = "200" ] || [ "$http_code" = "401" ] || [ "$http_code" = "503" ]; then
  echo "   Status: $http_code"
  echo "   ✓ Endpoint responding"
else
  echo "   Status: $http_code"
  echo "   ✗ Unexpected status"
fi
echo ""

# Test 2: Agentic RAG
echo "2. POST /rag/retrieve - Agentic RAG retrieval"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/rag/retrieve" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How to implement authentication?",
    "max_results": 5,
    "use_personalization": true
  }')
http_code=$(echo "$response" | tail -n1)

if [ "$http_code" = "200" ] || [ "$http_code" = "401" ] || [ "$http_code" = "503" ]; then
  echo "   Status: $http_code"
  echo "   ✓ Endpoint responding"
else
  echo "   Status: $http_code"
  echo "   ✗ Unexpected status"
fi
echo ""

# Test 3: Search explanation
echo "3. GET /rag/search/explain - Search explanation"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/rag/search/explain?query=test" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json")
http_code=$(echo "$response" | tail -n1)

if [ "$http_code" = "200" ] || [ "$http_code" = "401" ] || [ "$http_code" = "503" ]; then
  echo "   Status: $http_code"
  echo "   ✓ Endpoint responding"
else
  echo "   Status: $http_code"
  echo "   ✗ Unexpected status"
fi
echo ""

echo "=================================="
echo "RAG ENDPOINTS: COMPLETE"
echo "=================================="
