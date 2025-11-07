import { useState, useMemo } from 'react';
import { Link } from 'react-router-dom';
import { useAlerts } from '../hooks/useAlerts';
import { AlertCard } from '../components/alert-card';

type StatusFilter = 'active' | 'acknowledged' | 'snoozed' | 'dismissed' | 'resolved';
type SeverityFilter = 'all' | 'info' | 'warning' | 'critical';

export function AlertsPage() {
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('active');
  const [severityFilter, setSeverityFilter] = useState<SeverityFilter>('all');

  const { data, isLoading, error } = useAlerts(undefined, { refetchInterval: 30000 });

  const alerts = data?.alerts || [];

  const filteredAlerts = useMemo(() => {
    return alerts.filter((alert) => {
      const severityMatch = severityFilter === 'all' || alert.severity === severityFilter;

      let statusMatch = false;
      switch (statusFilter) {
        case 'active':
          statusMatch =
            !alert.resolved_at &&
            !alert.dismissed_at &&
            !alert.acknowledged_at &&
            (!alert.snoozed_until || new Date(alert.snoozed_until) <= new Date());
          break;
        case 'acknowledged':
          statusMatch = !!alert.acknowledged_at && !alert.resolved_at && !alert.dismissed_at;
          break;
        case 'snoozed':
          statusMatch = !!alert.snoozed_until && new Date(alert.snoozed_until) > new Date();
          break;
        case 'dismissed':
          statusMatch = !!alert.dismissed_at;
          break;
        case 'resolved':
          statusMatch = !!alert.resolved_at;
          break;
      }

      return severityMatch && statusMatch;
    });
  }, [alerts, statusFilter, severityFilter]);

  const statusCounts = useMemo(() => {
    const counts = {
      active: 0,
      acknowledged: 0,
      snoozed: 0,
      dismissed: 0,
      resolved: 0,
    };

    alerts.forEach((alert) => {
      if (alert.resolved_at) {
        counts.resolved++;
      } else if (alert.dismissed_at) {
        counts.dismissed++;
      } else if (alert.snoozed_until && new Date(alert.snoozed_until) > new Date()) {
        counts.snoozed++;
      } else if (alert.acknowledged_at) {
        counts.acknowledged++;
      } else {
        counts.active++;
      }
    });

    return counts;
  }, [alerts]);

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded">
        <p className="text-red-800 dark:text-red-200">
          Failed to load alerts: {(error as Error).message}
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Alerts</h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Real-time alert monitoring and management
          </p>
        </div>
        <Link
          to="/alerts/rules"
          className="px-4 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded hover:bg-blue-700 dark:hover:bg-blue-800 transition-colors"
        >
          Manage Rules
        </Link>
      </div>

      <div className="flex items-center gap-4 flex-wrap">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Status
          </label>
          <div className="flex gap-2">
            {(['active', 'acknowledged', 'snoozed', 'dismissed', 'resolved'] as StatusFilter[]).map(
              (status) => (
                <button
                  key={status}
                  onClick={() => setStatusFilter(status)}
                  className={`px-3 py-1.5 text-sm rounded transition-colors ${
                    statusFilter === status
                      ? 'bg-blue-600 dark:bg-blue-700 text-white'
                      : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'
                  }`}
                >
                  {status.charAt(0).toUpperCase() + status.slice(1)} ({statusCounts[status]})
                </button>
              )
            )}
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Severity
          </label>
          <select
            value={severityFilter}
            onChange={(e) => setSeverityFilter(e.target.value as SeverityFilter)}
            className="px-3 py-1.5 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400"
          >
            <option value="all">All Severities</option>
            <option value="info">Info</option>
            <option value="warning">Warning</option>
            <option value="critical">Critical</option>
          </select>
        </div>
      </div>

      {filteredAlerts.length === 0 ? (
        <div className="text-center py-12 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
          <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100 mb-4">No Alerts</h2>
          <p className="text-gray-600 dark:text-gray-400">No alerts match the current filters.</p>
        </div>
      ) : (
        <div className="space-y-4">
          {filteredAlerts.map((alert) => (
            <AlertCard key={alert.id} alert={alert} />
          ))}
        </div>
      )}
    </div>
  );
}
