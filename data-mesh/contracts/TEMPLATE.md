# Data Mesh Domain Contract Template

This template standardizes how all BusinessOS domains define, expose, and govern data assets.

## Format Options

Choose one or more formats:
- **YAML** (human-readable, recommended for operations)
- **RDF/Turtle** (machine-readable, semantic web compatible)
- **JSON** (programmatic consumption, API-first)

## Complete Contract Template (YAML)

```yaml
---
# ========================================================================
# [DOMAIN_NAME] Domain Data Mesh Contract
# ========================================================================
# Version: X.Y.Z
# Owner: [Owner Title]
# Created: [Date]
# ========================================================================

domain:
  name: "[Domain Name]"
  owner: "[Owner Title/Name]"
  contact_email: "[email]"
  contact_phone: "[phone]"
  department: "[Department]"
  business_unit: "[Business Unit]"
  cost_center: "[Cost Center]"

governance:
  data_steward: "[Name/Title]"
  data_steward_email: "[email]"
  technical_owner: "[Name/Title]"
  technical_owner_email: "[email]"
  security_owner: "[Name/Title]"
  privacy_officer: "[Name/Title (if applicable)]"
  compliance_frameworks:
    - "[Framework 1]"
    - "[Framework 2]"
  audit_frequency: "[Frequency]"
  audit_schedule: "[Details]"

# ========================================================================
# DATA ENTITIES
# ========================================================================

entities:
  - name: "[Entity Name]"
    identifier: "[domain.entity.version]"
    description: "[Business definition]"
    record_count: [number]
    growth_rate: "[percentage/timeframe]"
    business_criticality: "[CRITICAL|HIGH|MEDIUM|LOW]"

    # For sensitive data:
    pii_fields:
      - "[field_name]"
    pii_handling: "[How PII is encrypted/masked/minimized]"

    distribution:
      api:
        endpoint: "[URL]"
        format: "[JSON|XML|etc]"
        rate_limit: [number]
        rate_limit_window: "[timeframe]"
        authentication: "[method]"
        response_time_sla: "[latency SLA]"
        documentation: "[URL]"
        requires_justification: [true|false]

      database:
        host: "[hostname]"
        port: [number]
        database: "[db_name]"
        schema: "[schema_name]"
        table: "[table_name]"
        authentication: "[method]"
        replication_lag: "[latency]"
        encryption: "[type]"
        backup_frequency: "[frequency]"

    key_fields:
      - "[field_name]": "[description/type]"

    relationships:
      - to: "[Related Entity]"
        type: "[1:1|1:N|N:1|N:N]"
        foreign_key: "[key_name]"
        constraint: "[Description of relationship]"

    lineage:
      source_system: "[System name]"
      source_system_version: "[Version]"
      sync_frequency: "[Frequency]"
      sync_method: "[Change Data Capture|Batch|Event-driven]"
      last_sync: "[Timestamp]"
      records_synced: [number]
      sync_success_rate: "[percentage]"
      etl_process: "[Process name]"

    # More entities follow same pattern...

# ========================================================================
# DATA QUALITY METRICS (DQV - Data Quality Vocabulary)
# ========================================================================

quality_metrics:
  "[entity_name]":
    timeliness:
      metric: "[Description]"
      expected_value: "[Value]"
      actual_value: "[Value]"
      measurement_method: "[How measured]"
      monitoring_tool: "[Tool name]"
      status: "[PASSING|WARNING|FAILING]"

    accuracy:
      metric: "[Description]"
      expected_value: "[Value]"
      actual_value: "[Value]"
      validation_rules: "[Business rules]"
      status: "[PASSING|WARNING|FAILING]"

    completeness:
      metric: "[Description]"
      expected_value: "[Percentage]"
      actual_value: "[Percentage]"
      measured_fields: "[field1, field2]"
      required_fields: "[field1, field2]"
      status: "[PASSING|WARNING|FAILING]"

    consistency:
      metric: "[Description]"
      expected_value: "[Target]"
      actual_value: "[Actual]"
      validation_rules: "[Cross-entity rules]"
      status: "[PASSING|WARNING|FAILING]"

    validity:
      metric: "[Description]"
      expected_value: "[Target]"
      actual_value: "[Actual]"
      validation_rules: "[Schema rules]"
      status: "[PASSING|WARNING|FAILING]"

    overall_score: [0.0-1.0]

# ========================================================================
# ACCESS CONTROL (ODRL - Open Digital Rights Language)
# ========================================================================

access_control:
  "[role_or_team_name]":
    roles:
      - "[Role 1]"
      - "[Role 2]"
    permissions:
      "[entity]": "[READ|WRITE|DELETE|EXPORT|etc]"
    constraints:
      "[constraint_type]": "[constraint_value]"
      time_window: "[24/7|business-hours]"
      field_masking: "[list of masked fields]"
      volume_limit: "[limit or unlimited]"
      territory_scoping: "[description]"
      aggregation_required: "[true|false]"

# ========================================================================
# USAGE POLICIES (ODRL)
# ========================================================================

usage_policies:
  export_policy:
    description: "[Policy summary]"
    rules:
      - role: "[Role]"
        action: "[Action]"
        allowed: [true|false]
        constraints:
          - "[constraint]"

  retention_policy:
    "[entity]":
      retention_period: "[Duration]"
      reason: "[Why retained]"
      archival: "[true|false]"
      auto_deletion: "[true|false]"

  audit_policy:
    description: "[Audit requirements]"
    logging_level: "[DEBUG|INFO|WARN|ERROR]"
    log_retention: "[Duration]"
    log_fields: "[field1, field2]"

  encryption_policy:
    in_transit:
      protocol: "[TLS version]"
      min_version: "[version]"
    at_rest:
      algorithm: "[Algorithm]"
      key_rotation: "[Frequency]"
      pii_fields: "[encrypted|masked|minimized]"

  pii_policy:
    description: "[PII handling]"
    sensitive_fields:
      "[entity]":
        - "[field1]"
        - "[field2]"
    handling_method: "[Encryption|Masking|Minimization]"
    redaction_enabled: [true|false]

  data_minimization:
    rules:
      - "[Rule]"

  compliance_policy:
    frameworks:
      - "[Framework]"
    rules:
      - "[Rule]"

# ========================================================================
# SERVICE LEVEL AGREEMENTS (SLAs)
# ========================================================================

service_level_agreements:
  "[entity]":
    availability: "[Percentage]"
    max_latency: "[Duration]"
    max_latency_unit: "[minutes|hours|days]"
    data_refresh_frequency: "[Frequency]"
    rto_minutes: [number]
    rpo_minutes: [number]
    support_tier: "[Tier level]"
    escalation_contact: "[Contact info]"
    incident_response_time: "[Duration]"
    planned_maintenance_window: "[Time/Frequency]"
    backup_frequency: "[Frequency]"
    disaster_recovery_location: "[Location]"
    soc2_compliance: "[Type/Level]"

# ========================================================================
# INTEGRATION CONTRACTS (A2A - Agent-to-Agent)
# ========================================================================

integration_contracts:
  "[source]_to_[target]":
    protocol: "[MCP|JSON-RPC|REST|gRPC]"
    endpoint: "[URL/Path]"
    frequency: "[Real-time|Daily|Weekly]"
    sync_direction: "[Uni|Bi-directional]"
    contract_version: "[Version]"
    sla:
      availability: "[Percentage]"
      response_time: "[Latency]"
      rate_limit: "[limit/timeframe]"
    error_handling: "[Retry|Fallback|Alert]"

# ========================================================================
# MONITORING & ALERTING
# ========================================================================

monitoring_and_alerting:
  metrics:
    - name: "[metric_name]"
      threshold: [value]
      alert_severity: "[critical|warning|info]"

  alert_channels:
    - type: "[slack|email|pagerduty|etc]"
      recipients: "[recipients]"
      channel: "[channel_id]"

# ========================================================================
# DATA SHARING AGREEMENTS
# ========================================================================

data_sharing_agreements:
  with_[domain]:
    frequency: "[Frequency]"
    data_shared: "[What data]"
    quality_sla: "[SLA]"
    renewal_date: "[Date]"
    contact: "[Contact]"

# ========================================================================
# DOCUMENTATION
# ========================================================================

documentation:
  api_documentation: "[URL]"
  schema_documentation: "[URL]"
  governance_wiki: "[URL]"
  support_contact: "[Contact]"
  support_hours: "[Hours]"
  runbook_url: "[URL]"

# ========================================================================
# VERSIONING & REVIEW
# ========================================================================

versioning:
  contract_version: "[X.Y.Z]"
  schema_version: "[Version]"
  created_date: "[Date]"
  last_updated: "[Date]"
  last_audit_date: "[Date]"
  next_review_date: "[Date]"
  review_frequency: "[Frequency]"
  security_review_frequency: "[Frequency]"
  compliance_review_frequency: "[Frequency]"
```

