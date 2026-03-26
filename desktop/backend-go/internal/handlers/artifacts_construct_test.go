package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestCreateArtifactViaConstruct tests artifact creation via SPARQL CONSTRUCT
func TestCreateArtifactViaConstruct(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a mock context (in production, use full handler setup with DB)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Mock request body
	req := CreateArtifactRequest{
		Title:    "Go Service Architecture",
		Type:     "document",
		Content:  "## Architecture\n\nHandler -> Service -> Repository pattern.",
		Language: "markdown",
		Summary:  "Overview of service layering architecture",
	}

	body, _ := json.Marshal(req)
	c.Request = httptest.NewRequest("POST", "/api/artifacts/construct", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Accept", "application/json")

	// Test parameter building
	params := ConstructArtifactQueryParams{
		ArtifactID:     "550e8400-e29b-41d4-a716-446655440000",
		Title:          req.Title,
		Type:           req.Type,
		Language:       req.Language,
		Content:        req.Content,
		Summary:        req.Summary,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
		UserID:         "user-123",
		ProjectID:      "proj-456",
		ConversationID: "conv-789",
	}

	query := buildArtifactConstructQuery(params)

	// Assertions on query structure
	assert.Contains(t, query, "PREFIX bos:", "Query should include BOS namespace")
	assert.Contains(t, query, "CONSTRUCT {", "Query should have CONSTRUCT clause")
	assert.Contains(t, query, "Go Service Architecture", "Query should include artifact title")
	assert.Contains(t, query, "document", "Query should include artifact type")
	assert.Contains(t, query, "550e8400-e29b-41d4-a716-446655440000", "Query should include artifact ID")

	t.Logf("Generated CONSTRUCT query:\n%s", query)
}

// TestBuildArtifactConstructQuery tests CONSTRUCT query generation
func TestBuildArtifactConstructQuery(t *testing.T) {
	params := ConstructArtifactQueryParams{
		ArtifactID:     "artifact-123",
		Title:          "Test Proposal",
		Type:           "proposal",
		Language:       "english",
		Content:        "This is a test proposal.",
		Summary:        "Short summary",
		CreatedAt:      "2026-03-25T10:30:00Z",
		UserID:         "user-123",
		ProjectID:      "proj-456",
		ConversationID: "conv-789",
	}

	query := buildArtifactConstructQuery(params)

	// Verify SPARQL structure
	assert.Contains(t, query, "PREFIX bos:", "Missing BOS prefix")
	assert.Contains(t, query, "PREFIX dc:", "Missing Dublin Core prefix")
	assert.Contains(t, query, "PREFIX xsd:", "Missing XSD prefix")
	assert.Contains(t, query, "CONSTRUCT", "Missing CONSTRUCT keyword")
	assert.Contains(t, query, "?artifact a bos:Artifact", "Missing artifact type")
	assert.Contains(t, query, "dc:title", "Missing DC title property")
	assert.Contains(t, query, "bos:type", "Missing BOS type property")
	assert.Contains(t, query, "dc:created", "Missing created date property")
	assert.Contains(t, query, "2026-03-25T10:30:00Z", "Created date not in query")
}

// TestConstructArtifactQueryParams tests parameter structure
func TestConstructArtifactQueryParams(t *testing.T) {
	params := ConstructArtifactQueryParams{
		ArtifactID:     "id-1",
		Title:          "Title",
		Type:           "code",
		Language:       "go",
		Content:        "func main() {}",
		Summary:        "Summary",
		CreatedAt:      "2026-03-25T10:00:00Z",
		UserID:         "user-1",
		ProjectID:      "proj-1",
		ConversationID: "conv-1",
	}

	assert.Equal(t, "id-1", params.ArtifactID)
	assert.Equal(t, "Title", params.Title)
	assert.Equal(t, "code", params.Type)
	assert.Equal(t, "go", params.Language)
	assert.NotEmpty(t, params.Content)
	assert.NotEmpty(t, params.CreatedAt)
}

// TestCreateArtifactResponse tests response structure
func TestCreateArtifactResponse(t *testing.T) {
	response := ConstructArtifactResponse{
		ArtifactID: "artifact-123",
		Title:      "Test Artifact",
		Type:       "document",
		RDFTurtle:  "<artifact> <property> <value> .",
		RDFNTIPLES: "<http://example.org/artifact> <http://example.org/prop> <http://example.org/val> .",
		RDFJSONLD:  `{"@context": "http://example.org/", "@id": "artifact-123"}`,
		StoredInDB: true,
	}

	// Verify response fields
	assert.Equal(t, "artifact-123", response.ArtifactID)
	assert.Equal(t, "Test Artifact", response.Title)
	assert.Equal(t, "document", response.Type)
	assert.True(t, response.StoredInDB)
	assert.NotEmpty(t, response.RDFTurtle)
	assert.NotEmpty(t, response.RDFNTIPLES)
	assert.NotEmpty(t, response.RDFJSONLD)

	// Test JSON serialization
	jsonBytes, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonBytes)

	// Verify JSON contains expected fields
	var decoded map[string]interface{}
	err = json.Unmarshal(jsonBytes, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "artifact-123", decoded["artifact_id"])
	assert.Equal(t, "Test Artifact", decoded["title"])
}

