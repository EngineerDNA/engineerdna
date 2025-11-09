import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/client';
import type { WidgetProps } from './widget-registry';
import type { Engineer } from '../../api/types';

export function EngineerTableWidget({ config, onOpenModal }: WidgetProps) {
  const teamId = config.team_id as string | undefined;
  const columns = (config.columns as string[]) || ['name', 'role', 'team', 'status'];

  const [sortField, setSortField] = useState<keyof Engineer>('name');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');
  const [filterText, setFilterText] = useState('');

  const { data: engineersData, isLoading } = useQuery({
    queryKey: teamId ? ['team-members', teamId] : ['engineers'],
    queryFn: async () => {
      if (teamId) {
        const result = await api.getTeamMembers(teamId);
        return { engineers: result.members };
      }
      return api.getEngineers();
    },
    staleTime: 5 * 60 * 1000,
  });

  const engineers: Engineer[] = (engineersData as { engineers: Engineer[] })?.engineers || [];

  const filteredEngineers = engineers.filter((engineer: Engineer) =>
    engineer.name.toLowerCase().includes(filterText.toLowerCase())
  );

  const sortedEngineers = [...filteredEngineers].sort((a, b) => {
    const aVal = a[sortField];
    const bVal = b[sortField];
    const multiplier = sortDirection === 'asc' ? 1 : -1;

    if (typeof aVal === 'string' && typeof bVal === 'string') {
      return aVal.localeCompare(bVal) * multiplier;
    }
    if (typeof aVal === 'boolean' && typeof bVal === 'boolean') {
      return (aVal === bVal ? 0 : aVal ? 1 : -1) * multiplier;
    }
    return 0;
  });

  const handleSort = (field: keyof Engineer) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('asc');
    }
  };

  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="animate-pulse">
          <div className="h-10 bg-gray-300 dark:bg-gray-700 rounded mb-4" />
          <div className="space-y-3">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="h-12 bg-gray-300 dark:bg-gray-700 rounded" />
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full flex flex-col">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400">Engineers</h3>
        <input
          type="text"
          placeholder="Filter by name..."
          value={filterText}
          onChange={(e) => setFilterText(e.target.value)}
          className="px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />
      </div>

      <div className="flex-1 overflow-auto">
        <table className="w-full">
          <thead className="sticky top-0 bg-gray-50 dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700">
            <tr>
              {columns.includes('name') && (
                <th
                  onClick={() => handleSort('name')}
                  className="px-4 py-2 text-left text-xs font-medium text-gray-600 dark:text-gray-400 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800"
                >
                  <div className="flex items-center gap-1">
                    Name
                    {sortField === 'name' && <span>{sortDirection === 'asc' ? '↑' : '↓'}</span>}
                  </div>
                </th>
              )}
              {columns.includes('role') && (
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-600 dark:text-gray-400">
                  Role
                </th>
              )}
              {columns.includes('team') && (
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-600 dark:text-gray-400">
                  Team
                </th>
              )}
              {columns.includes('status') && (
                <th
                  onClick={() => handleSort('active')}
                  className="px-4 py-2 text-left text-xs font-medium text-gray-600 dark:text-gray-400 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800"
                >
                  <div className="flex items-center gap-1">
                    Status
                    {sortField === 'active' && <span>{sortDirection === 'asc' ? '↑' : '↓'}</span>}
                  </div>
                </th>
              )}
              <th className="px-4 py-2 text-right text-xs font-medium text-gray-600 dark:text-gray-400">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
            {sortedEngineers.length === 0 ? (
              <tr>
                <td
                  colSpan={columns.length + 1}
                  className="px-4 py-8 text-center text-gray-500 dark:text-gray-400"
                >
                  No engineers found
                </td>
              </tr>
            ) : (
              sortedEngineers.map((engineer) => (
                <tr
                  key={engineer.id}
                  className="hover:bg-gray-50 dark:hover:bg-gray-900 transition-colors"
                >
                  {columns.includes('name') && (
                    <td className="px-4 py-3 text-sm text-gray-900 dark:text-gray-100">
                      {engineer.name}
                    </td>
                  )}
                  {columns.includes('role') && (
                    <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">
                      {engineer.role_id || '-'}
                    </td>
                  )}
                  {columns.includes('team') && (
                    <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">
                      {engineer.manager || '-'}
                    </td>
                  )}
                  {columns.includes('status') && (
                    <td className="px-4 py-3 text-sm">
                      <span
                        className={`inline-flex px-2 py-1 rounded text-xs ${
                          engineer.active
                            ? 'bg-green-100 dark:bg-green-900/20 text-green-800 dark:text-green-200'
                            : 'bg-gray-100 dark:bg-gray-900/20 text-gray-800 dark:text-gray-200'
                        }`}
                      >
                        {engineer.active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                  )}
                  <td className="px-4 py-3 text-right text-sm">
                    <button
                      onClick={() => onOpenModal?.('engineer-detail', { engineerId: engineer.id })}
                      className="text-blue-600 dark:text-blue-400 hover:underline mr-3"
                    >
                      View
                    </button>
                    <button
                      onClick={() => onOpenModal?.('engineer-edit', { engineerId: engineer.id })}
                      className="text-gray-600 dark:text-gray-400 hover:underline"
                    >
                      Edit
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700 text-xs text-gray-500 dark:text-gray-400">
        Showing {sortedEngineers.length} of {engineers.length} engineers
      </div>
    </div>
  );
}
