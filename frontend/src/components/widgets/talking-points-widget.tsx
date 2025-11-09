import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/client';
import type { WidgetProps } from './widget-registry';

export function TalkingPointsWidget({ config }: WidgetProps) {
  const teamId = config.team_id as string | undefined;
  const weekStart = config.week_start_date as string | undefined;

  const { data, isLoading } = useQuery({
    queryKey: ['briefing', teamId, weekStart],
    queryFn: () => api.getWeeklyBriefing(teamId || '', weekStart),
    enabled: !!teamId,
    staleTime: 5 * 60 * 1000,
  });

  const [expandedSections, setExpandedSections] = useState<Set<number>>(new Set());

  const toggleSection = (index: number) => {
    const newExpanded = new Set(expandedSections);
    if (newExpanded.has(index)) {
      newExpanded.delete(index);
    } else {
      newExpanded.add(index);
    }
    setExpandedSections(newExpanded);
  };

  if (!teamId) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="flex items-center justify-center h-32 text-gray-500 dark:text-gray-400">
          Configure team_id to view talking points
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="animate-pulse space-y-3">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="h-12 bg-gray-300 dark:bg-gray-700 rounded" />
          ))}
        </div>
      </div>
    );
  }

  if (!data?.talking_points) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-4">
          Talking Points
        </h3>
        <div className="flex items-center justify-center h-32 text-gray-500 dark:text-gray-400">
          No talking points available
        </div>
      </div>
    );
  }

  const talkingPointsSections = data.talking_points
    .split('\n\n')
    .filter((section) => section.trim());

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full flex flex-col">
      <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-4">Talking Points</h3>

      <div className="flex-1 overflow-auto space-y-3">
        {talkingPointsSections.map((section, idx) => {
          const lines = section.split('\n');
          const title = lines[0];
          const content = lines.slice(1).join('\n');
          const isExpanded = expandedSections.has(idx);

          return (
            <div key={idx} className="border border-gray-200 dark:border-gray-700 rounded-lg">
              <button
                onClick={() => toggleSection(idx)}
                className="w-full px-4 py-3 flex items-center justify-between text-left hover:bg-gray-50 dark:hover:bg-gray-900 transition-colors rounded-lg"
              >
                <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
                  {title}
                </span>
                <svg
                  className={`w-5 h-5 text-gray-500 dark:text-gray-400 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M19 9l-7 7-7-7"
                  />
                </svg>
              </button>

              {isExpanded && content && (
                <div className="px-4 pb-3">
                  <div className="text-sm text-gray-700 dark:text-gray-300 whitespace-pre-line pl-4 border-l-2 border-blue-500 dark:border-blue-400">
                    {content}
                  </div>
                </div>
              )}
            </div>
          );
        })}
      </div>

      <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700 text-xs text-gray-500 dark:text-gray-400">
        Week of {new Date(data.week_start).toLocaleDateString()}
      </div>
    </div>
  );
}
