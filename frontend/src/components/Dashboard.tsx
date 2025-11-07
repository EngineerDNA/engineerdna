import { useState, useMemo, memo } from 'react';
import { useDebounce } from '../hooks/useDebounce';
import { ThroughputChart } from './charts/ThroughputChart';
import { InsightsPanel } from './InsightsPanel';
import { EventList } from './EventList';
import type { Event, Insight } from '../api/types';
import { RECENT_EVENTS_COUNT, THROUGHPUT_DAYS } from '../constants';

type ActivityFilter = 'all' | 'pull_request' | 'issue' | 'commit';

interface ThroughputDataPoint {
  date: string;
  pull_requests: number;
  issues: number;
}

interface DashboardProps {
  events: Event[];
  insights: Insight[];
  throughputData?: ThroughputDataPoint[];
  isLoading?: boolean;
  isThroughputLoading?: boolean;
}

export function Dashboard({
  events,
  insights,
  throughputData,
  isLoading,
  isThroughputLoading,
}: DashboardProps) {
  const [activityFilter, setActivityFilter] = useState<ActivityFilter>('all');
  const [activitySearch, setActivitySearch] = useState('');
  const debouncedSearch = useDebounce(activitySearch, 300);

  const filteredRecentEvents = useMemo(() => {
    let result = [...events].slice(0, RECENT_EVENTS_COUNT * 3);

    if (activityFilter !== 'all') {
      result = result.filter((e) => e.type === activityFilter);
    }

    if (debouncedSearch) {
      const query = debouncedSearch.toLowerCase();
      result = result.filter(
        (e) => e.actor.toLowerCase().includes(query) || e.source.toLowerCase().includes(query)
      );
    }

    return result.slice(0, RECENT_EVENTS_COUNT);
  }, [events, activityFilter, debouncedSearch]);

  const prCount = events.filter((e) => e.type === 'pull_request').length;
  const issueCount = events.filter((e) => e.type === 'issue').length;
  const teamSize = getUniqueActors(events).length;

  if (isLoading) {
    return (
      <div className="space-y-6">
        <LoadingSkeleton />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Dashboard</h1>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
            Team metrics and activity overview
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <MetricCard title="PRs Merged" value={prCount} icon="pr" />
        <MetricCard title="Issues Closed" value={issueCount} icon="issue" />
        <MetricCard title="Team Size" value={teamSize} icon="team" />
      </div>

      <InsightsPanel insights={insights} />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
          <ThroughputChart
            data={throughputData}
            isLoading={isThroughputLoading}
            days={THROUGHPUT_DAYS}
          />
        </div>

        <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
          <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3 mb-4">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
              Recent Activity
            </h3>
            <div className="flex flex-wrap gap-2 w-full sm:w-auto">
              <button
                onClick={() => setActivityFilter('all')}
                className={`px-3 py-1 text-sm rounded transition-colors ${
                  activityFilter === 'all'
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
                }`}
              >
                All
              </button>
              <button
                onClick={() => setActivityFilter('pull_request')}
                className={`px-3 py-1 text-sm rounded transition-colors ${
                  activityFilter === 'pull_request'
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
                }`}
              >
                PRs
              </button>
              <button
                onClick={() => setActivityFilter('issue')}
                className={`px-3 py-1 text-sm rounded transition-colors ${
                  activityFilter === 'issue'
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
                }`}
              >
                Issues
              </button>
              <button
                onClick={() => setActivityFilter('commit')}
                className={`px-3 py-1 text-sm rounded transition-colors ${
                  activityFilter === 'commit'
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
                }`}
              >
                Commits
              </button>
            </div>
          </div>

          <input
            type="text"
            placeholder="Filter by actor or source..."
            value={activitySearch}
            onChange={(e) => setActivitySearch(e.target.value)}
            className="w-full mb-4 px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />

          {filteredRecentEvents.length > 0 ? (
            <EventList events={filteredRecentEvents} />
          ) : (
            <div className="text-center py-8 text-gray-500 dark:text-gray-400">
              No activity matches your filters
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

const MetricCard = memo(function MetricCard({
  title,
  value,
  trend,
  icon,
}: {
  title: string;
  value: number;
  trend?: string;
  icon?: 'pr' | 'issue' | 'team';
}) {
  const iconPaths = {
    pr: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
    issue: 'M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
    team: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z',
  };

  return (
    <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow hover:shadow-md transition-shadow">
      <div className="flex items-center justify-between mb-2">
        <div className="text-sm text-gray-600 dark:text-gray-400">{title}</div>
        {icon && (
          <svg
            className="w-5 h-5 text-gray-400 dark:text-gray-600"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d={iconPaths[icon]}
            />
          </svg>
        )}
      </div>
      <div className="flex items-baseline gap-2">
        <div className="text-3xl font-bold text-gray-900 dark:text-gray-100">{value}</div>
        {trend && (
          <span
            className={`text-sm font-medium ${
              trend.startsWith('+')
                ? 'text-green-600 dark:text-green-400'
                : 'text-red-600 dark:text-red-400'
            }`}
          >
            {trend}
          </span>
        )}
      </div>
    </div>
  );
});

function LoadingSkeleton() {
  return (
    <>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow animate-pulse">
            <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-24 mb-3"></div>
            <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded w-16"></div>
          </div>
        ))}
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {[1, 2].map((i) => (
          <div key={i} className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow animate-pulse">
            <div className="h-6 bg-gray-200 dark:bg-gray-700 rounded w-32 mb-4"></div>
            <div className="h-64 bg-gray-200 dark:bg-gray-700 rounded"></div>
          </div>
        ))}
      </div>
    </>
  );
}

function getUniqueActors(events: Event[]): string[] {
  return Array.from(new Set(events.map((e) => e.actor)));
}
