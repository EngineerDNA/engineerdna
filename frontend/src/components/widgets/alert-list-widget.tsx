import { useState } from 'react';
import {
  useAlerts,
  useAcknowledgeAlert,
  useSnoozeAlert,
  useDismissAlert,
} from '../../hooks/useAlerts';
import type { WidgetProps } from './widget-registry';
import type { AlertInstance } from '../../api/types';

export function AlertListWidget({ config }: WidgetProps) {
  const teamId = config.team_id as string | undefined;
  const severityFilter = config.severity_filter as string | undefined;

  const [localSeverityFilter, setLocalSeverityFilter] = useState(severityFilter || '');

  const filters: Record<string, string> = {};
  if (teamId) filters.entity_id = teamId;
  if (localSeverityFilter) filters.severity = localSeverityFilter;

  const { data, isLoading } = useAlerts(filters, { refetchInterval: 30000 });
  const acknowledgeMutation = useAcknowledgeAlert();
  const snoozeMutation = useSnoozeAlert();
  const dismissMutation = useDismissAlert();

  const alerts = data?.alerts || [];

  const getSeverityColor = (severity: AlertInstance['severity']) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-100 dark:bg-red-900/20 border-red-300 dark:border-red-800 text-red-900 dark:text-red-200';
      case 'warning':
        return 'bg-yellow-100 dark:bg-yellow-900/20 border-yellow-300 dark:border-yellow-800 text-yellow-900 dark:text-yellow-200';
      default:
        return 'bg-blue-100 dark:bg-blue-900/20 border-blue-300 dark:border-blue-800 text-blue-900 dark:text-blue-200';
    }
  };

  const getSeverityIcon = (severity: AlertInstance['severity']) => {
    if (severity === 'critical') {
      return (
        <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
          <path
            fillRule="evenodd"
            d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
            clipRule="evenodd"
          />
        </svg>
      );
    }
    if (severity === 'warning') {
      return (
        <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
          <path
            fillRule="evenodd"
            d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
            clipRule="evenodd"
          />
        </svg>
      );
    }
    return (
      <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
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
          {[...Array(3)].map((_, i) => (
            <div key={i} className="h-20 bg-gray-300 dark:bg-gray-700 rounded" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full flex flex-col">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400">Active Alerts</h3>
        <select
          value={localSeverityFilter}
          onChange={(e) => setLocalSeverityFilter(e.target.value)}
          className="text-xs px-2 py-1 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100"
        >
          <option value="">All Severities</option>
          <option value="critical">Critical</option>
          <option value="warning">Warning</option>
          <option value="info">Info</option>
        </select>
      </div>

      <div className="flex-1 overflow-auto space-y-3">
        {alerts.length === 0 ? (
          <div className="flex items-center justify-center h-32 text-gray-500 dark:text-gray-400">
            No active alerts
          </div>
        ) : (
          alerts.map((alert) => (
            <div
              key={alert.id}
              className={`border rounded-lg p-3 ${getSeverityColor(alert.severity)}`}
            >
              <div className="flex items-start gap-3 mb-2">
                <div className="flex-shrink-0 mt-0.5">{getSeverityIcon(alert.severity)}</div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium mb-1">{alert.title}</p>
                  <p className="text-xs mb-2">{alert.message}</p>
                  <p className="text-xs opacity-75">
                    Fired {new Date(alert.fired_at).toLocaleString()}
                  </p>
                  {alert.acknowledged_at && (
                    <p className="text-xs opacity-75 mt-1">
                      Acknowledged {new Date(alert.acknowledged_at).toLocaleString()}
                    </p>
                  )}
                  {alert.snoozed_until && (
                    <p className="text-xs opacity-75 mt-1">
                      Snoozed until {new Date(alert.snoozed_until).toLocaleString()}
                    </p>
                  )}
                </div>
              </div>

              <div className="flex items-center gap-2 flex-wrap">
                {!alert.acknowledged_at && (
                  <button
                    onClick={() => acknowledgeMutation.mutate(alert.id)}
                    disabled={acknowledgeMutation.isPending}
                    className="text-xs px-3 py-1 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-900 disabled:opacity-50"
                  >
                    Acknowledge
                  </button>
                )}
                {!alert.snoozed_until && (
                  <button
                    onClick={() => snoozeMutation.mutate({ alertId: alert.id, duration: 60 })}
                    disabled={snoozeMutation.isPending}
                    className="text-xs px-3 py-1 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-900 disabled:opacity-50"
                  >
                    Snooze 1h
                  </button>
                )}
                <button
                  onClick={() => dismissMutation.mutate(alert.id)}
                  disabled={dismissMutation.isPending}
                  className="text-xs px-3 py-1 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-900 disabled:opacity-50"
                >
                  Dismiss
                </button>
              </div>
            </div>
          ))
        )}
      </div>

      {alerts.length > 0 && (
        <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700 text-xs text-gray-500 dark:text-gray-400 text-center">
          {alerts.length} active alert{alerts.length > 1 ? 's' : ''}
        </div>
      )}
    </div>
  );
}
