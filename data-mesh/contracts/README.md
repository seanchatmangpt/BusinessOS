# BusinessOS Data Mesh Contracts

**Complete machine-readable data mesh domain contracts for all 5 BusinessOS domains.**

This directory contains machine-readable, semantically-grounded contracts for all BusinessOS data domains using DCAT, schema.org, DQV, ODRL, and PROV-O standards.

---

## 📋 Quick Start

1. **Browse the catalog:** See [CATALOG.md](./CATALOG.md) for overview of all domains
2. **Read domain contracts:** Each domain has `domain-contract.yaml` (human-readable)
3. **Query with SPARQL:** Use [validation-queries.sparql](./validation-queries.sparql) for data quality/compliance checks
4. **Discover datasets:** Use RDF/Turtle versions for semantic web integration
5. **Implement your contract:** Use [TEMPLATE.md](./TEMPLATE.md) as a guide

---

## 📁 Directory Structure

```
data-mesh/contracts/
├── README.md                         ← You are here
├── CATALOG.md                        ← Master index of all domains
├── TEMPLATE.md                       ← How to create/update contracts
├── validation-queries.sparql         ← SPARQL validation queries (5 queries)
│
├── finance/                          ← Finance Domain
│   ├── domain-contract.yaml          ← Human-readable contract (YAML)
│   └── domain-contract.ttl           ← Machine-readable (RDF/Turtle)
│
├── operations/                       ← Operations Domain
│   └── domain-contract.yaml
│
├── marketing/                        ← Marketing Domain (GDPR/CCPA sensitive)
│   └── domain-contract.yaml
│
├── sales/                            ← Sales Domain
│   └── domain-contract.yaml
│
└── hr/                               ← Human Resources Domain (highest PII)
    └── domain-contract.yaml
```

---

## 📊 What's in Each Contract?

Every domain contract includes:

### 1. **Entity Definitions** (DCAT Dataset)
- Identifier, description, record count, growth rate
- Business criticality level
- Distribution endpoints (API + Database)
- Key fields with data types and constraints

**Example:**
```yaml
entities:
  - name: "Deal"
    identifier: "finance.deals.v1"
    record_count: 15847
    distribution:
      api:
        endpoint: "https://api.businessos.example.com/v1/finance/deals"
        format: "JSON-HAL"
        rate_limit: 1000
```

### 2. **Data Quality Metrics** (DQV - Data Quality Vocabulary)
- Timeliness (freshness), Accuracy, Completeness, Consistency, Validity
- Expected vs actual values
- Pass/Fail status
- Overall quality score (0.0-1.0)

**Example:**
```yaml
quality_metrics:
  deals:
    timeliness:
      expected_value: "< 5 minutes"
      actual_value: "1.2 minutes"
      status: "PASSING"
    accuracy:
      expected_value: "99.9%"
      actual_value: "99.95%"
      status: "PASSING"
    overall_score: 0.96
```

### 3. **Access Control Policies** (ODRL - Open Digital Rights Language)
- Role-based permissions (READ, WRITE, DELETE, EXPORT)
- Constraints (territory, time window, field masking, volume limits)
- Prohibitions (what users cannot do)
- Obligations (audit logging, data classification)

**Example:**
```yaml
access_control:
  sales_team:
    permissions:
      deals: "READ (assigned territory only)"
    constraints:
      territory_scoping: "strict"
      volume_limit: "unlimited for own territory"

  external_contractors:
    permissions:
      deals: "DENIED"
    reason: "Non-Finance personnel cannot access financial data"
```

### 4. **Service Level Agreements** (SLAs)
- Availability percentage
- Max latency
- Recovery Time Objective (RTO)
- Recovery Point Objective (RPO)
- Support tier (hours of operation)

**Example:**
```yaml
service_level_agreements:
  deals:
    availability: 99.5%
    max_latency: "5 minutes"
    rto_minutes: 15
    rpo_minutes: 1
    support_tier: "24/7 Level-2"
```

### 5. **Data Lineage** (PROV-O - Provenance)
- Source system
- Synchronization frequency and method
- ETL process name
- Records processed and success rate
- Last sync timestamp

