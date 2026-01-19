# BusinessOS Integration E2E Test Suite

Comprehensive end-to-end testing for the BusinessOS integration system covering all 9 providers (Google, Slack, Linear, HubSpot, Notion, ClickUp, Airtable, Fathom, Microsoft).

## 📋 Overview

This test suite validates the complete integration lifecycle:
- OAuth 2.0 flows
- Webhook signature verification
- Webhook event processing
- MCP tool execution
- Error handling and edge cases
- Database synchronization

## 🚀 Quick Start

### Prerequisites

1. **Server Running**: Ensure the BusinessOS backend is running at `http://localhost:8001` (or set `API_BASE_URL`)
2. **Dependencies**:
   - `jq` - JSON processing (`brew install jq` on macOS)
   - `openssl` - Cryptographic operations (usually pre-installed)
   - Go 1.24+ - For unit tests

### Run All Tests

```bash
# Run all test suites
./run_all_e2e_tests.sh

# Run with authentication (enables MCP tools tests)
export AUTH_TOKEN="your-jwt-token-here"
./run_all_e2e_tests.sh

# Test against different server
export API_BASE_URL="https://your-server.com"
export AUTH_TOKEN="your-token"
./run_all_e2e_tests.sh
```

## 📦 Test Suites

### 1. Unit Tests (Go)

**File**: `internal/webhooks/signature_test.go`
**Run**: `go test -v ./internal/webhooks`

Tests webhook signature verification for all 9 providers:
- Valid signature acceptance
- Invalid signature rejection
- Replay attack prevention (timestamp expiration)
- Development mode support (no secret configured)

**Coverage**: 11 test suites, 25+ individual tests

### 2. OAuth Flow Testing

**File**: `e2e_oauth_test.sh`
**Run**: `./e2e_oauth_test.sh`

Tests OAuth integration endpoints:
- List all providers (unauthenticated)
- OAuth initiation for each provider
- List user integrations (authenticated)
- Check specific integration status (authenticated)
- Database verification queries

**Providers Tested**: Google, Slack, Linear, HubSpot, Notion, ClickUp, Airtable, Fathom, Microsoft

**Example**:
```bash
# Basic test (unauthenticated endpoints)
./e2e_oauth_test.sh

# With authentication
export AUTH_TOKEN="your-jwt-token"
./e2e_oauth_test.sh

# Against production
export API_BASE_URL="https://api.businessos.com"
export AUTH_TOKEN="your-token"
./e2e_oauth_test.sh
```

### 3. MCP Tools Testing

**File**: `e2e_mcp_tools_test.sh`
**Run**: `./e2e_mcp_tools_test.sh`
**Requires**: `AUTH_TOKEN`

Tests MCP tool execution for integrated providers:

**Google Calendar**:
- `calendar_list_events` - List calendar events
- `calendar_create_event` - Create new event

**Slack**:
- `slack_list_channels` - List workspace channels
- `slack_send_message` - Send message (dry run)

**Notion**:
- `notion_list_databases` - List accessible databases
- `notion_search` - Search pages and databases

**Linear**:
- `linear_list_issues` - List project issues

**Example**:
```bash
export AUTH_TOKEN="your-jwt-token"
./e2e_mcp_tools_test.sh
```

### 4. Webhook Event Simulation

**File**: `e2e_webhook_test.sh`
**Run**: `./e2e_webhook_test.sh`

Simulates real webhook events with proper HMAC-SHA256 signatures:
- Slack message events
- Linear issue creation
- Google Calendar notifications
- HubSpot contact updates
- Notion page updates
- Airtable record changes
- Fathom meeting completions
- Microsoft calendar notifications
- ClickUp task events

**Example**:
```bash
# With default test secrets
./e2e_webhook_test.sh

# With production secrets
export SLACK_WEBHOOK_SECRET="your-real-slack-secret"
export LINEAR_WEBHOOK_SECRET="your-real-linear-secret"
./e2e_webhook_test.sh
```

### 5. Error Handling & Edge Cases

**File**: `e2e_error_handling_test.sh`
**Run**: `./e2e_error_handling_test.sh`

Tests error scenarios and security:
- Invalid webhook signatures (should reject with 401)
- Expired timestamps / replay attacks (should reject)
- Malformed JSON payloads (should reject with 400)
- Missing required headers (should reject)
- Unauthenticated access to protected endpoints (should reject with 401)
- Invalid authentication tokens (should reject with 401)
- Non-existent integration endpoints (should return 404)
- OAuth error callbacks (access denied, etc.)
- Large payload handling (5MB+ payloads)
- Concurrent webhook processing (race condition testing)
- Rate limiting (if implemented)

