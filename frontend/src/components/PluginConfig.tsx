import { useState } from 'react';
import type { Plugin } from '../api/types';
import { useConfigurePlugin } from '../hooks/usePlugins';

interface PluginConfigProps {
  plugin: Plugin;
  onClose: () => void;
}

export function PluginConfig({ plugin, onClose }: PluginConfigProps) {
  const [config, setConfig] = useState<Record<string, string>>({});
  const configureMutation = useConfigurePlugin();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await configureMutation.mutateAsync({ name: plugin.name, config });
      onClose();
    } catch (error) {
      console.error('Failed to configure plugin:', error);
    }
  };

  const fields = plugin.config_fields || [];

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 p-8 rounded-lg shadow-lg max-w-md w-full max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            Configure {plugin.name}
          </h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
            aria-label="Close"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        {fields.length === 0 ? (
          <div className="text-center py-8">
            <p className="text-gray-600 dark:text-gray-400 mb-4">
              No configuration fields available for this plugin.
            </p>
            <button
              onClick={onClose}
              className="px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
            >
              Close
            </button>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            {fields.map((field) => (
              <div key={field.name}>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  {field.description}
                  {field.required && <span className="text-red-500 ml-1">*</span>}
                </label>
                {field.type === 'select' && field.options ? (
                  <select
                    value={config[field.name] || field.default || ''}
                    onChange={(e) => setConfig({ ...config, [field.name]: e.target.value })}
                    required={field.required}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  >
                    <option value="">Select an option</option>
                    {field.options.map((option) => (
                      <option key={option} value={option}>
                        {option}
                      </option>
                    ))}
                  </select>
                ) : (
                  <input
                    type={field.secret ? 'password' : 'text'}
                    value={config[field.name] || field.default || ''}
                    onChange={(e) => setConfig({ ...config, [field.name]: e.target.value })}
                    required={field.required}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    placeholder={field.secret ? 'Enter value' : ''}
                  />
                )}
                {field.secret && (
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    This value will be encrypted
                  </p>
                )}
              </div>
            ))}

            {configureMutation.isError && (
              <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded">
                <p className="text-sm text-red-800 dark:text-red-200">
                  Failed to configure plugin. Please check your settings and try again.
                </p>
              </div>
            )}

            <div className="flex gap-2 pt-4">
              <button
                type="submit"
                disabled={configureMutation.isPending}
                className="flex-1 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800"
              >
                {configureMutation.isPending ? 'Saving...' : 'Save Configuration'}
              </button>
              <button
                type="button"
                onClick={onClose}
                disabled={configureMutation.isPending}
                className="px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded hover:bg-gray-300 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                Cancel
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}
