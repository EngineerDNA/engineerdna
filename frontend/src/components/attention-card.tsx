import { useState } from 'react';
import { Link } from 'react-router-dom';
import type { AttentionItem } from '../api/types';

interface AttentionCardProps {
  item: AttentionItem;
}

export function AttentionCard({ item }: AttentionCardProps) {
  const [expanded, setExpanded] = useState(false);

  const severityColors = {
    warning: {
      bg: 'bg-yellow-50 dark:bg-yellow-900/20',
      border: 'border-yellow-200 dark:border-yellow-800',
      text: 'text-yellow-800 dark:text-yellow-200',
      icon: 'text-yellow-600 dark:text-yellow-400',
    },
    critical: {
      bg: 'bg-red-50 dark:bg-red-900/20',
      border: 'border-red-200 dark:border-red-800',
      text: 'text-red-800 dark:text-red-200',
      icon: 'text-red-600 dark:text-red-400',
    },
  };

  const colors = severityColors[item.severity] || severityColors.warning;

  const icon =
    item.severity === 'critical' ? (
      <svg
        className={`w-6 h-6 ${colors.icon}`}
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        aria-hidden="true"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
        />
      </svg>
    ) : (
      <svg
        className={`w-6 h-6 ${colors.icon}`}
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        aria-hidden="true"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
        />
      </svg>
    );

  return (
    <div className={`${colors.bg} ${colors.border} border rounded-lg p-4`}>
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0 mt-0.5">{icon}</div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-2">
            <h3 className={`text-lg font-semibold ${colors.text}`}>{item.engineer_name}</h3>
            <span
              className={`px-2 py-0.5 text-xs font-medium ${colors.text} ${colors.bg} rounded-full`}
            >
              {item.severity?.toUpperCase() || 'UNKNOWN'}
            </span>
          </div>

          <p className={`text-sm ${colors.text} mb-3`}>{item.issue}</p>

          <div className={`${colors.bg} border ${colors.border} rounded p-3 mb-3`}>
            <div className="text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
              SUGGESTED 1:1 TOPIC
            </div>
            <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
              {item.suggested_one_on_one}
            </p>
          </div>

          <div className="flex items-center gap-4">
            <button
              onClick={() => setExpanded(!expanded)}
              className="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 font-medium transition-colors"
            >
              {expanded ? 'Hide evidence' : 'Show evidence'}
            </button>

            <Link
              to={`/team?engineer=${item.engineer_id}`}
              className="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 font-medium transition-colors"
            >
              View engineer data
            </Link>
          </div>

          {expanded && (item.evidence?.length ?? 0) > 0 && (
            <div className="mt-3 pt-3 border-t border-gray-300 dark:border-gray-600">
              <div className="text-xs font-medium text-gray-600 dark:text-gray-400 mb-2">
                EVIDENCE
              </div>
              <ul className="space-y-1">
                {item.evidence?.map((evidence, idx) => (
                  <li key={idx} className="text-sm text-gray-700 dark:text-gray-300">
                    <span className="text-gray-500 dark:text-gray-500">•</span> {evidence}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