## Minimal Contract (for simple domains)

If a domain has few entities or less complex governance, use this reduced template:

```yaml
---
domain:
  name: "[Domain Name]"
  owner: "[Owner]"
  contact_email: "[email]"

entities:
  - name: "[Entity]"
    identifier: "[domain.entity.v1]"
    description: "[Description]"
    record_count: [number]

    distribution:
      api:
        endpoint: "[URL]"
        authentication: "[method]"

quality_metrics:
  "[entity]":
    overall_score: [0.0-1.0]
    status: "[PASSING|WARNING|FAILING]"

access_control:
  "[role]":
    permissions:
      "[entity]": "[READ|WRITE|etc]"

service_level_agreements:
  "[entity]":
    availability: "[Percentage]"
    max_latency: "[Duration]"
    support_tier: "[Level]"

versioning:
  contract_version: "1.0.0"
  created_date: "[Date]"
  next_review_date: "[Date]"
```

## RDF/Turtle Format (Machine-Readable)

For semantic web / SPARQL integration:

```turtle
@prefix dcat: <http://www.w3.org/ns/dcat#> .
@prefix dct: <http://purl.org/dc/terms/> .
@prefix dqv: <http://www.w3.org/ns/dqv#> .
@prefix odrl: <http://www.w3.org/ns/odrl/2/> .
@prefix prov: <http://www.w3.org/ns/prov#> .
@prefix ex: <http://businessos.example.com/> .

ex:[DomainName]Catalog a dcat:Catalog ;
  dct:title "[Domain Name] Catalog" ;
  dcat:dataset ex:[Entity1], ex:[Entity2] ;
  dct:issued "[Date]"^^xsd:date ;
  dct:modified "[Date]"^^xsd:date ;
  dct:publisher ex:[Owner] .

ex:[Entity] a dcat:Dataset ;
  dct:title "[Entity Title]" ;
  dct:description "[Description]" ;
  dct:identifier "[domain.entity.v1]" ;
  dcat:distribution ex:[Distribution] ;
  dqv:hasQualityMeasurement ex:[QualityMetric] ;
  odrl:hasPolicy ex:[AccessPolicy] ;
  prov:hadPrimarySource ex:[SourceSystem] .
```

