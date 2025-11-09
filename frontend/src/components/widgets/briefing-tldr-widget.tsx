import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../../api/client';
import type { WidgetProps } from './widget-registry';

export function BriefingTldrWidget({ config, onOpenModal }: WidgetProps) {
  const teamId = config.team_id as string | undefined;
  const weekStart = config.week_start_date as string | undefined;

  const { data, isLoading, error } = useQuery({
    queryKey: ['briefing', teamId, weekStart],
    queryFn: () => api.getWeeklyBriefing(teamId || '', weekStart),
    enabled: !!teamId,
    staleTime: 5 * 60 * 1000,
  });

  const queryClient = useQueryClient();
  const generateMutation = useMutation({
    mutationFn: () => api.generateWeeklyBriefing(teamId || '', weekStart),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['briefing', teamId, weekStart] });
    },
  });

  if (!teamId) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="flex items-center justify-center h-32 text-gray-500 dark:text-gray-400">
          Configure team_id to view briefing
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-300 dark:bg-gray-700 rounded w-1/3 mb-4" />
          <div className="h-20 bg-gray-300 dark:bg-gray-700 rounded mb-4" />
          <div className="h-10 bg-gray-300 dark:bg-gray-700 rounded w-32" />
        </div>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-4">
          Team Briefing TL;DR
        </h3>
        <div className="text-center py-8">
          <p className="text-gray-500 dark:text-gray-400 mb-4">No briefing available</p>
          <button
            onClick={() => generateMutation.mutate()}
            disabled={generateMutation.isPending}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-400"
          >
            {generateMutation.isPending ? 'Generating...' : 'Generate Briefing'}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full flex flex-col">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400">
          Team Briefing TL;DR
        </h3>
        <button
          onClick={() => generateMutation.mutate()}
          disabled={generateMutation.isPending}
          className="text-xs px-3 py-1 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-400"
        >
          {generateMutation.isPending ? 'Regenerating...' : 'Regenerate'}
        </button>
      </div>

      <div className="flex-1 overflow-auto">
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4 mb-4">
          <p className="text-sm text-gray-900 dark:text-gray-100 leading-relaxed">{data.tldr}</p>
        </div>

        <p className="text-xs text-gray-500 dark:text-gray-400 mb-3">
          Week of {new Date(data.week_start).toLocaleDateString()} -{' '}
          {new Date(data.week_end).toLocaleDateString()}
        </p>

        <button
          onClick={() => onOpenModal?.('briefing-detail', { briefingData: data })}
          className="text-sm text-blue-600 dark:text-blue-400 hover:underline"
        >
          View Full Briefing
        </button>
      </div>
    </div>
  );
}
