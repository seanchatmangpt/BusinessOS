# How-To: SPARQL CONSTRUCT Artifact Handler

Create artifacts via SPARQL CONSTRUCT queries, generating RDF triples while storing metadata in PostgreSQL.

## Overview

The artifact CONSTRUCT handler (`artifacts_construct.go`) provides a unified endpoint for creating business artifacts and automatically generating their RDF representation via parameterized SPARQL CONSTRUCT queries.

**Flow:**
```
POST /api/artifacts/construct (JSON)
  ↓
Validate request + generate artifact UUID
  ↓
Build parameterized CONSTRUCT query
  ↓
Execute CONSTRUCT against triplestore
  ↓
Insert metadata into PostgreSQL
  ↓
Return RDF (Turtle/N-Triples/JSON-LD)
```

## API Endpoint

### POST `/api/artifacts/construct`

**Authentication:** Required (Bearer token)

**Content-Type:** `application/json`

**Accept:** `text/turtle`, `application/n-triples`, `application/ld+json`, or `application/json` (default)

### Request Body

```go
type CreateArtifactRequest struct {
    Title         string  `json:"title" binding:"required"`
    Type          string  `json:"type" binding:"required"`           // code, document, markdown, react, html, svg
    Content       string  `json:"content" binding:"required"`
    Language      string  `json:"language"`                          // go, typescript, python, markdown, etc.
    Summary       string  `json:"summary"`                           // optional
    ProjectID     *string `json:"project_id"`                        // optional UUID
    ConversationID *string `json:"conversation_id"`                  // optional UUID
}
```

### Response Body

**Default (application/json):**
```go
type ConstructArtifactResponse struct {
    ArtifactID    string `json:"artifact_id"`    // Generated UUID
    Title         string `json:"title"`
    Type          string `json:"type"`
    RDFTurtle     string `json:"rdf_turtle"`     // Turtle format
    RDFNTIPLES    string `json:"rdf_ntriples"`   // N-Triples format
    RDFJSONLD     string `json:"rdf_jsonld"`     // JSON-LD format
    StoredInDB    bool   `json:"stored_in_db"`   // True if persisted
}
```

**With Accept: text/turtle:**
```turtle
<http://businessos.example.org/artifacts/550e8400-e29b-41d4-a716-446655440000> a bos:Artifact ;
  dc:title "Q1 2026 Business Strategy" ;
  bos:type "document" ;
  bos:language "markdown" ;
  dc:created "2026-03-25T10:30:00Z"^^xsd:dateTime ;
  bos:createdBy "user-123" .
```

## Usage Examples

### Example 1: Create a Markdown Document

**Request:**
```bash
curl -X POST http://localhost:8001/api/artifacts/construct \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "title": "Q1 2026 Business Strategy",
    "type": "document",
    "language": "markdown",
    "content": "# Q1 2026 Strategy\n\n## Objectives\n- Expand AI product offerings\n- Reduce operational costs by 15%\n\n## Timeline\n- Q1: Planning and setup\n- Q2: Implementation",
    "summary": "Strategic plan for Q1 2026",
    "project_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

**Response (200 Created):**
```json
{
  "artifact_id": "123e4567-e89b-12d3-a456-426614174000",
  "title": "Q1 2026 Business Strategy",
  "type": "document",
  "rdf_turtle": "<http://businessos.example.org/artifacts/123e4567-e89b-12d3-a456-426614174000> a bos:Artifact ...",
  "rdf_ntriples": "<http://businessos.example.org/artifacts/123e4567-e89b-12d3-a456-426614174000> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://businessos.example.org/ontology#Artifact> ...",
  "rdf_jsonld": "{\"@context\": ...}",
  "stored_in_db": true
}
```

### Example 2: Create a Code Artifact

**Request:**
```bash
curl -X POST http://localhost:8001/api/artifacts/construct \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: text/turtle" \
  -d '{
    "title": "Go Handler for CONSTRUCT Artifacts",
    "type": "code",
    "language": "go",
    "content": "package handlers\n\nfunc (h *ArtifactConstructHandler) CreateArtifactViaConstruct(c *gin.Context) {\n    // Handler implementation\n}",
    "summary": "Handler function for artifact creation via SPARQL CONSTRUCT",
    "conversation_id": "conv-789"
  }'
```

**Response (200 Created, text/turtle):**
```turtle
@prefix bos: <http://businessos.example.org/ontology#> .
@prefix dc: <http://purl.org/dc/elements/1.1/> .
@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .

