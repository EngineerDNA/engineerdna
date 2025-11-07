import { useState, useMemo } from 'react';
import type { Event } from '../api/types';

interface ActivityFeedProps {
  events: Event[];
}

interface GroupedEvents {
  morning: Event[];
  afternoon: Event[];
  evening: Event[];
}

export function ActivityFeed({ events }: ActivityFeedProps) {
  const [filterType, setFilterType] = useState<string>('all');

  const eventTypes = useMemo(() => {
    const types = new Set(events.map((e) => e.type));
    return ['all', ...Array.from(types)];
  }, [events]);

  const filteredEvents = useMemo(() => {
    if (filterType === 'all') return events;
    return events.filter((e) => e.type === filterType);
  }, [events, filterType]);

  const groupedEvents = useMemo(() => {
    const groups: GroupedEvents = {
      morning: [],
      afternoon: [],
      evening: [],
    };

    filteredEvents.forEach((event) => {
      const hour = new Date(event.timestamp).getHours();
      if (hour < 12) {
        groups.morning.push(event);
      } else if (hour < 18) {
        groups.afternoon.push(event);
      } else {
        groups.evening.push(event);
      }
    });

    return groups;
  }, [filteredEvents]);

  const getEventIcon = (type: string) => {
    if (type.includes('pull_request')) {
      return (
        <svg
          className="w-5 h-5 text-blue-600 dark:text-blue-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"
          />
        </svg>
      );
    }
    if (type.includes('issue')) {
      return (
        <svg
          className="w-5 h-5 text-yellow-600 dark:text-yellow-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      );
    }
    if (type.includes('commit')) {
      return (
        <svg
          className="w-5 h-5 text-green-600 dark:text-green-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      );
    }
    if (type.includes('review')) {
      return (
        <svg
          className="w-5 h-5 text-purple-600 dark:text-purple-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
          />
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
          />
        </svg>
      );
    }
    return (
      <svg
        className="w-5 h-5 text-gray-600 dark:text-gray-400"
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

  const renderEventGroup = (title: string, events: Event[]) => {
    if (events.length === 0) return null;

    return (
      <div className="mb-6">
        <h3 className="text-sm font-semibold text-gray-600 dark:text-gray-400 mb-3 uppercase">
          {title}
        </h3>
        <div className="space-y-2">
          {events.map((event) => (
            <div
              key={event.id}
              className="flex items-start gap-3 p-3 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded hover:bg-gray-50 dark:hover:bg-gray-750 transition-colors"
            >
              <div className="flex-shrink-0 mt-0.5">{getEventIcon(event.type)}</div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
                    {event.type}
                  </span>
                  <span className="text-xs text-gray-500 dark:text-gray-500">
                    {new Date(event.timestamp).toLocaleTimeString()}
                  </span>
                </div>
                <div className="text-sm text-gray-600 dark:text-gray-400">by {event.actor}</div>
                {event.source && (
                  <div className="text-xs text-gray-500 dark:text-gray-500 mt-1">
                    Source: {event.source}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Today's Activity</h2>
        <div className="flex items-center gap-2">
          <label className="text-sm text-gray-600 dark:text-gray-400">Filter:</label>
          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value)}
            className="px-3 py-1.5 text-sm border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400"
          >
            {eventTypes.map((type) => (
              <option key={type} value={type}>
                {type === 'all' ? 'All Events' : type}
              </option>
            ))}
          </select>
        </div>
      </div>

      {filteredEvents.length === 0 ? (
        <div className="text-center py-12 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
          <p className="text-gray-600 dark:text-gray-400">No activity today</p>
        </div>
      ) : (
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          {renderEventGroup('Morning', groupedEvents.morning)}
          {renderEventGroup('Afternoon', groupedEvents.afternoon)}
          {renderEventGroup('Evening', groupedEvents.evening)}
        </div>
      )}
    </div>
  );
}
