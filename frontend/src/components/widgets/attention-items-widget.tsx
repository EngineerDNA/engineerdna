import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/client';
import type { WidgetProps } from './widget-registry';
import type { AttentionItem } from '../../api/types';

export function AttentionItemsWidget({ config, onOpenModal }: WidgetProps) {
  const teamId = config.team_id as string | undefined;
  const limit = (config.limit as number) || 5;

  const { data, isLoading } = useQuery({
    queryKey: ['briefing', teamId],
    queryFn: () => api.getWeeklyBriefing(teamId || ''),
    enabled: !!teamId,
    staleTime: 5 * 60 * 1000,
  });

  const attentionItems = data?.needs_attention?.slice(0, limit) || [];

  if (!teamId) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="flex items-center justify-center h-32 text-gray-500 dark:text-gray-400">
          Configure team_id to view attention items
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="animate-pulse space-y-3">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="h-16 bg-gray-300 dark:bg-gray-700 rounded" />
          ))}
        </div>
      </div>
    );
  }

  const getSeverityColor = (severity: AttentionItem['severity']) => {
    return severity === 'critical'
      ? 'bg-red-100 dark:bg-red-900/20 border-red-300 dark:border-red-800 text-red-900 dark:text-red-200'
      : 'bg-yellow-100 dark:bg-yellow-900/20 border-yellow-300 dark:border-yellow-800 text-yellow-900 dark:text-yellow-200';
  };

  const getSeverityIcon = (severity: AttentionItem['severity']) => {
    return severity === 'critical' ? (
      <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
        <path
          fillRule="evenodd"
          d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
          clipRule="evenodd"
        />
      </svg>
    ) : (
      <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
        <path
          fillRule="evenodd"
          d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
          clipRule="evenodd"
        />
      </svg>
    );
  };

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full flex flex-col">
      <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-4">Needs Attention</h3>

      <div className="flex-1 overflow-auto space-y-3">
        {attentionItems.length === 0 ? (
          <div className="flex items-center justify-center h-32 text-gray-500 dark:text-gray-400">
            No items need attention
          </div>
        ) : (
          attentionItems.map((item, idx) => (
            <div
              key={idx}
              className={`border rounded-lg p-3 cursor-pointer hover:shadow-md transition-shadow ${getSeverityColor(item.severity)}`}
              onClick={() => onOpenModal?.('attention-detail', { item })}
            >
              <div className="flex items-start gap-3">
                <div className="flex-shrink-0 mt-0.5">{getSeverityIcon(item.severity)}</div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium mb-1">{item.engineer_name}</p>
                  <p className="text-xs mb-2">{item.issue}</p>
                  {item.evidence.length > 0 && (
                    <p className="text-xs opacity-75">
                      {item.evidence.length} evidence item{item.evidence.length > 1 ? 's' : ''}
                    </p>
                  )}
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {attentionItems.length > 0 && (
        <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700 text-center">
          <button
            onClick={() => onOpenModal?.('briefing-detail', { briefingData: data })}
            className="text-sm text-blue-600 dark:text-blue-400 hover:underline"
          >
            View Full Briefing
          </button>
        </div>
      )}
    </div>
  );
}
