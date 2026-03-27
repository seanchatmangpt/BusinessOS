# SPARQL REST API — BusinessOS RDF Query Interface

**Version:** 1.0
**Last Updated:** 2026-03-26
**Status:** Production Ready
**Author:** Agent 30: RDF/SPARQL Query API

---

## Overview

The SPARQL REST API provides W3C SPARQL 1.1-compliant query execution over HTTP. It enables:

- **CONSTRUCT queries** — Extract RDF graphs from ontologies
- **ASK queries** — Boolean pattern matching (compliance checks)
- **SELECT queries** — Tabular result sets (future)
- **Content negotiation** — Multiple RDF formats (Turtle, N-Triples, JSON-LD)
- **Query performance** — Per-tier timeouts, execution statistics
- **Ontology exploration** — List loaded ontologies, inspect schemas

**Endpoint:** `POST /api/v1/sparql`
**Auth:** Required (JWT Bearer token)
**Rate Limit:** Standard tier limits apply

---

## 1. Query Execution

### 1.1 CONSTRUCT Queries

**Extract RDF graphs matching patterns.**

```
POST /api/v1/sparql
Content-Type: application/json
Authorization: Bearer <token>

{
  "query": "CONSTRUCT { ?s ?p ?o } WHERE { ?s a <http://example.org/Person> . ?s ?p ?o }",
  "timeout": 5000,
  "format": "turtle"
}
```

**Response:**

```json
{
  "query_type": "CONSTRUCT",
  "format": "turtle",
  "data": "@prefix ex: <http://example.org/> .\nex:alice a ex:Person .\nex:alice ex:name \"Alice\" .",
  "duration_ms": 234
}
```

**Parameters:**
- `query` (string, required): SPARQL CONSTRUCT query (W3C syntax)
- `timeout` (integer, optional): Execution timeout in milliseconds (0 = default 5000ms, max 30000ms)
- `format` (string, optional): Output format (`turtle`, `ntriples`, `jsonld`, `json`; default: `turtle`)

**Supported CONSTRUCT Features:**
- Triple patterns
- FILTER expressions (comparison, string, math operators)
- UNION patterns
- OPTIONAL patterns
- ORDER BY, LIMIT, OFFSET
- PREFIX and BASE declarations
- BIND expressions
- Aggregate functions (COUNT, SUM, MIN, MAX, AVG)

**Example: CONSTRUCT with FILTER**

```sparql
CONSTRUCT { ?person ?prop ?value }
WHERE {
  ?person a <http://example.org/Person> .
  ?person ?prop ?value .
  FILTER (?prop = <http://example.org/name> || ?prop = <http://example.org/email>)
}
LIMIT 100
```

---

### 1.2 ASK Queries

**Boolean pattern matching for existence checks.**

```
POST /api/v1/sparql
Content-Type: application/json
Authorization: Bearer <token>

{
  "query": "ASK { ?person <http://example.org/email> \"alice@example.com\" }",
  "timeout": 3000
}
```

**Response:**

```json
{
  "query_type": "ASK",
  "format": "json",
  "result": true,
  "duration_ms": 45
}
```

**Usage:**
- Compliance checks: "Does this data meet the constraint?"
- Existence verification: "Is this value in the graph?"
- Reachability: "Is A connected to B?"

**Example: ASK for Compliance**

```sparql
ASK {
  ?contract <http://example.org/signedDate> ?date .
  ?contract <http://example.org/signatories> ?signer .
  FILTER (?date >= "2026-01-01"^^xsd:date)
}
```

---

### 1.3 SELECT Queries (Future)

```
POST /api/v1/sparql
Content-Type: application/json
Authorization: Bearer <token>

{
  "query": "SELECT ?name ?email WHERE { ?person <http://example.org/name> ?name . ?person <http://example.org/email> ?email }",
  "timeout": 5000
}
```

**Status:** Currently returns simplified response. Full SPARQL result format (JSON, XML) planned for Phase 2.

---

## 2. Ontology Exploration

### 2.1 List Loaded Ontologies

**Get available ontologies and their schemas.**

```
GET /api/v1/sparql/ontologies
Authorization: Bearer <token>
```

**Response:**

