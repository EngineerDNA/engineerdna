import { useState } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { api } from '../../api/client';

interface FirstSyncStepProps {
  onComplete: () => void;
  onBack: () => void;
}

export function FirstSyncStep({ onComplete, onBack }: FirstSyncStepProps) {
  const [syncedPlugins, setSyncedPlugins] = useState<Set<string>>(new Set());

  const { data: pluginsData } = useQuery({
    queryKey: ['plugins'],
    queryFn: api.getPlugins,
  });

  const syncMutation = useMutation({
    mutationFn: (pluginName: string) => api.syncPlugin(pluginName),
    onSuccess: (_, pluginName) => {
      setSyncedPlugins((prev) => new Set(prev).add(pluginName));
    },
  });

  const configuredPlugins =
    pluginsData?.plugins.filter((p) => p.type === 'source' && p.enabled) || [];

  const handleSyncPlugin = (pluginName: string) => {
    syncMutation.mutate(pluginName);
  };

  const handleSyncAll = () => {
    configuredPlugins.forEach((plugin) => {
      if (!syncedPlugins.has(plugin.name)) {
        handleSyncPlugin(plugin.name);
      }
    });
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">
          Run Your First Sync
        </h2>
        <p className="text-gray-600 dark:text-gray-400">
          Sync data from your configured sources to start analyzing
        </p>
      </div>

      {configuredPlugins.length === 0 ? (
        <div className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 text-center">
          <p className="text-gray-600 dark:text-gray-400 mb-4">
            No sources configured yet. You can configure sources later from the Plugins page.
          </p>
          <button
            onClick={onComplete}
            className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 transition-colors focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-900"
          >
            Finish Setup
          </button>
        </div>
      ) : (
        <>
          <div className="space-y-3">
            {configuredPlugins.map((plugin) => {
              const isSyncing = syncMutation.isPending && syncMutation.variables === plugin.name;
              const isSynced = syncedPlugins.has(plugin.name);

              return (
                <div
                  key={plugin.name}
                  className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4 flex items-center justify-between"
                >
                  <div className="flex items-center gap-3">
                    {isSynced ? (
                      <svg
                        className="w-6 h-6 text-green-600 dark:text-green-400"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                        />
                      </svg>
                    ) : isSyncing ? (
                      <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600 dark:border-blue-400" />
                    ) : (
                      <div className="w-6 h-6 rounded-full border-2 border-gray-300 dark:border-gray-600" />
                    )}
                    <div>
                      <h3 className="font-medium text-gray-900 dark:text-gray-100 capitalize">
                        {plugin.name}
                      </h3>
                      {isSyncing && (
                        <p className="text-sm text-gray-600 dark:text-gray-400">Syncing...</p>
                      )}
                      {isSynced && (
                        <p className="text-sm text-green-600 dark:text-green-400">Sync completed</p>
                      )}
                    </div>
                  </div>
                  {!isSynced && !isSyncing && (
                    <button
                      onClick={() => handleSyncPlugin(plugin.name)}
                      className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 transition-colors text-sm"
                    >
                      Sync Now
                    </button>
                  )}
                </div>
              );
            })}
          </div>

          {syncedPlugins.size === 0 && (
            <button
              onClick={handleSyncAll}
              disabled={syncMutation.isPending}
              className="w-full px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 transition-colors focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-900 font-medium disabled:opacity-50"
            >
              Sync All Sources
            </button>
          )}

          <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Syncing may take a few moments depending on the amount of data. You can skip this step
              and sync later.
            </p>
          </div>
        </>
      )}

      <div className="flex justify-between pt-4 border-t border-gray-200 dark:border-gray-700">
        <button
          onClick={onBack}
          className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100 transition-colors"
        >
          Back
        </button>
        <button
          onClick={onComplete}
          className="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 dark:bg-green-700 dark:hover:bg-green-800 transition-colors focus:ring-2 focus:ring-green-500 focus:ring-offset-2 dark:focus:ring-offset-gray-900"
        >
          Finish Setup
        </button>
      </div>
    </div>
  );
}
