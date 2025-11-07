# chrome-mcp-server - Browser Automation Reference

## Overview

chrome-mcp-server provides browser automation for UI testing, authenticated API calls, and visual verification.

## When to Use

- UI testing and verification
- Authenticated API calls (browser has valid session)
- Visual regression checking
- User flow testing
- Console error detection

## Tool Categories

### Navigation Tools
- `chrome_navigate`: Load URL or refresh page
- `chrome_go_back_or_forward`: Browser history navigation
- `get_windows_and_tabs`: List all open windows/tabs
- `chrome_close_tabs`: Close specific tabs

### Interaction Tools
- `chrome_click_element`: Click elements by selector or coordinates
- `chrome_fill_or_select`: Fill form inputs or select options
- `chrome_keyboard`: Simulate keyboard input
- `chrome_get_interactive_elements`: Find clickable elements

### Inspection Tools
- `chrome_screenshot`: Capture page or element screenshots
- `chrome_console`: Get console messages and errors
- `chrome_get_web_content`: Extract page text or HTML

### Scripting Tools
- `chrome_inject_script`: Run custom JavaScript in page
- `chrome_send_command_to_inject_script`: Trigger events in injected scripts

### Network Tools
- `chrome_network_request`: Make HTTP requests with browser context
- `chrome_network_debugger_start`: Start capturing network traffic
- `chrome_network_debugger_stop`: Stop and retrieve network data
- `chrome_network_capture_start`: Lightweight network monitoring
- `chrome_network_capture_stop`: Stop lightweight monitoring

### History & Bookmarks
- `chrome_history`: Search browsing history
- `chrome_bookmark_search`: Find bookmarks
- `chrome_bookmark_add`: Create bookmarks
- `chrome_bookmark_delete`: Remove bookmarks
- `search_tabs_content`: Search across open tabs

## Usage Patterns

### Basic UI Testing

```typescript
// 1. Navigate to page
chrome_navigate({
  url: "http://localhost:3000/dashboard",
  width: 1280,
  height: 720
})

// 2. Take screenshot
chrome_screenshot({
  fullPage: true,
  savePng: false,
  storeBase64: true  // View immediately
})

// 3. Check console for errors (start small to avoid token limits)
chrome_console({
  includeExceptions: true,
  maxMessages: 20  // Start with 10-20, increase only if needed
})
```

### Form Interaction Testing

```typescript
// 1. Navigate
chrome_navigate({ url: "http://localhost:3000/signup" })

// 2. Fill form
chrome_fill_or_select({
  selector: "input[name='email']",
  value: "test@example.com"
})

chrome_fill_or_select({
  selector: "input[name='password']",
  value: "password123"
})

// 3. Click submit
chrome_click_element({
  selector: "button[type='submit']",
  waitForNavigation: true,
  timeout: 5000
})

// 4. Verify success
chrome_get_web_content({
  textContent: true
})
```

### Authenticated API Testing

```typescript
// 1. Navigate to establish session
chrome_navigate({ url: "http://localhost:3000/dashboard" })

// 2. Inject fetch call
chrome_inject_script({
  type: "MAIN",
  jsScript: `
    fetch('/api/projects', {
      method: 'GET',
      credentials: 'include'  // Include auth cookies
    })
    .then(r => r.json())
    .then(data => console.log('Projects:', data))
    .catch(err => console.error('Error:', err));
  `
})

// 3. Check results
chrome_console()
```

### Network Inspection

```typescript
// 1. Start network capture
chrome_network_debugger_start()

// 2. Navigate to page that makes API calls
chrome_navigate({ url: "http://localhost:3000/dashboard" })

// 3. Stop and get all network activity
const networkData = chrome_network_debugger_stop()

// Returns: All requests/responses including headers, status, body
```

### Finding Elements

```typescript
// Get all interactive elements
chrome_get_interactive_elements({
  includeCoordinates: true
})

// Search for specific text
chrome_get_interactive_elements({
  textQuery: "Save Changes"
})

// Filter by selector
chrome_get_interactive_elements({
  selector: "button[data-testid]"
})
```