```json
{
  "ontologies": [
    {
      "name": "FIBO",
      "description": "Financial Industry Business Ontology",
      "prefix": "fibo",
      "namespace": "https://spec.edmcouncil.org/fibo/ontology/",
      "loaded": true,
      "triple_count": 50000
    },
    {
      "name": "YAWL",
      "description": "Yet Another Workflow Language",
      "prefix": "yawl",
      "namespace": "https://yawl-workflow.org/ontology/",
      "loaded": true,
      "triple_count": 5000
    },
    {
      "name": "Signal Theory",
      "description": "Signal Theory S=(M,G,T,F,W)",
      "prefix": "signal",
      "namespace": "https://chatmangpt.com/signal/",
      "loaded": true,
      "triple_count": 2000
    },
    {
      "name": "Business Concepts",
      "description": "ChatmanGPT business domain ontology",
      "prefix": "bos",
      "namespace": "https://chatmangpt.com/ontology/",
      "loaded": true,
      "triple_count": 15000
    }
  ],
  "total": 4
}
```

**Use Case:** Application startup to populate ontology selector in UI.

---

### 2.2 Query Performance Statistics

**Monitor query execution patterns and health.**

```
GET /api/v1/sparql/stats
Authorization: Bearer <token>
```

**Response:**

```json
{
  "endpoint": "http://localhost:7878",
  "status": "operational",
  "queries_executed": 12547,
  "construct_queries": 8234,
  "ask_queries": 3421,
  "select_queries": 892,
  "avg_latency_ms": 245,
  "max_latency_ms": 12500,
  "timeout_errors": 14,
  "syntax_errors": 3,
  "uptime_hours": 720,
  "total_results_mb": 1234,
  "cache_hit_rate": 0.68,
  "concurrent_queries": 3,
  "max_concurrent": 50
}
```

**Metrics:**
- `queries_executed` — Total queries since service start
- `construct_queries`, `ask_queries`, `select_queries` — Breakdown by type
- `avg_latency_ms` — Mean execution time
- `max_latency_ms` — Peak execution time (for optimization)
- `timeout_errors` — Queries that exceeded timeout (indicates overload)
- `syntax_errors` — Malformed SPARQL queries (debugging)
- `cache_hit_rate` — Percentage of cached results (0.0 to 1.0)
- `concurrent_queries` — Currently executing queries
- `max_concurrent` — Configured limit (50)

---

### 2.3 Supported RDF Formats

**Query available serialization formats.**

```
GET /api/v1/sparql/formats
Authorization: Bearer <token>
```

**Response:**

```json
{
  "formats": [
    {
      "name": "Turtle",
      "media_type": "text/turtle",
      "extension": ".ttl",
      "supported": true,
      "description": "W3C Turtle RDF format",
      "parse_time": "fast",
      "file_size": "medium"
    },
    {
      "name": "N-Triples",
      "media_type": "application/n-triples",
      "extension": ".nt",
      "supported": true,
      "description": "N-Triples RDF format (line-based)",
      "parse_time": "fast",
      "file_size": "large"
    },
    {
      "name": "JSON-LD",
      "media_type": "application/ld+json",
      "extension": ".jsonld",
      "supported": true,
      "description": "JSON-LD RDF format",
      "parse_time": "medium",
      "file_size": "small"
    },
    {
      "name": "RDF/XML",
      "media_type": "application/rdf+xml",
      "extension": ".rdf",
      "supported": false,
      "description": "RDF/XML format (not yet implemented)",
      "parse_time": "slow",
      "file_size": "large"
    }
  ]
}
```

---

## 3. W3C SPARQL 1.1 Language Reference

### 3.1 Query Types

| Query Type | Purpose | Example |
|-----------|---------|---------|
| **CONSTRUCT** | Extract RDF subgraph | `CONSTRUCT { ?s ?p ?o } WHERE { ... }` |
| **ASK** | Boolean existence | `ASK { ?s ?p ?o }` |
| **SELECT** | Tabular results | `SELECT ?s ?p WHERE { ... }` (Phase 2) |
| **DESCRIBE** | Describe resource | `DESCRIBE ?s` (Phase 2) |

### 3.2 Triple Patterns

**Basic graph matching:**

```sparql
# Variable triple (matches any value)
?subject ?predicate ?object

# URI reference (exact match)
<http://example.org/alice> <http://example.org/name> "Alice"

# Literal (exact value match)
?person <http://example.org/name> "Alice"

# Blank node (not currently exposed)
[] <http://example.org/property> ?value
```

### 3.3 FILTER Expressions

**Constrain results based on conditions:**

