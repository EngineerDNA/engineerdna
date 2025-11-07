import { useState, useMemo } from 'react';
import type { EngineerPerformanceRow } from '../../api/types';

interface TableProps {
  data: EngineerPerformanceRow[];
  isLoading?: boolean;
}

type SortField = keyof EngineerPerformanceRow;
type SortOrder = 'asc' | 'desc';

/**
 * Table displays engineer performance data in a sortable table format.
 * Supports client-side sorting on all columns with visual indicators.
 * Includes loading state with skeleton and empty state handling.
 */
export function Table({ data, isLoading = false }: TableProps) {
  const [sortField, setSortField] = useState<SortField>('total_score');
  const [sortOrder, setSortOrder] = useState<SortOrder>('desc');

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortOrder('desc');
    }
  };

  const sortedData = useMemo(() => {
    if (!data) return [];
    return [...data].sort((a, b) => {
      const aValue = a[sortField];
      const bValue = b[sortField];

      if (typeof aValue === 'string' && typeof bValue === 'string') {
        return sortOrder === 'asc' ? aValue.localeCompare(bValue) : bValue.localeCompare(aValue);
      }

      const aNum = typeof aValue === 'number' ? aValue : 0;
      const bNum = typeof bValue === 'number' ? bValue : 0;
      return sortOrder === 'asc' ? aNum - bNum : bNum - aNum;
    });
  }, [data, sortField, sortOrder]);

  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full overflow-auto">
        <div className="animate-pulse">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="h-12 bg-gray-300 dark:bg-gray-700 rounded mb-2" />
          ))}
        </div>
      </div>
    );
  }

  if (!data || data.length === 0) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="flex items-center justify-center h-64 text-gray-500 dark:text-gray-400">
          No data available
        </div>
      </div>
    );
  }

  const SortIcon = ({ field }: { field: SortField }) => {
    if (sortField !== field) {
      return (
        <svg
          className="w-4 h-4 text-gray-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4"
          />
        </svg>
      );
    }
    return sortOrder === 'asc' ? (
      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 15l7-7 7 7" />
      </svg>
    ) : (
      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
      </svg>
    );
  };

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full overflow-auto">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-900">
            <tr>
              <th
                className="px-4 py-3 text-left text-xs font-medium text-gray-700 dark:text-gray-300 uppercase tracking-wider cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800"
                onClick={() => handleSort('engineer_name')}
              >
                <div className="flex items-center gap-1">
                  Engineer
                  <SortIcon field="engineer_name" />
                </div>
              </th>
              <th
                className="px-4 py-3 text-left text-xs font-medium text-gray-700 dark:text-gray-300 uppercase tracking-wider cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800"
                onClick={() => handleSort('total_score')}
              >
                <div className="flex items-center gap-1">
                  Total
                  <SortIcon field="total_score" />
                </div>
              </th>
              <th
                className="px-4 py-3 text-left text-xs font-medium text-gray-700 dark:text-gray-300 uppercase tracking-wider cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800"
                onClick={() => handleSort('throughput_score')}
              >
                <div className="flex items-center gap-1">
                  Throughput
                  <SortIcon field="throughput_score" />
                </div>
              </th>
              <th
                className="px-4 py-3 text-left text-xs font-medium text-gray-700 dark:text-gray-300 uppercase tracking-wider cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800"
                onClick={() => handleSort('quality_score')}
              >
                <div className="flex items-center gap-1">
                  Quality
                  <SortIcon field="quality_score" />
                </div>
              </th>
              <th
                className="px-4 py-3 text-left text-xs font-medium text-gray-700 dark:text-gray-300 uppercase tracking-wider cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800"
                onClick={() => handleSort('prs_merged')}
              >
                <div className="flex items-center gap-1">
                  PRs
                  <SortIcon field="prs_merged" />
                </div>
              </th>
              <th
                className="px-4 py-3 text-left text-xs font-medium text-gray-700 dark:text-gray-300 uppercase tracking-wider cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800"
                onClick={() => handleSort('cycle_time_days')}
              >
                <div className="flex items-center gap-1">
                  Cycle Time
                  <SortIcon field="cycle_time_days" />
                </div>
              </th>
            </tr>
          </thead>
          <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            {sortedData.map((row) => (
              <tr key={row.engineer_id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                <td className="px-4 py-3 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-gray-100">
                  {row.engineer_name}
                </td>
                <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-700 dark:text-gray-300">
                  {row.total_score.toFixed(1)}
                </td>
                <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-700 dark:text-gray-300">
                  {row.throughput_score.toFixed(1)}
                </td>
                <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-700 dark:text-gray-300">
                  {row.quality_score.toFixed(1)}
                </td>
                <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-700 dark:text-gray-300">
                  {row.prs_merged}
                </td>
                <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-700 dark:text-gray-300">
                  {row.cycle_time_days.toFixed(1)}d
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
