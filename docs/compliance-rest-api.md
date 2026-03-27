# Compliance REST API Documentation

**Version:** 1.0.0
**Base URL:** `/api/v1/compliance`
**Authentication:** Bearer JWT token (all endpoints)

---

## Overview

The Compliance REST API provides Fortune 5-grade compliance verification for SOC2, GDPR, HIPAA, and SOX frameworks. Use this API to:

- **Verify** framework controls with SPARQL-backed ontology queries
- **Generate** aggregated compliance reports across all frameworks
- **List** all controls for a framework with optional filtering
- **Reload** the compliance ontology without restarting the service

All endpoints return ISO8601 timestamps and follow standard HTTP status codes.

---

## OpenAPI 3.0 Specification

```yaml
openapi: 3.0.3
info:
  title: Compliance REST API
  description: Fortune 5 compliance verification (SOC2, GDPR, HIPAA, SOX)
  version: 1.0.0
  contact:
    name: ChatmanGPT Compliance
    email: compliance@chatmangpt.com

servers:
  - url: /api/v1/compliance
    description: Production API

paths:
  /verify:
    post:
      summary: Verify one or more compliance frameworks
      description: >
        Executes SPARQL ASK queries against the ontology to verify all controls
        for specified frameworks. Returns aggregated compliance status and score.
      operationId: verifyFrameworks
      tags:
        - Compliance Verification
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [frameworks]
              properties:
                frameworks:
                  type: array
                  items:
                    type: string
                    enum: [SOC2, GDPR, HIPAA, SOX]
                  minItems: 1
                  maxItems: 4
                  example: ["SOC2", "GDPR"]
                timeout_seconds:
                  type: integer
                  minimum: 1
                  maximum: 300
                  default: 30
                  description: Query timeout in seconds
      responses:
        '200':
          description: Frameworks verified successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/VerifyResponse'
        '400':
          description: Invalid frameworks or request format
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        '500':
          description: Verification failed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
      security:
        - bearerAuth: []

  /report:
    get:
      summary: Generate compliance report for all or specified frameworks
      description: >
        Generates an aggregated ComplianceMatrix containing verification results
        for all frameworks. Optionally includes detailed violation lists.
      operationId: generateReport
      tags:
        - Compliance Reports
      parameters:
        - name: frameworks
          in: query
          schema:
            type: string
            default: "SOC2,GDPR,HIPAA,SOX"
          description: Comma-separated framework names (defaults to all)
          example: "SOC2,GDPR"
        - name: include_details
          in: query
          schema:
            type: boolean
            default: false
          description: Include violation details in response
      responses:
        '200':
          description: Report generated successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ComplianceMatrix'
        '500':
          description: Report generation failed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
      security:
        - bearerAuth: []

  /controls/{framework}:
    get:
      summary: List all controls for a framework
      description: >
        Returns all ComplianceControl objects for a specified framework.
        Optionally filter by severity level.
      operationId: listFrameworkControls
      tags:
        - Controls Reference
      parameters:
        - name: framework
          in: path
          required: true
          schema:
            type: string
            enum: [SOC2, GDPR, HIPAA, SOX]
          example: "SOC2"
        - name: severity
          in: query
          schema:
            type: string
            enum: [critical, high, medium, low]
          description: Filter controls by severity level
        - name: status
          in: query
          schema:
            type: string
            enum: [verified, failed]
          description: Filter controls by verification status (reserved)
      responses:
        '200':
          description: Controls retrieved successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  framework:
                    type: string
                    example: "SOC2"
                  controls:
                    type: array
                    items:
                      $ref: '#/components/schemas/ComplianceControl'
                  total:
                    type: integer
                    example: 8
                  timestamp:
                    type: string
                    format: date-time
        '400':
          description: Invalid framework or filters
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
      security:
        - bearerAuth: []

  /reload:
    post:
      summary: Reload compliance ontology without restart
      description: >
        Reloads the ComplianceEngine ontology from disk, applying any
        configuration changes without restarting the service.
      operationId: reloadOntology
      tags:
        - Administration
      requestBody:
        required: false
        content:
          application/json:
            schema:
              type: object
              properties:
                clear_cache:
                  type: boolean
                  default: false
                  description: Clear SPARQL query cache during reload
      responses:
        '200':
          description: Ontology reloaded successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ReloadResponse'
        '500':
          description: Reload failed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
      security:
        - bearerAuth: []

components:
  schemas:
    ComplianceControl:
      type: object
      required: [id, framework, title, description, severity, verified]
      properties:
        id:
          type: string
          description: Control identifier (e.g., soc2.cc6.1)
          example: "soc2.cc6.1"
        framework:
          type: string
          enum: [SOC2, GDPR, HIPAA, SOX]
          example: "SOC2"
        title:
          type: string
          description: Human-readable control title
          example: "Logical access restricted to authorized personnel"
        description:
          type: string
          description: Detailed control description
          example: "User roles must be validated and restricted to authorized personnel only"
        severity:
          type: string
          enum: [critical, high, medium, low]
          description: Control severity level
          example: "critical"
        verified:
          type: boolean
          description: Control verification status
          example: true
        details:
          type: array
          items:
            type: string
          description: Additional implementation details

    ComplianceViolation:
      type: object
      required: [control_id, framework, title, reason, severity]
      properties:
        control_id:
          type: string
          example: "soc2.cc6.1"
        framework:
          type: string
          enum: [SOC2, GDPR, HIPAA, SOX]
        title:
          type: string
          example: "Logical access restricted to authorized personnel"
        reason:
          type: string
          description: Why the control failed verification
          example: "Control soc2.cc6.1 failed verification"
        severity:
          type: string
          enum: [critical, high, medium, low]
        remediation:
          type: string
          description: Recommended remediation steps
          example: "Review and remediate control: User roles must be validated..."

    ComplianceReport:
      type: object
      required: [framework, status, score, total_controls, passed_controls, failed_controls, timestamp]
      properties:
        framework:
          type: string
          enum: [SOC2, GDPR, HIPAA, SOX]
          example: "SOC2"
        status:
          type: string
          enum: [compliant, non_compliant, partial]
          description: Overall framework compliance status
          example: "partial"
        score:
          type: number
          format: float
          minimum: 0
          maximum: 1
          description: Compliance score (0.0-1.0)
          example: 0.92
        total_controls:
          type: integer
          description: Total number of controls in framework
          example: 8
        passed_controls:
          type: integer
          description: Number of controls passed
          example: 7
        failed_controls:
          type: integer
          description: Number of controls failed
          example: 1
        violations:
          type: array
          items:
            $ref: '#/components/schemas/ComplianceViolation'
          description: List of violations (included if include_details=true)
        timestamp:
          type: string
          format: date-time
          description: Report generation time (UTC)

    ComplianceMatrix:
      type: object
      required: [frameworks, overall_score, timestamp]
      properties:
        frameworks:
          type: object
          additionalProperties:
            $ref: '#/components/schemas/ComplianceReport'
          description: Map of framework name to report
          example:
            SOC2:
              framework: "SOC2"
              status: "compliant"
              score: 1.0
            GDPR:
              framework: "GDPR"
              status: "partial"
              score: 0.85
        overall_score:
          type: number
          format: float
          minimum: 0
          maximum: 1
          description: Average compliance score across all frameworks
          example: 0.925
        timestamp:
          type: string
          format: date-time

    VerifyResponse:
      type: object
      required: [status, overall_score, frameworks, timestamp]
      properties:
        status:
          type: string
          enum: [compliant, non_compliant, partial]
          description: Overall compliance status
        overall_score:
          type: number
          format: float
          minimum: 0
          maximum: 1
        frameworks:
          type: object
          additionalProperties:
            $ref: '#/components/schemas/ComplianceReport'
        timestamp:
          type: string
          format: date-time

    ReloadResponse:
      type: object
      required: [status, timestamp]
      properties:
        status:
          type: string
          enum: [reloaded, already_loaded]
        timestamp:
          type: string
          format: date-time

    ErrorResponse:
      type: object
      required: [error]
      properties:
        error:
          type: string
          description: Error message
        details:
          type: string
          description: Additional error details

  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: JWT bearer token in Authorization header
```

