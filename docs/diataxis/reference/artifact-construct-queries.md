# Reference: SPARQL CONSTRUCT Artifact Queries

Complete reference for SPARQL CONSTRUCT query patterns used in artifact creation.

## Query Template

### Basic Artifact CONSTRUCT

```sparql
PREFIX bos: <http://businessos.example.org/ontology#>
PREFIX dc: <http://purl.org/dc/elements/1.1/>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>

CONSTRUCT {
  ?artifact a bos:Artifact ;
    dc:title ?title ;
    bos:type ?type ;
    bos:language ?language ;
    bos:content ?content ;
    bos:summary ?summary ;
    dc:created ?created ;
    bos:createdBy ?createdBy ;
    bos:projectId ?projectId ;
    bos:conversationId ?conversationId .
}
WHERE {
  ?artifact a bos:Artifact ;
    dc:title ?title ;
    bos:type ?type ;
    bos:language ?language ;
    bos:content ?content ;
    bos:summary ?summary ;
    dc:created ?created ;
    bos:createdBy ?createdBy ;
    bos:projectId ?projectId ;
    bos:conversationId ?conversationId .
}
```

## Namespace Prefixes

| Prefix | URI | Purpose |
|--------|-----|---------|
| `bos` | `http://businessos.example.org/ontology#` | BusinessOS domain ontology |
| `dc` | `http://purl.org/dc/elements/1.1/` | Dublin Core (standard metadata) |
| `xsd` | `http://www.w3.org/2001/XMLSchema#` | XML Schema datatypes |
| `rdf` | `http://www.w3.org/1999/02/22-rdf-syntax-ns#` | RDF core (type, comment, etc.) |
| `rdfs` | `http://www.w3.org/2000/01/rdf-schema#` | RDF Schema (labels, comments) |

## Core Properties

### Artifact Class

```sparql
?artifact a bos:Artifact
```

**Instance:**
```turtle
<http://businessos.example.org/artifacts/550e8400-e29b-41d4-a716-446655440000>
  a bos:Artifact .
```

### Dublin Core Properties

| Property | Type | Example | Notes |
|----------|------|---------|-------|
| `dc:title` | String literal | `"Q1 2026 Strategy"` | Required, human-readable name |
| `dc:created` | DateTime | `"2026-03-25T10:30:00Z"^^xsd:dateTime` | ISO 8601 format |
| `dc:creator` | String/URI | `"user-123"` | User ID who created artifact |
| `dc:description` | String | `"Strategic plan"` | Optional summary/description |

### BusinessOS-Specific Properties

| Property | Type | Range | Required? |
|----------|------|-------|-----------|
| `bos:type` | String enum | code, document, markdown, react, html, svg | Yes |
| `bos:language` | String enum | go, typescript, python, markdown, sql, json | No |
| `bos:content` | String (large) | Any text | Yes |
| `bos:summary` | String | 0-500 characters | No |
| `bos:createdBy` | String | User UUID | Yes |
| `bos:projectId` | String | Project UUID | No (optional) |
| `bos:conversationId` | String | Conversation UUID | No (optional) |
| `bos:version` | Integer | ≥1 | Yes (default: 1) |
| `bos:accessCount` | Integer | ≥0 | No (tracking) |

## Query Examples

### Example 1: Simple Document Artifact

**Input Parameters:**
```json
{
  "title": "Q1 Strategy",
  "type": "document",
  "language": "markdown",
  "content": "# Q1 Plan\n\n...",
  "summary": "Strategic plan",
  "user_id": "user-123",
  "artifact_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Generated Query:**
```sparql
PREFIX bos: <http://businessos.example.org/ontology#>
PREFIX dc: <http://purl.org/dc/elements/1.1/>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>

