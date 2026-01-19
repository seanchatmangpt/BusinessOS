# Frontend Integration UI Verification Guide

Complete manual testing guide for the BusinessOS frontend integration system.

## 📋 Overview

The frontend integration UI is **fully implemented** in SvelteKit. This guide walks through comprehensive verification testing to ensure all components work correctly with the backend integration system.

## 🎯 Test Environment

**Frontend URL**: `http://localhost:5173` (or your deployed frontend URL)
**Backend URL**: `http://localhost:8001` (must be running)

### Prerequisites

1. ✅ Backend server running with integration system
2. ✅ Database with migration 053 applied
3. ✅ OAuth credentials configured in backend `.env`
4. ✅ User account created (for authentication)
5. ✅ Browser with DevTools (Chrome/Firefox recommended)

## 🧪 Test Suites

---

## SUITE 1: Integration Hub Page

### Test 1.1: Navigation & Page Load

**Steps**:
1. Log into BusinessOS frontend
2. Navigate to `/integrations`
3. Open browser DevTools (F12) → Console tab

**Expected**:
- ✅ Page loads without errors
- ✅ No console errors
- ✅ URL is `/integrations` or `/app/integrations`
- ✅ Page title displays "Integrations" or similar
- ✅ Loading state appears briefly before content

**Screenshot**: Capture full page

---

### Test 1.2: Tab Navigation

**Steps**:
1. On `/integrations` page
2. Observe the 4 tabs at the top

**Expected**:
- ✅ 4 tabs visible:
  - "Connected" (or "My Integrations")
  - "Available" (or "Explore")
  - "AI Models"
  - "Decisions"
- ✅ Active tab is highlighted
- ✅ Click each tab switches view
- ✅ URL updates with tab selection (e.g., `?tab=available`)
- ✅ No errors in console when switching tabs

**Screenshot**: Capture each tab

---

### Test 1.3: Provider Browser (Available Tab)

**Steps**:
1. Navigate to "Available" tab
2. Scroll through provider list

**Expected**:
- ✅ Provider cards displayed in grid layout
- ✅ Each card shows:
  - Provider logo/icon
  - Provider name
  - Brief description
  - "Connect" button
  - Category badge (Communication, CRM, Calendar, etc.)
- ✅ Minimum 9 providers visible (Google, Slack, Linear, HubSpot, Notion, ClickUp, Airtable, Fathom, Microsoft)
- ✅ Cards are responsive (resize browser to test)

**Providers to verify**:
```
✅ Google (Calendar, Gmail)
✅ Slack (Communication)
✅ Linear (Project Management)
✅ HubSpot (CRM)
✅ Notion (Documentation)
✅ ClickUp (Project Management)
✅ Airtable (Database)
✅ Fathom (Meetings)
✅ Microsoft (Calendar, Outlook)
```

**Screenshot**: Full provider grid

---

### Test 1.4: Category Filtering

**Steps**:
1. On "Available" tab
2. Look for category filter (dropdown, tabs, or sidebar)
3. Select different categories

**Expected**:
- ✅ Category filter UI is visible
- ✅ Categories available:
  - All
  - Communication (Slack, Microsoft)
  - CRM (HubSpot)
  - Calendar (Google, Microsoft)
  - Project Management (Linear, ClickUp, Airtable)
  - Meetings (Fathom)
  - Documentation (Notion)
- ✅ Selecting category filters provider list
- ✅ Provider count updates
- ✅ "All" shows all providers

**Screenshot**: Each category selected

---

### Test 1.5: Search Functionality (if implemented)

**Steps**:
1. On "Available" tab
2. Look for search input
3. Type provider names

**Expected**:
- ✅ Search input is visible
- ✅ Typing filters provider list in real-time
- ✅ Case-insensitive search
- ✅ Searches provider name and description
- ✅ Empty search shows all providers
- ✅ No results shows "No providers found" message

**Test queries**:
- "slack" → Slack card appears
- "calendar" → Google, Microsoft appear
- "xyz123" → No results

