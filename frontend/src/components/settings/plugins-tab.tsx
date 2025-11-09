import { useState, useMemo } from 'react';
import { usePlugins, useSyncPlugin } from '../../hooks/usePlugins';
import { PluginCard } from '../PluginCard';
import { PluginConfig } from '../PluginConfig';
import { ErrorAlert } from '../ErrorAlert';
import { ScheduleModal } from '../schedule-modal';

type SortOption = 'name' | 'type' | 'health';
type PluginType = 'all' | 'source' | 'destination' | 'processor';
type HealthFilter = 'all' | 'healthy' | 'error' | 'unknown';

export function PluginsTab() {
  const { data, isLoading, error: fetchError } = usePlugins();
  const syncMutation = useSyncPlugin();
  const [configuring, setConfiguring] = useState<string | null>(null);
  const [schedulingPlugin, setSchedulingPlugin] = useState<string | null>(null);

  const [searchQuery, setSearchQuery] = useState('');
  const [sortBy, setSortBy] = useState<SortOption>('name');
  const [typeFilter, setTypeFilter] = useState<PluginType>('all');
  const [healthFilter, setHealthFilter] = useState<HealthFilter>('all');

  const plugins = data?.plugins || [];

  const filteredAndSortedPlugins = useMemo(() => {
    let result = [...plugins];

    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      result = result.filter((plugin) => plugin.name.toLowerCase().includes(query));
    }

    if (typeFilter !== 'all') {
      result = result.filter((plugin) => plugin.type === typeFilter);
    }

    if (healthFilter !== 'all') {
      result = result.filter((plugin) => plugin.health === healthFilter);
    }

    result.sort((a, b) => {
      if (sortBy === 'name') {
        return a.name.localeCompare(b.name);
      }
      if (sortBy === 'type') {
        return a.type.localeCompare(b.type);
      }
      if (sortBy === 'health') {
        return a.health.localeCompare(b.health);
      }
      return 0;
    });

    return result;
  }, [plugins, searchQuery, typeFilter, healthFilter, sortBy]);

  // Group plugins by type
  const groupedPlugins = useMemo(() => {
    const groups: Record<string, typeof filteredAndSortedPlugins> = {
      source: [],
      processor: [],
      destination: [],
    };

    filteredAndSortedPlugins.forEach((plugin) => {
      if (groups[plugin.type]) {
        groups[plugin.type].push(plugin);
      }
    });

    return groups;
  }, [filteredAndSortedPlugins]);

  if (isLoading) {
    return <LoadingSkeleton />;
  }

  return (
    <div className="space-y-6">
      {fetchError && (
        <ErrorAlert error={fetchError as Error} onRetry={() => window.location.reload()} />
      )}

      {syncMutation.error && (
        <ErrorAlert
          error={syncMutation.error as Error}
          onRetry={() => syncMutation.reset()}
          onDismiss={() => syncMutation.reset()}
        />
      )}

      {/* Search and Filter */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="flex-1">
            <input
              type="text"
              placeholder="Search plugins..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>

          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as SortOption)}
            className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="name">Sort by Name</option>
            <option value="type">Sort by Type</option>
            <option value="health">Sort by Health</option>
          </select>

          <select
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value as PluginType)}
            className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="all">All Types</option>
            <option value="source">Source</option>
            <option value="destination">Destination</option>
            <option value="processor">Processor</option>
          </select>

          <select
            value={healthFilter}
            onChange={(e) => setHealthFilter(e.target.value as HealthFilter)}
            className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="all">All Status</option>
            <option value="healthy">Healthy</option>
            <option value="error">Error</option>
            <option value="unknown">Unknown</option>
          </select>
        </div>

        <div className="mt-4 text-sm text-gray-600 dark:text-gray-400">
          Showing {filteredAndSortedPlugins.length} of {plugins.length} plugins
        </div>
      </div>

      {/* Grouped Plugin Lists */}
      {Object.entries(groupedPlugins).map(([type, pluginList]) =>
        pluginList.length > 0 ? (
          <div key={type} className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4 capitalize">
              {type} Plugins ({pluginList.length})
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {pluginList.map((plugin) => (
                <PluginCard
                  key={plugin.name}
                  plugin={plugin}
                  onConfigure={() => setConfiguring(plugin.name)}
                  onSync={() => syncMutation.mutate(plugin.name)}
                  onSchedule={
                    plugin.type === 'destination'
                      ? () => setSchedulingPlugin(plugin.name)
                      : undefined
                  }
                />
              ))}
            </div>
          </div>
        ) : null
      )}

      {filteredAndSortedPlugins.length === 0 && (
        <EmptyState
          hasFilters={searchQuery !== '' || typeFilter !== 'all' || healthFilter !== 'all'}
        />
      )}

      {configuring && (
        <PluginConfig
          plugin={plugins.find((p) => p.name === configuring)!}
          onClose={() => setConfiguring(null)}
        />
      )}

      {schedulingPlugin && (
        <ScheduleModal
          isOpen={true}
          onClose={() => setSchedulingPlugin(null)}
          plugins={plugins}
          preSelectedPlugin={schedulingPlugin}
        />
      )}
    </div>
  );
}

function LoadingSkeleton() {
  return (
    <div className="space-y-6">
      {[1, 2, 3].map((i) => (
        <div key={i} className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 animate-pulse">
          <div className="h-6 bg-gray-200 dark:bg-gray-700 rounded w-32 mb-4"></div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {[1, 2, 3].map((j) => (
              <div key={j} className="p-4 border border-gray-200 dark:border-gray-700 rounded">
                <div className="h-6 bg-gray-200 dark:bg-gray-700 rounded w-24 mb-2"></div>
                <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-16 mb-4"></div>
                <div className="h-10 bg-gray-200 dark:bg-gray-700 rounded"></div>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}

function EmptyState({ hasFilters }: { hasFilters: boolean }) {
  return (
    <div className="bg-white dark:bg-gray-800 p-12 rounded-lg shadow text-center">
      <svg
        className="mx-auto h-12 w-12 text-gray-400 dark:text-gray-600 mb-4"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"
        />
      </svg>
      <p className="text-gray-600 dark:text-gray-400 text-lg mb-2">
        {hasFilters ? 'No plugins match your filters' : 'No plugins available'}
      </p>
      <p className="text-sm text-gray-500 dark:text-gray-500">
        {hasFilters
          ? 'Try adjusting your search or filters'
          : 'Plugins will appear here once discovered'}
      </p>
    </div>
  );
}
