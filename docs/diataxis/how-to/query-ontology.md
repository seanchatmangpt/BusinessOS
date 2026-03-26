# How to Query the Ontology — Business Intelligence from RDF

This guide shows how to use BusinessOS ontology endpoints to discover agents, check compliance policies, explore organization structure, and trace artifact lineage.

## Quick Start: REST API

All ontology endpoints require authentication. Include your JWT token in the Authorization header.

### List All Agents

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8001/api/ontology/agents
```

Response:
```json
{
  "agents": [
    {
      "id": "agent-1",
      "name": "Data Pipeline Agent",
      "type": "osa",
      "status": "active",
      "last_heartbeat": "2026-03-26T10:30:45Z",
      "capabilities": ["process_mining", "discovery"]
    }
  ],
  "count": 1
}
```

### Check Compliance Policies

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8001/api/ontology/compliance
```

Response:
```json
{
  "policies": [
    {
      "id": "soc2-cc6",
      "framework": "SOC2",
      "title": "Logical Access Control",
      "status": "verified",
      "controls": ["access_list", "mfa_enabled"]
    }
  ],
  "count": 1
}
```

### Verify Compliance Framework

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"frameworks": ["SOC2", "HIPAA"]}' \
  http://localhost:8001/api/ontology/compliance/check
```

Response:
```json
{
  "compliant": true,
  "score": 0.95,
  "details": {
    "frameworks": ["SOC2", "HIPAA"]
  }
}
```

### Explore Organization Structure

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8001/api/ontology/org
```

Response:
```json
{
  "organization": "Default Organization",
  "departments": [
    {
      "id": "dept-1",
      "name": "Engineering",
      "parent_id": null,
      "manager": "alice@company.com"
    }
  ],
  "roles": [
    {
      "id": "role-1",
      "title": "Senior Engineer",
      "department": "dept-1",
      "permissions": ["read", "write", "admin"]
    }
  ],
  "reporting_lines": []
}
```

### Trace Artifact Lineage

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8001/api/ontology/provenance/artifact-123
```

Response:
```json
{
  "artifact_id": "artifact-123",
  "artifact_name": "Report Q1 2026",
  "origins": [
    {
      "id": "source-1",
      "type": "Dataset",
      "name": "Sales Data",
      "timestamp": "2026-03-01T00:00:00Z"
    }
  ],
  "derivations": [],
  "agents": [
    {
      "id": "agent-report",
      "name": "Report Generator",
      "role": "Creator"
    }
  ]
}
```

### Emit Provenance Triple

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "http://example.com/artifact-123",
    "predicate": "http://www.w3.org/ns/prov#wasDerivedFrom",
    "object": "http://example.com/source-1",
    "agent": "http://example.com/agent-1",
    "activity": "2026-03-26T10:00:00Z"
  }' \
  http://localhost:8001/api/ontology/provenance
```

Response:
```json
{
  "status": "emitted",
  "triple_id": "http://example.com/artifact-123-wasDerivedFrom-http://example.com/source-1",
  "timestamp": "2026-03-26T10:00:00Z"
}
```

### Discover Available Tools

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8001/api/ontology/tools
```

Response:
```json
{
  "tools": [
    {
      "id": "tool-pm4py",
      "name": "Process Mining",
      "description": "Discover and analyze process flows",
      "category": "data",
      "version": "1.0.0",
      "endpoint": "http://localhost:8090",
      "parameters": {
        "max_variants": 100,
        "min_support": 0.01
      },
      "status": "available"
    }
  ],
  "count": 1
}
```

## CLI Commands

Use `bos ontology` subcommands to manage the ontology from the command line.

### List Agents

```bash
bos ontology agents --endpoint http://localhost:7878
```

Output:
```json
{
  "agents": [],
  "count": 0
}
```

### Check Compliance Policies

```bash
bos ontology compliance --endpoint http://localhost:7878
```

Output:
```json
{
  "policies": [],
  "count": 0
}
```

### List Tools

```bash
bos ontology tools --endpoint http://localhost:7878
```

Output:
```json
{
  "tools": [],
  "count": 0
}
```

## Understanding Compliance Policies

Compliance policies are SPARQL-queryable RDF triples stored in Oxigraph. Each policy has:

- **Framework**: SOC2, HIPAA, GDPR, SOX, or custom
- **Controls**: Specific technical/organizational controls
- **Status**: verified (passing), failed (non-compliant), pending (awaiting evidence)

Example query (SPARQL):
```sparql
PREFIX bo: <http://businessos.example/ontology/>

SELECT ?policyId ?framework ?title
WHERE {
  ?policy a bo:CompliancePolicy ;
          bo:policyId ?policyId ;
          bo:framework ?framework ;
          bo:title ?title ;
          bo:status "verified" .
}
ORDER BY ?framework
```

## Understanding Provenance Lineage

Provenance (PROV-O) traces artifact origin and derivation. Key relationships:

- **wasGeneratedBy**: Artifact generated by activity
- **wasDerivedFrom**: Artifact derived from input
- **wasAssociatedWith**: Activity associated with agent
- **used**: Activity used input entity

Example query (SPARQL):
```sparql
PREFIX prov: <http://www.w3.org/ns/prov#>

SELECT ?entityId ?agentName
WHERE {
  <http://example.com/artifact-123>
    prov:wasGeneratedBy ?activity ;
    prov:wasDerivedFrom ?origin .
  ?activity prov:wasAssociatedWith ?agent .
  ?agent prov:label ?agentName .
}
```

## Troubleshooting

### Ontology Service Unavailable (503)

**Cause**: Oxigraph is not running or not reachable.

**Solution**:
```bash
# Check if Oxigraph is running
curl http://localhost:7878

# Start Oxigraph
docker-compose up oxigraph

# Verify BusinessOS can reach it
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8001/api/ontology/agents
```

### No Results from Compliance Check

**Cause**: Compliance policies have not been loaded into Oxigraph.

**Solution**:
```bash
# Load compliance policies
bos ontology execute \
  --mapping compliance-mappings.json \
  --database postgresql://localhost/businessos

# Verify policies are loaded
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8001/api/ontology/compliance
```

### Provenance Triple Not Emitted

**Cause**: Invalid SPARQL CONSTRUCT query or Oxigraph write access denied.

**Solution**:
```bash
# Verify Oxigraph accepts writes
curl -X POST \
  -H "Content-Type: application/sparql-update" \
  -d "INSERT DATA { <http://example.com/test> <http://example.com/prop> \"value\" . }" \
  http://localhost:7878/update

# Check BusinessOS logs
docker-compose logs businessos-backend | grep "emit provenance"
```

## See Also

- [Signal Theory Explanation](../explanation/signal-theory-complete.md) — How every output is structured
- [7-Layer Architecture](../explanation/seven-layer-architecture.md) — Data layer storage patterns
- [YAWL 43 Patterns](../reference/yawl-43-patterns.md) — Coordination behavior patterns