<http://businessos.example.org/artifacts/abc12345-def6-7890-ghij-klmnopqr1234>
  a bos:Artifact ;
  dc:title "Go Handler for CONSTRUCT Artifacts"^^xsd:string ;
  bos:type "code"^^xsd:string ;
  bos:language "go"^^xsd:string ;
  dc:created "2026-03-25T10:30:00Z"^^xsd:dateTime ;
  bos:createdBy "user-123"^^xsd:string .
```

### Example 3: Create with JSON-LD Response

**Request:**
```bash
curl -X POST http://localhost:8001/api/artifacts/construct \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/ld+json" \
  -d '{
    "title": "Project Requirements",
    "type": "document",
    "content": "- User authentication\n- Admin dashboard\n- Reporting suite",
    "language": "markdown"
  }'
```

**Response (200 Created, application/ld+json):**
```json
{
  "@context": "http://businessos.example.org/context/artifact",
  "@id": "http://businessos.example.org/artifacts/550e8400-e29b-41d4-a716-446655440000",
  "type": "Artifact",
  "title": "Project Requirements",
  "artifactType": "document",
  "created": "2026-03-25T10:30:00Z"
}
```

## CONSTRUCT Query Template

The handler automatically builds parameterized CONSTRUCT queries:

```sparql
PREFIX bos: <http://businessos.example.org/ontology#>
PREFIX dc: <http://purl.org/dc/elements/1.1/>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>

CONSTRUCT {
  ?artifact a bos:Artifact ;
    dc:title "TITLE"^^xsd:string ;
    bos:type "TYPE"^^xsd:string ;
    bos:language "LANGUAGE"^^xsd:string ;
    bos:content "CONTENT"^^xsd:string ;
    bos:summary "SUMMARY"^^xsd:string ;
    dc:created "CREATED_AT"^^xsd:dateTime ;
    bos:createdBy "USER_ID"^^xsd:string ;
    bos:projectId "PROJECT_ID"^^xsd:string ;
    bos:conversationId "CONVERSATION_ID"^^xsd:string .
}
WHERE {
  BIND(IRI(CONCAT("http://businessos.example.org/artifacts/", "ARTIFACT_ID")) AS ?artifact)
}
```

**Parameters:**
- `ARTIFACT_ID`: Generated UUID (e.g., `550e8400-e29b-41d4-a716-446655440000`)
- `TITLE`: User-provided title (escaped for RDF literal)
- `TYPE`: Artifact type (code, document, markdown, react, html, svg)
- `LANGUAGE`: Programming language or format (go, typescript, python, markdown)
- `CONTENT`: Full artifact content (escaped for RDF literal)
- `SUMMARY`: Optional summary (escaped for RDF literal)
- `CREATED_AT`: ISO 8601 timestamp with timezone
- `USER_ID`: ID of creating user
- `PROJECT_ID`: Optional project UUID
- `CONVERSATION_ID`: Optional conversation UUID

## Integration

### Register Handler in Router

```go
// main.go or routes.go
func setupRoutes(engine *gin.Engine, db *pgxpool.Pool, bosService *services.BosOntologyService) {
    api := engine.Group("/api")
    api.Use(authMiddleware())

    // Artifact CONSTRUCT handler
    artifactConstructHandler := handlers.NewArtifactConstructHandler(db, bosService)
    artifacts := api.Group("/artifacts")
    {
        artifacts.POST("/construct", artifactConstructHandler.CreateArtifactViaConstruct)
    }
}
```

### With Authentication Middleware

```go
func authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        user := middleware.GetCurrentUser(c)
        if user == nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## Data Flow

### 1. Request Validation
- Verify user is authenticated
- Validate required fields (title, type, content)
- Parse project_id and conversation_id (optional)

### 2. Generate Artifact ID
- Create new UUID for artifact
- Assign creation timestamp

### 3. Build CONSTRUCT Query
- Create parameterized SPARQL CONSTRUCT
- Include all artifact properties as RDF triples
- Prepare for execution against triplestore

### 4. Execute CONSTRUCT
- Execute query against Oxigraph triplestore
- Receive N-Triples output
- Convert to requested format (Turtle, JSON-LD)

### 5. Store in PostgreSQL
- Insert artifact metadata into `artifacts` table
- Store user_id, project_id, conversation_id (foreign keys)
- Set version = 1