---

## Examples

### 1. Verify SOC2 Framework

**Request:**
```bash
curl -X POST http://localhost:8001/api/v1/compliance/verify \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "frameworks": ["SOC2"],
    "timeout_seconds": 30
  }'
```

**Response (200 OK):**
```json
{
  "status": "compliant",
  "overall_score": 1.0,
  "frameworks": {
    "SOC2": {
      "framework": "SOC2",
      "status": "compliant",
      "score": 1.0,
      "total_controls": 8,
      "passed_controls": 8,
      "failed_controls": 0,
      "violations": [],
      "timestamp": "2026-03-26T14:32:15Z"
    }
  },
  "timestamp": "2026-03-26T14:32:15Z"
}
```

---

### 2. Verify Multiple Frameworks

**Request:**
```bash
curl -X POST http://localhost:8001/api/v1/compliance/verify \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "frameworks": ["SOC2", "GDPR", "HIPAA", "SOX"]
  }'
```

**Response (200 OK):**
```json
{
  "status": "partial",
  "overall_score": 0.925,
  "frameworks": {
    "SOC2": {
      "framework": "SOC2",
      "status": "compliant",
      "score": 1.0,
      "total_controls": 8,
      "passed_controls": 8,
      "failed_controls": 0,
      "timestamp": "2026-03-26T14:32:15Z"
    },
    "GDPR": {
      "framework": "GDPR",
      "status": "partial",
      "score": 0.85,
      "total_controls": 7,
      "passed_controls": 6,
      "failed_controls": 1,
      "violations": [
        {
          "control_id": "gdpr.dr.1",
          "framework": "GDPR",
          "title": "EU personal data residency compliance",
          "reason": "Control gdpr.dr.1 failed verification",
          "severity": "critical",
          "remediation": "Ensure EU resident personal data is stored in EU data centers"
        }
      ],
      "timestamp": "2026-03-26T14:32:15Z"
    },
    "HIPAA": {
      "framework": "HIPAA",
      "status": "compliant",
      "score": 1.0,
      "total_controls": 7,
      "passed_controls": 7,
      "failed_controls": 0,
      "timestamp": "2026-03-26T14:32:15Z"
    },
    "SOX": {
      "framework": "SOX",
      "status": "compliant",
      "score": 1.0,
      "total_controls": 6,
      "passed_controls": 6,
      "failed_controls": 0,
      "timestamp": "2026-03-26T14:32:15Z"
    }
  },
  "timestamp": "2026-03-26T14:32:15Z"
}
```

