import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';

export function IgnoredIdentitiesList() {
  const queryClient = useQueryClient();

  const { data: ignoredData, isLoading: ignoredLoading } = useQuery({
    queryKey: ['ignored-identities'],
    queryFn: api.getIgnoredIdentities,
  });

  const unignoreMutation = useMutation({
    mutationFn: (unresolvedId: string) => api.ignoreIdentity(unresolvedId, false),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ignored-identities'] });
      queryClient.invalidateQueries({ queryKey: ['unresolved'] });
    },
  });

  const ignored = ignoredData?.ignored || [];

  return (
    <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
      <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">
        Ignored Identities
      </h3>
      <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
        These identities have been ignored and will not appear in the Team page. You can restore
        them if needed.
      </p>

      {ignoredLoading ? (
        <div className="text-gray-500 dark:text-gray-400">Loading...</div>
      ) : ignored.length === 0 ? (
        <div className="text-gray-500 dark:text-gray-400 text-sm">No ignored identities</div>
      ) : (
        <div className="space-y-2">
          {ignored.map((identity) => (
            <div
              key={identity.id}
              className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 p-3 bg-gray-50 dark:bg-gray-700/50 rounded border border-gray-200 dark:border-gray-600"
            >
              <div className="flex-1 min-w-0">
                <div className="font-mono text-sm text-gray-900 dark:text-gray-100 truncate">
                  {identity.identifier}
                </div>
                <div className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                  from {identity.source} • {identity.event_count} events • ignored since{' '}
                  {new Date(identity.first_seen).toLocaleDateString()}
                </div>
              </div>
              <button
                onClick={() => unignoreMutation.mutate(identity.id)}
                disabled={unignoreMutation.isPending}
                className="px-3 py-1 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors whitespace-nowrap"
              >
                {unignoreMutation.isPending ? 'Restoring...' : 'Restore'}
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