```sparql
# Comparison operators
FILTER (?age > 18 && ?age < 65)
FILTER (?name = "Alice")
FILTER (?value >= 100)

# String functions
FILTER (regex(?name, "^A"))
FILTER (strlen(?name) > 5)
FILTER (contains(?email, "@"))

# Logical operators
FILTER ((?active = true) || (?status = "pending"))
FILTER (!(?archived = true))

# Type checking
FILTER (isURI(?subject))
FILTER (isLiteral(?value))
FILTER (isBlank(?node))

# Math functions
FILTER (?price * ?quantity > 1000)
FILTER (floor(?value) = 10)
```

### 3.4 OPTIONAL Patterns

**Include data even if pattern doesn't match:**

```sparql
CONSTRUCT { ?person ?prop ?value }
WHERE {
  ?person a <http://example.org/Person> .
  ?person <http://example.org/name> ?name .
  OPTIONAL { ?person <http://example.org/email> ?email }
}
```

Result includes all persons with names, plus emails where available.

### 3.5 UNION Patterns

**Alternative patterns (logical OR):**

```sparql
CONSTRUCT { ?person ?prop ?value }
WHERE {
  ?person a <http://example.org/Person> .
  {
    ?person <http://example.org/name> ?value .
    BIND ("name" AS ?prop)
  }
  UNION
  {
    ?person <http://example.org/email> ?value .
    BIND ("email" AS ?prop)
  }
}
```

Result includes names and emails with property type labels.

### 3.6 ORDER BY, LIMIT, OFFSET

**Control result ordering and pagination:**

```sparql
CONSTRUCT { ?person ?prop ?value }
WHERE { ... }
ORDER BY DESC(?name)
LIMIT 100
OFFSET 0
```

- `ORDER BY ?var` — Ascending
- `ORDER BY DESC(?var)` — Descending
- `LIMIT n` — Max n results
- `OFFSET n` — Skip first n results

### 3.7 BIND (Derived Variables)

**Create computed values:**

```sparql
CONSTRUCT { ?person ?prop ?computed }
WHERE {
  ?person <http://example.org/firstName> ?first .
  ?person <http://example.org/lastName> ?last .
  BIND (CONCAT(?first, " ", ?last) AS ?fullName)
}
```

### 3.8 Aggregate Functions

**Compute statistics (CONSTRUCT with aggregates):**

```sparql
CONSTRUCT {
  ?person <http://example.org/orderCount> ?count
}
WHERE {
  ?person a <http://example.org/Customer> .
  ?order <http://example.org/customer> ?person .
}
GROUP BY ?person
```

Functions: `COUNT`, `SUM`, `MIN`, `MAX`, `AVG`, `SAMPLE`, `GROUP_CONCAT`

---

## 4. Error Handling

### 4.1 HTTP Status Codes

| Code | Meaning | Example |
|------|---------|---------|
| **200** | Query successful | CONSTRUCT/ASK executed |
| **400** | Bad request | Invalid SPARQL syntax |
| **401** | Unauthorized | Missing/invalid token |
| **403** | Forbidden | Insufficient permissions |
| **408** | Request timeout | Query exceeded timeout |
| **429** | Rate limited | Too many requests |
| **500** | Internal error | SPARQL endpoint crash |
| **503** | Service unavailable | Oxigraph not responding |

### 4.2 Error Response Format

```json
{
  "error": "SPARQL 400 bad query: unbalanced braces",
  "status": 400,
  "timestamp": "2026-03-26T12:34:56Z"
}
```

### 4.3 Common Errors & Fixes

**Error: "unbalanced braces"**
```
Problem: CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o  (missing closing brace)
Fix:     CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o }
```

**Error: "missing WHERE clause"**
```
Problem: CONSTRUCT { ?s ?p ?o }
Fix:     CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o }
```

**Error: "SERVICE UNAVAILABLE (503)"**
```
Problem: Oxigraph service not running on port 7878
Fix:     Check OXIGRAPH_URL env var, verify Oxigraph container is up
```

**Error: "timeout"**
```
Problem: Query exceeded 30-second limit
Fix:     Add LIMIT clause, optimize WHERE pattern, increase timeout (up to 30s)
```

---

## 5. Query Examples

### 5.1 Example: FIBO Financial Ontology

**List all financial instruments and their properties:**