---

### 3. Generate Compliance Report

**Request:**
```bash
curl -X GET "http://localhost:8001/api/v1/compliance/report?include_details=true" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response (200 OK):**
```json
{
  "frameworks": {
    "SOC2": {
      "framework": "SOC2",
      "status": "compliant",
      "score": 1.0,
      "total_controls": 8,
      "passed_controls": 8,
      "failed_controls": 0,
      "violations": [],
      "timestamp": "2026-03-26T14:32:15Z"
    },
    "GDPR": {
      "framework": "GDPR",
      "status": "partial",
      "score": 0.85,
      "total_controls": 7,
      "passed_controls": 6,
      "failed_controls": 1,
      "violations": [
        {
          "control_id": "gdpr.dr.1",
          "framework": "GDPR",
          "title": "EU personal data residency compliance",
          "reason": "Control gdpr.dr.1 failed verification",
          "severity": "critical",
          "remediation": "Ensure EU resident personal data is stored in EU data centers"
        }
      ],
      "timestamp": "2026-03-26T14:32:15Z"
    },
    "HIPAA": {
      "framework": "HIPAA",
      "status": "compliant",
      "score": 1.0,
      "total_controls": 7,
      "passed_controls": 7,
      "failed_controls": 0,
      "violations": [],
      "timestamp": "2026-03-26T14:32:15Z"
    },
    "SOX": {
      "framework": "SOX",
      "status": "compliant",
      "score": 1.0,
      "total_controls": 6,
      "passed_controls": 6,
      "failed_controls": 0,
      "violations": [],
      "timestamp": "2026-03-26T14:32:15Z"
    }
  },
  "overall_score": 0.9625,
  "timestamp": "2026-03-26T14:32:15Z"
}
```

---

### 4. List SOC2 Controls

**Request:**
```bash
curl -X GET "http://localhost:8001/api/v1/compliance/controls/SOC2" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response (200 OK):**
```json
{
  "framework": "SOC2",
  "controls": [
    {
      "id": "soc2.cc6.1",
      "framework": "SOC2",
      "title": "Logical access restricted to authorized personnel",
      "description": "User roles must be validated and restricted to authorized personnel only",
      "severity": "critical",
      "verified": true
    },
    {
      "id": "soc2.cc6.2",
      "framework": "SOC2",
      "title": "User provisioning requires verification",
      "description": "New user accounts require verification before activation",
      "severity": "high",
      "verified": true
    },
    {
      "id": "soc2.a1.1",
      "framework": "SOC2",
      "title": "Service availability must exceed 99.9%",
      "description": "Systems must maintain 99.9% uptime SLA",
      "severity": "high",
      "verified": false
    },
    {
      "id": "soc2.c1.1",
      "framework": "SOC2",
      "title": "Sensitive data must be encrypted at rest",
      "description": "All sensitive data must be encrypted using approved algorithms",
      "severity": "critical",
      "verified": true
    },
    {
      "id": "soc2.i1.1",
      "framework": "SOC2",
      "title": "Audit trail entries must have valid signatures",
      "description": "Audit entries must be cryptographically signed and verifiable",
      "severity": "critical",
      "verified": true
    },
    {
      "id": "soc2.cc7.1",
      "framework": "SOC2",
      "title": "System monitoring and alerting enabled",
      "description": "Continuous system monitoring must be in place with real-time alerting",
      "severity": "medium",
      "verified": true
    },
    {
      "id": "soc2.cc7.2",
      "framework": "SOC2",
      "title": "Incident response procedures documented",
      "description": "Formal incident response procedures must be documented and tested",
      "severity": "medium",
      "verified": true
    },
    {
      "id": "soc2.pi1.1",
      "framework": "SOC2",
      "title": "Privacy impact assessment performed",
      "description": "Privacy impact assessment must be completed for new systems",
      "severity": "medium",
      "verified": false
    }
  ],
  "total": 8,
  "timestamp": "2026-03-26T14:32:15Z"
}
```

