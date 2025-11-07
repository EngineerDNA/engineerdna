import { memo } from 'react';
import type { Plugin } from '../api/types';

interface PluginCardProps {
  plugin: Plugin;
  onConfigure: () => void;
  onSync: () => void;
  onSchedule?: () => void;
}

export const PluginCard = memo(function PluginCard({
  plugin,
  onConfigure,
  onSync,
  onSchedule,
}: PluginCardProps) {
  return (
    <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow hover:shadow-md transition-shadow">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">{plugin.name}</h3>
          <div className="text-sm text-gray-600 dark:text-gray-400 capitalize">{plugin.type}</div>
        </div>
        <HealthBadge health={plugin.health} />
      </div>

      {plugin.last_sync && (
        <div className="text-sm text-gray-600 dark:text-gray-400 mb-4">
          Last sync: {new Date(plugin.last_sync).toLocaleString()}
        </div>
      )}

      <div className="flex flex-col gap-2">
        <div className="flex gap-2">
          <button
            onClick={onConfigure}
            className="flex-1 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 transition-colors focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800"
          >
            Configure
          </button>
          {plugin.enabled && plugin.type === 'source' && (
            <button
              onClick={onSync}
              className="flex-1 px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 dark:bg-green-700 dark:hover:bg-green-800 transition-colors focus:ring-2 focus:ring-green-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800"
            >
              Sync Now
            </button>
          )}
        </div>
        {plugin.enabled && plugin.type === 'destination' && onSchedule && (
          <button
            onClick={onSchedule}
            className="w-full px-4 py-2 bg-purple-600 text-white rounded hover:bg-purple-700 dark:bg-purple-700 dark:hover:bg-purple-800 transition-colors focus:ring-2 focus:ring-purple-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800"
          >
            Schedule Export
          </button>
        )}
      </div>
    </div>
  );
});

function HealthBadge({ health }: { health: string }) {
  const colors = {
    healthy: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
    error: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
    unknown: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300',
  };

  return (
    <span
      className={`px-2 py-1 rounded text-sm font-medium ${colors[health as keyof typeof colors] || colors.unknown}`}
    >
      {health}
    </span>
  );
}