CONSTRUCT {
  ?artifact a bos:Artifact ;
    dc:title "Q1 Strategy"^^xsd:string ;
    bos:type "document"^^xsd:string ;
    bos:language "markdown"^^xsd:string ;
    bos:content "# Q1 Plan\n\n..."^^xsd:string ;
    bos:summary "Strategic plan"^^xsd:string ;
    dc:created "2026-03-25T10:30:00Z"^^xsd:dateTime ;
    bos:createdBy "user-123"^^xsd:string ;
    bos:projectId ""^^xsd:string ;
    bos:conversationId ""^^xsd:string .
}
WHERE {
  BIND(IRI(CONCAT("http://businessos.example.org/artifacts/", "123e4567-e89b-12d3-a456-426614174000")) AS ?artifact)
}
```

**Output (N-Triples):**
```ntriples
<http://businessos.example.org/artifacts/123e4567-e89b-12d3-a456-426614174000> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://businessos.example.org/ontology#Artifact> .
<http://businessos.example.org/artifacts/123e4567-e89b-12d3-a456-426614174000> <http://purl.org/dc/elements/1.1/title> "Q1 Strategy" .
<http://businessos.example.org/artifacts/123e4567-e89b-12d3-a456-426614174000> <http://businessos.example.org/ontology#type> "document" .
<http://businessos.example.org/artifacts/123e4567-e89b-12d3-a456-426614174000> <http://businessos.example.org/ontology#language> "markdown" .
<http://businessos.example.org/artifacts/123e4567-e89b-12d3-a456-426614174000> <http://businessos.example.org/ontology#content> "# Q1 Plan\n\n..." .
<http://businessos.example.org/artifacts/123e4567-e89b-12d3-a456-426614174000> <http://purl.org/dc/elements/1.1/created> "2026-03-25T10:30:00Z" .
<http://businessos.example.org/artifacts/123e4567-e89b-12d3-a456-426614174000> <http://businessos.example.org/ontology#createdBy> "user-123" .
```

**Output (Turtle):**
```turtle
@prefix bos: <http://businessos.example.org/ontology#> .
@prefix dc: <http://purl.org/dc/elements/1.1/> .
@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .

<http://businessos.example.org/artifacts/123e4567-e89b-12d3-a456-426614174000>
  a bos:Artifact ;
  dc:title "Q1 Strategy"^^xsd:string ;
  bos:type "document"^^xsd:string ;
  bos:language "markdown"^^xsd:string ;
  bos:content "# Q1 Plan\n\n..."^^xsd:string ;
  bos:summary "Strategic plan"^^xsd:string ;
  dc:created "2026-03-25T10:30:00Z"^^xsd:dateTime ;
  bos:createdBy "user-123"^^xsd:string ;
  bos:projectId ""^^xsd:string ;
  bos:conversationId ""^^xsd:string .
```

### Example 2: Code Artifact with Project Reference

**Input Parameters:**
```json
{
  "title": "Go Handler",
  "type": "code",
  "language": "go",
  "content": "func main() { ... }",
  "summary": "Main entry point",
  "user_id": "user-456",
  "project_id": "proj-789",
  "artifact_id": "abc12345-def6-7890-ghij-klmnopqr1234"
}
```

**Generated Query:**
```sparql
PREFIX bos: <http://businessos.example.org/ontology#>
PREFIX dc: <http://purl.org/dc/elements/1.1/>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>

