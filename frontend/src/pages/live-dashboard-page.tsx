import { useLiveMetrics, useSprintBurndown, useTodayActivity } from '../hooks/useLiveMetrics';
import { LiveMetrics } from '../components/live-metrics';
import { SprintBurndownChart } from '../components/charts/sprint-burndown-chart';
import { ActivityFeed } from '../components/activity-feed';

export function LiveDashboardPage() {
  const { data: metrics, isLoading: metricsLoading, error: metricsError } = useLiveMetrics(30000);

  const {
    data: burndown,
    isLoading: burndownLoading,
    error: burndownError,
  } = useSprintBurndown(undefined, 30000);

  const {
    data: activityData,
    isLoading: activityLoading,
    error: activityError,
  } = useTodayActivity(30000);

  if (metricsLoading && burndownLoading && activityLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400 mx-auto mb-4" />
          <p className="text-gray-600 dark:text-gray-400">Loading live dashboard...</p>
        </div>
      </div>
    );
  }

  const hasError = metricsError || burndownError || activityError;

  if (hasError) {
    return (
      <div className="space-y-4">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Live Dashboard</h1>
        {metricsError && (
          <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded">
            <p className="text-red-800 dark:text-red-200">
              Failed to load metrics: {(metricsError as Error).message}
            </p>
          </div>
        )}
        {burndownError && (
          <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded">
            <p className="text-red-800 dark:text-red-200">
              Failed to load burndown: {(burndownError as Error).message}
            </p>
          </div>
        )}
        {activityError && (
          <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded">
            <p className="text-red-800 dark:text-red-200">
              Failed to load activity: {(activityError as Error).message}
            </p>
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Live Dashboard</h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Real-time metrics and activity feed
          </p>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
            <div className="w-2 h-2 bg-green-500 dark:bg-green-400 rounded-full animate-pulse" />
            <span>Auto-refresh: 30s</span>
          </div>
        </div>
      </div>

      {metrics && <LiveMetrics metrics={metrics} lastUpdated={new Date().toISOString()} />}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div>
          {burndown ? (
            <SprintBurndownChart data={burndown} />
          ) : (
            <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">
                Sprint Burndown
              </h2>
              <p className="text-gray-600 dark:text-gray-400">
                No active sprint. Create a sprint in the Planning page to see burndown data.
              </p>
            </div>
          )}
        </div>

        <div>
          {activityData?.events ? (
            <ActivityFeed events={activityData.events} />
          ) : (
            <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">
                Today's Activity
              </h2>
              <p className="text-gray-600 dark:text-gray-400">No activity today.</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
