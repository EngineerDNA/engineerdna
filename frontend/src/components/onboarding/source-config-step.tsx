import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/client';
import type { Plugin } from '../../api/types';
import { PluginConfig } from '../PluginConfig';

interface SourceConfigStepProps {
  onNext: () => void;
  onSkip: () => void;
  onBack: () => void;
}

export function SourceConfigStep({ onNext, onSkip, onBack }: SourceConfigStepProps) {
  const [selectedPlugin, setSelectedPlugin] = useState<Plugin | null>(null);

  const { data: pluginsData, isLoading } = useQuery({
    queryKey: ['plugins'],
    queryFn: api.getPlugins,
  });

  const sourcePlugins = pluginsData?.plugins.filter((p) => p.type === 'source') || [];

  const handleConfigurePlugin = (plugin: Plugin) => {
    setSelectedPlugin(plugin);
  };

  const handleCloseConfig = () => {
    setSelectedPlugin(null);
  };

  return (
    <>
      <div className="space-y-6">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">
            Configure Data Sources
          </h2>
          <p className="text-gray-600 dark:text-gray-400">
            Connect to GitHub, Jira, or other tools to sync engineering data
          </p>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 dark:border-blue-400" />
          </div>
        ) : sourcePlugins.length === 0 ? (
          <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4">
            <p className="text-sm text-yellow-800 dark:text-yellow-200">
              No source plugins found. You can add plugins later from the Plugins page.
            </p>
          </div>
        ) : (
          <div className="space-y-4">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Select a source to configure (you can add more later):
            </p>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {sourcePlugins.map((plugin) => (
                <div
                  key={plugin.name}
                  className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:border-blue-500 dark:hover:border-blue-500 transition-colors cursor-pointer"
                  onClick={() => handleConfigurePlugin(plugin)}
                >
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-medium text-gray-900 dark:text-gray-100 capitalize">
                      {plugin.name}
                    </h3>
                    {plugin.enabled ? (
                      <span className="px-2 py-1 bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-200 text-xs rounded font-medium">
                        Configured
                      </span>
                    ) : (
                      <span className="px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 text-xs rounded font-medium">
                        Not configured
                      </span>
                    )}
                  </div>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    Click to configure this source
                  </p>
                </div>
              ))}
            </div>
          </div>
        )}

        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
          <h4 className="text-sm font-medium text-gray-900 dark:text-gray-100 mb-2">
            Don't worry if you're not ready
          </h4>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            You can configure data sources later from the Plugins page. This step is optional.
          </p>
        </div>

        <div className="flex justify-between pt-4 border-t border-gray-200 dark:border-gray-700">
          <button
            onClick={onBack}
            className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100 transition-colors"
          >
            Back
          </button>
          <div className="flex gap-3">
            <button
              onClick={onSkip}
              className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100 transition-colors"
            >
              Skip this step
            </button>
            <button
              onClick={onNext}
              className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 transition-colors focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-900"
            >
              Continue
            </button>
          </div>
        </div>
      </div>

      {selectedPlugin && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <h3 className="text-xl font-bold text-gray-900 dark:text-gray-100 mb-4">
              Configure {selectedPlugin.name}
            </h3>
            <PluginConfig plugin={selectedPlugin} onClose={handleCloseConfig} />
          </div>
        </div>
      )}
    </>
  );
}