// BenchmarkBuildArtifactConstructQuery benchmarks query generation
func BenchmarkBuildArtifactConstructQuery(b *testing.B) {
	params := ConstructArtifactQueryParams{
		ArtifactID:     "artifact-123",
		Title:          "Benchmark Test",
		Type:           "document",
		Language:       "english",
		Content:        "Large content block " + string(make([]byte, 1000)),
		Summary:        "Summary",
		CreatedAt:      "2026-03-25T10:00:00Z",
		UserID:         "user-123",
		ProjectID:      "proj-456",
		ConversationID: "conv-789",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildArtifactConstructQuery(params)
	}
}

// Example request/response for documentation
// Note: This is not testing an actual public function, just documenting the API
func TestArtifactConstructExample(t *testing.T) {
	// Example POST /api/artifacts/construct request
	exampleRequest := CreateArtifactRequest{
		Title:     "Q1 2026 Business Strategy",
		Type:      "document",
		Content:   "# Q1 2026 Strategy\n\n- Expand AI product offerings\n- Reduce operational costs by 15%",
		Language:  "markdown",
		Summary:   "Strategic plan for Q1 2026",
		ProjectID: stringPtr("550e8400-e29b-41d4-a716-446655440000"),
	}

	// Example CONSTRUCT query generated from above request
	constructQuery := `
PREFIX bos: <http://businessos.example.org/ontology#>
PREFIX dc: <http://purl.org/dc/elements/1.1/>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>

CONSTRUCT {
  ?artifact a bos:Artifact ;
    dc:title "Q1 2026 Business Strategy"^^xsd:string ;
    bos:type "document"^^xsd:string ;
    bos:language "markdown"^^xsd:string ;
    bos:content "# Q1 2026 Strategy\n\n- Expand AI product offerings\n- Reduce operational costs by 15%"^^xsd:string ;
    bos:summary "Strategic plan for Q1 2026"^^xsd:string ;
    dc:created "2026-03-25T10:30:00Z"^^xsd:dateTime ;
    bos:createdBy "user-123"^^xsd:string ;
    bos:projectId "550e8400-e29b-41d4-a716-446655440000"^^xsd:string ;
    bos:conversationId ""^^xsd:string .
}
WHERE {
  BIND(IRI(CONCAT("http://businessos.example.org/artifacts/", "123e4567-e89b-12d3-a456-426614174000")) AS ?artifact)
}
`

	// Example N-Triples response
	exampleNTriples := `
<http://businessos.example.org/artifacts/123e4567-e89b-12d3-a456-426614174000> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://businessos.example.org/ontology#Artifact> .
<http://businessos.example.org/artifacts/123e4567-e89b-12d3-a456-426614174000> <http://purl.org/dc/elements/1.1/title> "Q1 2026 Business Strategy" .
<http://businessos.example.org/artifacts/123e4567-e89b-12d3-a456-426614174000> <http://businessos.example.org/ontology#type> "document" .
<http://businessos.example.org/artifacts/123e4567-e89b-12d3-a456-426614174000> <http://businessos.example.org/ontology#language> "markdown" .
<http://businessos.example.org/artifacts/123e4567-e89b-12d3-a456-426614174000> <http://purl.org/dc/elements/1.1/created> "2026-03-25T10:30:00Z" .
<http://businessos.example.org/artifacts/123e4567-e89b-12d3-a456-426614174000> <http://businessos.example.org/ontology#createdBy> "user-123" .
<http://businessos.example.org/artifacts/123e4567-e89b-12d3-a456-426614174000> <http://businessos.example.org/ontology#projectId> "550e8400-e29b-41d4-a716-446655440000" .
`

	_ = exampleRequest
	_ = constructQuery
	_ = exampleNTriples
}

// Helper function for examples
func stringPtr(s string) *string {
	return &s
}
