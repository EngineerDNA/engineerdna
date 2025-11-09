import { useState } from 'react';
import type { WidgetProps } from './widget-registry';

interface QuickAction {
  id: string;
  label: string;
  icon: string;
  action: string;
  variant?: 'primary' | 'secondary' | 'danger';
}

export function QuickActionsWidget({ config, onOpenModal }: WidgetProps) {
  const configuredActions = (config.actions as QuickAction[]) || [];

  const defaultActions: QuickAction[] = [
    {
      id: 'generate-briefing',
      label: 'Generate Briefing',
      icon: 'document',
      action: 'generate-briefing',
      variant: 'primary',
    },
    {
      id: 'sync-plugins',
      label: 'Sync All Plugins',
      icon: 'refresh',
      action: 'sync-all-plugins',
      variant: 'secondary',
    },
    {
      id: 'import-team',
      label: 'Import Team CSV',
      icon: 'upload',
      action: 'import-team',
      variant: 'secondary',
    },
    {
      id: 'resolve-identities',
      label: 'Auto-Resolve Identities',
      icon: 'users',
      action: 'auto-resolve',
      variant: 'secondary',
    },
    {
      id: 'export-data',
      label: 'Export Dashboard',
      icon: 'download',
      action: 'export-dashboard',
      variant: 'secondary',
    },
    {
      id: 'view-audit-log',
      label: 'View Audit Log',
      icon: 'clipboard',
      action: 'view-audit-log',
      variant: 'secondary',
    },
  ];

  const actions = configuredActions.length > 0 ? configuredActions : defaultActions;

  const [executing, setExecuting] = useState<string | null>(null);

  const getIcon = (iconName: string) => {
    switch (iconName) {
      case 'document':
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
        );
      case 'refresh':
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
            />
          </svg>
        );
      case 'upload':
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
            />
          </svg>
        );
      case 'download':
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10"
            />
          </svg>
        );
      case 'users':
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
            />
          </svg>
        );
      case 'clipboard':
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
            />
          </svg>
        );
      default:
        return (
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M13 10V3L4 14h7v7l9-11h-7z"
            />
          </svg>
        );
    }
  };

  const getButtonClass = (variant?: string) => {
    switch (variant) {
      case 'primary':
        return 'bg-blue-600 hover:bg-blue-700 text-white border-transparent';
      case 'danger':
        return 'bg-red-600 hover:bg-red-700 text-white border-transparent';
      default:
        return 'bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-900 text-gray-700 dark:text-gray-300 border-gray-300 dark:border-gray-600';
    }
  };

  const handleAction = async (action: QuickAction) => {
    setExecuting(action.id);

    // Simulate action execution
    await new Promise((resolve) => setTimeout(resolve, 1000));

    setExecuting(null);

    // Trigger modal or action
    if (onOpenModal) {
      onOpenModal('quick-action-result', { action: action.action });
    }
  };

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full flex flex-col">
      <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-4">Quick Actions</h3>

      <div className="flex-1 overflow-auto">
        <div className="grid grid-cols-2 gap-3">
          {actions.map((action) => (
            <button
              key={action.id}
              onClick={() => handleAction(action)}
              disabled={executing === action.id}
              className={`flex flex-col items-center justify-center gap-2 p-4 border rounded-lg transition-all disabled:opacity-50 disabled:cursor-not-allowed ${getButtonClass(action.variant)}`}
            >
              {executing === action.id ? (
                <svg className="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle
                    className="opacity-25"
                    cx="12"
                    cy="12"
                    r="10"
                    stroke="currentColor"
                    strokeWidth="4"
                  />
                  <path
                    className="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  />
                </svg>
              ) : (
                getIcon(action.icon)
              )}
              <span className="text-xs font-medium text-center">{action.label}</span>
            </button>
          ))}
        </div>
      </div>

      <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700 text-xs text-gray-500 dark:text-gray-400 text-center">
        Customize actions in widget settings
      </div>
    </div>
  );
}