## Common Use Cases

### UI Regression Testing

```typescript
// Before change
chrome_navigate({ url: "http://localhost:3000/dashboard" })
const before = chrome_screenshot({ savePng: true, name: "before" })

// Make code changes...

// After change
chrome_navigate({ url: "http://localhost:3000/dashboard", refresh: true })
const after = chrome_screenshot({ savePng: true, name: "after" })

// Compare screenshots manually
```

### User Flow Verification

```typescript
// 1. Login
chrome_navigate({ url: "http://localhost:3000" })
chrome_fill_or_select({ selector: "#email", value: "user@example.com" })
chrome_fill_or_select({ selector: "#password", value: "pass123" })
chrome_click_element({ selector: "button[type='submit']", waitForNavigation: true })

// 2. Navigate to feature
chrome_click_element({ selector: "a[href='/projects']" })

// 3. Create project
chrome_click_element({ selector: "button[data-testid='new-project']" })
chrome_fill_or_select({ selector: "input[name='name']", value: "Test Project" })
chrome_click_element({ selector: "button[type='submit']" })

// 4. Verify success
chrome_get_web_content({ textContent: true })
```

### Console Error Detection

```typescript
// Navigate and run app
chrome_navigate({ url: "http://localhost:3000/dashboard" })

// Interact with features
chrome_click_element({ selector: "button[data-testid='load-data']" })

// Check for errors - START SMALL to avoid token limits
const console = chrome_console({
  includeExceptions: true,
  maxMessages: 20  // Start with fewer messages
})

// Parse results for errors
// Report any console.error or exceptions found
```

## Best Practices

1. **Navigate first** - Establish page context
2. **Use credentials: 'include'** - For authenticated API calls
3. **Check console** - Always verify no errors
4. **Start with small maxMessages** - Use 10-20 for console, increase only if needed
5. **Adapt to token limits** - Reduce parameters when hitting MCP token errors, never give up
6. **Full page screenshots** - Better for UI verification
7. **Network debugger for APIs** - Full request/response inspection
8. **Inject scripts for testing** - Run custom verification logic

## Troubleshooting

### Problem: MCP tool response exceeds token limit
**Error**: "MCP tool 'X' response (XXXXX tokens) exceeds maximum allowed tokens (25000)"

**Root Cause**: The error message says "use pagination, filtering, or limit parameters" but not all tools have these features. Check what the tool ACTUALLY supports.

**chrome_console Parameters** (limited options):
- `maxMessages` - ONLY size control available
- `includeExceptions` - Boolean, not a filter
- `url` - Navigation, not filtering

**Solution for chrome_console**:
```typescript
// ONLY option: Reduce maxMessages
chrome_console({ maxMessages: 10 })  // Start very small

// If you need more, increase gradually
chrome_console({ maxMessages: 20 })

// For exceptions specifically
chrome_console({ maxMessages: 20, includeExceptions: true })
```

**Solution for tools WITH filtering** (like chrome_history):
```typescript
// chrome_history has real filters:
chrome_history({
  maxResults: 20,           // Limit
  text: "specific search",  // Content filter
  startTime: "1 day ago",   // Time filter
  endTime: "now"
})
```

**General Strategy**:
1. **Check tool parameters** - What does it actually support?
2. **Use filters if available** - Narrow results by content, time, type
3. **Reduce limits** - maxMessages, maxResults, etc.
4. **Never give up** - Adapt parameters and retry

### Problem: Element not found
**Solution**: Use chrome_get_interactive_elements first

### Problem: Authentication fails
**Solution**: Ensure navigated to app first to establish session

### Problem: Network capture empty
**Solution**: Check timing - may need to trigger actions after start

### Problem: Script injection fails
**Solution**: Verify script syntax, check console for errors

## Performance

- **Speed**: Slow - Browser operations take time
- **Network dependent**: Page load speed varies
- **Recommendation**: Use for verification, not exploration
