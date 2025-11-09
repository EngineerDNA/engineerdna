import { useState } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { api } from '../../api/client';
import type { UserRole } from '../../api/types';

interface FirstSyncStepProps {
  selectedRole: UserRole | null;
  onComplete: () => void;
  onBack: () => void;
}

export function FirstSyncStep({ selectedRole, onComplete, onBack }: FirstSyncStepProps) {
  const [syncedPlugins, setSyncedPlugins] = useState<Set<string>>(new Set());
  const [isCreatingDashboards, setIsCreatingDashboards] = useState(false);
  const [dashboardsCreated, setDashboardsCreated] = useState(false);
  const [createdDashboardCount, setCreatedDashboardCount] = useState(0);

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

  const setRoleMutation = useMutation({
    mutationFn: (role: UserRole) => api.setUserRole(role),
    onSuccess: () => {
      setDashboardsCreated(true);
      setIsCreatingDashboards(false);
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

  const handleCreateDashboards = async () => {
    if (!selectedRole) return;

    setIsCreatingDashboards(true);

    // Fetch templates for the role to count them
    try {
      const { templates } = await api.getDashboardTemplates(selectedRole);
      setCreatedDashboardCount(templates.length);
    } catch (error) {
      console.error('Failed to fetch templates:', error);
    }

    // Set role (this will auto-create dashboards from templates on backend)
    setRoleMutation.mutate(selectedRole);
  };

  const handleFinish = () => {
    // Always ensure role is set and dashboards created before completing
    if (!dashboardsCreated && selectedRole) {
      handleCreateDashboards();
    } else if (selectedRole && !setRoleMutation.isSuccess) {
      // If dashboards already created but role wasn't saved via mutation, save it now
      setRoleMutation.mutate(selectedRole);
    } else {
      onComplete();
    }
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

      {isCreatingDashboards ? (
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-6 text-center">
          <div className="flex items-center justify-center mb-4">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400" />
          </div>
          <h3 className="font-medium text-gray-900 dark:text-gray-100 mb-2">
            Creating your dashboards...
          </h3>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Setting up {createdDashboardCount} dashboards based on your role
          </p>
        </div>
      ) : dashboardsCreated ? (
        <div className="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg p-6 text-center">
          <svg
            className="w-12 h-12 text-green-600 dark:text-green-400 mx-auto mb-4"
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
          <h3 className="font-medium text-gray-900 dark:text-gray-100 mb-2">
            All set! Created {createdDashboardCount} dashboards for you
          </h3>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Your dashboards are ready to use. Click below to get started.
          </p>
        </div>
      ) : configuredPlugins.length === 0 ? (
        <div className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 text-center">
          <p className="text-gray-600 dark:text-gray-400 mb-4">
            No sources configured yet. You can configure sources later from the Plugins page.
          </p>
          <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
            We'll create your dashboards now based on your selected role.
          </p>
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
          disabled={isCreatingDashboards || dashboardsCreated}
          className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Back
        </button>
        <button
          onClick={handleFinish}
          disabled={isCreatingDashboards || !selectedRole}
          className="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 dark:bg-green-700 dark:hover:bg-green-800 transition-colors focus:ring-2 focus:ring-green-500 focus:ring-offset-2 dark:focus:ring-offset-gray-900 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {dashboardsCreated ? 'Get Started' : 'Finish Setup'}
        </button>
      </div>
    </div>
  );
}
