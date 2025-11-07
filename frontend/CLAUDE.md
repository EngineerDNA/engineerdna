# Frontend - React Application

EngineerDNA's web interface for viewing metrics, configuring plugins, and analyzing engineering data.

## Architecture

**Stack**: Vite 7 + React 19 + TypeScript 5.9 + TanStack Query v5 + React Router v7 + Recharts v3 + Tailwind CSS v4

**Deployment**: Compiled to `frontend/dist/` and embedded in Go binary via `//go:embed`.

## Directory Structure

```
frontend/
├── package.json, vite.config.ts, tsconfig.json
├── src/
│   ├── main.tsx, App.tsx          # Entry and routing
│   ├── api/                        # API client + types
│   ├── hooks/                      # useEvents, usePlugins, useInsights
│   ├── components/                 # Dashboard, EventList, PluginCard, charts/
│   └── pages/                      # DashboardPage, PluginsPage, EventsPage
```

## Critical Rules

1. **Container/Presentational Pattern** - ENFORCE separation of concerns:
   - **Presentational Components** (`components/`) - Accept only props, NO useQuery/useMutation/API calls
   - **Container Components** (`pages/`) - Handle all data fetching, state management
   - Pure functions: props → UI for testability
2. **Strict TypeScript** - No `any`, strict mode
3. **Localhost API** - Proxy to `http://127.0.0.1:3847/api`
4. **Error boundaries** - Wrap async components
5. **UTC timestamps** - Display in user timezone, send UTC
6. **Component limit** - Max 300 lines
7. **Accessible UI** - ARIA, keyboard nav
8. **Responsive** - Mobile-first Tailwind (sm: 640px, md: 768px, lg: 1024px)
9. **Theme-aware** - All components support dark mode with `dark:` classes
10. **Loading states** - Always show loading/error (use skeletons, not "Loading...")
11. **Optimistic updates** - TanStack Query mutations
12. **No inline styles** - Tailwind only
13. **File naming** - PascalCase components, camelCase utilities
14. **Test coverage** - Unit + integration tests
15. **No secrets** - Never store keys/tokens
16. **Hot reload** - Vite HMR
17. **Performance** - Use React.memo for heavy list components, useMemo for expensive operations

## Quick Start

```bash
cd frontend
npm install
npm run dev          # Dev server: http://localhost:5173
npm run typecheck    # TypeScript check
npm run lint         # ESLint + Prettier
npm run build        # Production build → dist/
```

Production build:

```bash
cd frontend && npm run build
cd .. && make build  # Embeds frontend/dist/ in Go binary
./bin/engineerdna    # Access at http://127.0.0.1:3847
```

## Key Patterns

### Container/Presentational Component Pattern (REQUIRED)

**Rule**: Separate data fetching from UI rendering

**Presentational Components** (`components/`):

```typescript
// CORRECT - Pure presentational component
interface DashboardProps {
  events: Event[];
  insights: Insight[];
  isLoading?: boolean;
}

export function Dashboard({ events, insights, isLoading }: DashboardProps) {
  // Only UI logic, no data fetching
  const prCount = events.filter(e => e.type === 'pull_request').length;

  return <div>...</div>;
}
```

**Container Components** (`pages/`):

```typescript
// CORRECT - Container handles data fetching
export function DashboardPage() {
  const { data: eventsData, isLoading } = useEvents();
  const { data: insightsData } = useInsights();

  return (
    <Dashboard
      events={eventsData?.events || []}
      insights={insightsData?.insights || []}
      isLoading={isLoading}
    />
  );
}
```

**Anti-Patterns** (DO NOT DO):

```typescript
// WRONG - Component doing data fetching
export function Dashboard() {
  const { data } = useQuery(...);  // NO! Move to page
  const mutation = useMutation(...);  // NO! Move to page

  return <div>...</div>;
}
```

**Benefits**:

- Components testable in isolation with mock props
- Reusable in different contexts
- Clear separation of concerns
- Easier to refactor

### API Client

