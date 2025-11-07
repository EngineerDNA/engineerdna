import type { Event } from '../../api/types';

interface ActivityFeedProps {
  events: Event[];
  isLoading?: boolean;
  limit?: number;
}

/**
 * ActivityFeed displays a chronological list of recent engineering events.
 * Shows event type icons, actor information, and relative timestamps.
 * Supports limiting displayed events with overflow indicator.
 */
export function ActivityFeed({ events, isLoading = false, limit = 10 }: ActivityFeedProps) {
  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="animate-pulse space-y-3">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="h-16 bg-gray-300 dark:bg-gray-700 rounded" />
          ))}
        </div>
      </div>
    );
  }

  if (!events || events.length === 0) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="flex items-center justify-center h-64 text-gray-500 dark:text-gray-400">
          No recent activity
        </div>
      </div>
    );
  }

  const displayEvents = events.slice(0, limit);

  const getEventIcon = (type: string) => {
    if (type.includes('pull_request')) {
      return (
        <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
          <path
            fillRule="evenodd"
            d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zm-2 4a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"
            clipRule="evenodd"
          />
        </svg>
      );
    }
    if (type.includes('issue')) {
      return (
        <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
          <path
            fillRule="evenodd"
            d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z"
            clipRule="evenodd"
          />
        </svg>
      );
    }
    if (type.includes('commit')) {
      return (
        <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
          <path
            fillRule="evenodd"
            d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z"
            clipRule="evenodd"
          />
        </svg>
      );
    }
    return (
      <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
        <path
          fillRule="evenodd"
          d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z"
          clipRule="evenodd"
        />
      </svg>
    );
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  };

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full overflow-auto">
      <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-4">Recent Activity</h3>
      <div className="space-y-3">
        {displayEvents.map((event) => (
          <div
            key={event.id}
            className="flex items-start gap-3 p-3 bg-gray-50 dark:bg-gray-900 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          >
            <div className="flex-shrink-0 text-blue-600 dark:text-blue-400 mt-0.5">
              {getEventIcon(event.type)}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
                {event.type.replace(/_/g, ' ')}
              </p>
              <p className="text-xs text-gray-600 dark:text-gray-400 truncate">{event.actor}</p>
              <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">
                {formatTimestamp(event.timestamp)}
              </p>
            </div>
          </div>
        ))}
      </div>
      {events.length > limit && (
        <p className="text-xs text-gray-500 dark:text-gray-400 mt-4 text-center">
          Showing {limit} of {events.length} events
        </p>
      )}
    </div>
  );
}