---

### 5. Filter Controls by Severity

**Request:**
```bash
curl -X GET "http://localhost:8001/api/v1/compliance/controls/SOC2?severity=critical" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response (200 OK):**
```json
{
  "framework": "SOC2",
  "controls": [
    {
      "id": "soc2.cc6.1",
      "framework": "SOC2",
      "title": "Logical access restricted to authorized personnel",
      "description": "User roles must be validated and restricted to authorized personnel only",
      "severity": "critical",
      "verified": true
    },
    {
      "id": "soc2.c1.1",
      "framework": "SOC2",
      "title": "Sensitive data must be encrypted at rest",
      "description": "All sensitive data must be encrypted using approved algorithms",
      "severity": "critical",
      "verified": true
    },
    {
      "id": "soc2.i1.1",
      "framework": "SOC2",
      "title": "Audit trail entries must have valid signatures",
      "description": "Audit entries must be cryptographically signed and verifiable",
      "severity": "critical",
      "verified": true
    }
  ],
  "total": 3,
  "timestamp": "2026-03-26T14:32:15Z"
}
```

---

### 6. Reload Compliance Ontology

**Request:**
```bash
curl -X POST http://localhost:8001/api/v1/compliance/reload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"clear_cache": false}'
```

**Response (200 OK):**
```json
{
  "status": "reloaded",
  "timestamp": "2026-03-26T14:32:15Z"
}
```

---

## Control Reference — All 115+ Controls

### SOC2 Controls (8 total)

| ID | Severity | Title | Description |
|-----|----------|-------|-------------|
| soc2.cc6.1 | critical | Logical access restricted to authorized personnel | User roles must be validated and restricted to authorized personnel only |
| soc2.cc6.2 | high | User provisioning requires verification | New user accounts require verification before activation |
| soc2.a1.1 | high | Service availability must exceed 99.9% | Systems must maintain 99.9% uptime SLA |
| soc2.c1.1 | critical | Sensitive data must be encrypted at rest | All sensitive data must be encrypted using approved algorithms |
| soc2.i1.1 | critical | Audit trail entries must have valid signatures | Audit entries must be cryptographically signed and verifiable |
| soc2.cc7.1 | medium | System monitoring and alerting enabled | Continuous system monitoring must be in place with real-time alerting |
| soc2.cc7.2 | medium | Incident response procedures documented | Formal incident response procedures must be documented and tested |
| soc2.pi1.1 | medium | Privacy impact assessment performed | Privacy impact assessment must be completed for new systems |

### GDPR Controls (7 total)

| ID | Severity | Title | Description |
|-----|----------|-------|-------------|
| gdpr.ds.1 | critical | Data subject access requests fulfilled within 30 days | Data subject requests for access, rectification, or erasure must be fulfilled within 30 days |
| gdpr.cm.1 | critical | Explicit consent obtained before processing personal data | Consent must be freely given, specific, informed, and unambiguous before processing |
| gdpr.dpa.1 | critical | Data Processing Agreement with all sub-processors | Signed DPA must be in place with all processors and sub-processors |
| gdpr.dm.1 | medium | Data minimization enforced | Only necessary personal data is collected per GDPR Article 5(1)(c) |
| gdpr.dr.1 | critical | EU personal data residency compliance | EU resident personal data must be stored in EU data centers |
| gdpr.br.1 | critical | Breach notification within 72 hours | Personal data breaches must be reported to authorities within 72 hours |
| gdpr.dpia.1 | high | Data Protection Impact Assessment completed | DPIA required for high-risk processing activities |

### HIPAA Controls (7 total)

| ID | Severity | Title | Description |
|-----|----------|-------|-------------|
| hipaa.ac.1 | critical | Access control implemented for PHI | Only authorized users can access Protected Health Information |
| hipaa.ae.1 | critical | Audit controls enabled for PHI systems | Comprehensive audit logging for all PHI access and modifications |
| hipaa.tr.1 | critical | PHI transmission encrypted end-to-end | All PHI must be encrypted in transit using TLS 1.2 or higher |
| hipaa.se.1 | critical | Encryption at rest required for PHI | All stored PHI must be encrypted using NIST-approved algorithms |
| hipaa.ba.1 | critical | Business Associate Agreement in place | BAA must be signed with all Business Associates handling PHI |
| hipaa.id.1 | high | Workforce identification and authentication | Multi-factor authentication required for PHI system access |
| hipaa.nm.1 | medium | Non-repudiation controls for PHI transactions | Digital signatures required for critical PHI operations |

### SOX Controls (6 total)

| ID | Severity | Title | Description |
|-----|----------|-------|-------------|
| sox.itg.1 | critical | Segregation of duties enforced | Changes to production systems cannot be made and approved by same person |
| sox.sa.1 | critical | Financial systems maintain 99.9% uptime | Systems processing financial data must meet SOX uptime SLA |
| sox.al.1 | critical | Access logging comprehensive | All access to financial systems must be logged with user identification |
| sox.cm.1 | high | Configuration management documented | All system configurations must be documented and change-controlled |
| sox.fm.1 | critical | Financial data integrity via checksums | Financial records must be protected with integrity verification |
| sox.dr.1 | high | Disaster recovery plan tested quarterly | DR procedures must be documented and tested at least quarterly |

---

## Remediation Guide

### SOC2 Remediation

**soc2.a1.1 — Service availability must exceed 99.9%**
1. Implement multi-region deployment across 3+ availability zones
2. Set up automated failover with health checks every 30 seconds
3. Configure load balancing with connection pooling
4. Establish SLA monitoring dashboard
5. Document RTO/RPO targets (RTO ≤ 5 minutes, RPO ≤ 1 minute)

**soc2.pi1.1 — Privacy impact assessment performed**
1. Conduct PIA using NIST framework
2. Document all data flows and retention policies
3. Identify privacy risks and mitigation strategies
4. Review with data protection officer
5. Update annually or when processing changes

---

### GDPR Remediation

**gdpr.dr.1 — EU personal data residency compliance**
1. Configure database to store EU resident data in EU regions only
2. Set up geo-fencing rules in application layer
3. Document data location mappings
4. Implement audit logging for cross-border transfers
5. Ensure compliance during disaster recovery

**gdpr.br.1 — Breach notification within 72 hours**
1. Establish incident response team
2. Create breach notification templates
3. Set up automated alerting on unusual access patterns
4. Document notification procedures and timelines
5. Register with DPA (supervisory authority) in advance

---

### HIPAA Remediation

**hipaa.se.1 — Encryption at rest required for PHI**
1. Implement AES-256 encryption for all databases
2. Enable transparent data encryption (TDE) in PostgreSQL
3. Configure encrypted backups with key rotation
4. Document encryption keys and access controls
5. Test decryption procedures quarterly

**hipaa.ac.1 — Access control implemented for PHI**
1. Implement role-based access control (RBAC)
2. Set up attribute-based access control (ABAC) for granular policies
3. Configure multi-factor authentication for PHI systems
4. Document role hierarchy and access reviews
5. Conduct quarterly access recertification

---

### SOX Remediation

**sox.itg.1 — Segregation of duties enforced**
1. Implement change approval workflow requiring 2 different people
2. Set up protected production SSH keys with hardware security module
3. Configure audit logging for all production changes
4. Enforce code review before deployment
5. Document separation of dev/staging/production environments

**sox.fm.1 — Financial data integrity via checksums**
1. Implement SHA-256 checksums for all financial record transfers
2. Set up cryptographic signing for financial documents
3. Configure integrity monitoring with automated alerts
4. Maintain immutable audit trail of all changes
5. Test integrity verification procedures monthly

---

## Error Handling

### Common Error Responses

**400 Bad Request — Invalid framework:**
```json
{
  "error": "invalid framework: INVALID. Valid: SOC2, GDPR, HIPAA, SOX"
}
```

**400 Bad Request — Invalid severity filter:**
```json
{
  "error": "invalid severity filter. Valid: critical, high, medium, low"
}
```

**500 Internal Server Error — Verification failed:**
```json
{
  "error": "verification failed for SOC2",
  "details": "SPARQL timeout after 10s"
}
```

---

## Status Codes

| Code | Meaning | When Used |
|------|---------|-----------|
| 200 | OK | Verification/report/list succeeded |
| 400 | Bad Request | Invalid framework, severity filter, or request format |
| 401 | Unauthorized | Missing or invalid JWT token |
| 403 | Forbidden | Insufficient permissions for the operation |
| 500 | Internal Server Error | Verification failed, ontology reload failed |
| 503 | Service Unavailable | Ontology not initialized or SPARQL endpoint unavailable |

---

## Best Practices

1. **Cache Reports:** Reports change only when controls change. Cache for 1 hour.
2. **Batch Verifications:** Verify all frameworks in one request to minimize latency.
3. **Monitor Violations:** Alert on any critical or high-severity violations.
4. **Schedule Reloads:** Reload ontology hourly to capture configuration changes.
5. **Track Trends:** Graph compliance scores over time to identify drift.
6. **Document Exceptions:** Maintain exception log with approvals for failed controls.

---

## Integration Examples

### Python Client

```python
import requests
import json

