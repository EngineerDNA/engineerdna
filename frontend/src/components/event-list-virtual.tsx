import { memo, useRef } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import type { Event } from '../api/types';

interface EventListVirtualProps {
  events: Event[];
  height?: number;
}

/**
 * Virtualized EventList component for rendering large lists of events efficiently.
 * Only renders visible items in the viewport, significantly improving performance
 * for lists with thousands of events.
 *
 * Use this component instead of EventList when rendering >100 events.
 */
export const EventListVirtual = memo(function EventListVirtual({
  events,
  height = 600,
}: EventListVirtualProps) {
  const parentRef = useRef<HTMLDivElement>(null);

  const virtualizer = useVirtualizer({
    count: events.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 100,
    overscan: 5,
  });

  if (events.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500 dark:text-gray-400">No events to display</div>
    );
  }

  return (
    <div ref={parentRef} style={{ height: `${height}px`, overflow: 'auto' }} className="space-y-2">
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          width: '100%',
          position: 'relative',
        }}
      >
        {virtualizer.getVirtualItems().map((virtualItem) => {
          const event = events[virtualItem.index];
          return (
            <div
              key={virtualItem.key}
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                transform: `translateY(${virtualItem.start}px)`,
              }}
            >
              <div className="border-l-4 border-blue-500 dark:border-blue-400 pl-4 py-2 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors rounded-r mb-2">
                <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1">
                  <div className="font-medium text-gray-900 dark:text-gray-100">{event.actor}</div>
                  <div className="text-sm text-gray-600 dark:text-gray-400">
                    {new Date(event.timestamp).toLocaleString()}
                  </div>
                </div>
                <div className="text-sm text-gray-700 dark:text-gray-300 mt-1">
                  <span className="font-medium capitalize">{event.type.replace('_', ' ')}</span>
                  {event.data.title ? `: ${String(event.data.title)}` : null}
                  {event.data.action && !event.data.title ? `: ${String(event.data.action)}` : null}
                </div>
                <div className="flex flex-wrap gap-2 mt-1">
                  <span className="text-xs px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded">
                    {event.source}
                  </span>
                  {event.anonymized && (
                    <span className="text-xs px-2 py-1 bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-200 rounded">
                      Anonymized
                    </span>
                  )}
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
});
