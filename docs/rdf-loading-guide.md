# RDF Loading Guide — Oxigraph Integration

**Version:** 1.0
**Last Updated:** 2026-03-26
**Scope:** Ontology and RDF data loading into Oxigraph triplestore

---

## Quick Start

### 1. Start Oxigraph

```bash
# Using Docker
docker run -d -p 8890:8890 --name oxigraph oxigraph/oxigraph

# Verify connectivity
curl http://localhost:8890/query
```

### 2. Load Ontologies

```bash
# From project root
./scripts/load-ontologies.sh

# With custom Oxigraph URL
./scripts/load-ontologies.sh http://oxigraph:8890

# From custom ontology directory
./scripts/load-ontologies.sh http://localhost:8890 /path/to/ontologies
```

### 3. Seed Test Data

```bash
# Inserts FIBO deals, datasets, and healthcare records
./scripts/seed-rdf-data.sh

# With custom Oxigraph URL
./scripts/seed-rdf-data.sh http://oxigraph:8890
```

### 4. Query Data

```bash
# Example: Query all deals
curl -X POST http://localhost:8890/query \
  -H "Content-Type: application/sparql-query" \
  -d 'SELECT ?deal WHERE { ?deal a <https://ontology.chatmangpt.com/fibo/execution#CreditDefaultSwap> . } LIMIT 10'
```

---

## Architecture Overview

### System Components

```
┌─────────────────────────────────────────────────────────────────┐
│                        RDF Loading Pipeline                     │
└─────────────────────────────────────────────────────────────────┘

  ┌──────────────────┐          ┌─────────────────┐
  │  Ontology Files  │          │  FIBO/Domain    │
  │  (.ttl format)   │          │  Data (SPARQL)  │
  └────────┬─────────┘          └────────┬────────┘
           │                             │
           │ Validation                  │ Validation
           ├─> Turtle syntax check       ├─> Query syntax
           ├─> Prefix definitions        └─> Data constraints
           └─> Import statements
           │                             │
           ▼                             ▼
  ┌──────────────────┐          ┌─────────────────┐
  │ load-ontologies  │          │ seed-rdf-data   │
  │     (bash)       │          │     (bash)      │
  └────────┬─────────┘          └────────┬────────┘
           │                             │
           │ POST /store                 │ POST /query
           │ Content-Type: text/turtle   │ Content-Type: application/sparql-query
           └────────────┬────────────────┘
                        │
                        ▼
                  ┌──────────────┐
                  │  Oxigraph    │
                  │  Triplestore │
                  │  (:8890)     │
                  └──────┬───────┘
                         │
            ┌────────────┼────────────┐
            │            │            │
            ▼            ▼            ▼
        Subjects     Predicates   Objects
         (URIs)       (URIs)      (Literals)
```

### Ontology Dependency Order

1. **Layer 1: W3C Standards** (loaded automatically by ontology imports)
   - RDF/OWL meta-ontologies
   - PROV-O provenance
   - SHACL shapes
   - OWL-Time temporal

2. **Layer 2: ChatmanGPT Core** (must load first)
   - `chatman-core.ttl` — Foundation
   - `chatman-org.ttl` — Organizational structure
   - `chatman-signal.ttl` — Signal Theory

3. **Layer 3: Domain Ontologies** (load after core)
   - `chatman-agents.ttl` — Agent definitions
   - `chatman-process.ttl` — Process mining
   - `chatman-compliance.ttl` — Compliance rules
   - `chatman-healthcare.ttl` — Healthcare domain

The load scripts enforce this order automatically.

---

## Detailed Usage

### load-ontologies.sh

Loads all `.ttl` files from the ontologies directory into Oxigraph.

**Signature:**
```bash
./scripts/load-ontologies.sh [oxigraph_url] [ontology_dir]
```