### 6. Return Response
- Content negotiation: return RDF in requested format
- Default: JSON response with all RDF formats
- Set HTTP 201 (Created) status

## Testing

### Unit Tests

```bash
cd BusinessOS/desktop/backend-go
go test ./internal/handlers/... -run TestCreateArtifactViaConstruct -v
go test ./internal/handlers/... -run TestBuildArtifactConstructQuery -v
go test ./internal/handlers/... -run TestConstructArtifactQueryParams -v
```

### Integration Test

```bash
# Start services
make dev

# Create artifact via CONSTRUCT
curl -X POST http://localhost:8001/api/artifacts/construct \
  -H "Authorization: Bearer $(get_test_token)" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Artifact",
    "type": "document",
    "content": "Test content"
  }'

# Verify in PostgreSQL
psql $DATABASE_URL -c "SELECT id, title, type FROM artifacts ORDER BY created_at DESC LIMIT 1;"
```

## Error Handling

### 400 Bad Request
**Cause:** Missing required fields or invalid JSON

```json
{
  "error": "Invalid request body"
}
```

**Fix:**
- Ensure `title`, `type`, `content` are provided
- Verify JSON is valid
- Check Content-Type header

### 401 Unauthorized
**Cause:** Missing or invalid authentication token

```json
{
  "error": "Unauthorized"
}
```

**Fix:**
- Add `Authorization: Bearer YOUR_TOKEN` header
- Verify token is not expired

### 500 Internal Server Error
**Cause:** CONSTRUCT execution failed or database error

```json
{
  "error": "CONSTRUCT execution failed"
}
```

**Fix:**
- Check triplestore is running on port 8080 (default Oxigraph)
- Verify PostgreSQL connection is valid
- Check logs for detailed error message

## Content Negotiation

The handler supports multiple response formats:

| Accept Header | Response Type | Use Case |
|--------------|--------------|----------|
| `text/turtle` | Turtle RDF | Human-readable RDF, SPARQL tools |
| `application/n-triples` | N-Triples | Streaming, integration |
| `application/ld+json` | JSON-LD | Web applications, JSON APIs |
| `application/json` (default) | JSON | Full response with all RDF formats |

**Example:**
```bash
# Request Turtle format
curl -H "Accept: text/turtle" http://localhost:8001/api/artifacts/construct

# Request JSON-LD
curl -H "Accept: application/ld+json" http://localhost:8001/api/artifacts/construct
```

## Extending the Handler

### Add Custom RDF Properties

Edit `buildArtifactConstructQuery()` to include additional properties:

```go
func buildArtifactConstructQuery(params ConstructArtifactQueryParams) string {
    // Add custom property:
    // bos:customField "%s"^^xsd:string ;

    // Extend WHERE clause with filters:
    // WHERE {
    //   ...filter by date range, user, etc.
    // }
}
```

### Add Pre-Processing

Transform artifact content before CONSTRUCT:

```go
// Markdown → HTML conversion
if req.Type == "markdown" {
    req.Content = markdownToHTML(req.Content)
}

// Code formatting
if req.Type == "code" {
    req.Content = formatCode(req.Content, req.Language)
}
```

### Add Post-Processing

Convert RDF after CONSTRUCT execution:

```go
// N-Triples → Turtle
if acceptHeader == "text/turtle" {
    rdfTurtle, err := convertNTriplesToTurtle(rdfNTriples)
    c.String(http.StatusCreated, rdfTurtle)
}
```

## Performance Considerations

### Query Execution Time
- CONSTRUCT execution: 100-500ms (typical)
- N-Triples generation: 50-200ms
- PostgreSQL insert: 10-50ms
- **Total:** 160-750ms per request

### Optimization Strategies
1. **Connection pooling:** Use `pgxpool` (configured)
2. **Query caching:** Cache frequently executed queries
3. **Batch inserts:** Group multiple artifacts when possible
4. **Async CONSTRUCT:** Execute CONSTRUCT asynchronously for bulk operations

### Backpressure
- Max concurrent requests: 10 (semaphore-controlled)
- Request timeout: 15 seconds
- Database connection pool: 100 connections

## References

- SPARQL CONSTRUCT: https://www.w3.org/TR/sparql11-query/#construct
- RDF/Turtle: https://www.w3.org/TR/turtle/
- JSON-LD: https://json-ld.org/
- Oxigraph: https://github.com/oxigraph/oxigraph
- Gin HTTP Framework: https://github.com/gin-gonic/gin
