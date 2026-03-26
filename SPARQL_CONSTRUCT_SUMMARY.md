# BusinessOS SPARQL CONSTRUCT Queries — Summary

**Quick reference for organizational data generation.**

---

## Four CONSTRUCT Queries

| # | Query | Purpose | Input | Output Triples | File |
|----|-------|---------|-------|---|------|
| **1** | Team Members | Employee RDF (foaf:Person + org:Post) | `team_members[]` | 48 (3 employees) | `sparql/organizational-data-construct.sparql` |
| **2** | Clients | Customer RDF (schema:Organization + ContactPoint) | `clients[]` | 24 (1 customer) | `sparql/organizational-data-construct.sparql` |
| **3** | Organizational Hierarchies | Reporting chains (org:reportsTo + transitive) | `team_members[{id, manager_id}]` | 18 | `sparql/organizational-data-construct.sparql` |
| **4** | Projects | Project RDF (schema:Project + TeamAssignments) | `projects[]`, `assignments[]` | 42 (1 project, 3 assignments) | `sparql/organizational-data-construct.sparql` |

---

## Parameters at a Glance

### Query 1: Team Members
```
Required: personId, name, email, role, departmentLabel, hireDate, status, orgUri
Optional: managerId
Output Classes: foaf:Person, org:Post, org:Organization
```

### Query 2: Clients
```
Required: clientId, clientName, industry, tier, contractValue, accountOwnerId, createdAt, contactName, contactEmail, areaServed
Optional: contactPhone
Output Classes: schema:Organization, schema:ContactPoint, org:Classification
```

### Query 3: Organizational Hierarchies
```
Input: Same as Query 1 (team members)
Derived: reportingLevel, chainDepth, chainMembers, transitivelReportsTo
Output Classes: bos:ReportingHierarchy, org:Organization
```

### Query 4: Projects
```
Required: projectId, projectName, clientId, accountOwnerId, budget, createdAt, projectStatus
Nested: memberId, memberName, memberRole, hoursAllocated, allocationPercentage
Output Classes: schema:Project, bos:ProjectTeamAssignment, bos:BudgetAllocation
```

---

## RDF Output Structure

### Team Members
```
foaf:Person
├── foaf:name, foaf:mbox, foaf:givenName, foaf:familyName
├── org:holds → org:Post
├── foaf:reports → foaf:Person (manager)
└── dcterms:created, bos:employeeStatus

org:Post
├── org:role
├── org:organization
├── org:reportsTo → org:Post (manager post)
└── dcterms:issued

org:Organization (department)
└── org:isPartOf → org:Organization (company)
```

### Clients
```
schema:Organization
├── schema:name, schema:url, schema:industry
├── schema:contactPoint → schema:ContactPoint
├── schema:areaServed
├── bos:contractValue, bos:customerTier
└── bos:accountOwner → foaf:Person

schema:ContactPoint
├── schema:name, schema:email, schema:telephone
└── dcterms:created

org:Classification (tier)
└── schema:potentialAction (service level)
```

### Organizational Hierarchies
```
foaf:Person
├── org:reportsTo → foaf:Person (direct manager)
├── bos:reportingLevel (0=IC, 1=Manager, 2=Director, 3+=Exec)
├── bos:chainDepth (distance to CEO)
├── bos:transitivelReportsTo → foaf:Person (any ancestor)
└── org:memberOf → org:Organization (team)

bos:ReportingHierarchy
├── bos:root → foaf:Person (CEO)
├── bos:leaf → foaf:Person (employee)
├── bos:chainMembers (serialized path)
└── bos:totalDepth (longest chain)
```

### Projects
```
schema:Project
├── schema:name, schema:url
├── schema:customer → schema:Organization (client)
├── schema:provider → foaf:Person (owner)
├── schema:budget, bos:budgetAmount, bos:budgetCurrency
├── bos:projectStatus, bos:teamSize
├── schema:startDate, schema:endDate
└── org:hasPost → org:Post (project lead)

bos:ProjectTeamAssignment
├── bos:assignee → foaf:Person
├── bos:role, bos:hoursAllocated, bos:allocationPercentage
├── bos:assignmentStatus
└── org:memberOf → schema:Project

bos:BudgetAllocation
├── bos:allocatedTo → schema:Project
├── bos:allocatedFrom → schema:Organization (client)
├── bos:allocatedAmount, bos:remainingBudget
└── bos:allocationStatus
```

---

## Vocabulary Prefixes