**Example:**
```yaml
lineage:
  source_system: "Salesforce CRM"
  sync_frequency: "Real-time (sub-minute)"
  sync_method: "Kafka CDC (Change Data Capture)"
  last_sync: "2026-03-25T08:05:00Z"
  records_synced: 15847
  records_error: 0
```

### 6. **Usage Policies**
- Export policies (who can export, in what format)
- Retention policies (how long to keep data)
- Encryption requirements
- PII handling (masking, minimization)
- Audit logging rules
- Compliance frameworks (SOC2, GDPR, SOX, etc.)

---

## 🌍 Domains at a Glance

### Finance (CFO)
- **Entities:** 4 (Deal, Invoice, Payment, Account)
- **Records:** 173,289
- **Quality Score:** 0.96 ⭐⭐⭐⭐⭐
- **SLA:** 99.5-99.99% (payments highest)
- **Criticality:** 🔴 CRITICAL
- **Compliance:** SOC2, SOX, GDPR, PCI-DSS
- **Key Challenge:** Payment data encryption + audit trail

### Operations (COO)
- **Entities:** 4 (Project, Process, Resource, KPI)
- **Records:** 31,122
- **Quality Score:** 0.95
- **SLA:** 98-99.9%
- **Criticality:** 🟠 HIGH
- **Compliance:** SOC2, GDPR, ISO 27001
- **Key Challenge:** Real-time KPI freshness

### Marketing (CMO)
- **Entities:** 4 (Campaign, Lead, Engagement, Content)
- **Records:** 9,147,450
- **Quality Score:** 0.94
- **SLA:** 98-99.9%
- **Criticality:** 🟠 HIGH
- **Compliance:** **GDPR, CCPA** (customer emails + behavior)
- **Key Challenge:** PII minimization + consent tracking

### Sales (CRO)
- **Entities:** 4 (Opportunity, Account, Quote, Activity)
- **Records:** 590,847
- **Quality Score:** 0.96 ⭐⭐⭐⭐⭐
- **SLA:** 99-99.9%
- **Criticality:** 🔴 CRITICAL
- **Compliance:** SOC2, GDPR, CCPA
- **Key Challenge:** Territory-scoped access control

### Human Resources (CHRO)
- **Entities:** 5 (Employee, Compensation, Leave, Performance, Training)
- **Records:** 44,368
- **Quality Score:** 0.98 ⭐⭐⭐⭐⭐
- **SLA:** 98-99.95%
- **Criticality:** 🔴 CRITICAL
- **Compliance:** **GDPR, CCPA, SOC2, FMLA, ADA, HIPAA**
- **Key Challenge:** Highest PII + encryption + access control

---

## 🔍 How to Use This

### For Data Consumers (Developers)

**I want to consume the Deals dataset:**

1. Check [CATALOG.md](./CATALOG.md) → Finance domain → Deals entity
2. Note the API endpoint: `https://api.businessos.example.com/v1/finance/deals`
3. Check access control matrix: Can you access it?
4. Review SLA: Max 5-minute latency, 99.5% availability
5. Check quality metrics: 0.96 score, all PASSING ✅
6. Implement with:
   ```bash
   curl -H "Authorization: Bearer $TOKEN" \
     "https://api.businessos.example.com/v1/finance/deals?limit=100"
   ```

**I'm building a cross-domain integration:**

1. Check [validation-queries.sparql](./validation-queries.sparql) → Query 4 (Cross-Domain Integration)
2. Load all RDF contracts into Oxigraph
3. Run query to see all existing integrations
4. Define new integration contract in your domain's contract
5. Get approval from both domain owners
6. Test using validation queries

### For Data Owners (Domain Stewards)

**I need to update my domain contract:**

1. Edit `[domain]/domain-contract.yaml`
2. Update quality metrics (timeliness, accuracy, completeness)
3. Add new entity if applicable
4. Update SLA if changed
5. Increment `contract_version` (semantic versioning)
6. Update `last_updated` date
7. Submit PR with changes

**My dataset's quality dropped below 0.90:**

1. Run [validation-queries.sparql](./validation-queries.sparql) → Query 1 (Data Quality Check)
2. Identify which metric failed (timeliness? accuracy?)
3. Investigate root cause:
   - Timeliness: Check ETL sync frequency
   - Accuracy: Check validation rules in source
   - Completeness: Check for null values in required fields
4. Apply fix
5. Re-run validation query to confirm passing
6. Update quality_metrics in contract
7. Open incident if SLA violated

