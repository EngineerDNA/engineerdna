import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import type { SprintBurndownData } from '../../api/types';

interface SprintBurndownChartProps {
  data: SprintBurndownData;
}

export function SprintBurndownChart({ data }: SprintBurndownChartProps) {
  const riskColors = {
    on_track: {
      bg: 'bg-green-50 dark:bg-green-900/20',
      border: 'border-green-200 dark:border-green-800',
      text: 'text-green-800 dark:text-green-200',
    },
    at_risk: {
      bg: 'bg-yellow-50 dark:bg-yellow-900/20',
      border: 'border-yellow-200 dark:border-yellow-800',
      text: 'text-yellow-800 dark:text-yellow-200',
    },
    high_risk: {
      bg: 'bg-red-50 dark:bg-red-900/20',
      border: 'border-red-200 dark:border-red-800',
      text: 'text-red-800 dark:text-red-200',
    },
  };

  const colors = riskColors[data.risk_level];

  const getRiskLabel = () => {
    switch (data.risk_level) {
      case 'on_track':
        return 'On Track';
      case 'at_risk':
        return 'At Risk';
      case 'high_risk':
        return 'High Risk';
      default:
        return 'Unknown';
    }
  };

  const getRiskIcon = () => {
    if (data.risk_level === 'on_track') {
      return (
        <svg
          className="w-5 h-5 text-green-600 dark:text-green-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
      );
    }
    if (data.risk_level === 'at_risk') {
      return (
        <svg
          className="w-5 h-5 text-yellow-600 dark:text-yellow-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
      );
    }
    return (
      <svg
        className="w-5 h-5 text-red-600 dark:text-red-400"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M6 18L18 6M6 6l12 12"
        />
      </svg>
    );
  };

  const completionPercentage = ((data.completed_points / data.total_points) * 100).toFixed(1);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            {data.sprint_name}
          </h2>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            {new Date(data.start_date).toLocaleDateString()} -{' '}
            {new Date(data.end_date).toLocaleDateString()}
          </p>
        </div>
        <div
          className={`flex items-center gap-2 px-3 py-2 ${colors.bg} ${colors.border} border rounded-lg`}
        >
          {getRiskIcon()}
          <span className={`font-semibold ${colors.text}`}>{getRiskLabel()}</span>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-4">
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
          <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Total Points</div>
          <div className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            {data.total_points}
          </div>
        </div>
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
          <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Completed</div>
          <div className="text-2xl font-bold text-green-600 dark:text-green-400">
            {data.completed_points}
          </div>
        </div>
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
          <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Progress</div>
          <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">
            {completionPercentage}%
          </div>
        </div>
      </div>

      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
          Burndown Chart
        </h3>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={data.burndown_points}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis
              dataKey="date"
              stroke="#9CA3AF"
              tickFormatter={(value) =>
                new Date(value).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
              }
            />
            <YAxis stroke="#9CA3AF" />
            <Tooltip
              contentStyle={{
                backgroundColor: '#1F2937',
                border: '1px solid #374151',
                borderRadius: '0.5rem',
                color: '#F9FAFB',
              }}
              labelFormatter={(value) => new Date(value).toLocaleDateString()}
            />
            <Legend />
            <Line
              type="monotone"
              dataKey="ideal"
              stroke="#6B7280"
              strokeDasharray="5 5"
              name="Ideal"
              dot={false}
            />
            <Line
              type="monotone"
              dataKey="actual"
              stroke="#3B82F6"
              strokeWidth={2}
              name="Actual"
              dot={{ fill: '#3B82F6', r: 4 }}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
