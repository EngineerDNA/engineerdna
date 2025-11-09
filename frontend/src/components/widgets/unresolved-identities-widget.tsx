import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../../api/client';
import type { WidgetProps } from './widget-registry';

export function UnresolvedIdentitiesWidget({ config }: WidgetProps) {
  const limit = (config.limit as number) || 10;

  const { data: unresolvedData, isLoading } = useQuery({
    queryKey: ['unresolved-identities'],
    queryFn: api.getUnresolved,
    staleTime: 60 * 1000,
  });

  const { data: engineersData } = useQuery({
    queryKey: ['engineers'],
    queryFn: api.getEngineers,
    staleTime: 5 * 60 * 1000,
  });

  const { data: suggestionsData } = useQuery({
    queryKey: ['identity-suggestions'],
    queryFn: api.getSuggestions,
    staleTime: 5 * 60 * 1000,
  });

  const queryClient = useQueryClient();

  const resolveMutation = useMutation({
    mutationFn: ({
      unresolvedId,
      engineerId,
    }: {
      unresolvedId: string;
      engineerId: string | null;
    }) => api.resolveIdentity(unresolvedId, engineerId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['unresolved-identities'] });
      queryClient.invalidateQueries({ queryKey: ['identity-suggestions'] });
    },
  });

  const unresolved = unresolvedData?.unresolved?.slice(0, limit) || [];
  const engineers = engineersData?.engineers || [];
  const suggestions = suggestionsData?.suggestions || [];

  const [selectedEngineer, setSelectedEngineer] = useState<Record<string, string>>({});

  const getSuggestionForIdentity = (unresolvedId: string) => {
    return suggestions.find((s) => s.unresolved_id === unresolvedId);
  };

  const handleResolve = (unresolvedId: string) => {
    const engineerId = selectedEngineer[unresolvedId];
    if (engineerId) {
      resolveMutation.mutate({ unresolvedId, engineerId });
    }
  };

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

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full flex flex-col">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400">
          Unresolved Identities
        </h3>
        {unresolved.length > 0 && (
          <span className="inline-flex items-center px-2 py-1 rounded-full text-xs bg-yellow-100 dark:bg-yellow-900/20 text-yellow-800 dark:text-yellow-200">
            {unresolved.length}
          </span>
        )}
      </div>

      <div className="flex-1 overflow-auto space-y-3">
        {unresolved.length === 0 ? (
          <div className="flex items-center justify-center h-32 text-gray-500 dark:text-gray-400">
            All identities resolved
          </div>
        ) : (
          unresolved.map((identity) => {
            const suggestion = getSuggestionForIdentity(identity.id);

            return (
              <div
                key={identity.id}
                className="border border-gray-200 dark:border-gray-700 rounded-lg p-3"
              >
                <div className="flex items-start justify-between mb-2">
                  <div className="flex-1">
                    <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                      {identity.identifier}
                    </p>
                    <p className="text-xs text-gray-500 dark:text-gray-400">
                      {identity.source} • {identity.event_count} events
                    </p>
                  </div>
                  {suggestion && (
                    <div className="ml-2">
                      <span className="inline-flex items-center px-2 py-1 rounded text-xs bg-blue-100 dark:bg-blue-900/20 text-blue-800 dark:text-blue-200">
                        {Math.round(suggestion.confidence * 100)}% match
                      </span>
                    </div>
                  )}
                </div>

                {suggestion && (
                  <p className="text-xs text-gray-600 dark:text-gray-400 mb-2">
                    Suggested: {suggestion.engineer_name} ({suggestion.reason})
                  </p>
                )}

                <div className="flex items-center gap-2">
                  <select
                    value={selectedEngineer[identity.id] || suggestion?.engineer_id || ''}
                    onChange={(e) =>
                      setSelectedEngineer({ ...selectedEngineer, [identity.id]: e.target.value })
                    }
                    className="flex-1 text-sm px-2 py-1 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100"
                  >
                    <option value="">Select engineer...</option>
                    {engineers.map((eng) => (
                      <option key={eng.id} value={eng.id}>
                        {eng.name}
                      </option>
                    ))}
                  </select>
                  <button
                    onClick={() => handleResolve(identity.id)}
                    disabled={!selectedEngineer[identity.id] && !suggestion}
                    className="text-sm px-3 py-1 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed"
                  >
                    Resolve
                  </button>
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
