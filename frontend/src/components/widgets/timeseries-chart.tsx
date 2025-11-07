import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts';
import type { AlertMarker } from '../../api/types';

interface TimeseriesChartProps {
  data: Array<{ timestamp: string; value: number }>;
  unit?: string;
  alerts?: AlertMarker[];
  isLoading?: boolean;
  chartType?: 'line' | 'area';
}

/**
 * TimeseriesChart displays metric values over time as a line chart.
 * Supports alert overlay markers and responsive sizing.
 * Built with Recharts and supports dark mode.
 */
export function TimeseriesChart({
  data,
  unit = '',
  alerts = [],
  isLoading = false,
}: TimeseriesChartProps) {
  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="animate-pulse flex items-center justify-center h-64">
          <div className="h-full w-full bg-gray-300 dark:bg-gray-700 rounded" />
        </div>
      </div>
    );
  }

  if (!data || data.length === 0) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="flex items-center justify-center h-64 text-gray-500 dark:text-gray-400">
          No data available
        </div>
      </div>
    );
  }

  const formattedData = data.map((d) => ({
    ...d,
    date: new Date(d.timestamp).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
  }));

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
      <ResponsiveContainer width="100%" height={250}>
        <LineChart data={formattedData}>
          <XAxis dataKey="date" stroke="#9CA3AF" style={{ fontSize: '12px' }} />
          <YAxis
            stroke="#9CA3AF"
            style={{ fontSize: '12px' }}
            tickFormatter={(value) => `${value}${unit}`}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: 'rgba(0, 0, 0, 0.8)',
              border: 'none',
              borderRadius: '8px',
              color: '#fff',
            }}
            formatter={(value: number) => [`${value.toLocaleString()}${unit}`, 'Value']}
          />
          {alerts.map((alert, idx) => (
            <ReferenceLine
              key={idx}
              x={new Date(alert.timestamp).toLocaleDateString('en-US', {
                month: 'short',
                day: 'numeric',
              })}
              stroke={
                alert.severity === 'critical'
                  ? '#DC2626'
                  : alert.severity === 'warning'
                    ? '#F59E0B'
                    : '#3B82F6'
              }
              strokeDasharray="3 3"
              label={{
                value: alert.title,
                position: 'top',
                fill:
                  alert.severity === 'critical'
                    ? '#DC2626'
                    : alert.severity === 'warning'
                      ? '#F59E0B'
                      : '#3B82F6',
                fontSize: 10,
              }}
            />
          ))}
          <Line
            type="monotone"
            dataKey="value"
            stroke="#3B82F6"
            strokeWidth={2}
            dot={{ fill: '#3B82F6', r: 3 }}
            activeDot={{ r: 5 }}
          />
        </LineChart>
      </ResponsiveContainer>
      {alerts.length > 0 && (
        <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
          <p className="text-xs text-gray-600 dark:text-gray-400 mb-2">Alerts on timeline:</p>
          <div className="flex flex-wrap gap-2">
            {alerts.map((alert, idx) => (
              <span
                key={idx}
                className={`inline-flex items-center px-2 py-1 rounded text-xs ${
                  alert.severity === 'critical'
                    ? 'bg-red-100 dark:bg-red-900/20 text-red-800 dark:text-red-200'
                    : alert.severity === 'warning'
                      ? 'bg-yellow-100 dark:bg-yellow-900/20 text-yellow-800 dark:text-yellow-200'
                      : 'bg-blue-100 dark:bg-blue-900/20 text-blue-800 dark:text-blue-200'
                }`}
              >
                {alert.title}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