CONSTRUCT {
  ?artifact a bos:Artifact ;
    dc:title "Go Handler"^^xsd:string ;
    bos:type "code"^^xsd:string ;
    bos:language "go"^^xsd:string ;
    bos:content "func main() { ... }"^^xsd:string ;
    bos:summary "Main entry point"^^xsd:string ;
    dc:created "2026-03-25T10:30:00Z"^^xsd:dateTime ;
    bos:createdBy "user-456"^^xsd:string ;
    bos:projectId "proj-789"^^xsd:string ;
    bos:conversationId ""^^xsd:string ;
    bos:linksTo ?project .
}
WHERE {
  BIND(IRI(CONCAT("http://businessos.example.org/artifacts/", "abc12345-def6-7890-ghij-klmnopqr1234")) AS ?artifact)
  OPTIONAL { BIND(IRI(CONCAT("http://businessos.example.org/projects/", "proj-789")) AS ?project) }
}
```

**Output (Turtle):**
```turtle
<http://businessos.example.org/artifacts/abc12345-def6-7890-ghij-klmnopqr1234>
  a bos:Artifact ;
  dc:title "Go Handler"^^xsd:string ;
  bos:type "code"^^xsd:string ;
  bos:language "go"^^xsd:string ;
  dc:created "2026-03-25T10:30:00Z"^^xsd:dateTime ;
  bos:createdBy "user-456"^^xsd:string ;
  bos:projectId "proj-789"^^xsd:string ;
  bos:linksTo <http://businessos.example.org/projects/proj-789> .
```

## Query Patterns

### Pattern 1: Basic CONSTRUCT

```sparql
CONSTRUCT {
  ?artifact a bos:Artifact ;
    dc:title ?title ;
    dc:created ?created .
}
WHERE {
  ?artifact a bos:Artifact ;
    dc:title ?title ;
    dc:created ?created .
}
```

**Use Case:** Simple 1:1 mapping from database to RDF

---

### Pattern 2: Property Inference

```sparql
CONSTRUCT {
  ?artifact a bos:Artifact ;
    dc:title ?title ;
    bos:hasVersion ?version .
}
WHERE {
  ?artifact a bos:Artifact ;
    dc:title ?title .
  BIND(1 AS ?version)
}
```

**Use Case:** Add computed properties (e.g., version count, derived status)

---

### Pattern 3: Conditional Properties

```sparql
CONSTRUCT {
  ?artifact a bos:Artifact ;
    dc:title ?title ;
    bos:type ?type ;
    bos:isCode ?isCode .
}
WHERE {
  ?artifact a bos:Artifact ;
    dc:title ?title ;
    bos:type ?type .
  FILTER(?type IN ("code", "react", "html", "svg"))
  BIND(true AS ?isCode)
}
```

**Use Case:** Add derived properties based on conditions

---

### Pattern 4: Property Linking

```sparql
CONSTRUCT {
  ?artifact a bos:Artifact ;
    dc:title ?title ;
    bos:project ?project ;
    bos:creator ?creator .
}
WHERE {
  ?artifact a bos:Artifact ;
    dc:title ?title ;
    bos:projectId ?projectId ;
    bos:createdBy ?userId .
  OPTIONAL { BIND(IRI(CONCAT("http://businessos.example.org/projects/", ?projectId)) AS ?project) }
  OPTIONAL { BIND(IRI(CONCAT("http://businessos.example.org/users/", ?userId)) AS ?creator) }
}
```

**Use Case:** Create references to related resources

---

## Datatype Handling

### String Literals

```sparql
"Hello World"^^xsd:string
```

Escape rules:
- Newlines: `\n`
- Quotes: `\"`
- Backslash: `\\`

**Example:**
```sparql
"Line 1\nLine 2"^^xsd:string
"She said \"hello\""^^xsd:string
```

### DateTime Literals

```sparql
"2026-03-25T10:30:00Z"^^xsd:dateTime
```

Format: ISO 8601 (YYYY-MM-DDTHH:MM:SSZ)

### Integer Literals

```sparql
"1"^^xsd:integer
```

### Boolean Literals

```sparql
"true"^^xsd:boolean
"false"^^xsd:boolean
```

## Query Optimization

### 1. Use BIND for Computed Values

**Instead of:**
```sparql
CONSTRUCT {
  ?artifact bos:createdYear ?year .
}
WHERE {
  ?artifact dc:created ?created .
  # String manipulation in filter
}
```

**Use:**
```sparql
CONSTRUCT {
  ?artifact bos:createdYear ?year .
}
WHERE {
  ?artifact dc:created ?created .
  BIND(YEAR(?created) AS ?year)
}
```

