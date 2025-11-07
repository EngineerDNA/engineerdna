interface NumberCardProps {
  title: string;
  value: number | string;
  previousValue?: number;
  changePercentage?: number;
  changeDirection?: 'up' | 'down' | 'stable';
  unit?: string;
  isLoading?: boolean;
}

/**
 * NumberCard displays a single KPI metric value with optional comparison to previous period.
 * Includes trend indicator (up/down arrow) and change percentage.
 * Supports dark mode and loading state with skeleton.
 */
export function NumberCard({
  title,
  value,
  previousValue,
  changePercentage,
  changeDirection = 'stable',
  unit = '',
  isLoading = false,
}: NumberCardProps) {
  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-300 dark:bg-gray-700 rounded w-1/2 mb-4" />
          <div className="h-8 bg-gray-300 dark:bg-gray-700 rounded w-3/4 mb-2" />
          <div className="h-4 bg-gray-300 dark:bg-gray-700 rounded w-1/3" />
        </div>
      </div>
    );
  }

  const getChangeColor = () => {
    if (changeDirection === 'up') return 'text-green-600 dark:text-green-400';
    if (changeDirection === 'down') return 'text-red-600 dark:text-red-400';
    return 'text-gray-600 dark:text-gray-400';
  };

  const getChangeIcon = () => {
    if (changeDirection === 'up') {
      return (
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M5 10l7-7m0 0l7 7m-7-7v18"
          />
        </svg>
      );
    }
    if (changeDirection === 'down') {
      return (
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14" />
      </svg>
    );
  };

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
      <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-2">{title}</h3>
      <div className="flex items-baseline gap-2">
        <p className="text-3xl font-bold text-gray-900 dark:text-gray-100">
          {typeof value === 'number' ? value.toLocaleString() : value}
          {unit && <span className="text-lg text-gray-600 dark:text-gray-400 ml-1">{unit}</span>}
        </p>
      </div>
      {changePercentage !== undefined && (
        <div className={`flex items-center gap-1 mt-2 text-sm ${getChangeColor()}`}>
          {getChangeIcon()}
          <span>{Math.abs(changePercentage).toFixed(1)}%</span>
          {previousValue !== undefined && (
            <span className="text-gray-500 dark:text-gray-400">
              vs {previousValue.toLocaleString()}
              {unit}
            </span>
          )}
        </div>
      )}
    </div>
  );
}