**Parameters:**
- `oxigraph_url` — Oxigraph HTTP endpoint (default: `http://localhost:8890`)
- `ontology_dir` — Directory containing `.ttl` files (default: `./ontologies`)

**Features:**
- **Order-aware loading**: Core ontologies first, dependencies resolved
- **Retry logic**: 3 attempts with 2-second backoff for transient failures
- **Validation**: Checks Turtle syntax before loading
- **Idempotent**: Safe to re-run (no duplication)
- **Progress logging**: Logs each ontology loaded
- **Connectivity check**: Verifies Oxigraph is ready before starting

**Exit Codes:**
| Code | Meaning | Action |
|------|---------|--------|
| 0 | Success | All ontologies loaded |
| 1 | Oxigraph unreachable | Check docker: `docker ps`, ensure port 8890 open |
| 2 | Directory not found | Check ontology path |
| 3 | Validation error | Check .ttl file syntax |
| 4 | Network error (retries) | Check network, Oxigraph stability |

**Example Output:**
```
[INFO] === Ontology Loader (Oxigraph) ===
[INFO] Oxigraph URL: http://localhost:8890
[INFO] Ontology directory: ./ontologies
[INFO] Checking Oxigraph connectivity...
[SUCCESS] Oxigraph is reachable
[INFO] Found 8 ontology files to load

[INFO] Loading ontology: chatman-core.ttl
[INFO]   Attempt 1/3...
[SUCCESS]   ✓ chatman-core.ttl loaded (HTTP 204)

[INFO] Loading ontology: chatman-org.ttl
[SUCCESS]   ✓ chatman-org.ttl loaded (HTTP 204)

...

[SUCCESS] === Summary ===
[SUCCESS] Loaded: 8 / 8
[SUCCESS] All ontologies loaded successfully
```

**Log Files:**
Logs are written to `./logs/load-ontologies_YYYYMMDD_HHMMSS.log`

---

### seed-rdf-data.sh

Seeds test data into Oxigraph using SPARQL INSERT queries.

**Signature:**
```bash
./scripts/seed-rdf-data.sh [oxigraph_url]
```

**Parameters:**
- `oxigraph_url` — Oxigraph HTTP endpoint (default: `http://localhost:8890`)

**Data Inserted:**
- **3 FIBO Deals**: CDS Swap, Interest Rate Swap, FX Forward
- **5 Datasets**: Process Events, Compliance Records, and 3 derived datasets
- **2 Healthcare Records**: Patient record, Clinical observation

**Features:**
- **Idempotent**: Checks for duplicate data before inserting
- **Independent**: Runs independently of ontologies (can insert before/after)
- **Retry logic**: 3 attempts with 2-second backoff
- **Progress logging**: Logs each insert attempt
- **SPARQL INSERT DATA**: Uses standard SPARQL protocol

**Exit Codes:**
Same as `load-ontologies.sh`

**Example Output:**
```
[INFO] === RDF Data Seeder (Oxigraph) ===
[INFO] Oxigraph URL: http://localhost:8890
[INFO] Checking Oxigraph connectivity...
[SUCCESS] Oxigraph is reachable

[INFO] === Loading FIBO Deals ===
[INFO] Inserting: FIBO Deal 1: CDS Swap
[INFO]   Attempt 1/3...
[SUCCESS]   ✓ FIBO Deal 1: CDS Swap inserted (HTTP 204)

[INFO] Inserting: FIBO Deal 2: IRS Swap
[SUCCESS]   ✓ FIBO Deal 2: IRS Swap inserted (HTTP 204)

...

[SUCCESS] === Summary ===
[SUCCESS] Inserted: 8 / 8 items
[SUCCESS] All data seeded successfully
```

---

## Using the Go Client

The `rdf.Client` in `internal/rdf/loader.go` provides programmatic access to Oxigraph.

### Initialize Client

