import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/client';
import type { WidgetProps } from './widget-registry';

export function TeamSelectorWidget({ widgetId, config, onUpdateWidget }: WidgetProps) {
  const defaultTeamId = config.default_team_id as string | undefined;

  const [selectedTeamId, setSelectedTeamId] = useState(defaultTeamId || '');
  const [startDate, setStartDate] = useState(() => {
    const date = new Date();
    date.setDate(date.getDate() - 7);
    return date.toISOString().split('T')[0];
  });
  const [endDate, setEndDate] = useState(() => new Date().toISOString().split('T')[0]);

  const { data: teamsData, isLoading } = useQuery({
    queryKey: ['teams'],
    queryFn: api.getTeams,
    staleTime: 5 * 60 * 1000,
  });

  const teams = teamsData?.teams || [];

  const handleApply = () => {
    if (onUpdateWidget) {
      onUpdateWidget(widgetId, {
        selected_team_id: selectedTeamId,
        start_date: startDate,
        end_date: endDate,
      });
    }
  };

  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="animate-pulse">
          <div className="h-10 bg-gray-300 dark:bg-gray-700 rounded mb-3" />
          <div className="h-10 bg-gray-300 dark:bg-gray-700 rounded mb-3" />
          <div className="h-10 bg-gray-300 dark:bg-gray-700 rounded" />
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
      <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-4">
        Team & Date Selection
      </h3>

      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Team
          </label>
          <select
            value={selectedTeamId}
            onChange={(e) => setSelectedTeamId(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="">Select a team...</option>
            {teams.map((team) => (
              <option key={team.id} value={team.id}>
                {team.name}
              </option>
            ))}
          </select>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Start Date
            </label>
            <input
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              End Date
            </label>
            <input
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
        </div>

        <div className="flex gap-2">
          <button
            onClick={handleApply}
            disabled={!selectedTeamId}
            className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
          >
            Apply Selection
          </button>
          <button
            onClick={() => {
              setSelectedTeamId(defaultTeamId || '');
              setStartDate(() => {
                const date = new Date();
                date.setDate(date.getDate() - 7);
                return date.toISOString().split('T')[0];
              });
              setEndDate(new Date().toISOString().split('T')[0]);
            }}
            className="px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-900 transition-colors"
          >
            Reset
          </button>
        </div>
      </div>
    </div>
  );
}
