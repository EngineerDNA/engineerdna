import { useState, useMemo } from 'react';
import { useEvents } from '../hooks/useEvents';
import { EventList } from '../components/EventList';
import { Link } from 'react-router-dom';
import { EVENTS_PER_PAGE } from '../constants';

type DateRange = '7d' | '30d' | '90d' | 'all';

export function EventsPage() {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(EVENTS_PER_PAGE);
  const [searchQuery, setSearchQuery] = useState('');
  const [typeFilter, setTypeFilter] = useState<string>('all');
  const [sourceFilter, setSourceFilter] = useState<string>('all');
  const [dateRange, setDateRange] = useState<DateRange>('all');

  const sinceDate = useMemo(() => {
    if (dateRange === 'all') return undefined;
    const days = parseInt(dateRange);
    const date = new Date();
    date.setDate(date.getDate() - days);
    return date.toISOString();
  }, [dateRange]);

  const { data, isLoading, error } = useEvents({
    limit: 1000,
    since: sinceDate,
  });

  const allEvents = data?.events || [];

  const filteredEvents = useMemo(() => {
    let result = [...allEvents];

    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      result = result.filter(
        (event) =>
          event.actor.toLowerCase().includes(query) ||
          event.source_id.toLowerCase().includes(query) ||
          event.type.toLowerCase().includes(query)
      );
    }

    if (typeFilter !== 'all') {
      result = result.filter((event) => event.type === typeFilter);
    }

    if (sourceFilter !== 'all') {
      result = result.filter((event) => event.source === sourceFilter);
    }

    return result;
  }, [allEvents, searchQuery, typeFilter, sourceFilter]);

  const totalEvents = filteredEvents.length;
  const totalPages = totalEvents > 0 ? Math.ceil(totalEvents / pageSize) : 1;
  const paginatedEvents = filteredEvents.slice((page - 1) * pageSize, page * pageSize);

  const uniqueTypes = Array.from(new Set(allEvents.map((e) => e.type))).sort();
  const uniqueSources = Array.from(new Set(allEvents.map((e) => e.source))).sort();

  const handlePageSizeChange = (newSize: number) => {
    setPageSize(newSize);
    setPage(1);
  };

  const handleFiltersChange = () => {
    setPage(1);
  };

  if (isLoading) {
    return (
      <div>
        <h2 className="text-2xl font-bold mb-6 text-gray-900 dark:text-gray-100">Events</h2>
        <LoadingSkeleton />
      </div>
    );
  }

  if (error) {
    return (
      <div>
        <h2 className="text-2xl font-bold mb-6 text-gray-900 dark:text-gray-100">Events</h2>
        <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
          <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded">
            <p className="text-red-600 dark:text-red-400">Failed to load events: {error.message}</p>
          </div>
        </div>
      </div>
    );
  }

  if (allEvents.length === 0) {
    return (
      <div>
        <h2 className="text-2xl font-bold mb-6 text-gray-900 dark:text-gray-100">Events</h2>
        <div className="bg-white dark:bg-gray-800 p-12 rounded-lg shadow text-center">
          <svg
            className="mx-auto h-12 w-12 text-gray-400 dark:text-gray-600 mb-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
            />
          </svg>
          <p className="text-gray-600 dark:text-gray-400 text-lg mb-2">No events yet</p>
          <p className="text-sm text-gray-500 dark:text-gray-500 mb-6">
            Configure a plugin and sync to see events, or use the seed command to generate demo data
          </p>
          <Link
            to="/plugins"
            className="inline-block px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 transition-colors"
          >
            View Plugins
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Events</h2>
        <p className="text-gray-600 dark:text-gray-400 mt-1">
          View and filter engineering activity events
        </p>
      </div>

      <div className="mb-6 space-y-4">
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="flex-1">
            <input
              type="text"
              placeholder="Search by actor, source ID, or type..."
              value={searchQuery}
              onChange={(e) => {
                setSearchQuery(e.target.value);
                handleFiltersChange();
              }}
              className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>

          <select
            value={dateRange}
            onChange={(e) => {
              setDateRange(e.target.value as DateRange);
              handleFiltersChange();
            }}
            className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="all">All Time</option>
            <option value="7d">Last 7 Days</option>
            <option value="30d">Last 30 Days</option>
            <option value="90d">Last 90 Days</option>
          </select>

          <select
            value={typeFilter}
            onChange={(e) => {
              setTypeFilter(e.target.value);
              handleFiltersChange();
            }}
            className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="all">All Types</option>
            {uniqueTypes.map((type) => (
              <option key={type} value={type}>
                {type}
              </option>
            ))}
          </select>

          <select
            value={sourceFilter}
            onChange={(e) => {
              setSourceFilter(e.target.value);
              handleFiltersChange();
            }}
            className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="all">All Sources</option>
            {uniqueSources.map((source) => (
              <option key={source} value={source}>
                {source}
              </option>
            ))}
          </select>
        </div>

        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
          <div className="text-sm text-gray-600 dark:text-gray-400">
            Showing {paginatedEvents.length === 0 ? 0 : (page - 1) * pageSize + 1}-
            {Math.min(page * pageSize, totalEvents)} of {totalEvents} events
          </div>

          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-600 dark:text-gray-400">Per page:</span>
            <select
              value={pageSize}
              onChange={(e) => handlePageSizeChange(Number(e.target.value))}
              className="px-3 py-1 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value={10}>10</option>
              <option value={20}>20</option>
              <option value={50}>50</option>
            </select>
          </div>
        </div>
      </div>

      {paginatedEvents.length === 0 ? (
        <div className="bg-white dark:bg-gray-800 p-12 rounded-lg shadow text-center">
          <svg
            className="mx-auto h-12 w-12 text-gray-400 dark:text-gray-600 mb-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
          <p className="text-gray-600 dark:text-gray-400 text-lg mb-2">
            No events match your filters
          </p>
          <p className="text-sm text-gray-500 dark:text-gray-500">
            Try adjusting your search criteria or filters
          </p>
        </div>
      ) : (
        <>
          <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow overflow-x-auto">
            <EventList events={paginatedEvents} />
          </div>

          {totalPages > 1 && (
            <div className="flex flex-col sm:flex-row justify-between items-center gap-4 mt-6">
              <div className="text-sm text-gray-600 dark:text-gray-400">
                Page {page} of {totalPages}
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page === 1}
                  className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
                >
                  Previous
                </button>
                <span className="px-4 py-2 text-gray-700 dark:text-gray-300">
                  Page {page} of {totalPages}
                </span>
                <button
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page === totalPages}
                  className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}

function LoadingSkeleton() {
  return (
    <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
      <div className="space-y-4">
        {[1, 2, 3, 4, 5].map((i) => (
          <div key={i} className="animate-pulse">
            <div className="flex justify-between items-start">
              <div className="flex-1">
                <div className="h-5 bg-gray-200 dark:bg-gray-700 rounded w-1/4 mb-2"></div>
                <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4"></div>
              </div>
              <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-32"></div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