```go
package main

import (
  "github.com/rhl/businessos-backend/internal/config"
  "github.com/rhl/businessos-backend/internal/rdf"
)

func main() {
  cfg := config.Load()

  // Create RDF client
  client := rdf.NewClient(cfg)

  // Check connectivity
  ctx := context.Background()
  if err := client.Health(ctx); err != nil {
    log.Fatalf("Oxigraph is down: %v", err)
  }
}
```

### Load Ontology File

```go
ctx := context.Background()

if err := client.LoadOntology(ctx, "./ontologies/chatman-core.ttl"); err != nil {
  log.Fatalf("Failed to load ontology: %v", err)
}
log.Println("Ontology loaded")
```

### Load Turtle Data (bytes)

```go
turtleData := []byte(`
@prefix ex: <https://example.com/> .
ex:Subject ex:predicate ex:Object .
`)

if err := client.LoadTurtleData(ctx, turtleData); err != nil {
  log.Fatalf("Failed to load data: %v", err)
}
```

### Execute SPARQL INSERT

```go
query := `
PREFIX fibo-exec: <https://ontology.chatmangpt.com/fibo/execution#>
PREFIX dcterms: <http://purl.org/dc/terms/>

INSERT DATA {
  <https://data.chatmangpt.com/deals/deal-001> a fibo-exec:Deal ;
    dcterms:title "Deal 001" .
}
`

if err := client.LoadData(ctx, query); err != nil {
  log.Fatalf("Failed to insert data: %v", err)
}
```

### Execute SPARQL SELECT

```go
query := `
PREFIX fibo-exec: <https://ontology.chatmangpt.com/fibo/execution#>

SELECT ?deal WHERE {
  ?deal a fibo-exec:CreditDefaultSwap .
  ?deal <http://purl.org/dc/terms/title> ?title .
}
LIMIT 10
`

results, err := client.QuerySPARQL(ctx, query)
if err != nil {
  log.Fatalf("Query failed: %v", err)
}

// results is JSON string
log.Println(results)
```

### Configuration

Set `OxigraphURL` in `.env` or config:

```bash
# .env
OXIGRAPH_URL=http://oxigraph:8890
```

Default if not set: `http://localhost:8890`

---

## Troubleshooting

### Error: "Oxigraph unreachable"

**Symptom:**
```
[ERROR] Oxigraph unreachable at http://localhost:8890
```

**Diagnosis:**
```bash
# Check if Oxigraph container is running
docker ps | grep oxigraph

# Check logs
docker logs oxigraph

# Test connectivity
curl -v http://localhost:8890/query
```

**Fix:**
```bash
# Start Oxigraph if not running
docker run -d -p 8890:8890 --name oxigraph oxigraph/oxigraph

# If already running, restart it
docker restart oxigraph

# Check port is correct in script
# Default: http://localhost:8890
```

---

### Error: "Invalid Turtle file"

**Symptom:**
```
[ERROR] Invalid Turtle file (missing prefixes): chatman-core.ttl
```

**Diagnosis:**
Check the file has required Turtle syntax:

```bash
head -20 ontologies/chatman-core.ttl | grep "@prefix"
```

**Fix:**
Ensure file starts with prefix declarations:

```turtle
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix owl: <http://www.w3.org/2002/07/owl#> .
```

---

### Error: "HTTP 400 Bad Request"

**Symptom:**
```
[ERROR] HTTP 400: Turtle parse error at line 42
```

**Diagnosis:**
The Turtle file has syntax errors.

**Fix:**
Validate using a Turtle validator:

```bash
# Using rapper (RDF toolkit)
rapper -c ontologies/chatman-core.ttl

# Or check syntax manually
cat ontologies/chatman-core.ttl | grep -n "error\|missing"
```

---

### Error: "HTTP 502 Bad Gateway"

**Symptom:**
```
[WARN] HTTP 502 (transient error, will retry)
```

**Diagnosis:**
Oxigraph is overloaded or temporarily unavailable.

