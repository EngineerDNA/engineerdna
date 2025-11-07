import { useState } from 'react';
import type { AlertRule } from '../api/types';
import { useUpdateAlertRule, useDeleteAlertRule } from '../hooks/useAlertRules';

interface AlertRuleCardProps {
  rule: AlertRule;
  onEdit?: (rule: AlertRule) => void;
}

export function AlertRuleCard({ rule, onEdit }: AlertRuleCardProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [editedThreshold, setEditedThreshold] = useState(rule.threshold_value?.toString() || '');

  const updateMutation = useUpdateAlertRule();
  const deleteMutation = useDeleteAlertRule();

  const severityColors = {
    info: 'text-blue-600 dark:text-blue-400 bg-blue-100 dark:bg-blue-900/30',
    warning: 'text-yellow-600 dark:text-yellow-400 bg-yellow-100 dark:bg-yellow-900/30',
    critical: 'text-red-600 dark:text-red-400 bg-red-100 dark:bg-red-900/30',
  };

  const handleToggleEnabled = () => {
    updateMutation.mutate({ id: rule.id, data: { enabled: !rule.enabled } });
  };

  const handleSaveThreshold = () => {
    const threshold = parseFloat(editedThreshold);
    if (!isNaN(threshold)) {
      updateMutation.mutate({ id: rule.id, data: { threshold_value: threshold } });
      setIsEditing(false);
    }
  };

  const handleDelete = () => {
    if (confirm(`Delete alert rule "${rule.name}"?`)) {
      deleteMutation.mutate(rule.id);
    }
  };

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
      <div className="flex items-start justify-between mb-3">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-1">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">{rule.name}</h3>
            <span
              className={`px-2 py-0.5 text-xs font-medium rounded-full ${severityColors[rule.severity]}`}
            >
              {rule.severity.toUpperCase()}
            </span>
            {rule.enabled ? (
              <span className="px-2 py-0.5 text-xs font-medium bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-200 rounded-full">
                Enabled
              </span>
            ) : (
              <span className="px-2 py-0.5 text-xs font-medium bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-200 rounded-full">
                Disabled
              </span>
            )}
          </div>
          {rule.description && (
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">{rule.description}</p>
          )}
          <div className="text-xs text-gray-500 dark:text-gray-500">
            Type: {rule.alert_type}
            {rule.target_entity && <span className="ml-3">Target: {rule.target_entity}</span>}
          </div>
        </div>

        <div className="flex items-center gap-2 ml-4">
          <label className="relative inline-flex items-center cursor-pointer">
            <input
              type="checkbox"
              checked={rule.enabled}
              onChange={handleToggleEnabled}
              disabled={updateMutation.isPending}
              className="sr-only peer"
            />
            <div className="w-11 h-6 bg-gray-200 dark:bg-gray-700 peer-focus:ring-2 peer-focus:ring-blue-500 dark:peer-focus:ring-blue-400 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600 dark:peer-checked:bg-blue-500"></div>
          </label>
        </div>
      </div>

      {rule.threshold_value !== undefined && rule.threshold_operator && (
        <div className="mb-3 p-3 bg-gray-50 dark:bg-gray-900/50 rounded border border-gray-200 dark:border-gray-700">
          <div className="text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">THRESHOLD</div>
          {isEditing ? (
            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-700 dark:text-gray-300">
                {rule.threshold_operator}
              </span>
              <input
                type="number"
                value={editedThreshold}
                onChange={(e) => setEditedThreshold(e.target.value)}
                className="px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded w-24"
              />
              <button
                onClick={handleSaveThreshold}
                className="px-2 py-1 text-xs bg-blue-600 dark:bg-blue-700 text-white rounded hover:bg-blue-700 dark:hover:bg-blue-800 transition-colors"
              >
                Save
              </button>
              <button
                onClick={() => {
                  setIsEditing(false);
                  setEditedThreshold(rule.threshold_value?.toString() || '');
                }}
                className="px-2 py-1 text-xs bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
              >
                Cancel
              </button>
            </div>
          ) : (
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
                {rule.threshold_operator} {rule.threshold_value}
              </span>
              <button
                onClick={() => setIsEditing(true)}
                className="text-xs text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 transition-colors"
              >
                Edit
              </button>
            </div>
          )}
        </div>
      )}

      <div className="flex items-center gap-2">
        {onEdit && (
          <button
            onClick={() => onEdit(rule)}
            className="px-3 py-1.5 text-sm bg-blue-600 dark:bg-blue-700 text-white rounded hover:bg-blue-700 dark:hover:bg-blue-800 transition-colors"
          >
            Configure
          </button>
        )}
        <button
          onClick={handleDelete}
          disabled={deleteMutation.isPending}
          className="px-3 py-1.5 text-sm bg-red-600 dark:bg-red-700 text-white rounded hover:bg-red-700 dark:hover:bg-red-800 disabled:opacity-50 transition-colors"
        >
          Delete
        </button>
      </div>
    </div>
  );
}
