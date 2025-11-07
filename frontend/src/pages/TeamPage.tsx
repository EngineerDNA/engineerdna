import { useState, useCallback } from 'react';
import { useQuery, useQueries } from '@tanstack/react-query';
import type { Engineer } from '../api/types';
import { api } from '../api/client';
import { EngineerCard } from '../components/EngineerCard';
import { EngineerForm } from '../components/EngineerForm';
import { UnresolvedPanel } from '../components/UnresolvedPanel';
import { ImportTeamModal } from '../components/ImportTeamModal';
import { ErrorBoundary } from '../components/error-boundary';

export function TeamPage() {
  const [showEngineerForm, setShowEngineerForm] = useState(false);
  const [editingEngineer, setEditingEngineer] = useState<Engineer | undefined>(undefined);
  const [showImportModal, setShowImportModal] = useState(false);
  const [expandedEngineers, setExpandedEngineers] = useState<Set<string>>(new Set());

  const {
    data: engineersData,
    isLoading: engineersLoading,
    error: engineersError,
  } = useQuery({
    queryKey: ['engineers'],
    queryFn: api.getEngineers,
    staleTime: 30000,
  });

  const { data: unresolvedData, isLoading: unresolvedLoading } = useQuery({
    queryKey: ['unresolved'],
    queryFn: api.getUnresolved,
    staleTime: 30000,
  });

  const { data: suggestionsData } = useQuery({
    queryKey: ['suggestions'],
    queryFn: api.getSuggestions,
    staleTime: 30000,
  });

  const engineers = engineersData?.engineers || [];
  const unresolved = unresolvedData?.unresolved || [];
  const suggestions = suggestionsData?.suggestions || [];

  const handleAddPerson = () => {
    setEditingEngineer(undefined);
    setShowEngineerForm(true);
  };

  const handleEditPerson = (engineer: Engineer) => {
    setEditingEngineer(engineer);
    setShowEngineerForm(true);
  };

  const handleCloseForm = () => {
    setShowEngineerForm(false);
    setEditingEngineer(undefined);
  };

  const handleActivityToggle = useCallback((engineerId: string, show: boolean) => {
    setExpandedEngineers((prev) => {
      const next = new Set(prev);
      if (show) {
        next.add(engineerId);
      } else {
        next.delete(engineerId);
      }
      return next;
    });
  }, []);

  const activityQueries = useQueries({
    queries: Array.from(expandedEngineers).map((engineerId) => ({
      queryKey: ['engineer-activity', engineerId],
      queryFn: () => api.getEngineerActivity(engineerId, 30),
      staleTime: 5 * 60 * 1000,
      retry: 1,
    })),
  });

  const activityByEngineerId = new Map(
    activityQueries.map((query, index) => {
      const engineerId = Array.from(expandedEngineers)[index];
      return [engineerId, query];
    })
  );

  if (engineersLoading || unresolvedLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400" />
      </div>
    );
  }

  if (engineersError) {
    return (
      <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded">
        <p className="text-red-800 dark:text-red-200">
          Failed to load team data: {(engineersError as Error).message}
        </p>
      </div>
    );
  }

  return (
    <ErrorBoundary>
      <div>
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Team</h1>
          <div className="flex gap-3">
            <button
              onClick={() => setShowImportModal(true)}
              className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 dark:bg-green-700 dark:hover:bg-green-800"
            >
              Import CSV
            </button>
            <button
              onClick={handleAddPerson}
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800"
            >
              + Add Person
            </button>
          </div>
        </div>

        {unresolved.length > 0 && (
          <ErrorBoundary>
            <UnresolvedPanel
              unresolved={unresolved}
              suggestions={suggestions}
              engineers={engineers}
            />
          </ErrorBoundary>
        )}

        {engineers.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-600 dark:text-gray-400 mb-4">No team members yet.</p>
            <button
              onClick={handleAddPerson}
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800"
            >
              Add Your First Team Member
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {engineers.map((engineer) => {
              const activityQuery = activityByEngineerId.get(engineer.id);
              return (
                <ErrorBoundary key={engineer.id}>
                  <EngineerCard
                    engineer={engineer}
                    onEdit={handleEditPerson}
                    activity={activityQuery?.data?.metrics}
                    isActivityLoading={activityQuery?.isLoading}
                    activityError={activityQuery?.isError}
                    onActivityToggle={handleActivityToggle}
                  />
                </ErrorBoundary>
              );
            })}
          </div>
        )}

        {showEngineerForm && (
          <ErrorBoundary>
            <EngineerForm engineer={editingEngineer} onClose={handleCloseForm} />
          </ErrorBoundary>
        )}

        {showImportModal && (
          <ErrorBoundary>
            <ImportTeamModal onClose={() => setShowImportModal(false)} />
          </ErrorBoundary>
        )}
      </div>
    </ErrorBoundary>
  );
}
