import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/client';

interface SprintDetailModalProps {
  sprint_id: string;
  isOpen: boolean;
  onClose: () => void;
}

export function SprintDetailModal({ sprint_id, isOpen, onClose }: SprintDetailModalProps) {
  const { data: sprint, isLoading: sprintLoading } = useQuery({
    queryKey: ['sprint', sprint_id],
    queryFn: () => api.getSprint(sprint_id),
    enabled: isOpen,
  });

  const { data: health, isLoading: healthLoading } = useQuery({
    queryKey: ['sprint-health', sprint_id],
    queryFn: () => api.getSprintHealth(sprint_id),
    enabled: isOpen,
  });

  const { data: burndown, isLoading: burndownLoading } = useQuery({
    queryKey: ['sprint-burndown', sprint_id],
    queryFn: () => api.getSprintBurndown(sprint_id),
    enabled: isOpen,
  });

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      return () => document.removeEventListener('keydown', handleEscape);
    }
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const isLoading = sprintLoading || healthLoading || burndownLoading;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="fixed inset-0 bg-black bg-opacity-50" onClick={onClose} />

      <div className="relative min-h-screen flex items-center justify-center p-4">
        <div className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] flex flex-col">
          <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
            <div>
              <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                {sprintLoading ? 'Loading...' : sprint?.name || 'Sprint Details'}
              </h2>
              {sprint && (
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                  {new Date(sprint.start_date).toLocaleDateString()} -{' '}
                  {new Date(sprint.end_date).toLocaleDateString()}
                </p>
              )}
            </div>
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

          <div className="flex-1 overflow-y-auto p-6">
            {isLoading ? (
              <LoadingSkeleton />
            ) : (
              <div className="space-y-6">
                {health && (
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
                      Sprint Health
                    </h3>
                    <div className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                      <div className="flex items-center justify-between mb-4">
                        <span className="text-gray-900 dark:text-gray-100 font-medium">Status</span>
                        <span
                          className={`px-3 py-1 rounded-full text-sm font-medium ${
                            health.risk_level === 'on_track'
                              ? 'bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-400'
                              : health.risk_level === 'at_risk'
                                ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400'
                                : 'bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-400'
                          }`}
                        >
                          {health.risk_level.replace('_', ' ').toUpperCase()}
                        </span>
                      </div>

                      <div className="mb-4">
                        <div className="flex items-center justify-between mb-2">
                          <span className="text-sm text-gray-600 dark:text-gray-400">Progress</span>
                          <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
                            {health.completion_percentage.toFixed(0)}%
                          </span>
                        </div>
                        <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3">
                          <div
                            className={`h-3 rounded-full ${
                              health.risk_level === 'on_track'
                                ? 'bg-green-600 dark:bg-green-500'
                                : health.risk_level === 'at_risk'
                                  ? 'bg-yellow-600 dark:bg-yellow-500'
                                  : 'bg-red-600 dark:bg-red-500'
                            }`}
                            style={{ width: `${health.completion_percentage}%` }}
                          />
                        </div>
                      </div>

                      <div className="grid grid-cols-3 gap-4">
                        <div>
                          <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">
                            Committed
                          </dt>
                          <dd className="mt-1 text-lg font-semibold text-gray-900 dark:text-gray-100">
                            {health.committed_points}
                          </dd>
                        </div>
                        <div>
                          <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">
                            Completed
                          </dt>
                          <dd className="mt-1 text-lg font-semibold text-gray-900 dark:text-gray-100">
                            {health.completed_points}
                          </dd>
                        </div>
                        <div>
                          <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">
                            Remaining
                          </dt>
                          <dd className="mt-1 text-lg font-semibold text-gray-900 dark:text-gray-100">
                            {health.remaining_points}
                          </dd>
                        </div>
                      </div>

                      {health.alerts && health.alerts.length > 0 && (
                        <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
                          <h4 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-2">
                            Alerts
                          </h4>
                          <ul className="space-y-1">
                            {health.alerts.map((alert, idx) => (
                              <li key={idx} className="text-sm text-red-600 dark:text-red-400">
                                {alert}
                              </li>
                            ))}
                          </ul>
                        </div>
                      )}
                    </div>
                  </div>
                )}

                {burndown && burndown.burndown_points && burndown.burndown_points.length > 0 && (
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
                      Burndown Chart
                    </h3>
                    <div className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                      <div className="h-64 flex items-end justify-between space-x-1">
                        {burndown.burndown_points.map((point, idx) => {
                          const maxPoints = burndown.total_points;
                          const idealHeight = (point.ideal / maxPoints) * 100;
                          const actualHeight = (point.actual / maxPoints) * 100;

                          return (
                            <div key={idx} className="flex-1 flex flex-col items-center">
                              <div className="w-full flex items-end justify-center h-48 relative">
                                <div
                                  className="w-1/3 bg-gray-300 dark:bg-gray-600 rounded-t opacity-50"
                                  style={{ height: `${idealHeight}%` }}
                                  title={`Ideal: ${point.ideal}`}
                                />
                                <div
                                  className="w-1/3 bg-blue-600 dark:bg-blue-500 rounded-t ml-1"
                                  style={{ height: `${actualHeight}%` }}
                                  title={`Actual: ${point.actual}`}
                                />
                              </div>
                              <div className="text-xs text-gray-500 dark:text-gray-400 mt-1 rotate-45 origin-left">
                                {new Date(point.date).toLocaleDateString('en-US', {
                                  month: 'short',
                                  day: 'numeric',
                                })}
                              </div>
                            </div>
                          );
                        })}
                      </div>
                      <div className="flex items-center justify-center mt-8 space-x-6">
                        <div className="flex items-center">
                          <div className="w-4 h-4 bg-gray-300 dark:bg-gray-600 opacity-50 rounded mr-2" />
                          <span className="text-sm text-gray-600 dark:text-gray-400">Ideal</span>
                        </div>
                        <div className="flex items-center">
                          <div className="w-4 h-4 bg-blue-600 dark:bg-blue-500 rounded mr-2" />
                          <span className="text-sm text-gray-600 dark:text-gray-400">Actual</span>
                        </div>
                      </div>
                    </div>
                  </div>
                )}

                {sprint && (
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
                      Sprint Information
                    </h3>
                    <dl className="grid grid-cols-2 gap-4">
                      <div className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                        <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">
                          Status
                        </dt>
                        <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100 capitalize">
                          {sprint.status}
                        </dd>
                      </div>
                      <div className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                        <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">
                          Duration
                        </dt>
                        <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100">
                          {Math.ceil(
                            (new Date(sprint.end_date).getTime() -
                              new Date(sprint.start_date).getTime()) /
                              (1000 * 60 * 60 * 24)
                          )}{' '}
                          days
                        </dd>
                      </div>
                    </dl>
                  </div>
                )}
              </div>
            )}
          </div>

          <div className="flex gap-3 justify-end p-6 border-t border-gray-200 dark:border-gray-700">
            <button
              onClick={onClose}
              className="px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

function LoadingSkeleton() {
  return (
    <div className="space-y-4">
      {[1, 2, 3].map((i) => (
        <div key={i} className="animate-pulse">
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/4 mb-2" />
          <div className="h-32 bg-gray-200 dark:bg-gray-700 rounded w-full" />
        </div>
      ))}
    </div>
  );
}
