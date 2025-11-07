import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { UnresolvedIdentity, Engineer } from '../api/types';
import { api } from '../api/client';

interface AssignModalProps {
  unresolved: UnresolvedIdentity;
  engineers: Engineer[];
  onClose: () => void;
}

export function AssignModal({ unresolved, engineers, onClose }: AssignModalProps) {
  const queryClient = useQueryClient();
  const [mode, setMode] = useState<'existing' | 'new'>('existing');
  const [selectedEngineerId, setSelectedEngineerId] = useState('');
  const [newEngineerName, setNewEngineerName] = useState('');
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: async () => {
      if (mode === 'existing') {
        if (!selectedEngineerId) {
          throw new Error('Please select an engineer');
        }
        return api.resolveIdentity(unresolved.id, selectedEngineerId);
      } else {
        if (!newEngineerName.trim()) {
          throw new Error('Please enter a name');
        }
        return api.resolveIdentity(unresolved.id, null, newEngineerName.trim());
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['unresolved'] });
      queryClient.invalidateQueries({ queryKey: ['suggestions'] });
      queryClient.invalidateQueries({ queryKey: ['engineers'] });
      onClose();
    },
    onError: (err: Error) => {
      setError(err.message);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    mutation.mutate();
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-lg w-full">
        <h2 className="text-xl font-bold mb-4 text-gray-900 dark:text-gray-100">
          Assign Identity: {unresolved.source} / {unresolved.identifier}
        </h2>

        <form onSubmit={handleSubmit}>
          <div className="space-y-4">
            <div>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  value="existing"
                  checked={mode === 'existing'}
                  onChange={() => setMode('existing')}
                  className="w-4 h-4"
                />
                <span className="text-gray-900 dark:text-gray-100">Assign to existing person</span>
              </label>

              {mode === 'existing' && (
                <div className="mt-2 ml-6">
                  <select
                    value={selectedEngineerId}
                    onChange={(e) => setSelectedEngineerId(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="">Select an engineer...</option>
                    {engineers.map((eng) => (
                      <option key={eng.id} value={eng.id}>
                        {eng.name}
                      </option>
                    ))}
                  </select>
                </div>
              )}
            </div>

            <div>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  value="new"
                  checked={mode === 'new'}
                  onChange={() => setMode('new')}
                  className="w-4 h-4"
                />
                <span className="text-gray-900 dark:text-gray-100">Create new person</span>
              </label>

              {mode === 'new' && (
                <div className="mt-2 ml-6">
                  <input
                    type="text"
                    value={newEngineerName}
                    onChange={(e) => setNewEngineerName(e.target.value)}
                    placeholder="Enter name..."
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              )}
            </div>
          </div>

          {error && (
            <div className="mt-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-red-800 dark:text-red-200">
              {error}
            </div>
          )}

          <div className="mt-6 flex gap-3 justify-end">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={mutation.isPending}
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 disabled:opacity-50"
            >
              {mutation.isPending ? 'Assigning...' : 'Assign'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
