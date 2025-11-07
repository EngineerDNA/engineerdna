import {
  BarChart as RechartsBarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';

interface BarChartProps {
  data: Array<{ entity_name: string; value: number }>;
  unit?: string;
  isLoading?: boolean;
  orientation?: 'horizontal' | 'vertical';
}

/**
 * BarChart displays comparative metric values across entities (engineers or teams).
 * Supports horizontal and vertical orientations with responsive sizing.
 * Built with Recharts and supports dark mode.
 */
export function BarChart({
  data,
  unit = '',
  isLoading = false,
  orientation = 'vertical',
}: BarChartProps) {
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

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
      <ResponsiveContainer width="100%" height={250}>
        <RechartsBarChart
          data={data}
          layout={orientation === 'horizontal' ? 'horizontal' : 'vertical'}
          margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
        >
          {orientation === 'horizontal' ? (
            <>
              <XAxis type="number" stroke="#9CA3AF" style={{ fontSize: '12px' }} />
              <YAxis
                type="category"
                dataKey="entity_name"
                stroke="#9CA3AF"
                style={{ fontSize: '12px' }}
                width={100}
              />
            </>
          ) : (
            <>
              <XAxis
                dataKey="entity_name"
                stroke="#9CA3AF"
                style={{ fontSize: '12px' }}
                angle={-45}
                textAnchor="end"
                height={80}
              />
              <YAxis stroke="#9CA3AF" style={{ fontSize: '12px' }} />
            </>
          )}
          <Tooltip
            contentStyle={{
              backgroundColor: 'rgba(0, 0, 0, 0.8)',
              border: 'none',
              borderRadius: '8px',
              color: '#fff',
            }}
            formatter={(value: number) => [`${value.toLocaleString()}${unit}`, 'Value']}
          />
          <Bar dataKey="value" fill="#3B82F6" radius={[4, 4, 0, 0]} />
        </RechartsBarChart>
      </ResponsiveContainer>
    </div>
  );
}
