import { useState, useMemo, memo } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { UnresolvedIdentity, MatchSuggestion, Engineer } from '../api/types';
import { api } from '../api/client';
import { HIGH_CONFIDENCE_THRESHOLD } from '../constants';
import { AssignModal } from './AssignModal';

interface UnresolvedPanelProps {
  unresolved: UnresolvedIdentity[];
  suggestions: MatchSuggestion[];
  engineers: Engineer[];
}

const ITEMS_PER_PAGE = 10;

export function UnresolvedPanel({ unresolved, suggestions, engineers }: UnresolvedPanelProps) {
  const [assignModalOpen, setAssignModalOpen] = useState(false);
  const [selectedUnresolved, setSelectedUnresolved] = useState<UnresolvedIdentity | null>(null);
  const [visibleCount, setVisibleCount] = useState(ITEMS_PER_PAGE);

  const openAssignModal = (identity: UnresolvedIdentity) => {
    setSelectedUnresolved(identity);
    setAssignModalOpen(true);
  };

  const closeAssignModal = () => {
    setSelectedUnresolved(null);
    setAssignModalOpen(false);
  };

  const suggestionMap = useMemo(() => {
    const map = new Map<string, MatchSuggestion>();
    suggestions.forEach((s) => map.set(s.unresolved_id, s));
    return map;
  }, [suggestions]);

  const visibleUnresolved = useMemo(
    () => unresolved.slice(0, visibleCount),
    [unresolved, visibleCount]
  );

  const hasMore = unresolved.length > visibleCount;

  if (unresolved.length === 0) {
    return null;
  }

  return (
    <>
      <div className="bg-yellow-50 dark:bg-yellow-900/20 border-l-4 border-yellow-400 p-4 mb-6">
        <div className="flex items-start">
          <div className="flex-shrink-0">
            <svg className="h-5 w-5 text-yellow-400" fill="currentColor" viewBox="0 0 20 20">
              <path
                fillRule="evenodd"
                d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                clipRule="evenodd"
              />
            </svg>
          </div>
          <div className="ml-3 flex-1">
            <h3 className="text-sm font-medium text-yellow-800 dark:text-yellow-200">
              Unresolved Identities ({unresolved.length})
            </h3>
            <div className="mt-2 text-sm text-yellow-700 dark:text-yellow-300">
              <p className="mb-2">
                The following identities from your data sources need to be assigned to team members:
              </p>
              <div className="space-y-3">
                {visibleUnresolved.map((identity) => {
                  const suggestion = suggestionMap.get(identity.id);
                  return (
                    <UnresolvedIdentityRow
                      key={identity.id}
                      identity={identity}
                      suggestion={suggestion}
                      onAssign={() => openAssignModal(identity)}
                    />
                  );
                })}
              </div>
              {hasMore && (
                <button
                  onClick={() => setVisibleCount((prev) => prev + ITEMS_PER_PAGE)}
                  className="mt-4 px-4 py-2 text-sm font-medium text-yellow-800 dark:text-yellow-200 bg-yellow-100 dark:bg-yellow-900/40 rounded hover:bg-yellow-200 dark:hover:bg-yellow-900/60 transition-colors"
                >
                  Show {Math.min(ITEMS_PER_PAGE, unresolved.length - visibleCount)} more (
                  {unresolved.length - visibleCount} remaining)
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      {assignModalOpen && selectedUnresolved && (
        <AssignModal
          unresolved={selectedUnresolved}
          engineers={engineers}
          onClose={closeAssignModal}
        />
      )}
    </>
  );
}

interface UnresolvedIdentityRowProps {
  identity: UnresolvedIdentity;
  suggestion?: MatchSuggestion;
  onAssign: () => void;
}

const UnresolvedIdentityRow = memo(function UnresolvedIdentityRow({
  identity,
  suggestion,
  onAssign,
}: UnresolvedIdentityRowProps) {
  const queryClient = useQueryClient();
  const [error, setError] = useState<string | null>(null);

  const acceptMutation = useMutation({
    mutationFn: () => {
      if (!suggestion) {
        throw new Error('No suggestion available');
      }
      return api.resolveIdentity(identity.id, suggestion.engineer_id);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['unresolved'] });
      queryClient.invalidateQueries({ queryKey: ['suggestions'] });
      queryClient.invalidateQueries({ queryKey: ['engineers'] });
    },
    onError: (err: Error) => {
      setError(err.message);
    },
  });

  const ignoreMutation = useMutation({
    mutationFn: () => api.ignoreIdentity(identity.id, true),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['unresolved'] });
      queryClient.invalidateQueries({ queryKey: ['suggestions'] });
    },
    onError: (err: Error) => {
      setError(err.message);
    },
  });

  return (
    <div className="bg-white dark:bg-gray-800 p-3 rounded border border-yellow-200 dark:border-yellow-700">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="font-medium text-gray-900 dark:text-gray-100">
            {identity.source}: {identity.identifier}
          </div>
          <div className="text-xs text-gray-600 dark:text-gray-400 mt-1">
            {identity.event_count} events since {new Date(identity.first_seen).toLocaleDateString()}
          </div>

          {suggestion && (
            <div className="mt-2 flex items-center gap-2">
              <span className="text-xs bg-blue-100 dark:bg-blue-900/40 text-blue-800 dark:text-blue-200 px-2 py-1 rounded">
                AI Suggestion: {suggestion.engineer_name} ({Math.round(suggestion.confidence * 100)}
                %)
              </span>
              <span className="text-xs text-gray-600 dark:text-gray-400">{suggestion.reason}</span>
            </div>
          )}

          {error && <div className="mt-2 text-xs text-red-600 dark:text-red-400">{error}</div>}
        </div>

        <div className="flex gap-2 ml-4">
          {suggestion && suggestion.confidence >= HIGH_CONFIDENCE_THRESHOLD && (
            <button
              onClick={() => acceptMutation.mutate()}
              disabled={acceptMutation.isPending}
              className="px-3 py-1 text-sm bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50"
            >
              {acceptMutation.isPending ? 'Accepting...' : 'Accept'}
            </button>
          )}
          <button
            onClick={onAssign}
            className="px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Assign
          </button>
          <button
            onClick={() => ignoreMutation.mutate()}
            disabled={ignoreMutation.isPending}
            className="px-3 py-1 text-sm bg-gray-400 text-white rounded hover:bg-gray-500 disabled:opacity-50"
          >
            {ignoreMutation.isPending ? 'Ignoring...' : 'Ignore'}
          </button>
        </div>
      </div>
    </div>
  );
});