| Prefix | Namespace | Usage |
|--------|-----------|-------|
| `foaf:` | http://xmlns.com/foaf/0.1/ | Person, workplace, reporting |
| `org:` | http://www.w3.org/ns/org# | Organization, Post, reportsTo |
| `schema:` | https://schema.org/ | Organization, Project, ContactPoint |
| `dcterms:` | http://purl.org/dc/terms/ | created, issued, valid |
| `prov:` | http://www.w3.org/ns/prov# | Provenance tracking |
| `bos:` | https://chatmangpt.com/ontology/businessos/ | BusinessOS custom terms |
| `time:` | http://www.w3.org/2006/time# | Time intervals |
| `rdfs:` | http://www.w3.org/2000/01/rdf-schema# | Labels |
| `xsd:` | http://www.w3.org/2001/XMLSchema# | Data types |

---

## Usage Pipeline

```bash
# 1. Load SPARQL from file
QUERY=$(cat BusinessOS/sparql/organizational-data-construct.sparql)

# 2. Build input JSON with team_members, clients, projects
cat > /tmp/input.json << 'EOF'
{
  "team_members": [...],
  "clients": [...],
  "projects": [...]
}
EOF

# 3. Execute CONSTRUCT query via Go handler
curl -X POST http://localhost:8001/api/sparql/construct \
  -H "Content-Type: application/json" \
  -d '{
    "query": "CONSTRUCT { ... } WHERE { VALUES ... { ... } }",
    "format": "turtle"
  }' | jq '.triples'

# 4. Load into Oxigraph
obsr load \
  --store /var/lib/oxigraph/businessos.db \
  --input output.ttl \
  --format turtle

# 5. Verify with SPARQL SELECT
obsr query \
  --store /var/lib/oxigraph/businessos.db \
  --query 'SELECT (COUNT(?person) AS ?count) WHERE { ?person a foaf:Person }'
# Output: count=3
```

---

## Validation Checklist

- [ ] All primary keys (personId, clientId, projectId) are unique
- [ ] All enum fields (status, tier, projectStatus) use allowed values
- [ ] All dates in ISO8601 format (YYYY-MM-DD)
- [ ] All numeric fields (contractValue, budget, hoursAllocated) >= 0
- [ ] All email addresses valid format
- [ ] All manager IDs reference existing persons (no orphans)
- [ ] No circular reporting relationships
- [ ] Generated timestamps use ISO8601 with timezone
- [ ] Provenance metadata complete (wasGeneratedBy with Activity, startedAtTime)
- [ ] Output passes SPARQL ASK integrity checks

---

## Documentation

| Document | Purpose |
|----------|---------|
| `docs/diataxis/how-to/sparql-construct-organizational-data.md` | Complete usage guide with parameters, examples, and troubleshooting |
| `docs/diataxis/reference/sparql-construct-rdf-output-examples.md` | Ready-to-load RDF/Turtle examples for all 4 queries |
| `BusinessOS/sparql/organizational-data-construct.sparql` | Source SPARQL file with all 4 CONSTRUCT queries |

---

## Error Resolution

| Error | Cause | Fix |
|-------|-------|-----|
| `Undefined variable in CONSTRUCT` | Variable in CONSTRUCT not in WHERE | Ensure all vars bound in WHERE |
| `Type mismatch: expected xsd:date` | Date string without type | Add `^^xsd:date` suffix |
| `Duplicate rows in VALUES` | Same value appears twice | Deduplicate by primary key |
| `Multiple values for same var` | Cartesian product in WHERE | Move FILTER to VALUES or join more specifically |
| `Null/empty RDF output` | All FILTER conditions failed | Check FILTER logic; use OPTIONAL for nullable fields |
| `Slow performance on large dataset` | Inefficient triple patterns | Reorder WHERE clauses (most selective first) |

---

## Key Rules

1. **CONSTRUCT only, never INSERT** — All data generation via CONSTRUCT queries
2. **SPARQL CONSTRUCT → Oxigraph** — No ad hoc RDF manipulation
3. **Validate output** — Run SPARQL ASK checks before loading
4. **Provenance always** — Every generated triple has prov:wasGeneratedBy
5. **Type safety** — All literal values cast to correct xsd: type
6. **Unique identifiers** — All IRIs deterministic from primary keys

---

**Created:** 2026-03-25
**Status:** Production (v1.2.0)
**Maintained By:** Data Operations (Sean Chatman)
**Last Updated:** 2026-03-25
