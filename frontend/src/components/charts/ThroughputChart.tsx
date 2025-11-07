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
import { CHART_HEIGHT } from '../../constants';
import { useTheme } from '../../contexts/ThemeContext';

interface ThroughputDataPoint {
  date: string;
  pull_requests: number;
  issues: number;
}

interface ThroughputChartProps {
  data?: ThroughputDataPoint[];
  isLoading?: boolean;
  days?: number;
}

export function ThroughputChart({ data, isLoading, days = 30 }: ThroughputChartProps) {
  const { theme } = useTheme();

  if (isLoading || !data) {
    return (
      <div className="animate-pulse">
        <div className="h-6 bg-gray-200 dark:bg-gray-700 rounded w-48 mb-4"></div>
        <div className="h-64 bg-gray-200 dark:bg-gray-700 rounded"></div>
      </div>
    );
  }

  const isDark = theme === 'dark';
  const gridColor = isDark ? '#374151' : '#e5e7eb';
  const textColor = isDark ? '#9ca3af' : '#6b7280';

  return (
    <div>
      <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">
        Throughput (Last {days} Days)
      </h3>
      <div className="flex items-center justify-center">
        <ResponsiveContainer width="100%" height={CHART_HEIGHT}>
          <LineChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke={gridColor} />
            <XAxis dataKey="date" stroke={textColor} />
            <YAxis stroke={textColor} />
            <Tooltip
              contentStyle={{
                backgroundColor: isDark ? '#1f2937' : '#ffffff',
                border: `1px solid ${isDark ? '#374151' : '#e5e7eb'}`,
                borderRadius: '0.5rem',
                color: isDark ? '#f3f4f6' : '#111827',
              }}
            />
            <Legend wrapperStyle={{ color: textColor }} />
            <Line type="monotone" dataKey="pull_requests" stroke="#3b82f6" name="PRs" />
            <Line type="monotone" dataKey="issues" stroke="#10b981" name="Issues" />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
