# Security Test Suite

Comprehensive security testing for BusinessOS backend, validating OWASP Top 10 protections and security best practices.

## Test Coverage

### 1. Injection Prevention (`injection_test.go`)
Tests that prevent various injection attacks:

- **SQL Injection**: Validates that sqlc's parameterized queries prevent SQL injection
- **Command Injection**: Tests terminal and container endpoint sanitization
- **LDAP Injection**: Validates LDAP filter escaping (if applicable)
- **NoSQL Injection**: Ensures PostgreSQL JSONB doesn't execute NoSQL operators

**Test Cases:**
- `' OR '1'='1`
- `'; DROP TABLE users; --`
- `$(whoami)`, `` `whoami` ``, `|ls -la`
- `*)(uid=*` (LDAP)
- `{"$gt": ""}` (NoSQL)

### 2. Cross-Site Scripting (XSS) Prevention (`xss_test.go`)
Tests XSS prevention across different contexts:

- **Reflected XSS**: Query parameter sanitization
- **Stored XSS**: User content escaping (messages, names, descriptions)
- **DOM-based XSS**: URL fragment and innerHTML usage
- **CSP**: Content-Security-Policy header configuration

**Test Cases:**
- `<script>alert('XSS')</script>`
- `<img src=x onerror=alert('XSS')>`
- `javascript:alert('XSS')`
- JSON encoding validation
- HTML entity encoding

### 3. Authentication & Session Security (`auth_test.go`)
Tests authentication mechanisms:

- **Password Hashing**: bcrypt with cost factor 12
- **Password Policy**: Minimum 8 chars, complexity requirements
- **JWT Security**: Signature validation, expiration, algorithm verification
- **Session Fixation**: New session on login
- **Credential Stuffing**: Rate limiting (5 attempts, 15 min lockout)
- **Refresh Token Rotation**: One-time use, automatic rotation

**Test Cases:**
- Password strength validation
- JWT manipulation attempts
- Token expiration handling
- Brute force protection

### 4. CSRF Protection (`csrf_test.go`)
Tests Cross-Site Request Forgery protection:

- **Token Generation**: Cryptographically random, unique
- **Token Validation**: Required for POST/PUT/DELETE
- **Token Reuse**: One-time use tokens
- **SameSite Cookies**: Lax/Strict configuration
- **Origin Validation**: Header checking

**Test Cases:**
- Valid/invalid CSRF tokens
- Missing tokens on state-changing operations
- SameSite=Lax/Strict enforcement
- Origin header validation

### 5. Access Control (`access_control_test.go`)
Tests authorization and access control:

- **Horizontal Privilege Escalation**: User A cannot access User B's resources
- **Vertical Privilege Escalation**: Regular users cannot perform admin actions
- **Path Traversal**: `../../../etc/passwd` prevention
- **IDOR**: Insecure Direct Object Reference prevention
- **Workspace Isolation**: Multi-tenant resource isolation
- **Role-Based Access**: Admin/Member/Viewer permissions

**Test Cases:**
- Cross-user resource access
- Admin-only endpoint protection
- Path traversal patterns
- UUID-based ID enumeration prevention

### 6. Rate Limiting (`rate_limiting_test.go`)
Tests rate limiting enforcement:

- **API Rate Limiting**: 100 requests/minute general limit
- **Login Rate Limiting**: 5 attempts per 15 minutes
- **Agent Requests**: 100 requests/hour per user
- **OSA Generation**: 10 apps/day per user
- **Rate Limit Headers**: X-RateLimit-Limit, Remaining, Reset, Retry-After

**Test Cases:**
- Request counting in time windows
- 429 Too Many Requests responses
- Lockout duration enforcement
- Rate limit header validation

### 7. Sensitive Data Protection (`sensitive_data_test.go`)
Tests sensitive data handling:

- **Logging Security**: Secrets masked in logs (use `slog`, not `fmt.Printf`)
- **API Responses**: API keys and tokens excluded/masked
- **Password Hashing**: bcrypt cost 12+
- **Error Masking**: Credentials not in error messages
- **JWT Secrets**: Not in environment dumps
- **Encryption at Rest**: OAuth tokens and API keys encrypted

**Test Cases:**
- Token masking in logs
- Sensitive field filtering
- Database credential masking
- slog structured logging validation

### 8. Security Headers (`headers_test.go`)
Tests HTTP security headers:

- **X-Frame-Options**: DENY (clickjacking prevention)
- **X-Content-Type-Options**: nosniff
- **Strict-Transport-Security**: max-age=31536000, includeSubDomains
- **Content-Security-Policy**: Restrictive CSP without unsafe-inline/unsafe-eval
- **Referrer-Policy**: strict-origin-when-cross-origin
- **Permissions-Policy**: Restrict geolocation, camera, microphone, etc.
- **CORS Headers**: No wildcard with credentials