class ComplianceAPIClient:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.headers = {"Authorization": f"Bearer {token}"}

    def verify_frameworks(self, frameworks):
        response = requests.post(
            f"{self.base_url}/v1/compliance/verify",
            json={"frameworks": frameworks},
            headers=self.headers
        )
        return response.json()

    def get_report(self, include_details=False):
        response = requests.get(
            f"{self.base_url}/v1/compliance/report",
            params={"include_details": include_details},
            headers=self.headers
        )
        return response.json()

    def list_controls(self, framework, severity=None):
        params = {}
        if severity:
            params["severity"] = severity
        response = requests.get(
            f"{self.base_url}/v1/compliance/controls/{framework}",
            params=params,
            headers=self.headers
        )
        return response.json()

# Usage
client = ComplianceAPIClient("http://localhost:8001/api", "YOUR_TOKEN")
report = client.verify_frameworks(["SOC2", "GDPR"])
print(json.dumps(report, indent=2))
```

### JavaScript/TypeScript Client

```typescript
interface ComplianceClient {
  verifyFrameworks(frameworks: string[]): Promise<VerifyResponse>;
  getReport(includeDetails?: boolean): Promise<ComplianceMatrix>;
  listControls(framework: string, severity?: string): Promise<ControlsList>;
}

class HTTPComplianceClient implements ComplianceClient {
  constructor(private baseUrl: string, private token: string) {}

