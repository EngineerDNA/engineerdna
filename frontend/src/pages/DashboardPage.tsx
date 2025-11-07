import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useEvents } from '../hooks/useEvents';
import { useInsights } from '../hooks/useInsights';
import { Dashboard } from '../components/Dashboard';
import { ErrorAlert } from '../components/ErrorAlert';
import { DASHBOARD_EVENT_LIMIT, THROUGHPUT_DAYS } from '../constants';
import { api } from '../api/client';

type DateRange = '7days' | '30days' | '90days' | 'all';

export function DashboardPage() {
  const [dateRange, setDateRange] = useState<DateRange>('30days');

  const getSinceDate = (range: DateRange): string | undefined => {
    if (range === 'all') return undefined;
    const days = range === '7days' ? 7 : range === '30days' ? 30 : 90;
    const date = new Date();
    date.setDate(date.getDate() - days);
    return date.toISOString();
  };

  const sinceDate = useMemo(() => getSinceDate(dateRange), [dateRange]);

  const {
    data: eventsData,
    isLoading: eventsLoading,
    error: eventsError,
  } = useEvents({
    limit: DASHBOARD_EVENT_LIMIT,
    since: sinceDate,
  });

  const { data: insightsData, error: insightsError } = useInsights();

  const { data: throughputData, isLoading: throughputLoading } = useQuery({
    queryKey: ['throughput'],
    queryFn: () => api.getThroughput(THROUGHPUT_DAYS),
  });

  const events = eventsData?.events || [];
  const insights = insightsData?.insights || [];

  return (
    <div className="space-y-6">
      {eventsError && (
        <ErrorAlert error={eventsError as Error} onRetry={() => window.location.reload()} />
      )}
      {insightsError && (
        <ErrorAlert error={insightsError as Error} onRetry={() => window.location.reload()} />
      )}

      <div className="flex flex-wrap gap-2">
        <button
          onClick={() => setDateRange('7days')}
          className={`px-4 py-2 text-sm rounded-lg transition-colors ${
            dateRange === '7days'
              ? 'bg-blue-600 dark:bg-blue-700 text-white'
              : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
          }`}
        >
          Last 7 days
        </button>
        <button
          onClick={() => setDateRange('30days')}
          className={`px-4 py-2 text-sm rounded-lg transition-colors ${
            dateRange === '30days'
              ? 'bg-blue-600 dark:bg-blue-700 text-white'
              : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
          }`}
        >
          Last 30 days
        </button>
        <button
          onClick={() => setDateRange('90days')}
          className={`px-4 py-2 text-sm rounded-lg transition-colors ${
            dateRange === '90days'
              ? 'bg-blue-600 dark:bg-blue-700 text-white'
              : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
          }`}
        >
          Last 90 days
        </button>
        <button
          onClick={() => setDateRange('all')}
          className={`px-4 py-2 text-sm rounded-lg transition-colors ${
            dateRange === 'all'
              ? 'bg-blue-600 dark:bg-blue-700 text-white'
              : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
          }`}
        >
          All time
        </button>
      </div>

      <Dashboard
        events={events}
        insights={insights}
        throughputData={throughputData?.data}
        isLoading={eventsLoading}
        isThroughputLoading={throughputLoading}
      />
    </div>
  );
}
