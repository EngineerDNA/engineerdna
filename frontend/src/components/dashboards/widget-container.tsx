import type { ReactNode } from 'react';

interface WidgetContainerProps {
  title: string;
  children: ReactNode;
  onEdit?: () => void;
  onDelete?: () => void;
  className?: string;
}

export function WidgetContainer({
  title,
  children,
  onEdit,
  onDelete,
  className = '',
}: WidgetContainerProps) {
  return (
    <div
      className={`h-full flex flex-col bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden ${className}`}
    >
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900">
        <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100">{title}</h3>
        <div className="flex items-center gap-2">
          {onEdit && (
            <button
              onClick={onEdit}
              className="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
              aria-label="Edit widget"
            >
              <svg
                className="w-4 h-4 text-gray-600 dark:text-gray-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                />
              </svg>
            </button>
          )}
          {onDelete && (
            <button
              onClick={onDelete}
              className="p-1 rounded hover:bg-red-100 dark:hover:bg-red-900/20 transition-colors"
              aria-label="Delete widget"
            >
              <svg
                className="w-4 h-4 text-red-600 dark:text-red-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                />
              </svg>
            </button>
          )}
        </div>
      </div>
      <div className="flex-1 overflow-auto">{children}</div>
    </div>
  );
}