**Example**:
```bash
# Basic security tests
./e2e_error_handling_test.sh

# With authentication for protected endpoint tests
export AUTH_TOKEN="your-jwt-token"
./e2e_error_handling_test.sh
```

## 🎯 Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `API_BASE_URL` | Backend server URL | `http://localhost:8001` | No |
| `AUTH_TOKEN` | JWT authentication token | (none) | For authenticated tests |
| `SLACK_WEBHOOK_SECRET` | Slack webhook secret | `test-slack-secret` | No |
| `LINEAR_WEBHOOK_SECRET` | Linear webhook secret | `test-linear-secret` | No |
| `HUBSPOT_WEBHOOK_SECRET` | HubSpot webhook secret | `test-hubspot-secret` | No |
| `NOTION_WEBHOOK_SECRET` | Notion webhook secret | `test-notion-secret` | No |
| `AIRTABLE_WEBHOOK_SECRET` | Airtable webhook secret | `test-airtable-secret` | No |
| `FATHOM_WEBHOOK_SECRET` | Fathom webhook secret | `test-fathom-secret` | No |
| `MICROSOFT_CLIENT_STATE` | Microsoft client state | `expected-client-state` | No |
| `CLICKUP_WEBHOOK_SECRET` | ClickUp webhook secret | `test-clickup-secret` | No |

## 📊 Test Results

### Expected Output

```
╔══════════════════════════════════════════════════════════════════════╗
║                                                                      ║
║   ██████╗ ██╗   ██╗███████╗██╗███╗   ██╗███████╗███████╗███████╗   ║
║   ██╔══██╗██║   ██║██╔════╝██║████╗  ██║██╔════╝██╔════╝██╔════╝   ║
║   ██████╔╝██║   ██║███████╗██║██╔██╗ ██║█████╗  ███████╗███████╗   ║
║   ██╔══██╗██║   ██║╚════██║██║██║╚██╗██║██╔══╝  ╚════██║╚════██║   ║
║   ██████╔╝╚██████╔╝███████║██║██║ ╚████║███████╗███████║███████║   ║
║   ╚═════╝  ╚═════╝ ╚══════╝╚═╝╚═╝  ╚═══╝╚══════╝╚══════╝╚══════╝   ║
║                                                                      ║
║              E2E INTEGRATION TESTING SUITE                           ║
║                                                                      ║
╚══════════════════════════════════════════════════════════════════════╝

CHECKING PREREQUISITES
✓ Server is running
✓ jq is installed
✓ openssl is available
✓ AUTH_TOKEN provided

PHASE 1: UNIT TESTS
✓ TestVerifySlackSignature - PASS
✓ TestVerifyLinearSignature - PASS
...
✓ Unit tests passed

PHASE 2: E2E INTEGRATION TESTS

┌─────────────────────────────────────────────────────────────────┐
│  SUITE 2: OAuth Flow Testing
└─────────────────────────────────────────────────────────────────┘

✓ Found all 9 providers
✓ OAuth initiation endpoints respond correctly
✓ User integrations endpoint works
✓ Suite completed successfully

...

FINAL TEST SUMMARY

Test Suite Results:

  ✓ PASSED  unit_tests
  ✓ PASSED  oauth_flows
  ✓ PASSED  mcp_tools
  ✓ PASSED  webhooks
  ✓ PASSED  error_handling

Overall Statistics:
  Total Suites:   5
  Passed:         5
  Failed:         0
  Skipped:        0
  Pass Rate:      100%

╔═══════════════════════════════════════════════════╗
║                                                   ║
║   ✓ ALL TESTS PASSED!                            ║
║                                                   ║
╚═══════════════════════════════════════════════════╝
```

## 🔍 Database Verification

After running tests, verify data was synced correctly:

```sql
-- Check OAuth tokens
SELECT provider, user_id, expires_at, created_at
FROM oauth_tokens
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- Check webhook subscriptions
SELECT provider, resource_type, status, event_count, last_event_at
FROM webhook_subscriptions
WHERE status = 'active'
ORDER BY last_event_at DESC;

-- Check synced data counts
SELECT 'Calendar Events' as entity, COUNT(*) as count FROM synced_calendar_events
UNION ALL
SELECT 'Tasks' as entity, COUNT(*) as count FROM synced_tasks
UNION ALL
SELECT 'Messages' as entity, COUNT(*) as count FROM synced_messages
UNION ALL
SELECT 'Contacts' as entity, COUNT(*) as count FROM synced_contacts
UNION ALL
SELECT 'Meetings' as entity, COUNT(*) as count FROM synced_meetings;

-- Check sync tokens
SELECT provider, resource_type, last_sync_at
FROM sync_tokens
ORDER BY last_sync_at DESC;
```