```typescript
// api/client.ts
export async function apiClient<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`/api${endpoint}`, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...options?.headers },
  });
  if (!response.ok) throw new Error(`API error: ${response.statusText}`);
  return response.json();
}
```

### Data Fetching Hooks

```typescript
// hooks/useEvents.ts
import { useQuery } from '@tanstack/react-query';

export function useEvents(filters?: EventFilters) {
  return useQuery({
    queryKey: ['events', filters],
    queryFn: () => apiClient<Event[]>('/events'),
    staleTime: 30000,
  });
}

// Usage
function EventsPage() {
  const { data: events, isLoading, error } = useEvents();
  if (isLoading) return <LoadingSpinner />;
  if (error) return <ErrorMessage error={error} />;
  return <EventList events={events} />;
}
```

### Components

```typescript
// components/EventList.tsx
interface EventListProps {
  events: Event[];
  onEventClick?: (event: Event) => void;
}

export function EventList({ events, onEventClick }: EventListProps) {
  return (
    <div className="space-y-2">
      {events.map((event) => (
        <div key={event.id} onClick={() => onEventClick?.(event)}
             className="p-4 border rounded hover:bg-gray-50">
          <h3 className="font-bold">{event.type}</h3>
          <p className="text-sm text-gray-600">{event.actor}</p>
          <time className="text-xs">{new Date(event.timestamp).toLocaleString()}</time>
        </div>
      ))}
    </div>
  );
}
```

### Routing

```typescript
// App.tsx
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';

export function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/plugins" element={<PluginsPage />} />
        <Route path="/events" element={<EventsPage />} />
        <Route path="/settings" element={<SettingsPage />} />
      </Routes>
    </BrowserRouter>
  );
}
```

### Visualizations

```typescript
// components/charts/ThroughputChart.tsx
import { LineChart, Line, XAxis, YAxis, Tooltip } from 'recharts';

export function ThroughputChart({ events }: { events: Event[] }) {
  const data = aggregateByDay(events);
  return (
    <LineChart width={600} height={300} data={data}>
      <XAxis dataKey="date" />
      <YAxis />
      <Tooltip />
      <Line type="monotone" dataKey="count" stroke="#8884d8" />
    </LineChart>
  );
}
```

## TypeScript Types

```typescript
// api/types.ts
export interface Event {
  id: string;
  type: string;
  source: string;
  source_id: string;
  timestamp: string; // ISO 8601 UTC
  actor: string;
  data: Record<string, unknown>;
  anonymized: boolean;
  created_at: string;
  updated_at: string;
}

export interface Plugin {
  name: string;
  version: string;
  type: 'source' | 'destination' | 'processor';
  description: string;
  status: 'configured' | 'configuring' | 'failed' | 'disabled';
  config_fields: ConfigField[];
  last_sync?: string;
}

export interface ConfigField {
  name: string;
  type: 'string' | 'password' | 'boolean' | 'select';
  required: boolean;
  description: string;
  secret: boolean;
  default?: string;
  options?: string[];
}

export interface Insight {
  severity: 'info' | 'warning' | 'critical';
  title: string;
  description: string;
  recommendation?: string;
  metrics?: Record<string, number>;
}
```

## Styling with Tailwind

```tsx
// Utility-first approach
<div className="p-4 bg-white border border-gray-200 rounded-lg shadow-sm">
  <h2 className="text-xl font-bold text-gray-900">Title</h2>
  <p className="mt-2 text-gray-600">Description</p>
</div>

// Responsive design
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
  {plugins.map((plugin) => <PluginCard key={plugin.name} plugin={plugin} />)}
</div>

// Dark mode support (REQUIRED for all components)
<div className="p-4 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700">
  <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">Title</h2>
  <p className="mt-2 text-gray-600 dark:text-gray-400">Description</p>
</div>
```

## Theme System

**Location**: Context provider in `src/contexts/ThemeContext.tsx`
**Configuration**: Tailwind config uses `darkMode: 'class'` strategy
**Storage**: User preference persisted to localStorage as `engineerdna-theme`