**Test Cases:**
- All security headers present
- CSP directive validation
- HSTS configuration
- Cookie flags (Secure, HttpOnly, SameSite)

## Running Tests

### Run all security tests
```bash
cd desktop/backend-go
go test -v ./tests/security/...
```

### Run specific test suite
```bash
go test -v ./tests/security/injection_test.go
go test -v ./tests/security/auth_test.go
go test -v ./tests/security/xss_test.go
```

### Run with coverage
```bash
go test -v -coverprofile=security-coverage.out ./tests/security/...
go tool cover -html=security-coverage.out
```

### Run with race detector
```bash
go test -v -race ./tests/security/...
```

## CI/CD Integration

Security tests run automatically on:
- Every push to main, main-dev, pedro-dev
- Every pull request to main, main-dev
- Weekly schedule (Monday 6 AM UTC)
- Manual workflow dispatch

### GitHub Actions Workflow
`.github/workflows/security-tests.yml` runs:
1. Security test suite
2. GoSec static analysis
3. Nancy dependency vulnerability scan
4. OWASP ZAP baseline scan (scheduled runs only)

### Additional Security Scans
Configured in `.github/workflows/security-scan.yml`:
- **Trivy**: Docker image vulnerability scanning
- **Snyk**: Dependency vulnerability scanning
- **govulncheck**: Go vulnerability database

## Security Tools

### GoSec - Static Analysis
```bash
go install github.com/securego/gosec/v2/cmd/gosec@latest
cd desktop/backend-go
gosec -fmt=json -out=gosec-report.json ./...
```

### Nancy - Dependency Vulnerabilities
```bash
go install github.com/sonatype-nexus-community/nancy@latest
cd desktop/backend-go
go list -json -m all | nancy sleuth
```

### OWASP ZAP - Dynamic Scanning
```bash
docker run -v $(pwd):/zap/wrk/:rw -t owasp/zap2docker-stable \
  zap-baseline.py -t http://localhost:8001 -r zap-report.html
```

### npm audit - Frontend Dependencies
```bash
cd frontend
npm audit --audit-level=high
```

## Test Failure Handling

### CRITICAL Severity
Tests that MUST pass before merge:
- SQL Injection prevention
- Authentication bypass
- Horizontal privilege escalation
- Sensitive data exposure in logs

### HIGH Severity
Should fix before merge:
- XSS vulnerabilities
- CSRF protection
- Vertical privilege escalation
- Missing security headers

### MEDIUM Severity
Fix soon:
- Rate limiting gaps
- Path traversal edge cases
- Incomplete CSP directives

## Security Checklist

Before deploying to production:

- [ ] All security tests pass
- [ ] GoSec shows no HIGH/CRITICAL issues
- [ ] Nancy shows no known vulnerabilities
- [ ] OWASP ZAP scan clean
- [ ] Security headers configured
- [ ] Rate limiting enabled
- [ ] HTTPS enforced (Strict-Transport-Security)
- [ ] Secrets not in logs (using `slog`)
- [ ] Database credentials encrypted
- [ ] JWT secrets rotated
- [ ] CORS configured (no wildcard with credentials)
- [ ] Cookie flags set (Secure, HttpOnly, SameSite)

## Adding New Security Tests

When adding new features, add corresponding security tests:

1. **Create test file**: `tests/security/feature_test.go`
2. **Follow naming convention**: `TestFeatureSecurity`
3. **Use table-driven tests**: For multiple test cases
4. **Document attack scenarios**: Explain what you're testing
5. **Add to CI**: Tests auto-run in security-tests.yml

### Example Test Structure
```go
func TestNewFeatureSecurity(t *testing.T) {
    t.Run("Attack scenario description", func(t *testing.T) {
        // Arrange: Setup attack
        maliciousInput := "..."

        // Act: Attempt attack
        result := performAction(maliciousInput)

        // Assert: Verify protection
        assert.Error(t, result, "Attack should be prevented")
    })
}
```

## References

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [OWASP Testing Guide](https://owasp.org/www-project-web-security-testing-guide/)
- [Go Security Checklist](https://github.com/golang/go/wiki/Security)
- [BusinessOS Security Audit Report](../SECURITY_AUDIT_CUS-119.md)

## Maintenance

### Update Test Cases
When new attack vectors discovered:
1. Add test case to appropriate test file
2. Verify existing protection or add mitigation
3. Document in this README

### Dependency Updates
```bash
# Update security tools
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install github.com/sonatype-nexus-community/nancy@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
```

### Review Schedule
- **Weekly**: Automated security scans (GitHub Actions)
- **Monthly**: Manual penetration testing review
- **Quarterly**: Security audit and test coverage review
- **Annually**: Third-party security assessment

---

**Last Updated**: 2026-01-23
**Test Coverage**: 8 test suites, 100+ test cases
**Compliance**: OWASP Top 10 2021
