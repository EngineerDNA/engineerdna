import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/client';
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import type { WidgetProps } from './widget-registry';

export function SprintBoardWidget({ config, onOpenModal }: WidgetProps) {
  const sprintId = (config.sprint_id as string) || 'current';

  const { data: sprintsData, isLoading: sprintsLoading } = useQuery({
    queryKey: ['sprints', 'active'],
    queryFn: () => api.getSprints(undefined, 'active'),
    enabled: sprintId === 'current',
  });

  const currentSprintId = sprintId === 'current' ? sprintsData?.sprints?.[0]?.id : sprintId;

  const { data: burndownData, isLoading: burndownLoading } = useQuery({
    queryKey: ['sprint-burndown', currentSprintId],
    queryFn: () => api.getSprintBurndown(currentSprintId),
    enabled: !!currentSprintId,
    staleTime: 60 * 1000,
  });

  const { data: healthData } = useQuery({
    queryKey: ['sprint-health', currentSprintId],
    queryFn: () => api.getSprintHealth(currentSprintId!),
    enabled: !!currentSprintId,
    staleTime: 60 * 1000,
  });

  const isLoading = sprintsLoading || burndownLoading;

  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="animate-pulse">
          <div className="h-6 bg-gray-300 dark:bg-gray-700 rounded w-1/3 mb-4" />
          <div className="h-32 bg-gray-300 dark:bg-gray-700 rounded mb-4" />
          <div className="h-4 bg-gray-300 dark:bg-gray-700 rounded w-1/2" />
        </div>
      </div>
    );
  }

  if (!burndownData) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-4">Sprint Board</h3>
        <div className="flex items-center justify-center h-32 text-gray-500 dark:text-gray-400">
          No active sprint
        </div>
      </div>
    );
  }

  const getRiskColor = (risk: string) => {
    switch (risk) {
      case 'on_track':
        return 'bg-green-100 dark:bg-green-900/20 text-green-800 dark:text-green-200';
      case 'at_risk':
        return 'bg-yellow-100 dark:bg-yellow-900/20 text-yellow-800 dark:text-yellow-200';
      case 'high_risk':
        return 'bg-red-100 dark:bg-red-900/20 text-red-800 dark:text-red-200';
      default:
        return 'bg-gray-100 dark:bg-gray-900/20 text-gray-800 dark:text-gray-200';
    }
  };

  const completionPercentage = (burndownData.completed_points / burndownData.total_points) * 100;

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full flex flex-col">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-sm font-medium text-gray-900 dark:text-gray-100">
            {burndownData.sprint_name}
          </h3>
          <p className="text-xs text-gray-500 dark:text-gray-400">
            {new Date(burndownData.start_date).toLocaleDateString()} -{' '}
            {new Date(burndownData.end_date).toLocaleDateString()}
          </p>
        </div>
        <span
          className={`inline-flex px-2 py-1 rounded text-xs ${getRiskColor(burndownData.risk_level)}`}
        >
          {burndownData.risk_level.replace(/_/g, ' ')}
        </span>
      </div>

      <div className="mb-4">
        <div className="flex items-center justify-between mb-2">
          <span className="text-sm text-gray-600 dark:text-gray-400">Progress</span>
          <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
            {burndownData.completed_points} / {burndownData.total_points} points (
            {completionPercentage.toFixed(0)}%)
          </span>
        </div>
        <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
          <div
            className="bg-blue-600 h-2 rounded-full transition-all"
            style={{ width: `${Math.min(completionPercentage, 100)}%` }}
          />
        </div>
      </div>

      <div className="flex-1 min-h-0">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={burndownData.burndown_points}>
            <XAxis
              dataKey="date"
              stroke="#9CA3AF"
              style={{ fontSize: '12px' }}
              tickFormatter={(date) =>
                new Date(date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
              }
            />
            <YAxis stroke="#9CA3AF" style={{ fontSize: '12px' }} />
            <Tooltip
              contentStyle={{
                backgroundColor: 'rgba(0, 0, 0, 0.8)',
                border: 'none',
                borderRadius: '8px',
                color: '#fff',
              }}
            />
            <Legend />
            <Line
              type="monotone"
              dataKey="ideal"
              stroke="#9CA3AF"
              strokeDasharray="5 5"
              name="Ideal"
            />
            <Line type="monotone" dataKey="actual" stroke="#3B82F6" strokeWidth={2} name="Actual" />
          </LineChart>
        </ResponsiveContainer>
      </div>

      {healthData && (
        <button
          onClick={() => onOpenModal?.('sprint-detail', { sprintId: currentSprintId })}
          className="mt-4 text-sm text-blue-600 dark:text-blue-400 hover:underline text-center"
        >
          View Sprint Details
        </button>
      )}
    </div>
  );
}