### For Security/Compliance

**I need to audit data access:**

1. Run [validation-queries.sparql](./validation-queries.sparql) → Query 2 (Policy Compliance)
2. Identify any policies marked "NON_COMPLIANT"
3. Check for missing audit logging obligations
4. Check for missing constraints (critical for access control)
5. Remediate and re-run

**I need to validate compliance frameworks are covered:**

1. Run [validation-queries.sparql](./validation-queries.sparql) → Query 5 (Compliance Framework Coverage)
2. Check each domain declares required frameworks:
   - All: SOC2 Type II, GDPR
   - Finance: SOX
   - Marketing: CCPA
   - HR: FMLA, ADA, HIPAA
3. Flag any gaps

**I need to trace data lineage:**

1. Run [validation-queries.sparql](./validation-queries.sparql) → Query 3 (Lineage & Provenance)
2. See source system, ETL process, last sync, freshness
3. Verify sync error rate < 1%
4. Check for stale data (freshness > SLA max_latency)

### For Platform Engineers

**I want to expose these contracts via API:**

See `GET /api/data-mesh/contracts/` section in [CATALOG.md](./CATALOG.md)

**I want to generate schemas from contracts:**

```python
# Example: Convert YAML contract to JSON Schema
import yaml
import json

with open('finance/domain-contract.yaml') as f:
    contract = yaml.safe_load(f)

for entity in contract['entities']:
    schema = {
        "type": "object",
        "properties": {},
        "required": []
    }
    for field_name, field_desc in entity['key_fields'].items():
        schema['properties'][field_name] = {
            "description": field_desc,
            "type": "string"  # Infer from key_fields description
        }

    print(json.dumps(schema, indent=2))
```

**I want to load contracts into a semantic data lake:**

```bash
# Convert all YAML contracts to RDF
for file in */domain-contract.yaml; do
  python3 convert_yaml_to_ttl.py "$file" > "${file%.yaml}.ttl"
done

# Load into Oxigraph
obsr load *.ttl --format turtle

# Query with SPARQL
obsr query validation-queries.sparql --query "Data Quality Check"
```

---

## 🔐 Security & Privacy

### PII Handling by Domain

| Domain | PII Fields | Handling | Retention |
|--------|-----------|----------|-----------|
| Finance | Payment info | Masked, encrypted | 7 years (SOX) |
| Operations | None | N/A | 2-3 years |
| Marketing | Email, phone, IP | Encrypted, consent tracked | 2 years or until opt-out |
| Sales | Contact info, territory | Masked for non-sales | 3+ years (contract) |
| HR | SSN, DOB, health info | End-to-end encrypted | 7 years post-termination |

### Compliance Frameworks Covered

- ✅ **SOC2 Type II:** All domains (security audited)
- ✅ **GDPR:** All domains (right to deletion, portability, access)
- ✅ **SOX:** Finance domain (10-year audit trail)
- ✅ **PCI-DSS:** Finance payments (Level 1 certification)
- ✅ **CCPA:** Marketing + Sales (opt-out honored within 45 days)
- ✅ **FMLA:** HR domain (leave entitlement tracking)
- ✅ **HIPAA:** HR domain if benefits included (health info segregation)
- ✅ **ADA:** HR domain (accessibility accommodations)

---

## 📈 Quality Metrics Summary

**Current Status (as of 2026-03-25):**

| Domain | Overall Score | Trend | Status |
|--------|---------------|-------|--------|
| Finance | 0.96 | ↗️ +0.02 | ✅ EXCELLENT |
| Operations | 0.95 | → | ✅ EXCELLENT |
| Marketing | 0.94 | ↘️ -0.01 | ✅ GOOD |
| Sales | 0.96 | ↗️ +0.01 | ✅ EXCELLENT |
| HR | 0.98 | ↗️ +0.03 | ✅ EXCELLENT |

**Metric Definitions:**

- **Timeliness:** Time from source change to data mesh update (target: < 5 min)
- **Accuracy:** % of records with correct values (target: ≥ 99.9%)
- **Completeness:** % of non-null values in required fields (target: ≥ 99%)
- **Consistency:** % of records passing referential integrity (target: 100%)
- **Validity:** % of records passing schema validation (target: 100%)

