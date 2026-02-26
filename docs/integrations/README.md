# Integrations

BusinessOS connects to 9 external services via OAuth 2.0 and API key authentication. This guide explains how the OAuth flow works, how to configure each integration, and how to add a new one.

---

## How the OAuth Flow Works

For OAuth-based integrations, the flow is:

1. The user clicks **Connect** on the Integrations page.
2. The frontend calls `GET /api/v1/integrations/{provider}/connect`.
3. The backend builds the provider's authorization URL with the required scopes and a CSRF state token, then redirects the browser.
4. The user authorizes the app on the provider's consent screen.
5. The provider redirects back to `GET /api/v1/auth/{provider}/callback` with an authorization code.
6. The backend exchanges the code for access and refresh tokens.
7. Tokens are encrypted with `TOKEN_ENCRYPTION_KEY` and stored in the database.
8. Subsequent API calls use the stored tokens and refresh them automatically when they expire.

All OAuth redirect URIs follow the pattern: `{BASE_URL}/api/auth/{provider}/callback`

---

## Integration Setup

### Google

**What it syncs:** Gmail, Google Calendar, Google Contacts, Google Drive, Google Tasks, Google Meet.

**Create credentials:**
1. Go to [console.cloud.google.com/apis/credentials](https://console.cloud.google.com/apis/credentials).
2. Create an **OAuth 2.0 Client ID** with type **Web application**.
3. Add your redirect URI to the **Authorized redirect URIs** list.
4. Copy the Client ID and Client Secret.

**Scopes required:**
- `https://www.googleapis.com/auth/gmail.readonly`
- `https://www.googleapis.com/auth/calendar`
- `https://www.googleapis.com/auth/contacts.readonly`
- `https://www.googleapis.com/auth/drive.readonly`
- `https://www.googleapis.com/auth/tasks`

**Environment variables:**
```env
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URI=http://localhost:8001/api/auth/google/callback
```

**Testing:** After connecting, go to **Settings > Integrations**, click the Google card, and run **Sync Now**. Check the activity log for sync results.

---

### Microsoft 365

**What it syncs:** Outlook email, Exchange Calendar, Microsoft Teams, OneDrive.

**Create credentials:**
1. Go to [portal.azure.com](https://portal.azure.com) > **Azure Active Directory > App registrations**.
2. Create a new registration.
3. Under **Certificates & secrets**, create a new client secret.
4. Under **Authentication**, add your redirect URI.
5. Copy the Application (client) ID and the client secret value.

**Scopes required:**
- `Mail.Read`
- `Calendars.ReadWrite`
- `Contacts.Read`
- `Files.Read`
- `offline_access`

**Environment variables:**
```env
MICROSOFT_CLIENT_ID=your_azure_application_id
MICROSOFT_CLIENT_SECRET=your_azure_client_secret
MICROSOFT_REDIRECT_URI=http://localhost:8001/api/auth/microsoft/callback
```

**Testing:** After connecting, trigger a manual sync via the Integrations page and confirm events appear in the Calendar module.

---

### Slack

**What it syncs:** Channel messages, direct messages, workspace members, file attachments.

**Create credentials:**
1. Go to [api.slack.com/apps](https://api.slack.com/apps) and create a new app.
2. Under **OAuth & Permissions**, add your redirect URI.
3. Under **Basic Information > App Credentials**, copy the Client ID, Client Secret, and Signing Secret.

**Scopes required (Bot Token):**
- `channels:read`
- `channels:history`
- `chat:write`
- `im:read`
- `im:history`
- `users:read`
- `files:read`

**Environment variables:**
```env
SLACK_CLIENT_ID=your_slack_client_id
SLACK_CLIENT_SECRET=your_slack_client_secret
SLACK_SIGNING_SECRET=your_slack_signing_secret
SLACK_REDIRECT_URI=http://localhost:8001/api/auth/slack/callback
```

**Testing:** After connecting, open the Communication module and confirm your Slack channels appear in the sidebar.

---

### Notion

**What it syncs:** Pages, databases, blocks. Enables searching and reading Notion content from BusinessOS chat.

**Create credentials:**
1. Go to [notion.so/my-integrations](https://www.notion.so/my-integrations) and create a new integration.
2. Under **OAuth Domain & URIs**, add your redirect URI.
3. Copy the OAuth client ID and client secret.

**Environment variables:**
```env
NOTION_CLIENT_ID=your_notion_oauth_client_id
NOTION_CLIENT_SECRET=your_notion_oauth_client_secret
NOTION_REDIRECT_URI=http://localhost:8001/api/auth/notion/callback
```

**Testing:** After connecting, ask OSA "Find my Notion pages about [topic]" to verify the integration works.

---

### Linear

**What it syncs:** Issues, projects, teams, cycles. Allows OSA to create and update Linear issues from chat.

**Create credentials:**
1. Go to [linear.app/settings/api](https://linear.app/settings/api) > **OAuth Applications**.
2. Create a new application.
3. Set the redirect URI and copy the Client ID and Client Secret.
4. Optionally configure a webhook for real-time sync.

**Environment variables:**
```env
LINEAR_CLIENT_ID=your_linear_client_id
LINEAR_CLIENT_SECRET=your_linear_client_secret
LINEAR_REDIRECT_URI=http://localhost:8001/api/auth/linear/callback
LINEAR_WEBHOOK_SECRET=your_linear_webhook_secret  # optional, for real-time sync
```

**Testing:** After connecting, ask OSA "What are my open Linear issues?" to verify read access.

---

### HubSpot

**What it syncs:** Contacts, companies, deals pipeline, activities. Bidirectional CRM sync with the CRM module.

**Create credentials:**
1. Go to [developers.hubspot.com](https://developers.hubspot.com) and create an app.
2. Under **Auth**, configure OAuth settings and add your redirect URI.
3. Copy the Client ID and Client Secret.

**Scopes required:**
- `crm.objects.contacts.read`
- `crm.objects.contacts.write`
- `crm.objects.companies.read`
- `crm.objects.deals.read`

**Environment variables:**
```env
HUBSPOT_CLIENT_ID=your_hubspot_client_id
HUBSPOT_CLIENT_SECRET=your_hubspot_client_secret
HUBSPOT_REDIRECT_URI=http://localhost:8001/api/auth/hubspot/callback
```

**Testing:** After connecting, open the CRM module and trigger a sync. Confirm contacts appear from HubSpot.

---

### ClickUp

**What it syncs:** Tasks, lists, spaces, and team workspaces. Enables task creation and status updates from BusinessOS.

**Create credentials:**
1. Go to [app.clickup.com/settings/apps](https://app.clickup.com/settings/apps) and create an app.
2. Add your redirect URI and copy the Client ID and Client Secret.

**Environment variables:**
```env
CLICKUP_CLIENT_ID=your_clickup_client_id
CLICKUP_CLIENT_SECRET=your_clickup_client_secret
CLICKUP_REDIRECT_URI=http://localhost:8001/api/auth/clickup/callback
```

**Testing:** After connecting, ask OSA "Show me my ClickUp tasks due this week."

---

### Airtable

**What it syncs:** Bases, tables, records. Allows reading and writing Airtable data from BusinessOS.

**Create credentials:**
1. Go to [airtable.com/create/oauth](https://airtable.com/create/oauth) and register an integration.
2. Add your redirect URI and copy the Client ID and Client Secret.

**Scopes required:**
- `data.records:read`
- `data.records:write`
- `schema.bases:read`

**Environment variables:**
```env
AIRTABLE_CLIENT_ID=your_airtable_client_id
AIRTABLE_CLIENT_SECRET=your_airtable_client_secret
AIRTABLE_REDIRECT_URI=http://localhost:8001/api/auth/airtable/callback
```

**Testing:** After connecting, ask OSA "List my Airtable bases."

---

### Fathom Analytics

**What it syncs:** Site analytics including page views, unique visitors, average time on page. Data is displayed in the Analytics module.

**This integration uses API key authentication (not OAuth).**

**Get your API key:**
1. Go to [app.usefathom.com/api](https://app.usefathom.com/api).
2. Generate or copy your API key.

**Environment variables:**
```env
FATHOM_API_KEY=your_fathom_api_key
```

**Testing:** After adding the key, open the Analytics module and confirm Fathom data appears in the dashboard widgets.

---

## Adding a New Integration

To add a new integration to BusinessOS:

### 1. Add Environment Variables

Add the new provider's credentials to `.env.example` and document them in [docs/integrations/ENVIRONMENT.md](ENVIRONMENT.md).

### 2. Implement the OAuth Handler

Create a handler file at `backend/internal/handlers/{provider}_oauth.go`. The file should implement:

- `ConnectHandler(c *gin.Context)` — builds the authorization URL and redirects
- `CallbackHandler(c *gin.Context)` — exchanges the code for tokens, encrypts them, and stores them

Use `backend/internal/handlers/auth_google.go` as a reference implementation.

### 3. Store the Tokens

Tokens are encrypted using the `TOKEN_ENCRYPTION_KEY` before being written to the `integration_tokens` table. Use the existing `EncryptToken` / `DecryptToken` helpers in the services layer.

### 4. Register the Routes

Add the new routes to the router registration function:

```go
// In backend/internal/handlers/ router setup
auth.GET("/{provider}/connect", handler.ConnectHandler)
auth.GET("/{provider}/callback", handler.CallbackHandler)
```

### 5. Implement the Sync Service

Create `backend/internal/services/{provider}_sync.go` to implement data fetching and mapping from the provider's API to BusinessOS entities.

### 6. Register with the Integration Router

Add the new provider to the integration router in `backend/internal/handlers/integration_router.go` so it appears in the Integrations UI with connection status.

### 7. Add Frontend UI

Add the integration card to `frontend/src/routes/(app)/integrations/+page.svelte` with the provider name, logo, description, and connect button.

---

## See Also

- [Environment Variable Reference](ENVIRONMENT.md) — complete list of all variables
- [Architecture Overview](../architecture/README.md) — how the backend layers work
- [Getting Started](../getting-started/README.md) — initial setup guide