### How to Switch Themes

**For Users**:

- Click the theme toggle button in the top-right corner of the navigation bar
- Sun icon = currently in light mode (click to switch to dark)
- Moon icon = currently in dark mode (click to switch to light)
- Preference is saved automatically and persists across sessions
- On first visit, defaults to system preference

**For Developers**:

```typescript
// Using the theme context
import { useTheme } from '../contexts/ThemeContext';

function MyComponent() {
  const { theme, setTheme } = useTheme();

  return (
    <button onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}>
      Toggle theme
    </button>
  );
}
```

### Dark Mode Implementation Rules

**REQUIRED for all new components**:

1. **Background colors**: Always provide dark variants

   ```tsx
   bg-white dark:bg-gray-800
   bg-gray-100 dark:bg-gray-900
   bg-gray-50 dark:bg-gray-800
   ```

2. **Text colors**: Always provide dark variants

   ```tsx
   text-gray-900 dark:text-gray-100
   text-gray-600 dark:text-gray-400
   text-gray-500 dark:text-gray-500
   ```

3. **Borders**: Always provide dark variants

   ```tsx
   border-gray-300 dark:border-gray-600
   border-gray-200 dark:border-gray-700
   ```

4. **Shadows**: Optional in dark mode (often disabled)

   ```tsx
   shadow-sm dark:shadow-none
   ```

5. **State colors**: Maintain accessibility in both themes
   ```tsx
   bg-blue-600 dark:bg-blue-500
   text-red-600 dark:text-red-400
   bg-yellow-50 dark:bg-yellow-900/20
   ```

### Loading Skeletons (Dark Mode)

```tsx
<div className="animate-pulse">
  <div className="h-4 bg-gray-300 dark:bg-gray-700 rounded w-3/4 mb-2" />
  <div className="h-4 bg-gray-300 dark:bg-gray-700 rounded w-1/2" />
</div>
```

### Testing Dark Mode

1. Build and run application
2. Toggle theme using nav bar button
3. Verify all components have proper contrast in both modes
4. Check WCAG AA contrast ratio (4.5:1 for normal text, 3:1 for large text)
5. Test focus states and interactive elements

## State Management

### TanStack Query for Server State

```typescript
// Query
const { data, isLoading, error, refetch } = useQuery({
  queryKey: ['events'],
  queryFn: fetchEvents,
});

// Mutation
const mutation = useMutation({
  mutationFn: (config: PluginConfig) =>
    apiClient('/plugins/configure', {
      method: 'POST',
      body: JSON.stringify(config),
    }),
  onSuccess: () => queryClient.invalidateQueries({ queryKey: ['plugins'] }),
});

// Optimistic update
const mutation = useMutation({
  mutationFn: updateEvent,
  onMutate: async (newEvent) => {
    await queryClient.cancelQueries({ queryKey: ['events'] });
    const previous = queryClient.getQueryData(['events']);
    queryClient.setQueryData(['events'], (old) => [...old, newEvent]);
    return { previous };
  },
  onError: (err, newEvent, context) => {
    queryClient.setQueryData(['events'], context.previous);
  },
});
```

### React State for UI State

```typescript
const [isOpen, setIsOpen] = useState(false);
const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
```

## Testing

### Unit Tests

```typescript
// hooks/useEvents.test.ts
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

function wrapper({ children }) {
  return <QueryClientProvider client={new QueryClient()}>{children}</QueryClientProvider>;
}

test('useEvents fetches events', async () => {
  const { result } = renderHook(() => useEvents(), { wrapper });
  await waitFor(() => expect(result.current.isSuccess).toBe(true));
  expect(result.current.data).toHaveLength(10);
});
```

### Integration Tests

```typescript
// components/EventList.test.tsx
import { render, screen } from '@testing-library/react';

test('renders events', () => {
  const events = [{id: '1', type: 'pull_request', actor: 'alice@example.com', ...}];
  render(<EventList events={events} />);
  expect(screen.getByText('pull_request')).toBeInTheDocument();
  expect(screen.getByText('alice@example.com')).toBeInTheDocument();
});
```