**SLA Alert Thresholds:**

- 🔴 **CRITICAL:** Quality score < 0.80 or SLA violated
- 🟠 **WARNING:** Quality score 0.80-0.90 or approaching SLA limit
- 🟢 **HEALTHY:** Quality score ≥ 0.90 and all SLAs met

---

## 📋 Standards Used

### DCAT (Data Catalog Vocabulary)
- Describes datasets, distributions, catalogs
- Machine-readable dataset metadata
- Standard: https://www.w3.org/TR/vocab-dcat-2/

### schema.org
- Defines entity data types and properties
- Structured data markup for web
- Standard: https://schema.org

### DQV (Data Quality Vocabulary)
- Measures data quality metrics
- Specifies quality dimensions (accuracy, timeliness, etc.)
- Standard: https://www.w3.org/TR/vocab-dqv/

### ODRL (Open Digital Rights Language)
- Expresses access control policies
- Defines permissions, prohibitions, obligations
- Standard: https://www.w3.org/TR/odrl-model/

### PROV-O (Provenance Ontology)
- Traces data lineage and transformations
- Captures source systems and processes
- Standard: https://www.w3.org/TR/prov-o/

---

## 🚀 Getting Started

### 1. **Read the Catalog**
```bash
cat CATALOG.md
```

### 2. **Explore a Domain**
```bash
cat finance/domain-contract.yaml
```

### 3. **Convert to RDF**
```bash
# Requires: PyYAML, rdflib
python3 -c "
import yaml
from rdflib import Graph, Namespace, URIRef, Literal

with open('finance/domain-contract.yaml') as f:
    data = yaml.safe_load(f)
    # Convert to RDF triples
"
```

### 4. **Query with SPARQL**
```bash
# Requires: Oxigraph / Virtuoso
obsr query validation-queries.sparql
```

### 5. **Validate Your Contract**
```bash
# Use validation-queries.sparql:
# - Query 1: Data Quality Check
# - Query 2: Policy Compliance Check
# - Query 3: Lineage Verification
# - Query 4: Cross-Domain Integration
# - Query 5: Compliance Framework Coverage
```

---

## 📞 Support

### Ownership
- **Catalog Steward:** Chief Data Officer (cdo@businessos.example.com)
- **Finance Domain:** CFO (cfo@businessos.example.com)
- **Operations Domain:** COO (coo@businessos.example.com)
- **Marketing Domain:** CMO (cmo@businessos.example.com)
- **Sales Domain:** CRO (cro@businessos.example.com)
- **HR Domain:** CHRO (chro@businessos.example.com)

### Questions?
- **Data Access:** Contact domain owner
- **Technical Implementation:** Contact technical_owner in domain contract
- **Security/Compliance:** Contact Chief Information Security Officer
- **Contract Updates:** Submit PR to this directory

---

## 📄 Files Reference

| File | Purpose | Format |
|------|---------|--------|
| [CATALOG.md](./CATALOG.md) | Master index of all domains | Markdown |
| [TEMPLATE.md](./TEMPLATE.md) | How to create/update contracts | Markdown |
| [validation-queries.sparql](./validation-queries.sparql) | Data quality/compliance checks | SPARQL 1.1 |
| `finance/domain-contract.yaml` | Finance domain contract | YAML |
| `finance/domain-contract.ttl` | Finance contract (semantic) | RDF/Turtle |
| `operations/domain-contract.yaml` | Operations domain contract | YAML |
| `marketing/domain-contract.yaml` | Marketing domain contract | YAML |
| `sales/domain-contract.yaml` | Sales domain contract | YAML |
| `hr/domain-contract.yaml` | HR domain contract | YAML |

---

## 📅 Maintenance Schedule

- **Weekly:** Monitor quality metrics (run Query 1)
- **Monthly:** Audit access policies (run Query 2)
- **Monthly:** Verify data lineage (run Query 3)
- **Quarterly:** Review compliance frameworks (run Query 5)
- **Quarterly:** Full contract review + update
- **Annually:** Security audit + framework re-certification

---

## 📝 Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-03-25 | Initial catalog: 5 domains, 20 entities, 22 quality metrics |

---

**Last Updated:** 2026-03-25
**Next Review:** 2026-06-25 (Quarterly)
**Status:** ✅ All contracts published and validated
