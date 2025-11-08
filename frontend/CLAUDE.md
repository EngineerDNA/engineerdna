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
  return <div>...</div>;
}
```

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
```

### TypeScript Types

```typescript
// api/types.ts
export interface Event {
  id: string;
  type: string;
  source: string;
  timestamp: string; // ISO 8601 UTC
  actor: string;
  data: Record<string, unknown>;
  anonymized: boolean;
}

export interface Plugin {
  name: string;
  version: string;
  type: 'source' | 'destination' | 'processor';
  description: string;
  status: 'configured' | 'configuring' | 'failed' | 'disabled';
  config_fields: ConfigField[];
}

export interface ConfigField {
  name: string;
  type: 'string' | 'password' | 'boolean' | 'select';
  required: boolean;
  secret: boolean;
  options?: string[];
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

### Dark Mode Implementation Rules

**REQUIRED for all new components**:

1. **Background colors**: Always provide dark variants
   ```tsx
   bg-white dark:bg-gray-800
   bg-gray-100 dark:bg-gray-900
   ```

2. **Text colors**: Always provide dark variants
   ```tsx
   text-gray-900 dark:text-gray-100
   text-gray-600 dark:text-gray-400
   ```

3. **Borders**: Always provide dark variants
   ```tsx
   border-gray-300 dark:border-gray-600
   ```

4. **State colors**: Maintain accessibility in both themes
   ```tsx
   bg-blue-600 dark:bg-blue-500
   text-red-600 dark:text-red-400
   ```

### Testing Dark Mode

1. Build and run application
2. Toggle theme using nav bar button
3. Verify all components have proper contrast in both modes
4. Check WCAG AA contrast ratio (4.5:1 for normal text, 3:1 for large text)

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

## Common Patterns

### Loading States

```tsx
function DashboardPage() {
  const { data: events, isLoading, error, refetch } = useEvents();

  if (isLoading) {
    return <div className="flex justify-center h-screen">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600" />
    </div>;
  }

  if (error) {
    return <div className="p-4 bg-red-50 border border-red-200 rounded">
      <p className="text-red-800">Failed: {error.message}</p>
      <button onClick={() => refetch()} className="mt-2 text-blue-600">Retry</button>
    </div>;
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
    <form onSubmit={(e) => { e.preventDefault(); mutation.mutate(config); }}>
      {plugin.config_fields.map((field) => (
        <div key={field.name}>
          <label className="block text-sm font-medium">{field.name}</label>
          <input
            type={field.secret ? 'password' : 'text'}
            value={config[field.name] || ''}
            onChange={(e) => setConfig({ ...config, [field.name]: e.target.value })}
            required={field.required}
            className="mt-1 block w-full rounded border-gray-300"
          />
        </div>
      ))}
      <button type="submit" disabled={mutation.isPending}
        className="px-4 py-2 bg-blue-600 text-white rounded">
        {mutation.isPending ? 'Saving...' : 'Save'}
      </button>
    </form>
  );
}
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

**No Secrets**: Backend handles auth, frontend doesn't need keys
**Input Validation**: Sanitize user inputs
**XSS Prevention**: React escapes by default (never use dangerouslySetInnerHTML)

## Troubleshooting

**API calls fail with CORS**: Ensure Vite proxy configured in vite.config.ts
**Build fails with TypeScript errors**: Run `npm run typecheck`
**Hot reload not working**: Restart dev server, check for syntax errors
**Blank page in production**: Verify `frontend/dist/` exists, rebuild Go binary

## References

- Rule: All files under 500 lines
- Root CLAUDE.md: Project-wide context
- Backend API: `internal/` directory
- Vite: https://vite.dev
- React: https://react.dev
- TanStack Query: https://tanstack.com/query
- Tailwind: https://tailwindcss.com