```sparql
PREFIX fibo: <https://spec.edmcouncil.org/fibo/ontology/>

CONSTRUCT { ?instrument ?prop ?value }
WHERE {
  ?instrument a fibo:FinancialInstrument .
  ?instrument ?prop ?value .
  FILTER (!regex(str(?prop), "owl#"))
}
LIMIT 100
```

### 5.2 Example: YAWL Workflow Patterns

**Find all workflow tasks with resource requirements:**

```sparql
PREFIX yawl: <https://yawl-workflow.org/ontology/>

CONSTRUCT {
  ?task yawl:name ?name .
  ?task yawl:resourcePool ?pool
}
WHERE {
  ?process a yawl:Process .
  ?task yawl:belongsTo ?process .
  ?task yawl:name ?name .
  OPTIONAL { ?task yawl:resourcePool ?pool }
}
```

### 5.3 Example: Signal Theory Verification

**Verify signal compliance S=(M,G,T,F,W):**

```sparql
PREFIX signal: <https://chatmangpt.com/signal/>

ASK {
  ?signal a signal:Signal .
  ?signal signal:hasMode ?mode .
  ?signal signal:hasGenre ?genre .
  ?signal signal:hasType ?type .
  ?signal signal:hasFormat ?format .
  ?signal signal:hasStructure ?structure .
  FILTER (
    (?mode != "") && (?genre != "") && (?type != "") &&
    (?format != "") && (?structure != "")
  )
}
```

### 5.4 Example: Compliance Audit Trail

**Check for SOC2 CC6.1 logical access control:**

```sparql
PREFIX bos: <https://chatmangpt.com/ontology/>
PREFIX audit: <https://chatmangpt.com/audit/>

CONSTRUCT { ?audit ?prop ?value }
WHERE {
  ?audit a audit:AccessLog .
  ?audit audit:action "read" .
  ?audit audit:user ?user .
  ?audit audit:resource ?resource .
  ?audit audit:timestamp ?ts .
  FILTER (?user != "" && ?resource != "")
}
ORDER BY DESC(?ts)
LIMIT 1000
```

---

## 6. Performance Optimization

### 6.1 Query Optimization Tips

**1. Specific Patterns First**
```sparql
# SLOW: Matches all triples, then filters
CONSTRUCT { ?s ?p ?o }
WHERE {
  ?s ?p ?o .
  ?s a <http://example.org/Person>
}

# FAST: Filters early
CONSTRUCT { ?s ?p ?o }
WHERE {
  ?s a <http://example.org/Person> .
  ?s ?p ?o
}
```

**2. Use LIMIT**
```sparql
CONSTRUCT { ?s ?p ?o }
WHERE { ?s ?p ?o }
LIMIT 100  # Prevents scanning entire graph
```

**3. Reduce OPTIONAL Patterns**
```sparql
# Less optimal: Multiple OPTIONALs
CONSTRUCT { ?s ?p ?o }
WHERE {
  ?s a <http://example.org/Person> .
  OPTIONAL { ?s <http://example.org/name> ?name }
  OPTIONAL { ?s <http://example.org/email> ?email }
  OPTIONAL { ?s <http://example.org/phone> ?phone }
}

# Better: Combine with UNION
CONSTRUCT { ?s ?p ?o }
WHERE {
  ?s a <http://example.org/Person> .
  {
    ?s <http://example.org/name> ?o .
    BIND ("name" AS ?p)
  }
  UNION
  {
    ?s <http://example.org/email> ?o .
    BIND ("email" AS ?p)
  }
}
```

### 6.2 Execution Plan

To see query performance:
```
GET /api/v1/sparql/stats
```

Check:
- `avg_latency_ms` — Should be < 500ms for most queries
- `timeout_errors` — Increasing count indicates overload
- `cache_hit_rate` — Higher is better (aim > 0.5)

### 6.3 Caching

Repeated queries are cached automatically:
- Cache TTL: 5 minutes per query
- Hit rate visible in `/stats` endpoint
- No explicit cache invalidation needed (automatic based on TTL)

---

## 7. Integration Examples

### 7.1 Node.js / TypeScript

```typescript
async function querySPARQL(query: string, format: string = "turtle") {
  const response = await fetch("/api/v1/sparql", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${token}`
    },
    body: JSON.stringify({
      query,
      timeout: 5000,
      format
    })
  });

  if (!response.ok) {
    throw new Error(`SPARQL error: ${response.statusText}`);
  }

  return response.json();
}

