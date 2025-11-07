import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { Engineer } from '../api/types';
import { api } from '../api/client';

interface EngineerFormProps {
  engineer?: Engineer;
  onClose: () => void;
}

interface IdentifierRow {
  id: string;
  source: string;
  identifier: string;
}

export function EngineerForm({ engineer, onClose }: EngineerFormProps) {
  const queryClient = useQueryClient();
  const [name, setName] = useState(engineer?.name || '');
  const [email, setEmail] = useState(engineer?.email || '');
  const [manager, setManager] = useState(engineer?.manager || '');
  const [identifiers, setIdentifiers] = useState<Array<IdentifierRow>>(
    Object.entries(engineer?.identifiers || {}).map(([source, identifier]) => ({
      id: crypto.randomUUID(),
      source,
      identifier,
    }))
  );
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: async (data: Partial<Engineer>) => {
      if (engineer) {
        return api.updateEngineer(engineer.id, data);
      } else {
        return api.createEngineer(data);
      }
    },
    onSuccess: () => {
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

    if (!name.trim()) {
      setError('Name is required');
      return;
    }

    // Validate identifiers - check for incomplete rows
    const identifiersObj: Record<string, string> = {};
    const hasIncompleteRows = identifiers.some(
      ({ source, identifier }) => (source && !identifier) || (!source && identifier)
    );

    if (hasIncompleteRows) {
      setError('All identifier rows must have both source and identifier, or be completely empty');
      return;
    }

    identifiers.forEach(({ source, identifier }) => {
      if (source && identifier) {
        identifiersObj[source] = identifier;
      }
    });

    mutation.mutate({
      name: name.trim(),
      email: email.trim(),
      manager: manager.trim(),
      identifiers: identifiersObj,
    });
  };

  const addIdentifier = () => {
    setIdentifiers([...identifiers, { id: crypto.randomUUID(), source: '', identifier: '' }]);
  };

  const removeIdentifier = (id: string) => {
    setIdentifiers(identifiers.filter((item) => item.id !== id));
  };

  const updateIdentifier = (id: string, field: 'source' | 'identifier', value: string) => {
    setIdentifiers(
      identifiers.map((item) => (item.id === id ? { ...item, [field]: value } : item))
    );
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <h2 className="text-2xl font-bold mb-4 text-gray-900 dark:text-gray-100">
          {engineer ? 'Edit Engineer' : 'Add New Engineer'}
        </h2>

        <form onSubmit={handleSubmit}>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Name <span className="text-red-600 dark:text-red-400">*</span>
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Manager
              </label>
              <input
                type="text"
                value={manager}
                onChange={(e) => setManager(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400"
              />
            </div>

            <div>
              <div className="flex items-center justify-between mb-2">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                  Identifiers
                </label>
                <button
                  type="button"
                  onClick={addIdentifier}
                  className="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300"
                >
                  + Add Identifier
                </button>
              </div>

              <div className="space-y-2">
                {identifiers.map((item) => (
                  <div key={item.id} className="flex gap-2">
                    <input
                      type="text"
                      placeholder="Source (e.g., github)"
                      value={item.source}
                      onChange={(e) => updateIdentifier(item.id, 'source', e.target.value)}
                      className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400"
                    />
                    <input
                      type="text"
                      placeholder="Identifier (e.g., username)"
                      value={item.identifier}
                      onChange={(e) => updateIdentifier(item.id, 'identifier', e.target.value)}
                      className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400"
                    />
                    <button
                      type="button"
                      onClick={() => removeIdentifier(item.id)}
                      className="px-3 py-2 text-red-600 dark:text-red-400 hover:text-red-700 dark:hover:text-red-300"
                    >
                      Remove
                    </button>
                  </div>
                ))}
              </div>
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
              className="px-4 py-2 text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={mutation.isPending}
              className="px-4 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded hover:bg-blue-700 dark:hover:bg-blue-800 disabled:opacity-50 transition-colors"
            >
              {mutation.isPending ? 'Saving...' : 'Save'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