## Build and Deployment

### Development

```bash
# Vite dev server with HMR
npm run dev

# vite.config.ts
server: {
  proxy: {
    '/api': {
      target: 'http://127.0.0.1:3847',
      changeOrigin: true,
    },
  },
}
```

### Production

```bash
npm run build  # Output: dist/

# Embedding in Go
//go:embed frontend/dist
var frontendFS embed.FS

frontend, _ := fs.Sub(frontendFS, "frontend/dist")
http.Handle("/", http.FileServer(http.FS(frontend)))
```

## Common Patterns

### Loading States

```tsx
function DashboardPage() {
  const { data: events, isLoading, error, refetch } = useEvents();

  if (isLoading) {
    return (
      <div className="flex justify-center h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 bg-red-50 border border-red-200 rounded">
        <p className="text-red-800">Failed: {error.message}</p>
        <button onClick={() => refetch()} className="mt-2 text-blue-600">
          Retry
        </button>
      </div>
    );
  }

  return <EventList events={events} />;
}
```

### Form Handling

```tsx
function PluginConfigForm({ plugin }: { plugin: Plugin }) {
  const [config, setConfig] = useState<Record<string, string>>({});
  const mutation = useMutation({
    mutationFn: (data) =>
      apiClient('/plugins/configure', {
        method: 'POST',
        body: JSON.stringify({ name: plugin.name, config: data }),
      }),
  });

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        mutation.mutate(config);
      }}
    >
      {plugin.config_fields.map((field) => (
        <div key={field.name}>
          <label className="block text-sm font-medium">{field.description}</label>
          <input
            type={field.secret ? 'password' : 'text'}
            value={config[field.name] || ''}
            onChange={(e) => setConfig({ ...config, [field.name]: e.target.value })}
            required={field.required}
            className="mt-1 block w-full rounded border-gray-300"
          />
        </div>
      ))}
      <button
        type="submit"
        disabled={mutation.isPending}
        className="px-4 py-2 bg-blue-600 text-white rounded"
      >
        {mutation.isPending ? 'Saving...' : 'Save'}
      </button>
    </form>
  );
}
```

## Performance

### Code Splitting

```typescript
import { lazy, Suspense } from 'react';

const DashboardPage = lazy(() => import('./pages/DashboardPage'));

function App() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <Routes>
        <Route path="/dashboard" element={<DashboardPage />} />
      </Routes>
    </Suspense>
  );
}
```

### Memoization

```typescript
import { memo, useMemo } from 'react';

export const EventList = memo(function EventList({ events }) {
  return <div>{/* ... */}</div>;
});

function Dashboard({ events }) {
  const stats = useMemo(() => calculateStats(events), [events]);
  return <div>{/* ... */}</div>;
}
```

## Security

### No Secrets

```typescript
// [BAD]
const API_KEY = 'sk-abc123';

// [GOOD]
// Backend handles auth, frontend doesn't need keys
```

### Input Validation

```typescript
function sanitizeInput(value: string): string {
  return value.trim().replace(/[<>]/g, '');
}
```

### XSS Prevention

```typescript
// React escapes by default
<div>{userInput}</div> // Safe

// Dangerous: avoid dangerouslySetInnerHTML
<div dangerouslySetInnerHTML={{ __html: userInput }} /> // Unsafe!
```

## Troubleshooting

**API calls fail with CORS**: Ensure Vite proxy configured in vite.config.ts

**Build fails with TypeScript errors**: Run `npm run typecheck`

**Hot reload not working**: Restart dev server, check for syntax errors

**Blank page in production**: Verify `frontend/dist/` exists, rebuild Go binary

## References

- Rule 35: All files under 500 lines
- Root CLAUDE.md: Project-wide context
- Backend API: `internal/` directory
- Vite: https://vite.dev
- React: https://react.dev
- TanStack Query: https://tanstack.com/query
- Tailwind: https://tailwindcss.com
