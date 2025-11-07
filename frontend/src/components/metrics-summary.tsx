import type { MetricSummary } from '../api/types';

interface MetricsSummaryProps {
  metrics: MetricSummary[];
}

export function MetricsSummary({ metrics }: MetricsSummaryProps) {
  const getChangeIcon = (direction: 'up' | 'down' | 'stable') => {
    if (direction === 'up') {
      return (
        <svg
          className="w-4 h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M5 10l7-7m0 0l7 7m-7-7v18"
          />
        </svg>
      );
    }
    if (direction === 'down') {
      return (
        <svg
          className="w-4 h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M19 14l-7 7m0 0l-7-7m7 7V3"
          />
        </svg>
      );
    }
    return (
      <svg
        className="w-4 h-4"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        aria-hidden="true"
      >
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14" />
      </svg>
    );
  };

  const getChangeColor = (direction: 'up' | 'down' | 'stable', positiveTrend: boolean) => {
    if (direction === 'stable') {
      return 'text-gray-600 dark:text-gray-400';
    }
    if (positiveTrend) {
      return 'text-green-600 dark:text-green-400';
    }
    return 'text-red-600 dark:text-red-400';
  };

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
      <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100 mb-4">Key Metrics</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {metrics.map((metric, index) => {
          const changeColor = getChangeColor(metric.change_direction, metric.positive_trend);
          const changeIcon = getChangeIcon(metric.change_direction);

          return (
            <div
              key={index}
              className="p-4 bg-gray-50 dark:bg-gray-900/50 rounded-lg border border-gray-200 dark:border-gray-700"
            >
              <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">{metric.label}</div>
              <div className="flex items-baseline gap-2">
                <span className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                  {metric.value}
                </span>
                <span className="text-sm text-gray-500 dark:text-gray-500">{metric.unit}</span>
              </div>
              {metric.change !== 0 && (
                <div className={`flex items-center gap-1 mt-2 ${changeColor}`}>
                  {changeIcon}
                  <span className="text-sm font-medium">
                    {Math.abs(metric.change) > 0
                      ? `${Math.abs(metric.change)}% vs last week`
                      : 'stable'}
                  </span>
                </div>
              )}
              {metric.change === 0 && (
                <div className="flex items-center gap-1 mt-2 text-gray-600 dark:text-gray-400">
                  {changeIcon}
                  <span className="text-sm font-medium">stable</span>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