---

### 2. Leverage OPTIONAL for Sparse Properties

```sparql
CONSTRUCT {
  ?artifact a bos:Artifact ;
    dc:title ?title ;
    bos:summary ?summary .
}
WHERE {
  ?artifact a bos:Artifact ;
    dc:title ?title .
  OPTIONAL { ?artifact bos:summary ?summary }
}
```

---

### 3. Index-Friendly Queries

```sparql
CONSTRUCT {
  ?artifact a bos:Artifact .
}
WHERE {
  ?artifact a bos:Artifact ;
    dc:created ?created .
  FILTER(?created > "2026-03-01"^^xsd:dateTime)
}
```

---

## Common Pitfalls

### Pitfall 1: Unbounded Variables

**Wrong:**
```sparql
CONSTRUCT {
  ?artifact ?property ?value .
}
WHERE {
  ?artifact ?property ?value .
}
```

**Risk:** Constructs ALL properties for ALL artifacts

**Right:**
```sparql
CONSTRUCT {
  ?artifact dc:title ?title ;
    dc:created ?created .
}
WHERE {
  ?artifact a bos:Artifact ;
    dc:title ?title ;
    dc:created ?created .
}
```

---

### Pitfall 2: Missing Type Checks

**Wrong:**
```sparql
CONSTRUCT {
  ?x a bos:Artifact ;
    dc:title ?title .
}
WHERE {
  ?x dc:title ?title .
}
```

**Risk:** Matches non-artifacts with titles

**Right:**
```sparql
CONSTRUCT {
  ?artifact a bos:Artifact ;
    dc:title ?title .
}
WHERE {
  ?artifact a bos:Artifact ;
    dc:title ?title .
}
```

---

### Pitfall 3: Literal Type Mismatch

**Wrong:**
```sparql
CONSTRUCT {
  ?artifact dc:created ?created .
}
WHERE {
  ?artifact dc:created "2026-03-25T10:30:00Z" .
}
```

**Risk:** Won't match if stored as xsd:dateTime

**Right:**
```sparql
CONSTRUCT {
  ?artifact dc:created ?created .
}
WHERE {
  ?artifact dc:created ?created .
  FILTER(DATATYPE(?created) = xsd:dateTime)
}
```

---

## Testing CONSTRUCT Queries

### Test with Oxigraph CLI

```bash
# Load data into Oxigraph
./oxigraph load --format turtle my_data.ttl

# Execute CONSTRUCT
./oxigraph query --format turtle my_construct.rq > output.ttl

# Verify output
cat output.ttl | head -20
```

### SPARQL Playground

- **Wikibase Query Service:** https://query.wikidata.org/
- **DBpedia:** https://dbpedia.org/sparql
- **Local Oxigraph:** http://localhost:8080/query

### Online Validators

- **RDF Turtle:** https://www.w3.org/2005/awebb/rdfxml.html
- **SPARQL:** https://query.wikidata.org/sparql

## Performance Metrics

| Operation | Time | Note |
|-----------|------|------|
| CONSTRUCT execution | 100-500ms | Typical for single artifact |
| N-Triples generation | 50-200ms | Format conversion |
| PostgreSQL insert | 10-50ms | Metadata storage |
| Content negotiation | 5-20ms | Format conversion (Turtle/JSON-LD) |
| **Total** | **165-770ms** | End-to-end creation |

---

## References

- [SPARQL 1.1 CONSTRUCT](https://www.w3.org/TR/sparql11-query/#construct)
- [RDF/Turtle](https://www.w3.org/TR/turtle/)
- [XML Schema Datatypes](https://www.w3.org/TR/xmlschema-2/#built-in-datatypes)
- [Dublin Core Metadata](https://purl.org/dc/elements/1.1/)
- [Oxigraph Documentation](https://github.com/oxigraph/oxigraph)
