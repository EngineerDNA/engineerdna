import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../../api/client';
import type { Event as EventType } from '../../api/types';

interface AlertDetailModalProps {
  alert_id: string;
  isOpen: boolean;
  onClose: () => void;
}

export function AlertDetailModal({ alert_id, isOpen, onClose }: AlertDetailModalProps) {
  const [snoozeDuration, setSnoozeDuration] = useState<number>(60);
  const queryClient = useQueryClient();

  const { data: alert, isLoading: alertLoading } = useQuery({
    queryKey: ['alert-instance', alert_id],
    queryFn: () => api.getAlertInstance(alert_id),
    enabled: isOpen,
  });

  const { data: rule, isLoading: ruleLoading } = useQuery({
    queryKey: ['alert-rule', alert?.rule_id],
    queryFn: () => api.getAlertRule(alert!.rule_id),
    enabled: isOpen && !!alert?.rule_id,
  });

  const { data: relatedEvents } = useQuery({
    queryKey: ['alert-events', alert?.entity_id],
    queryFn: () =>
      api.getEvents({
        limit: 10,
      }),
    enabled: isOpen && !!alert?.entity_id,
  });

  const acknowledgeMutation = useMutation({
    mutationFn: () => api.acknowledgeAlert(alert_id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alert-instance', alert_id] });
      queryClient.invalidateQueries({ queryKey: ['alerts'] });
    },
  });

  const snoozeMutation = useMutation({
    mutationFn: () => api.snoozeAlert(alert_id, snoozeDuration),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alert-instance', alert_id] });
      queryClient.invalidateQueries({ queryKey: ['alerts'] });
      onClose();
    },
  });

  const dismissMutation = useMutation({
    mutationFn: () => api.dismissAlert(alert_id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alert-instance', alert_id] });
      queryClient.invalidateQueries({ queryKey: ['alerts'] });
      onClose();
    },
  });

  const resolveMutation = useMutation({
    mutationFn: () => api.resolveAlert(alert_id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alert-instance', alert_id] });
      queryClient.invalidateQueries({ queryKey: ['alerts'] });
      onClose();
    },
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

  const isLoading = alertLoading || ruleLoading;
  const isActioning =
    acknowledgeMutation.isPending ||
    snoozeMutation.isPending ||
    dismissMutation.isPending ||
    resolveMutation.isPending;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="fixed inset-0 bg-black bg-opacity-50" onClick={onClose} />

      <div className="relative min-h-screen flex items-center justify-center p-4">
        <div className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-3xl w-full max-h-[90vh] flex flex-col">
          <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
            <div className="flex-1">
              <div className="flex items-center gap-3">
                <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                  {alertLoading ? 'Loading...' : alert?.title || 'Alert Details'}
                </h2>
                {alert && (
                  <span
                    className={`px-3 py-1 rounded-full text-sm font-medium ${
                      alert.severity === 'critical'
                        ? 'bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-400'
                        : alert.severity === 'warning'
                          ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400'
                          : 'bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400'
                    }`}
                  >
                    {alert.severity.toUpperCase()}
                  </span>
                )}
              </div>
            </div>
            <button
              onClick={onClose}
              className="ml-4 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
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
                {alert && (
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">
                      Alert Information
                    </h3>
                    <div className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg space-y-3">
                      <p className="text-gray-900 dark:text-gray-100">{alert.message}</p>

                      <div className="grid grid-cols-2 gap-4 pt-3 border-t border-gray-200 dark:border-gray-700">
                        <div>
                          <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">
                            Fired At
                          </dt>
                          <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100">
                            {new Date(alert.fired_at).toLocaleString()}
                          </dd>
                        </div>

                        {alert.acknowledged_at && (
                          <div>
                            <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">
                              Acknowledged At
                            </dt>
                            <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100">
                              {new Date(alert.acknowledged_at).toLocaleString()}
                            </dd>
                          </div>
                        )}

                        {alert.entity_type && (
                          <div>
                            <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">
                              Entity Type
                            </dt>
                            <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100 capitalize">
                              {alert.entity_type}
                            </dd>
                          </div>
                        )}

                        {alert.snoozed_until && (
                          <div>
                            <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">
                              Snoozed Until
                            </dt>
                            <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100">
                              {new Date(alert.snoozed_until).toLocaleString()}
                            </dd>
                          </div>
                        )}
                      </div>

                      {alert.context && Object.keys(alert.context).length > 0 && (
                        <div className="pt-3 border-t border-gray-200 dark:border-gray-700">
                          <dt className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">
                            Context
                          </dt>
                          <dd className="text-sm text-gray-900 dark:text-gray-100">
                            <pre className="bg-gray-50 dark:bg-gray-900 p-3 rounded overflow-x-auto">
                              {JSON.stringify(alert.context, null, 2)}
                            </pre>
                          </dd>
                        </div>
                      )}
                    </div>
                  </div>
                )}

                {rule && (
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">
                      Alert Rule
                    </h3>
                    <div className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                      <h4 className="font-medium text-gray-900 dark:text-gray-100 mb-2">
                        {rule.name}
                      </h4>
                      {rule.description && (
                        <p className="text-sm text-gray-600 dark:text-gray-400 mb-3">
                          {rule.description}
                        </p>
                      )}
                      <div className="grid grid-cols-2 gap-4">
                        <div>
                          <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">
                            Alert Type
                          </dt>
                          <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100">
                            {rule.alert_type}
                          </dd>
                        </div>
                        {rule.threshold_value && (
                          <div>
                            <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">
                              Threshold
                            </dt>
                            <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100">
                              {rule.threshold_operator} {rule.threshold_value}
                            </dd>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                )}

                {relatedEvents && relatedEvents.events && relatedEvents.events.length > 0 && (
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-3">
                      Related Events
                    </h3>
                    <div className="space-y-2">
                      {relatedEvents.events.slice(0, 5).map((event: EventType) => (
                        <div
                          key={event.id}
                          className="p-3 border border-gray-200 dark:border-gray-700 rounded-lg"
                        >
                          <div className="flex items-start justify-between">
                            <div>
                              <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                                {event.type}
                              </p>
                              <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">
                                by {event.actor} - {new Date(event.timestamp).toLocaleString()}
                              </p>
                            </div>
                            <span className="px-2 py-1 text-xs bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-400 rounded">
                              {event.source}
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>

          <div className="p-6 border-t border-gray-200 dark:border-gray-700">
            {alert && !alert.acknowledged_at && !alert.dismissed_at && !alert.resolved_at && (
              <div className="mb-4 p-4 bg-gray-50 dark:bg-gray-900 rounded-lg">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Snooze Duration (minutes)
                </label>
                <input
                  type="number"
                  min="5"
                  step="5"
                  value={snoozeDuration}
                  onChange={(e) => setSnoozeDuration(parseInt(e.target.value))}
                  className="w-32 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                />
              </div>
            )}

            <div className="flex gap-3 justify-end flex-wrap">
              {alert && !alert.acknowledged_at && !alert.dismissed_at && !alert.resolved_at && (
                <>
                  <button
                    onClick={() => acknowledgeMutation.mutate()}
                    disabled={isActioning}
                    className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 disabled:opacity-50 transition-colors"
                  >
                    Acknowledge
                  </button>
                  <button
                    onClick={() => snoozeMutation.mutate()}
                    disabled={isActioning}
                    className="px-4 py-2 bg-yellow-600 text-white rounded hover:bg-yellow-700 dark:bg-yellow-700 dark:hover:bg-yellow-800 disabled:opacity-50 transition-colors"
                  >
                    Snooze
                  </button>
                  <button
                    onClick={() => dismissMutation.mutate()}
                    disabled={isActioning}
                    className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 dark:bg-gray-700 dark:hover:bg-gray-800 disabled:opacity-50 transition-colors"
                  >
                    Dismiss
                  </button>
                  <button
                    onClick={() => resolveMutation.mutate()}
                    disabled={isActioning}
                    className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 dark:bg-green-700 dark:hover:bg-green-800 disabled:opacity-50 transition-colors"
                  >
                    Resolve
                  </button>
                </>
              )}
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
    </div>
  );
}

function LoadingSkeleton() {
  return (
    <div className="space-y-4">
      {[1, 2, 3].map((i) => (
        <div key={i} className="animate-pulse">
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/4 mb-2" />
          <div className="h-24 bg-gray-200 dark:bg-gray-700 rounded w-full" />
        </div>
      ))}
    </div>
  );
}
