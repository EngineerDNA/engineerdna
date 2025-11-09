import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/client';

interface EventDetailDrawerProps {
  event_id: string;
  isOpen: boolean;
  onClose: () => void;
}

export function EventDetailDrawer({ event_id, isOpen, onClose }: EventDetailDrawerProps) {
  const { data: eventsData } = useQuery({
    queryKey: ['events'],
    queryFn: () => api.getEvents({ limit: 1000 }),
    enabled: isOpen,
  });

  const event = eventsData?.events.find((e) => e.id === event_id);

  const { data: relatedEvents } = useQuery({
    queryKey: ['related-events', event?.actor, event?.source],
    queryFn: () =>
      api.getEvents({
        limit: 10,
      }),
    enabled: isOpen && !!event,
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

  return (
    <>
      {isOpen && <div className="fixed inset-0 bg-black bg-opacity-50 z-40" onClick={onClose} />}

      <div
        className={`fixed right-0 top-0 h-full w-full sm:w-[32rem] bg-white dark:bg-gray-800 shadow-xl transform transition-transform duration-300 z-50 ${
          isOpen ? 'translate-x-0' : 'translate-x-full'
        }`}
      >
        <div className="flex flex-col h-full">
          <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
            <div className="flex-1 min-w-0">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Event Details</h2>
              {event && (
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1 truncate">
                  {event.type}
                </p>
              )}
            </div>
            <button
              onClick={onClose}
              className="ml-4 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors flex-shrink-0"
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
            {!event ? (
              <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                Event not found
              </div>
            ) : (
              <div className="space-y-6">
                <div>
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
                    Event Information
                  </h3>
                  <div className="space-y-4">
                    <InfoRow label="Event Type" value={event.type} />
                    <InfoRow label="Source" value={event.source} />
                    <InfoRow label="Source ID" value={event.source_id} />
                    <InfoRow label="Actor" value={event.actor} />
                    <InfoRow label="Timestamp" value={new Date(event.timestamp).toLocaleString()} />
                    <InfoRow
                      label="Anonymized"
                      value={event.anonymized ? 'Yes' : 'No'}
                      badge={event.anonymized ? 'yes' : 'no'}
                    />
                  </div>
                </div>

                {event.data && Object.keys(event.data).length > 0 && (
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
                      Event Payload
                    </h3>
                    <div className="bg-gray-50 dark:bg-gray-900 p-4 rounded-lg overflow-x-auto">
                      <pre className="text-sm text-gray-900 dark:text-gray-100">
                        {JSON.stringify(event.data, null, 2)}
                      </pre>
                    </div>
                  </div>
                )}

                {relatedEvents && relatedEvents.events && relatedEvents.events.length > 0 && (
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
                      Related Events
                    </h3>
                    <div className="space-y-3">
                      {relatedEvents.events
                        .filter((e) => e.id !== event_id)
                        .slice(0, 5)
                        .map((relatedEvent) => (
                          <div
                            key={relatedEvent.id}
                            className="p-3 border border-gray-200 dark:border-gray-700 rounded-lg"
                          >
                            <div className="flex items-start justify-between">
                              <div className="flex-1 min-w-0">
                                <p className="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
                                  {relatedEvent.type}
                                </p>
                                <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">
                                  by {relatedEvent.actor}
                                </p>
                                <p className="text-xs text-gray-500 dark:text-gray-500">
                                  {new Date(relatedEvent.timestamp).toLocaleString()}
                                </p>
                              </div>
                              <span className="ml-2 px-2 py-1 text-xs bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400 rounded flex-shrink-0">
                                {relatedEvent.source}
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
    </>
  );
}

function InfoRow({ label, value, badge }: { label: string; value: string; badge?: 'yes' | 'no' }) {
  return (
    <div className="flex items-start justify-between py-3 border-b border-gray-200 dark:border-gray-700 last:border-b-0">
      <dt className="text-sm font-medium text-gray-500 dark:text-gray-400 w-1/3">{label}</dt>
      <dd className="text-sm text-gray-900 dark:text-gray-100 w-2/3 text-right">
        {badge ? (
          <span
            className={`px-2 py-1 rounded text-xs ${
              badge === 'yes'
                ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400'
                : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-400'
            }`}
          >
            {value}
          </span>
        ) : (
          <span className="break-words">{value}</span>
        )}
      </dd>
    </div>
  );
}
