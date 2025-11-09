import { useSystemInfo } from '../../hooks/useSystemInfo';
import { useSettings } from '../../hooks/useSettings';
import { useState, useEffect } from 'react';
import { IgnoredIdentitiesList } from '../IgnoredIdentitiesList';
import { BYTES_PER_KB } from '../../constants';

type SyncSchedule = 'manual' | 'hourly' | 'daily' | 'weekly';
type AnonymizationStrategy = 'sequential' | 'uuid' | 'hash';

export function SystemTab() {
  const { data: systemInfo, isLoading } = useSystemInfo();
  const { data: settings, isLoading: isLoadingSettings } = useSettings();

  const [syncSchedule, setSyncSchedule] = useState<SyncSchedule>('manual');
  const [anonymizationStrategy, setAnonymizationStrategy] =
    useState<AnonymizationStrategy>('sequential');
  const [port, setPort] = useState('3847');

  useEffect(() => {
    if (settings) {
      setSyncSchedule(settings.sync_schedule as SyncSchedule);
      setAnonymizationStrategy(settings.anonymization_strategy as AnonymizationStrategy);
      setPort(settings.port.toString());
    }
  }, [settings]);

  const [saveSuccess, setSaveSuccess] = useState<string | null>(null);
  const [saveError, setSaveError] = useState<string | null>(null);

  const handleSaveSchedule = async () => {
    setSaveSuccess(null);
    setSaveError(null);
    try {
      const response = await fetch('/api/settings/sync-schedule', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ schedule: syncSchedule }),
      });
      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to save sync schedule');
      }
      setSaveSuccess('Sync schedule saved successfully');
      setTimeout(() => setSaveSuccess(null), 3000);
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : 'Failed to save sync schedule');
      setTimeout(() => setSaveError(null), 5000);
    }
  };

  const handleSaveAnonymization = async () => {
    setSaveSuccess(null);
    setSaveError(null);
    try {
      const response = await fetch('/api/settings/anonymization-strategy', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ strategy: anonymizationStrategy }),
      });
      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to save anonymization strategy');
      }
      setSaveSuccess('Anonymization strategy saved successfully');
      setTimeout(() => setSaveSuccess(null), 3000);
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : 'Failed to save anonymization strategy');
      setTimeout(() => setSaveError(null), 5000);
    }
  };

  const handleSavePort = async () => {
    setSaveSuccess(null);
    setSaveError(null);
    const portNum = parseInt(port);
    if (isNaN(portNum) || portNum < 1024 || portNum > 65535) {
      setSaveError('Port must be between 1024 and 65535');
      setTimeout(() => setSaveError(null), 3000);
      return;
    }
    try {
      const response = await fetch('/api/settings/port', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ port: portNum }),
      });
      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to save port');
      }
      const data = await response.json();
      setSaveSuccess(data.message || 'Port saved successfully');
      setTimeout(() => setSaveSuccess(null), 5000);
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : 'Failed to save port');
      setTimeout(() => setSaveError(null), 5000);
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = BYTES_PER_KB;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
  };

  if (isLoading || isLoadingSettings) {
    return <LoadingSkeleton />;
  }

  return (
    <div className="space-y-6">
      {saveSuccess && (
        <div className="p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded">
          <p className="text-green-800 dark:text-green-200">{saveSuccess}</p>
        </div>
      )}

      {saveError && (
        <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded">
          <p className="text-red-800 dark:text-red-200">{saveError}</p>
        </div>
      )}

      {/* System Information */}
      <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
        <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">
          System Information
        </h3>
        <div className="space-y-2">
          <div className="flex justify-between">
            <span className="text-gray-600 dark:text-gray-400">Version:</span>
            <span className="text-gray-900 dark:text-gray-100">{systemInfo?.version}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-gray-600 dark:text-gray-400">Plugins Discovered:</span>
            <span className="text-gray-900 dark:text-gray-100">{systemInfo?.plugins.count}</span>
          </div>
        </div>
      </div>

      {/* Database */}
      <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
        <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">Database</h3>
        <div className="space-y-3">
          <div className="flex flex-col sm:flex-row sm:justify-between gap-2">
            <span className="text-gray-600 dark:text-gray-400">Location:</span>
            <span className="font-mono text-sm text-gray-900 dark:text-gray-100 break-all">
              {systemInfo?.database.path}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-gray-600 dark:text-gray-400">Size:</span>
            <span className="text-gray-900 dark:text-gray-100">
              {formatBytes(systemInfo?.database.size_bytes || 0)}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-gray-600 dark:text-gray-400">Events:</span>
            <span className="text-gray-900 dark:text-gray-100">
              {systemInfo?.database.event_count.toLocaleString()}
            </span>
          </div>
        </div>
      </div>

      {/* Encryption */}
      <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
        <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">Encryption</h3>
        <div className="space-y-3">
          <div className="flex justify-between">
            <span className="text-gray-600 dark:text-gray-400">Master Key Source:</span>
            <span className="capitalize text-gray-900 dark:text-gray-100">
              {systemInfo?.encryption.master_key_source}
            </span>
          </div>
        </div>
      </div>

      {/* Sync Schedule */}
      <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
        <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">
          Sync Schedule
        </h3>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
          Configure how often plugins should automatically sync data
        </p>
        <div className="flex flex-col sm:flex-row gap-4">
          <select
            value={syncSchedule}
            onChange={(e) => setSyncSchedule(e.target.value as SyncSchedule)}
            className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="manual">Manual Only</option>
            <option value="hourly">Every Hour</option>
            <option value="daily">Daily</option>
            <option value="weekly">Weekly</option>
          </select>
          <button
            onClick={handleSaveSchedule}
            className="px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 transition-colors focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800"
          >
            Save
          </button>
        </div>
      </div>

      {/* Anonymization */}
      <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
        <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">
          Anonymization
        </h3>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
          Choose how identities are anonymized when processing data externally
        </p>
        <div className="flex flex-col sm:flex-row gap-4">
          <select
            value={anonymizationStrategy}
            onChange={(e) => setAnonymizationStrategy(e.target.value as AnonymizationStrategy)}
            className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="sequential">Sequential (Engineer_A, B, C)</option>
            <option value="uuid">UUID (random identifiers)</option>
            <option value="hash">Hash (one-way hash)</option>
          </select>
          <button
            onClick={handleSaveAnonymization}
            className="px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 transition-colors focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800"
          >
            Save
          </button>
        </div>
      </div>

      {/* Port Configuration */}
      <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
        <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">
          Port Configuration
        </h3>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
          Change the port EngineerDNA runs on (requires restart)
        </p>
        <div className="flex flex-col sm:flex-row gap-4">
          <input
            type="number"
            value={port}
            onChange={(e) => setPort(e.target.value)}
            min="1024"
            max="65535"
            className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="3847"
          />
          <button
            onClick={handleSavePort}
            className="px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 transition-colors focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800"
          >
            Save
          </button>
        </div>
        <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">Valid range: 1024-65535</p>
      </div>

      {/* Ignored Identities */}
      <IgnoredIdentitiesList />

      {/* Data Management */}
      <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
        <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">
          Data Management
        </h3>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
          Advanced data cleanup and reset options
        </p>
        <div className="space-y-2">
          <button className="w-full sm:w-auto px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors">
            Vacuum Database
          </button>
          <button className="w-full sm:w-auto px-4 py-2 bg-yellow-600 dark:bg-yellow-700 text-white rounded hover:bg-yellow-700 dark:hover:bg-yellow-800 transition-colors ml-0 sm:ml-2">
            Clear Cache
          </button>
          <button className="w-full sm:w-auto px-4 py-2 bg-red-600 dark:bg-red-700 text-white rounded hover:bg-red-700 dark:hover:bg-red-800 transition-colors ml-0 sm:ml-2">
            Reset All Data
          </button>
        </div>
      </div>
    </div>
  );
}

function LoadingSkeleton() {
  return (
    <div className="space-y-6">
      {[1, 2, 3, 4, 5].map((i) => (
        <div key={i} className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow animate-pulse">
          <div className="h-6 bg-gray-200 dark:bg-gray-700 rounded w-32 mb-4"></div>
          <div className="space-y-2">
            <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
            <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4"></div>
          </div>
        </div>
      ))}
    </div>
  );
}