## 🐛 Troubleshooting

### Server Not Running
```
✗ FAILED: Server is not running at http://localhost:8001
```
**Solution**: Start the backend server:
```bash
cd desktop/backend-go
go run ./cmd/server
```

### jq Not Installed
```
✗ FAILED: jq is not installed (required for JSON parsing)
```
**Solution**:
```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# CentOS/RHEL
sudo yum install jq
```

### Webhook Signature Verification Fails
```
✗ FAILED: Slack webhook rejected (HTTP 401)
```
**Solution**: Ensure webhook secrets match between test scripts and `.env` file:
```bash
# Check your .env file
cat .env | grep WEBHOOK_SECRET

# Or use test secrets (development only)
export SLACK_WEBHOOK_SECRET="test-slack-secret"
```

### OAuth Tests Fail
```
✗ FAILED: OAuth initiation failed (HTTP 500)
```
**Solution**:
1. Check that OAuth credentials are configured in `.env`
2. Verify callback URLs are registered with providers
3. Check server logs for detailed error messages

### MCP Tools Tests Skipped
```
⚠ WARNING: Skipping - AUTH_TOKEN not provided
```
**Solution**: Provide a valid JWT token:
```bash
# Get token from login
export AUTH_TOKEN="your-jwt-token-here"
./e2e_mcp_tools_test.sh
```

### Integration Not Connected
```
⚠ WARNING: Google Calendar not connected
```
**Solution**: This is expected if you haven't connected the integration yet. Complete OAuth flow first:
```bash
# Open in browser
open http://localhost:8001/api/integrations/google/connect
```

## 📝 Manual Testing Checklist

Beyond automated tests, manually verify:

### Frontend Integration UI (CUS-111)
- [ ] Navigate to `/integrations` page loads without errors
- [ ] 4 tabs visible (Connected, Available, AI Models, Decisions)
- [ ] Provider browser shows all 9 providers with correct icons
- [ ] Category filtering works (Communication, CRM, Calendar, etc.)
- [ ] OAuth flow: Click "Connect" → Authorize → Redirect back works
- [ ] Integration appears in "Connected" tab after OAuth
- [ ] Integration settings page shows sync stats
- [ ] Disconnect button works
- [ ] Manual sync trigger works

### Real OAuth Flows
For each provider, manually test:
1. Start OAuth flow from frontend or curl
2. Complete authorization on provider's site
3. Verify redirect back to BusinessOS works
4. Check database: `SELECT * FROM oauth_tokens WHERE provider='google';`
5. Trigger a sync and verify data appears

### Real Webhook Events
For each provider with webhook support:
1. Configure webhook URL in provider dashboard
2. Trigger an event (create task, send message, etc.)
3. Check webhook subscription: `SELECT * FROM webhook_subscriptions WHERE provider='slack';`
4. Verify event was processed: `SELECT * FROM synced_messages WHERE provider='slack' ORDER BY created_at DESC LIMIT 1;`

## 📚 References

- **Main Test Guide**: `TESTING.md` (comprehensive testing documentation)
- **Webhook Handler Implementation**: `internal/webhooks/handler.go`
- **Webhook Signature Tests**: `internal/webhooks/signature_test.go`
- **Sync Service**: `internal/services/sync_service.go`
- **Polling Jobs**: `internal/jobs/sync_jobs.go`
- **Database Schema**: `internal/database/migrations/053_sync_tables.sql`

## 🎓 Best Practices

1. **Run tests locally** before deploying to production
2. **Use test secrets** in development (never commit real secrets)
3. **Verify database state** after running webhook/sync tests
4. **Check server logs** for detailed error messages
5. **Test OAuth flows** manually at least once per provider
6. **Monitor webhook subscriptions** for event counts and last event times
7. **Run full suite** before marking integration work as complete

## 🚨 CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y jq

      - name: Run migrations
        run: go run ./cmd/migrate

      - name: Start server
        run: go run ./cmd/server &

      - name: Wait for server
        run: |
          for i in {1..30}; do
            if curl -s http://localhost:8001/health; then
              break
            fi
            sleep 1
          done

      - name: Run E2E tests
        run: cd scripts/tests && ./run_all_e2e_tests.sh
```

## 📄 License

Part of BusinessOS - Internal testing documentation.

---

**Last Updated**: 2026-01-19
**Version**: 1.0.0
**Maintained By**: BusinessOS Engineering Team
