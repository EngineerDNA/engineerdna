import { useState, useMemo } from 'react';

export interface TableColumn {
  key: string;
  label: string;
  sortable?: boolean;
  format?: 'number' | 'decimal' | 'text' | 'date';
  decimals?: number;
  suffix?: string;
}

interface TableProps {
  data: Record<string, any>[];
  columns?: TableColumn[];
  isLoading?: boolean;
}

type SortOrder = 'asc' | 'desc';

/**
 * Generic table component that displays data in a sortable table format.
 * Supports configurable columns with custom formatting and sorting.
 * Includes loading state with skeleton and empty state handling.
 */
export function Table({ data, columns, isLoading = false }: TableProps) {
  const [sortField, setSortField] = useState<string>('');
  const [sortOrder, setSortOrder] = useState<SortOrder>('desc');

  // Auto-detect columns from data if not provided
  const tableColumns = useMemo(() => {
    if (columns && columns.length > 0) {
      return columns;
    }
    if (!data || data.length === 0) {
      return [];
    }
    // Generate columns from first data row
    const firstRow = data[0];
    return Object.keys(firstRow).map((key) => ({
      key,
      label: key
        .split('_')
        .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' '),
      sortable: true,
      format: (typeof firstRow[key] === 'number' ? 'number' : 'text') as TableColumn['format'],
    }));
  }, [data, columns]);

  const handleSort = (field: string) => {
    if (sortField === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortOrder('desc');
    }
  };

  const sortedData = useMemo(() => {
    if (!data || !sortField) return data || [];
    return [...data].sort((a, b) => {
      const aValue = a[sortField];
      const bValue = b[sortField];

      // Handle null/undefined
      if (aValue == null && bValue == null) return 0;
      if (aValue == null) return 1;
      if (bValue == null) return -1;

      if (typeof aValue === 'string' && typeof bValue === 'string') {
        return sortOrder === 'asc' ? aValue.localeCompare(bValue) : bValue.localeCompare(aValue);
      }

      const aNum = typeof aValue === 'number' ? aValue : 0;
      const bNum = typeof bValue === 'number' ? bValue : 0;
      return sortOrder === 'asc' ? aNum - bNum : bNum - aNum;
    });
  }, [data, sortField, sortOrder]);

  const formatValue = (value: any, column: TableColumn): string => {
    if (value == null) return '-';

    switch (column.format) {
      case 'decimal': {
        const decimals = column.decimals ?? 1;
        const num = typeof value === 'number' ? value : parseFloat(value);
        if (isNaN(num)) return '-';
        const formatted = num.toFixed(decimals);
        return column.suffix ? `${formatted}${column.suffix}` : formatted;
      }
      case 'number':
        if (typeof value === 'number') {
          const formatted = Math.round(value).toString();
          return column.suffix ? `${formatted}${column.suffix}` : formatted;
        }
        return String(value);
      case 'date':
        try {
          return new Date(value).toLocaleDateString();
        } catch {
          return String(value);
        }
      case 'text':
      default:
        return String(value);
    }
  };

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

  if (!data || data.length === 0 || tableColumns.length === 0) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="flex items-center justify-center h-64 text-gray-500 dark:text-gray-400">
          No data available
        </div>
      </div>
    );
  }

  const SortIcon = ({ field }: { field: string }) => {
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
              {tableColumns.map((column) => (
                <th
                  key={column.key}
                  className={`px-4 py-3 text-left text-xs font-medium text-gray-700 dark:text-gray-300 uppercase tracking-wider ${
                    column.sortable !== false
                      ? 'cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800'
                      : ''
                  }`}
                  onClick={() => column.sortable !== false && handleSort(column.key)}
                >
                  <div className="flex items-center gap-1">
                    {column.label}
                    {column.sortable !== false && <SortIcon field={column.key} />}
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            {sortedData.map((row, rowIndex) => (
              <tr key={rowIndex} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                {tableColumns.map((column, colIndex) => (
                  <td
                    key={column.key}
                    className={`px-4 py-3 whitespace-nowrap text-sm ${
                      colIndex === 0
                        ? 'font-medium text-gray-900 dark:text-gray-100'
                        : 'text-gray-700 dark:text-gray-300'
                    }`}
                  >
                    {formatValue(row[column.key], column)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
