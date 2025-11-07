import { useState } from 'react';
import type { AlertInstance } from '../api/types';
import {
  useAcknowledgeAlert,
  useSnoozeAlert,
  useDismissAlert,
  useResolveAlert,
} from '../hooks/useAlerts';

interface AlertCardProps {
  alert: AlertInstance;
}

export function AlertCard({ alert }: AlertCardProps) {
  const [expanded, setExpanded] = useState(false);
  const [snoozing, setSnoozing] = useState(false);

  const acknowledgeMutation = useAcknowledgeAlert();
  const snoozeMutation = useSnoozeAlert();
  const dismissMutation = useDismissAlert();
  const resolveMutation = useResolveAlert();

  const severityColors = {
    info: {
      bg: 'bg-blue-50 dark:bg-blue-900/20',
      border: 'border-blue-200 dark:border-blue-800',
      text: 'text-blue-800 dark:text-blue-200',
      icon: 'text-blue-600 dark:text-blue-400',
      badge: 'bg-blue-500 dark:bg-blue-600',
    },
    warning: {
      bg: 'bg-yellow-50 dark:bg-yellow-900/20',
      border: 'border-yellow-200 dark:border-yellow-800',
      text: 'text-yellow-800 dark:text-yellow-200',
      icon: 'text-yellow-600 dark:text-yellow-400',
      badge: 'bg-yellow-500 dark:bg-yellow-600',
    },
    critical: {
      bg: 'bg-red-50 dark:bg-red-900/20',
      border: 'border-red-200 dark:border-red-800',
      text: 'text-red-800 dark:text-red-200',
      icon: 'text-red-600 dark:text-red-400',
      badge: 'bg-red-500 dark:bg-red-600',
    },
  };

  const colors = severityColors[alert.severity];

  const getSeverityIcon = () => {
    if (alert.severity === 'critical') {
      return (
        <svg
          className={`w-6 h-6 ${colors.icon}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
      );
    }
    if (alert.severity === 'warning') {
      return (
        <svg
          className={`w-6 h-6 ${colors.icon}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
      );
    }
    return (
      <svg
        className={`w-6 h-6 ${colors.icon}`}
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
    );
  };

  const getStatusBadge = () => {
    if (alert.resolved_at) {
      return (
        <span className="px-2 py-0.5 text-xs font-medium bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-200 rounded-full">
          Resolved
        </span>
      );
    }
    if (alert.dismissed_at) {
      return (
        <span className="px-2 py-0.5 text-xs font-medium bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-200 rounded-full">
          Dismissed
        </span>
      );
    }
    if (alert.snoozed_until) {
      const snoozedUntil = new Date(alert.snoozed_until);
      if (snoozedUntil > new Date()) {
        return (
          <span className="px-2 py-0.5 text-xs font-medium bg-purple-100 dark:bg-purple-900/30 text-purple-800 dark:text-purple-200 rounded-full">
            Snoozed
          </span>
        );
      }
    }
    if (alert.acknowledged_at) {
      return (
        <span className="px-2 py-0.5 text-xs font-medium bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-200 rounded-full">
          Acknowledged
        </span>
      );
    }
    return (
      <span className="px-2 py-0.5 text-xs font-medium bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-200 rounded-full">
        Active
      </span>
    );
  };

  const handleSnooze = (minutes: number) => {
    snoozeMutation.mutate({ alertId: alert.id, duration: minutes });
    setSnoozing(false);
  };

  const isActionable = !alert.resolved_at && !alert.dismissed_at;

  return (
    <div className={`${colors.bg} ${colors.border} border rounded-lg p-4`}>
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0 mt-0.5">{getSeverityIcon()}</div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-2 flex-wrap">
            <h3 className={`text-lg font-semibold ${colors.text}`}>{alert.title}</h3>
            <span
              className={`px-2 py-0.5 text-xs font-medium text-white ${colors.badge} rounded-full`}
            >
              {alert.severity.toUpperCase()}
            </span>
            {getStatusBadge()}
          </div>

          <p className={`text-sm ${colors.text} mb-3`}>{alert.message}</p>

          <div className="text-xs text-gray-600 dark:text-gray-400 mb-3">
            Fired: {new Date(alert.fired_at).toLocaleString()}
            {alert.entity_type && (
              <span className="ml-3">
                Entity: {alert.entity_type} {alert.entity_id && `(${alert.entity_id})`}
              </span>
            )}
          </div>

          {isActionable && (
            <div className="flex items-center gap-2 mb-3 flex-wrap">
              {!alert.acknowledged_at && (
                <button
                  onClick={() => acknowledgeMutation.mutate(alert.id)}
                  disabled={acknowledgeMutation.isPending}
                  className="px-3 py-1.5 text-sm bg-blue-600 dark:bg-blue-700 text-white rounded hover:bg-blue-700 dark:hover:bg-blue-800 disabled:opacity-50 transition-colors"
                >
                  Acknowledge
                </button>
              )}
              <button
                onClick={() => setSnoozing(!snoozing)}
                className="px-3 py-1.5 text-sm bg-purple-600 dark:bg-purple-700 text-white rounded hover:bg-purple-700 dark:hover:bg-purple-800 transition-colors"
              >
                Snooze
              </button>
              <button
                onClick={() => dismissMutation.mutate(alert.id)}
                disabled={dismissMutation.isPending}
                className="px-3 py-1.5 text-sm bg-gray-600 dark:bg-gray-700 text-white rounded hover:bg-gray-700 dark:hover:bg-gray-800 disabled:opacity-50 transition-colors"
              >
                Dismiss
              </button>
              <button
                onClick={() => resolveMutation.mutate(alert.id)}
                disabled={resolveMutation.isPending}
                className="px-3 py-1.5 text-sm bg-green-600 dark:bg-green-700 text-white rounded hover:bg-green-700 dark:hover:bg-green-800 disabled:opacity-50 transition-colors"
              >
                Resolve
              </button>
            </div>
          )}

          {snoozing && isActionable && (
            <div className="mb-3 p-3 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded">
              <div className="text-sm font-medium text-gray-900 dark:text-gray-100 mb-2">
                Snooze for:
              </div>
              <div className="flex gap-2 flex-wrap">
                <button
                  onClick={() => handleSnooze(15)}
                  className="px-3 py-1.5 text-sm bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
                >
                  15 min
                </button>
                <button
                  onClick={() => handleSnooze(30)}
                  className="px-3 py-1.5 text-sm bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
                >
                  30 min
                </button>
                <button
                  onClick={() => handleSnooze(60)}
                  className="px-3 py-1.5 text-sm bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
                >
                  1 hour
                </button>
                <button
                  onClick={() => handleSnooze(240)}
                  className="px-3 py-1.5 text-sm bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
                >
                  4 hours
                </button>
              </div>
            </div>
          )}

          {alert.context && Object.keys(alert.context).length > 0 && (
            <button
              onClick={() => setExpanded(!expanded)}
              className="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 font-medium transition-colors"
            >
              {expanded ? 'Hide context' : 'Show context'}
            </button>
          )}

          {expanded && alert.context && (
            <div className="mt-3 pt-3 border-t border-gray-300 dark:border-gray-600">
              <div className="text-xs font-medium text-gray-600 dark:text-gray-400 mb-2">
                CONTEXT
              </div>
              <pre className="text-xs text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 p-2 rounded overflow-x-auto">
                {JSON.stringify(alert.context, null, 2)}
              </pre>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