**Screenshot**: Search with results and no results

---

## SUITE 2: OAuth Flow Testing

### Test 2.1: OAuth Initiation (Slack)

**Steps**:
1. On "Available" tab
2. Find Slack provider card
3. Click "Connect" button
4. Monitor browser DevTools → Network tab

**Expected**:
- ✅ Clicking "Connect" triggers loading state on button
- ✅ Browser opens OAuth authorization page (Slack's domain)
- ✅ OR: Opens in new tab/window
- ✅ Network tab shows request to `/api/integrations/slack/connect`
- ✅ No console errors

**Screenshot**: Slack OAuth authorization page

---

### Test 2.2: OAuth Authorization (Slack)

**Steps**:
1. On Slack OAuth page
2. Select workspace (if multiple)
3. Review permissions requested
4. Click "Allow" or "Authorize"

**Expected**:
- ✅ Slack OAuth page loads correctly
- ✅ Shows correct app name and icon
- ✅ Lists permissions requested (read messages, channels, etc.)
- ✅ "Allow" button is clickable
- ✅ No errors on Slack's page

**Screenshot**: Slack permission page

---

### Test 2.3: OAuth Callback & Redirect

**Steps**:
1. After clicking "Allow" on Slack OAuth
2. Monitor URL changes
3. Wait for redirect

**Expected**:
- ✅ Browser redirects back to BusinessOS
- ✅ URL includes `/auth/callback` or similar
- ✅ Query params include `code` and `state`
- ✅ Loading state shows "Connecting..." or similar
- ✅ Final redirect to `/integrations?tab=connected` (or similar)
- ✅ Success message/toast appears: "Slack connected successfully"
- ✅ No errors in console

**Screenshot**: Success message

---

### Test 2.4: Verify Connection in UI

**Steps**:
1. Navigate to "Connected" tab
2. Look for Slack in connected integrations list

**Expected**:
- ✅ Slack appears in "Connected" tab
- ✅ Shows connection details:
  - Provider name and icon
  - Workspace name (if available)
  - Connection status: "Active" or green indicator
  - Connected timestamp
  - "Settings" or "Manage" button
  - "Disconnect" button
- ✅ Sync status indicator (syncing, last synced, etc.)

**Screenshot**: Connected Slack card

---

### Test 2.5: Repeat for Other Providers

**Repeat Tests 2.1-2.4 for**:
- ✅ Google (Calendar)
- ✅ Linear
- ✅ Notion

**Notes**:
- Google requires selecting Google account
- Linear requires workspace selection
- Notion requires page selection

---

## SUITE 3: Integration Settings Page

### Test 3.1: Navigate to Integration Settings

**Steps**:
1. On "Connected" tab
2. Find connected Slack integration
3. Click "Settings", "Manage", or the integration card itself

**Expected**:
- ✅ Navigates to `/integrations/slack` or `/integrations/[integration-id]`
- ✅ Settings page loads without errors
- ✅ Page title shows provider name (e.g., "Slack Integration")
- ✅ Back button or breadcrumb to return to integrations list

**Screenshot**: Integration settings page

---

### Test 3.2: Sync Statistics Display

**Steps**:
1. On integration settings page (Slack)
2. Scroll to sync statistics section

**Expected**:
- ✅ Sync statistics section is visible
- ✅ Displays:
  - Last sync time (timestamp or "Never")
  - Sync status (Active, Paused, Error)
  - Items synced (message count, channel count)
  - Next sync time (if scheduled)
  - Event count (webhook events received)
- ✅ Data is formatted properly (dates, numbers)
- ✅ Shows loading state if fetching data

**Example**:
```
Last Sync: 2 minutes ago
Status: Active
Messages Synced: 1,234
Channels: 5
Events Received: 89
Next Sync: In 3 minutes
```

**Screenshot**: Sync statistics

---

### Test 3.3: Manual Sync Trigger

**Steps**:
1. On integration settings page
2. Find "Sync Now" or "Trigger Sync" button
3. Click the button
4. Monitor sync status

**Expected**:
- ✅ "Sync Now" button is visible
- ✅ Clicking shows loading state ("Syncing...")
- ✅ Network request to `/api/integrations/slack/sync` (check DevTools)
- ✅ Success/error message appears
- ✅ Sync statistics update after completion
- ✅ Last sync time updates to "Just now"

**Screenshot**: Sync in progress, then completed

---

### Test 3.4: Sync History (if implemented)

**Steps**:
1. On integration settings page
2. Look for sync history table/list

**Expected**:
- ✅ Sync history section is visible
- ✅ Shows recent sync attempts with:
  - Timestamp
  - Status (Success, Failed, In Progress)
  - Items synced count
  - Error message (if failed)
- ✅ Most recent sync at top
- ✅ Paginated if many entries

**Screenshot**: Sync history

---

### Test 3.5: Permission Management

**Steps**:
1. On integration settings page
2. Find permissions or scopes section

**Expected**:
- ✅ Permissions section shows what access was granted
- ✅ Lists specific permissions:
  - Read messages
  - Send messages
  - Read channels
  - etc.
- ✅ Shows if permission was granted or denied
- ✅ "Modify Permissions" button (re-triggers OAuth)

**Screenshot**: Permissions list

---

### Test 3.6: Skill Configuration (if applicable)

**Steps**:
1. On integration settings page
2. Look for skills or features section

**Expected**:
- ✅ Skills section is visible
- ✅ Shows available skills/features:
  - Send Message
  - List Channels
  - Create Channel
  - etc.
- ✅ Toggle switches to enable/disable skills
- ✅ Toggling skill updates immediately
- ✅ Success message on save

**Screenshot**: Skills configuration

---

### Test 3.7: Disconnect Integration

**Steps**:
1. On integration settings page
2. Find "Disconnect" button (usually at bottom)
3. Click "Disconnect"

**Expected**:
- ✅ Confirmation modal/dialog appears:
  - Warning message
  - "Are you sure?" prompt
  - "Cancel" and "Disconnect" buttons
- ✅ Clicking "Cancel" closes modal
- ✅ Clicking "Disconnect" shows loading state
- ✅ Network request to DELETE `/api/integrations/slack` (check DevTools)
- ✅ Success message: "Slack disconnected"
- ✅ Redirects back to `/integrations`
- ✅ Slack no longer in "Connected" tab
- ✅ Slack back in "Available" tab with "Connect" button

**Screenshot**: Disconnect confirmation modal

---

### Test 3.8: Reconnect After Disconnect

**Steps**:
1. After disconnecting Slack
2. On "Available" tab, find Slack
3. Click "Connect" again
4. Complete OAuth flow

**Expected**:
- ✅ OAuth flow works again
- ✅ Successfully reconnects
- ✅ Slack appears in "Connected" tab
- ✅ Previous sync data cleared or archived
- ✅ New connection timestamp

---

## SUITE 4: AI Models Tab

### Test 4.1: AI Models Display

**Steps**:
1. Navigate to "AI Models" tab
2. Observe content

**Expected**:
- ✅ Page shows AI model integrations
- ✅ Shows models like:
  - OpenAI (GPT-4, GPT-3.5)
  - Anthropic (Claude)
  - Local models (if configured)
- ✅ Each model card shows:
  - Model name
  - Provider
  - Status (Available, Connected)
  - "Configure" or "Connect" button
- ✅ No errors in console

**Screenshot**: AI Models tab

---

### Test 4.2: Model Configuration (if interactive)

**Steps**:
1. Click "Configure" on a model
2. Enter API key (if prompted)
3. Save configuration

**Expected**:
- ✅ Configuration modal opens
- ✅ API key input is secure (password field)
- ✅ Save button triggers validation
- ✅ Success message on save
- ✅ Model status updates to "Connected"

**Screenshot**: Model configuration

---

## SUITE 5: Decisions Tab

### Test 5.1: Decisions Display

**Steps**:
1. Navigate to "Decisions" tab
2. Observe content

**Expected**:
- ✅ Page shows pending decisions/approvals
- ✅ Empty state if no decisions: "No pending decisions"
- ✅ If decisions exist:
  - Decision description
  - Related integration
  - Timestamp
  - Approve/Reject buttons
- ✅ No errors in console

**Screenshot**: Decisions tab

---

## SUITE 6: Error Handling

### Test 6.1: OAuth Error - User Denies Access

**Steps**:
1. Start OAuth flow for Google
2. On Google consent screen, click "Deny" or "Cancel"

**Expected**:
- ✅ Redirects back to BusinessOS
- ✅ Error message displayed: "Authorization failed" or "Access denied"
- ✅ Integration not added to "Connected" tab
- ✅ User can try again
- ✅ No unhandled errors in console

**Screenshot**: Error message

---

### Test 6.2: OAuth Error - Invalid Callback

**Steps**:
1. Manually navigate to OAuth callback with invalid params:
   ```
   /auth/callback?error=invalid_request
   ```

**Expected**:
- ✅ Error page or message displayed
- ✅ Explains what went wrong
- ✅ Link back to `/integrations`
- ✅ No crash or white screen

**Screenshot**: Error handling

---

### Test 6.3: API Error - Backend Down

**Steps**:
1. Stop the backend server
2. Try to load `/integrations` page
3. Try to trigger sync on connected integration

**Expected**:
- ✅ Page shows loading state, then error state
- ✅ Error message: "Unable to connect to server" or similar
- ✅ Retry button available
- ✅ No unhandled promise rejections in console
- ✅ UI doesn't crash or freeze

**Screenshot**: Backend down error

---

### Test 6.4: API Error - Invalid Token

**Steps**:
1. Clear localStorage/cookies
2. Try to access `/integrations`

**Expected**:
- ✅ Redirects to login page
- ✅ After login, returns to `/integrations`
- ✅ No errors shown to user about invalid token

---

### Test 6.5: Network Error - Slow Connection

**Steps**:
1. Open DevTools → Network tab
2. Throttle to "Slow 3G"
3. Navigate to `/integrations`
4. Try to connect an integration

**Expected**:
- ✅ Loading states appear while waiting
- ✅ Eventually loads (or shows timeout error)
- ✅ Timeout errors are graceful: "Request timed out"
- ✅ Retry button available
- ✅ No infinite loading spinners

---

## SUITE 7: Responsive Design

### Test 7.1: Mobile View (375px width)

**Steps**:
1. Open DevTools → Device Toolbar (Ctrl+Shift+M)
2. Select iPhone SE or similar (375px)
3. Navigate through `/integrations`

**Expected**:
- ✅ Page is fully usable on mobile
- ✅ Provider cards stack vertically
- ✅ Tabs are accessible (may scroll horizontally)
- ✅ Buttons are tappable (min 44px height)
- ✅ Text is readable (no truncation issues)
- ✅ No horizontal scroll on page
- ✅ Images/logos scale appropriately

**Screenshot**: Mobile view of integrations page

---

### Test 7.2: Tablet View (768px width)

**Steps**:
1. Set device width to iPad (768px)
2. Navigate through `/integrations`

**Expected**:
- ✅ Provider cards in 2-column grid
- ✅ All features accessible
- ✅ Navigation works smoothly

**Screenshot**: Tablet view

---

### Test 7.3: Desktop View (1920px width)

**Steps**:
1. Set browser to full HD (1920px)
2. Navigate through `/integrations`

**Expected**:
- ✅ Provider cards in 3-4 column grid
- ✅ Content doesn't stretch too wide (max-width)
- ✅ Whitespace is balanced

**Screenshot**: Desktop view

---

## SUITE 8: Browser Compatibility

### Test 8.1: Chrome/Edge (Chromium)

**Steps**:
1. Test all above suites in Chrome or Edge

**Expected**:
- ✅ All tests pass
- ✅ No console errors
- ✅ OAuth popups work

---

### Test 8.2: Firefox

**Steps**:
1. Test core flows in Firefox (OAuth, sync, disconnect)

**Expected**:
- ✅ OAuth flows work
- ✅ UI renders correctly
- ✅ No Firefox-specific errors

---

### Test 8.3: Safari (macOS)

**Steps**:
1. Test core flows in Safari

**Expected**:
- ✅ OAuth flows work (popup blockers off)
- ✅ UI renders correctly
- ✅ WebSocket connections work (if used)

---

## SUITE 9: Performance

### Test 9.1: Page Load Time

**Steps**:
1. Open DevTools → Network tab
2. Hard refresh `/integrations` (Ctrl+Shift+R)
3. Check "DOMContentLoaded" and "Load" times

**Expected**:
- ✅ DOMContentLoaded < 1 second
- ✅ Full Load < 3 seconds (with 50+ providers)
- ✅ First Contentful Paint < 1 second

---

### Test 9.2: Memory Usage

**Steps**:
1. Open DevTools → Performance Monitor
2. Navigate through tabs multiple times
3. Monitor memory usage

**Expected**:
- ✅ Memory doesn't continuously increase (no memory leaks)
- ✅ Memory stays under 100MB for integrations page

---

## 📊 Test Results Template

Use this template to record your test results:

```markdown
# Frontend Integration UI Verification - Test Results

**Date**: YYYY-MM-DD
**Tester**: [Your Name]
**Frontend Version**: [commit hash or version]
**Backend Version**: [commit hash or version]
**Browser**: [Chrome 120, Firefox 121, etc.]

## Summary

- Total Tests: XX
- Passed: XX
- Failed: XX
- Blocked: XX (if backend issue, etc.)

## Test Results by Suite

### SUITE 1: Integration Hub Page
- [x] Test 1.1: Navigation & Page Load - PASS
- [x] Test 1.2: Tab Navigation - PASS
- [ ] Test 1.3: Provider Browser - FAIL (reason: ...)
- ...

### SUITE 2: OAuth Flow Testing
- ...

## Issues Found

### Issue 1: [Title]
- **Severity**: Critical / High / Medium / Low
- **Test**: SUITE X, Test Y.Z
- **Description**: [What went wrong]
- **Steps to Reproduce**: [How to reproduce]
- **Screenshot**: [Link or attachment]
- **Expected**: [What should happen]
- **Actual**: [What actually happened]

## Recommendations

1. [Issue to fix]
2. [Enhancement suggestion]

## Sign-off

- [ ] All critical tests passed
- [ ] No blocking issues found
- [ ] Frontend ready for production

Signed: __________________
```

## 🔧 Troubleshooting

### Issue: OAuth popup blocked

**Solution**:
1. Allow popups for your domain
2. Or use redirect flow instead of popup

### Issue: "Integration not found" error

**Solution**:
1. Check backend is running
2. Verify provider is configured in backend `.env`
3. Check database migration 053 is applied

### Issue: Sync doesn't trigger

**Solution**:
1. Check browser console for errors
2. Verify webhook subscriptions in database
3. Test with backend E2E scripts first

### Issue: Images/logos not loading

**Solution**:
1. Check CDN or image paths
2. Verify CORS settings on image hosts
3. Check browser console for 404 errors

## 📚 Related Documentation

- Backend E2E Tests: `scripts/tests/README.md`
- Backend Testing Guide: `TESTING.md`
- OAuth Implementation: `internal/integrations/*/provider.go`
- Frontend API Client: `frontend/src/lib/api/integrations/`

## ✅ Final Checklist

Before marking CUS-111 as complete:

- [ ] All 9 test suites completed
- [ ] At least 3 OAuth providers tested end-to-end
- [ ] All critical bugs documented
- [ ] Screenshots captured for each suite
- [ ] Test results documented
- [ ] Cross-browser testing completed (Chrome + 1 other)
- [ ] Mobile responsive testing completed
- [ ] Sign-off from team lead

---

**Last Updated**: 2026-01-19
**Version**: 1.0.0
**Maintained By**: BusinessOS Frontend Team