**Fix:**
Retry after a moment (script auto-retries 3x):

```bash
# Wait 10 seconds and re-run
sleep 10
./scripts/load-ontologies.sh

# Or scale Oxigraph resources
docker update --cpus 2 oxigraph
docker restart oxigraph
```

---

### Error: "HTTP 413 Payload Too Large"

**Symptom:**
```
[ERROR] HTTP 413: Request entity too large
```

**Diagnosis:**
Ontology file is too large for single request.

**Fix:**
Split large ontologies or increase Oxigraph limit:

```bash
# Increase Docker resource limit
docker update --memory 4g oxigraph
docker restart oxigraph
```

---

### Data Not Appearing in Query Results

**Symptom:**
```
[SUCCESS] Data inserted successfully
# But later:
SELECT ?s WHERE { ?s a <class> . }
# Returns 0 results
```

**Diagnosis:**
Data was inserted but query doesn't find it.

**Check:**
1. Verify insert actually succeeded (check logs for "HTTP 200" or "HTTP 204")
2. Check class URI is exactly correct (URIs are case-sensitive)
3. Check data actually exists:

```bash
# Query everything
curl -X POST http://localhost:8890/query \
  -H "Content-Type: application/sparql-query" \
  -d 'SELECT ?s ?p ?o WHERE { ?s ?p ?o . } LIMIT 5'
```

---

## Advanced Usage

### Batch Loading Multiple Directories

```bash
for dir in ontologies/*/; do
  ./scripts/load-ontologies.sh http://localhost:8890 "$dir"
done
```

### Load Only Specific Ontologies

```bash
# Copy specific files to temp dir
mkdir /tmp/ontologies
cp ontologies/chatman-core.ttl /tmp/ontologies/
cp ontologies/chatman-org.ttl /tmp/ontologies/

# Load only those
./scripts/load-ontologies.sh http://localhost:8890 /tmp/ontologies
```

### Custom SPARQL Queries

Create a file `custom-insert.rq`:

```sparql
PREFIX fibo-exec: <https://ontology.chatmangpt.com/fibo/execution#>
PREFIX dcterms: <http://purl.org/dc/terms/>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>

INSERT DATA {
  <https://data.chatmangpt.com/custom/deal-001> a fibo-exec:Deal ;
    dcterms:title "Custom Deal" ;
    dcterms:issued "2026-03-26"^^xsd:date ;
    fibo-exec:hasNotional "1000000"^^xsd:decimal .
}
```

Then insert:

```bash
curl -X POST http://localhost:8890/query \
  -H "Content-Type: application/sparql-query" \
  -d @custom-insert.rq
```

### Dump All Triples (Backup)

```bash
curl -X POST http://localhost:8890/query \
  -H "Content-Type: application/sparql-query" \
  -H "Accept: text/turtle" \
  -d 'CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o . }'  > backup.ttl
```

### Oxigraph Maintenance

```bash
# List all named graphs
curl -X POST http://localhost:8890/query \
  -H "Content-Type: application/sparql-query" \
  -d 'SELECT DISTINCT ?g WHERE { GRAPH ?g { ?s ?p ?o . } }'

# Clear all data (WARNING: deletes everything)
# docker run --rm -it -p 8890:8890 -v oxigraph_data:/var/lib/oxigraph \
#   oxigraph/oxigraph --reset

# Restart with empty store
docker restart oxigraph
rm -rf ~/oxigraph_data  # If using local volume
```

---

## Configuration Reference

### Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `OXIGRAPH_URL` | `http://localhost:8890` | Oxigraph HTTP endpoint |
| `ONTOLOGY_DIR` | `./ontologies` | Directory with .ttl files |

### File Locations

