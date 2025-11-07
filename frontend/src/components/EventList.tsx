import { memo, useState } from 'react';
import type { Event } from '../api/types';

interface EventListProps {
  events: Event[];
}

export const EventList = memo(function EventList({ events }: EventListProps) {
  const [expandedEvents, setExpandedEvents] = useState<Set<string>>(new Set());

  const toggleEvent = (eventId: string) => {
    setExpandedEvents((prev) => {
      const next = new Set(prev);
      if (next.has(eventId)) {
        next.delete(eventId);
      } else {
        next.add(eventId);
      }
      return next;
    });
  };

  if (events.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500 dark:text-gray-400">No events to display</div>
    );
  }

  return (
    <div className="space-y-2">
      {events.map((event) => {
        const isExpanded = expandedEvents.has(event.id);
        return (
          <div
            key={event.id}
            className="border-l-4 border-blue-500 dark:border-blue-400 pl-4 py-2 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors rounded-r cursor-pointer"
            onClick={() => toggleEvent(event.id)}
            role="button"
            tabIndex={0}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                toggleEvent(event.id);
              }
            }}
          >
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1">
              <div className="flex items-center gap-2">
                <span className="text-gray-600 dark:text-gray-400 font-mono text-sm">
                  {isExpanded ? '▼' : '▶'}
                </span>
                <div className="font-medium text-gray-900 dark:text-gray-100">{event.actor}</div>
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-400">
                {new Date(event.timestamp).toLocaleString()}
              </div>
            </div>
            <div className="text-sm text-gray-700 dark:text-gray-300 mt-1 ml-6">
              <span className="font-medium capitalize">{event.type.replace('_', ' ')}</span>
              {event.data.title ? `: ${String(event.data.title)}` : null}
              {event.data.action && !event.data.title ? `: ${String(event.data.action)}` : null}
            </div>
            <div className="flex flex-wrap gap-2 mt-1 ml-6">
              <span className="text-xs px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded">
                {event.source}
              </span>
              {event.anonymized && (
                <span className="text-xs px-2 py-1 bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-200 rounded">
                  Anonymized
                </span>
              )}
            </div>

            {isExpanded && (
              <div
                className="mt-3 ml-6 p-3 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded text-xs overflow-x-auto"
                onClick={(e) => e.stopPropagation()}
              >
                <div className="font-semibold text-gray-700 dark:text-gray-300 mb-2">
                  Event Metadata
                </div>
                <pre className="text-gray-800 dark:text-gray-200 whitespace-pre-wrap break-words">
                  {JSON.stringify(event.data, null, 2)}
                </pre>
                <div className="mt-2 pt-2 border-t border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-400">
                  <div>
                    <span className="font-semibold">ID:</span> {event.id}
                  </div>
                  <div>
                    <span className="font-semibold">Source ID:</span> {event.source_id}
                  </div>
                </div>
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
});