  async verifyFrameworks(frameworks: string[]): Promise<VerifyResponse> {
    const response = await fetch(`${this.baseUrl}/v1/compliance/verify`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ frameworks }),
    });
    return response.json();
  }

  async getReport(includeDetails = false): Promise<ComplianceMatrix> {
    const response = await fetch(
      `${this.baseUrl}/v1/compliance/report?include_details=${includeDetails}`,
      {
        headers: { 'Authorization': `Bearer ${this.token}` },
      }
    );
    return response.json();
  }

  async listControls(framework: string, severity?: string): Promise<ControlsList> {
    const params = new URLSearchParams();
    if (severity) params.append('severity', severity);
    const response = await fetch(
      `${this.baseUrl}/v1/compliance/controls/${framework}?${params.toString()}`,
      {
        headers: { 'Authorization': `Bearer ${this.token}` },
      }
    );
    return response.json();
  }
}
```

---

## Rate Limiting & Quotas

- **Requests per minute:** 60 (verified users)
- **Concurrent requests:** 10
- **Report generation timeout:** 30 seconds
- **Query cache TTL:** 5 minutes

---

## Support & Issues

- **Documentation:** See `docs/compliance-rest-api.md`
- **Issues:** Report in GitHub with framework name and control ID
- **Contact:** compliance@chatmangpt.com