| Path | Purpose |
|------|---------|
| `./ontologies/` | Core ontology files |
| `./ontologies/examples/` | Example RDF files for testing |
| `./ontologies/shacl/` | SHACL shape definitions |
| `./scripts/load-ontologies.sh` | Loader script |
| `./scripts/seed-rdf-data.sh` | Data seeder script |
| `./logs/load-ontologies_*.log` | Load operation logs |
| `./logs/seed-rdf-data_*.log` | Seed operation logs |

### Oxigraph Docker

```bash
# Run Oxigraph with persistent volume
docker run -d \
  --name oxigraph \
  -p 8890:8890 \
  -v oxigraph_data:/var/lib/oxigraph \
  oxigraph/oxigraph

# View logs
docker logs oxigraph

# Stats
docker stats oxigraph
```

---

## Integration Examples

### Use Case 1: Load Ontologies at Startup

In `cmd/main.go`:

```go
func init() {
  ctx := context.Background()
  client := rdf.NewClient(cfg)

  // Load core ontologies
  ontologies := []string{
    "./ontologies/chatman-core.ttl",
    "./ontologies/chatman-org.ttl",
    "./ontologies/chatman-signal.ttl",
  }

  for _, ont := range ontologies {
    if err := client.LoadOntology(ctx, ont); err != nil {
      log.Printf("Warning: failed to load %s: %v", ont, err)
    }
  }
}
```

### Use Case 2: Query Business Entities

```go
// Query all FIBO deals
query := `
PREFIX fibo-exec: <https://ontology.chatmangpt.com/fibo/execution#>
PREFIX dcterms: <http://purl.org/dc/terms/>

SELECT ?deal ?title WHERE {
  ?deal a fibo-exec:Deal ;
        dcterms:title ?title .
}
`

results, err := client.QuerySPARQL(ctx, query)
// Parse JSON results
```

### Use Case 3: Insert Deal Data from Handler

```go
func (h *Handler) CreateDeal(w http.ResponseWriter, r *http.Request) {
  // Parse request body
  var req DealRequest
  json.NewDecoder(r.Body).Decode(&req)

  // Generate SPARQL INSERT
  insert := fmt.Sprintf(`
    PREFIX fibo-exec: <https://ontology.chatmangpt.com/fibo/execution#>
    PREFIX dcterms: <http://purl.org/dc/terms/>
    PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>

    INSERT DATA {
      <https://data.chatmangpt.com/deals/%s> a fibo-exec:Deal ;
        dcterms:title "%s"@en ;
        fibo-exec:hasNotional "%f"^^xsd:decimal .
    }
  `, req.ID, req.Title, req.Notional)

  // Insert into RDF store
  if err := h.rdfClient.LoadData(r.Context(), insert); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  w.WriteHeader(http.StatusCreated)
}
```

---

## Performance Tuning

### Oxigraph Docker Resources

For production, allocate adequate resources:

```bash
docker run -d \
  --name oxigraph \
  -p 8890:8890 \
  --cpus 4 \
  --memory 8g \
  -v oxigraph_data:/var/lib/oxigraph \
  oxigraph/oxigraph
```

### Batch Inserts

For large data loads, use SPARQL INSERT with multiple triples:

```sparql
INSERT DATA {
  <uri1> <pred1> <obj1> .
  <uri2> <pred2> <obj2> .
  <uri3> <pred3> <obj3> .
  ... (many more triples)
}
```

Not individual INSERT queries (slower).

### Connection Pooling

The Go client pools HTTP connections (100 connections, 10 per host):

```go
// Already configured in rdf.Client
// No tuning needed unless very high concurrency (>1000 req/s)
```

---

## References

- **Oxigraph Documentation:** https://oxigraph.org/
- **SPARQL Specification:** https://www.w3.org/TR/sparql11-query/
- **Turtle Format:** https://www.w3.org/TR/turtle/
- **W3C PROV-O:** https://www.w3.org/TR/prov-o/
- **FIBO Ontology:** https://spec.edmcouncil.org/fibo/

---

*Last updated: 2026-03-26*