// Usage
const result = await querySPARQL(
  "CONSTRUCT { ?s ?p ?o } WHERE { ?s a <http://example.org/Person> . ?s ?p ?o }",
  "jsonld"
);
```

### 7.2 Python

```python
import requests

def query_sparql(query: str, format: str = "turtle"):
    response = requests.post(
        "http://localhost:8001/api/v1/sparql",
        json={
            "query": query,
            "timeout": 5000,
            "format": format
        },
        headers={"Authorization": f"Bearer {token}"}
    )
    response.raise_for_status()
    return response.json()

# Usage
result = query_sparql(
    "ASK { ?person <http://example.org/email> \"alice@example.com\" }"
)
print(f"Match found: {result['result']}")
```

### 7.3 cURL

```bash
# CONSTRUCT query
curl -X POST http://localhost:8001/api/v1/sparql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "query": "CONSTRUCT { ?s ?p ?o } WHERE { ?s a <http://example.org/Person> . ?s ?p ?o LIMIT 10 }",
    "timeout": 5000,
    "format": "turtle"
  }'

# ASK query
curl -X POST http://localhost:8001/api/v1/sparql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "query": "ASK { ?person <http://example.org/email> \"alice@example.com\" }",
    "timeout": 3000
  }'

# List ontologies
curl -X GET http://localhost:8001/api/v1/sparql/ontologies \
  -H "Authorization: Bearer $TOKEN"

# Get stats
curl -X GET http://localhost:8001/api/v1/sparql/stats \
  -H "Authorization: Bearer $TOKEN"
```

---

## 8. Deployment & Configuration

### 8.1 Environment Variables

```bash
# Oxigraph SPARQL endpoint
OXIGRAPH_URL=http://localhost:7878

# API timeout (seconds)
SPARQL_TIMEOUT_MAX=30

# Rate limiting
SPARQL_RATE_LIMIT=100  # Queries per minute

# Caching
SPARQL_CACHE_TTL=300   # Seconds
```

### 8.2 Oxigraph Setup

```bash
# Docker container (recommended)
docker run -d \
  --name oxigraph \
  -p 7878:7878 \
  ghcr.io/oxigraph/oxigraph:latest

# Load ontology data
docker exec oxigraph curl \
  -X POST http://localhost:7878/query \
  -H "Content-Type: text/turtle" \
  --data-binary @fibo.ttl
```

### 8.3 Monitoring

**Key metrics to track:**
- Query latency (p50, p95, p99)
- Timeout rate (should be < 0.1%)
- Cache hit rate (target > 0.5)
- Concurrent query count (max 50)

---

## 9. FAQ

**Q: How do I use custom prefixes?**
```sparql
PREFIX custom: <http://my.example.org/>
CONSTRUCT { ?s ?p ?o }
WHERE {
  ?s a custom:MyClass .
  ?s ?p ?o
}
```

**Q: Can I query multiple ontologies in one query?**
```sparql
PREFIX fibo: <https://spec.edmcouncil.org/fibo/ontology/>
PREFIX yawl: <https://yawl-workflow.org/ontology/>

CONSTRUCT { ?s ?p ?o }
WHERE {
  ?s a fibo:FinancialInstrument .
  ?workflow a yawl:Process .
  ...
}
```
Yes, all loaded ontologies are in the same graph.

**Q: What's the difference between CONSTRUCT and SELECT?**
- **CONSTRUCT** returns RDF/graph data (Turtle, JSON-LD, etc.)
- **SELECT** returns tabular data (variables + values)

**Q: How long can I set timeout?**
Maximum 30 seconds. Longer queries indicate missing LIMIT or inefficient pattern.

**Q: Does the API cache results?**
Yes, 5-minute TTL per unique query. See `cache_hit_rate` in `/stats`.

**Q: Can I modify the graph (INSERT/DELETE)?**
No, this API is read-only. RDF modifications go through data-modelling-cli or bos CLI.

---

## 10. References

- **W3C SPARQL 1.1 Specification:** https://www.w3.org/TR/sparql11-query/
- **Turtle RDF Format:** https://www.w3.org/TR/turtle/
- **JSON-LD Spec:** https://json-ld.org/
- **FIBO Ontology:** https://spec.edmcouncil.org/fibo/
- **Oxigraph Project:** https://github.com/oxigraph/oxigraph

---

**Support:** For issues, contact the DataOps team or file an issue in the main repository.
