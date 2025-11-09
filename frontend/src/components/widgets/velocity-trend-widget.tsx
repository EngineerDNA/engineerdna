import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/client';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts';
import type { WidgetProps } from './widget-registry';

export function VelocityTrendWidget({ config }: WidgetProps) {
  const teamId = config.team_id as string | undefined;
  const timeRange = (config.time_range as number) || 6;

  const { data, isLoading } = useQuery({
    queryKey: ['velocity', teamId, timeRange],
    queryFn: () => api.getVelocity(teamId || '', timeRange),
    enabled: !!teamId,
    staleTime: 5 * 60 * 1000,
  });

  if (!teamId) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="flex items-center justify-center h-32 text-gray-500 dark:text-gray-400">
          Configure team_id to view velocity
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="animate-pulse">
          <div className="h-6 bg-gray-300 dark:bg-gray-700 rounded w-1/3 mb-4" />
          <div className="h-48 bg-gray-300 dark:bg-gray-700 rounded" />
        </div>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-4">
          Velocity Trend
        </h3>
        <div className="flex items-center justify-center h-32 text-gray-500 dark:text-gray-400">
          No velocity data available
        </div>
      </div>
    );
  }

  const getTrendIcon = () => {
    if (data.trend === 'increasing') {
      return (
        <svg
          className="w-5 h-5 text-green-600 dark:text-green-400"
          fill="currentColor"
          viewBox="0 0 20 20"
        >
          <path
            fillRule="evenodd"
            d="M5 10l7-7m0 0l7 7m-7-7v18"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            stroke="currentColor"
            fill="none"
          />
        </svg>
      );
    }
    if (data.trend === 'decreasing') {
      return (
        <svg
          className="w-5 h-5 text-red-600 dark:text-red-400"
          fill="currentColor"
          viewBox="0 0 20 20"
        >
          <path
            fillRule="evenodd"
            d="M19 14l-7 7m0 0l-7-7m7 7V3"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            stroke="currentColor"
            fill="none"
          />
        </svg>
      );
    }
    return (
      <svg
        className="w-5 h-5 text-gray-600 dark:text-gray-400"
        fill="currentColor"
        viewBox="0 0 20 20"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M5 12h14"
          stroke="currentColor"
          fill="none"
        />
      </svg>
    );
  };

  const getTrendColor = () => {
    if (data.trend === 'increasing') return 'text-green-600 dark:text-green-400';
    if (data.trend === 'decreasing') return 'text-red-600 dark:text-red-400';
    return 'text-gray-600 dark:text-gray-400';
  };

  const chartData = data.recent_sprints.map((sprint) => ({
    name: sprint.sprint_name,
    points: sprint.points,
    average: data.average_points,
  }));

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full flex flex-col">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400">Velocity Trend</h3>
        <div className={`flex items-center gap-2 ${getTrendColor()}`}>
          {getTrendIcon()}
          <span className="text-sm font-medium capitalize">{data.trend}</span>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-4 mb-4">
        <div>
          <p className="text-xs text-gray-500 dark:text-gray-400">Average</p>
          <p className="text-lg font-bold text-gray-900 dark:text-gray-100">
            {data.average_points.toFixed(1)} <span className="text-sm text-gray-500">pts</span>
          </p>
        </div>
        <div>
          <p className="text-xs text-gray-500 dark:text-gray-400">Std Dev</p>
          <p className="text-lg font-bold text-gray-900 dark:text-gray-100">
            {data.std_dev.toFixed(1)} <span className="text-sm text-gray-500">pts</span>
          </p>
        </div>
        <div>
          <p className="text-xs text-gray-500 dark:text-gray-400">Confidence</p>
          <p className="text-lg font-bold text-gray-900 dark:text-gray-100">
            {(data.confidence * 100).toFixed(0)} <span className="text-sm text-gray-500">%</span>
          </p>
        </div>
      </div>

      <div className="flex-1 min-h-0">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={chartData}>
            <XAxis dataKey="name" stroke="#9CA3AF" style={{ fontSize: '12px' }} />
            <YAxis stroke="#9CA3AF" style={{ fontSize: '12px' }} />
            <Tooltip
              contentStyle={{
                backgroundColor: 'rgba(0, 0, 0, 0.8)',
                border: 'none',
                borderRadius: '8px',
                color: '#fff',
              }}
              formatter={(value: number) => [`${value.toFixed(1)} pts`, 'Points']}
            />
            <ReferenceLine
              y={data.average_points}
              stroke="#9CA3AF"
              strokeDasharray="3 3"
              label={{ value: 'Avg', position: 'right', fill: '#9CA3AF', fontSize: 10 }}
            />
            <Line
              type="monotone"
              dataKey="points"
              stroke="#3B82F6"
              strokeWidth={2}
              dot={{ fill: '#3B82F6', r: 4 }}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