## JSON Format (API-Driven)

For programmatic contract discovery and consumption:

```json
{
  "domain": {
    "name": "[Domain Name]",
    "owner": "[Owner]",
    "contact_email": "[email]"
  },
  "entities": [
    {
      "name": "[Entity]",
      "identifier": "[domain.entity.v1]",
      "distribution": {
        "api": {
          "endpoint": "[URL]",
          "format": "JSON",
          "authentication": "[method]"
        }
      }
    }
  ],
  "quality_metrics": {
    "[entity]": {
      "overall_score": 0.95,
      "status": "PASSING"
    }
  },
  "access_control": {
    "[role]": {
      "permissions": {
        "[entity]": "READ"
      }
    }
  },
  "versioning": {
    "contract_version": "1.0.0",
    "created_date": "[Date]"
  }
}
```

## Publishing Your Contract

1. **Version control**: Commit contract to `data-mesh/contracts/[domain]/`
2. **Register catalog**: Add to `data-mesh/contracts/CATALOG.md`
3. **DCAT index**: Generate RDF/Turtle for SPARQL queries
4. **API endpoint**: Expose via `/api/data-mesh/contracts/[domain]`
5. **Documentation**: Link from domain wiki

## Quality Checklist

Before publishing, ensure:

- [ ] All entities documented with business descriptions
- [ ] Quality metrics show passing status (green) or have action plan
- [ ] Access control reflects actual RBAC matrix
- [ ] SLAs are realistic and measurable
- [ ] Encryption policy covers all PII fields
- [ ] Compliance frameworks listed and verified
- [ ] Data owner/steward contact info valid
- [ ] Integration contracts specify protocol/endpoint/frequency
- [ ] Alert channels configured and tested
- [ ] Review dates scheduled
- [ ] Semantic versioning used (major.minor.patch)
- [ ] All contacts have manager approval

## Example: Minimal Entity Definition

```yaml
entities:
  - name: "Customer"
    identifier: "sales.customers.v1"
    description: "Customer master records"
    record_count: 25000
    business_criticality: "CRITICAL"

    distribution:
      api:
        endpoint: "https://api.businessos.example.com/v1/sales/customers"
        authentication: "OAuth2"
        rate_limit: 1000

    key_fields:
      - customer_id: "Unique identifier"
      - name: "Customer name"
      - email: "Contact email"

    quality_metrics:
      overall_score: 0.97
      status: "PASSING"

    access_control:
      sales_team:
        permissions:
          customer: "READ, WRITE"
        constraints:
          territory_scoping: "strict"

    service_level_agreements:
      availability: 99.5%
      max_latency: "5 minutes"
      support_tier: "24/7 Level-2"
```

## Lifecycle Management

1. **Draft** (version 0.x) - Under development, not for consumption
2. **Published** (version 1.x) - Available for integration
3. **Stable** (version 2.x+) - Backwards compatible changes
4. **Deprecated** (marked in next_review_date) - Plan migration
5. **Archived** (removed from catalog) - Historical reference only

Each domain contract has 6-month review cycle with quarterly quality audits.

See [CATALOG.md](./CATALOG.md) for all published contracts.
