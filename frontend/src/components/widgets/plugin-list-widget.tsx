import { useQuery } from '@tanstack/react-query';
import { useSyncPlugin } from '../../hooks/usePlugins';
import { api } from '../../api/client';
import type { WidgetProps } from './widget-registry';
import type { Plugin } from '../../api/types';

export function PluginListWidget({ config, onOpenModal }: WidgetProps) {
  const pluginTypeFilter = config.plugin_type as string | undefined;

  const { data, isLoading } = useQuery({
    queryKey: ['plugins'],
    queryFn: api.getPlugins,
    staleTime: 5 * 60 * 1000,
  });

  const syncMutation = useSyncPlugin();

  const plugins = data?.plugins || [];
  const filteredPlugins = pluginTypeFilter
    ? plugins.filter((p) => p.type === pluginTypeFilter)
    : plugins;

  const groupedPlugins = filteredPlugins.reduce(
    (acc, plugin) => {
      if (!acc[plugin.type]) acc[plugin.type] = [];
      acc[plugin.type].push(plugin);
      return acc;
    },
    {} as Record<string, Plugin[]>
  );

  const getStatusColor = (health: Plugin['health']) => {
    switch (health) {
      case 'healthy':
        return 'bg-green-100 dark:bg-green-900/20 text-green-800 dark:text-green-200';
      case 'error':
        return 'bg-red-100 dark:bg-red-900/20 text-red-800 dark:text-red-200';
      default:
        return 'bg-gray-100 dark:bg-gray-900/20 text-gray-800 dark:text-gray-200';
    }
  };

  const getStatusIcon = (health: Plugin['health']) => {
    if (health === 'healthy') {
      return (
        <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
          <path
            fillRule="evenodd"
            d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
            clipRule="evenodd"
          />
        </svg>
      );
    }
    if (health === 'error') {
      return (
        <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
          <path
            fillRule="evenodd"
            d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
            clipRule="evenodd"
          />
        </svg>
      );
    }
    return (
      <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
        <path
          fillRule="evenodd"
          d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
          clipRule="evenodd"
        />
      </svg>
    );
  };

  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="animate-pulse space-y-3">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="h-20 bg-gray-300 dark:bg-gray-700 rounded" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full flex flex-col">
      <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-4">Plugins</h3>

      <div className="flex-1 overflow-auto space-y-4">
        {Object.keys(groupedPlugins).length === 0 ? (
          <div className="flex items-center justify-center h-32 text-gray-500 dark:text-gray-400">
            No plugins found
          </div>
        ) : (
          Object.entries(groupedPlugins).map(([type, typePlugins]) => (
            <div key={type}>
              <h4 className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase mb-2">
                {type.replace('_', ' ')}
              </h4>
              <div className="space-y-2">
                {typePlugins.map((plugin) => (
                  <div
                    key={plugin.name}
                    className="border border-gray-200 dark:border-gray-700 rounded-lg p-3"
                  >
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex-1">
                        <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                          {plugin.name}
                        </p>
                        {plugin.last_sync && (
                          <p className="text-xs text-gray-500 dark:text-gray-400">
                            Last sync: {new Date(plugin.last_sync).toLocaleString()}
                          </p>
                        )}
                      </div>
                      <span
                        className={`inline-flex items-center gap-1 px-2 py-1 rounded text-xs ${getStatusColor(plugin.health)}`}
                      >
                        {getStatusIcon(plugin.health)}
                        {plugin.health}
                      </span>
                    </div>

                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => onOpenModal?.('plugin-config', { pluginName: plugin.name })}
                        className="text-xs px-3 py-1 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-900 transition-colors"
                      >
                        Configure
                      </button>
                      {plugin.type === 'source' && (
                        <button
                          onClick={() => syncMutation.mutate(plugin.name)}
                          disabled={syncMutation.isPending || !plugin.enabled}
                          className="text-xs px-3 py-1 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
                        >
                          {syncMutation.isPending ? 'Syncing...' : 'Sync'}
                        </button>
                      )}
                      <button
                        onClick={() => onOpenModal?.('plugin-detail', { pluginName: plugin.name })}
                        className="text-xs text-blue-600 dark:text-blue-400 hover:underline"
                      >
                        Details
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
