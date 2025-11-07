import { useState } from 'react';
import type { AIInsight } from '../api/types';

interface InsightCardProps {
  insight: AIInsight;
}

export function InsightCard({ insight }: InsightCardProps) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0 mt-0.5">
          <svg
            className="w-6 h-6 text-blue-600 dark:text-blue-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
            />
          </svg>
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-2">
            <span className="px-2 py-0.5 text-xs font-medium text-blue-800 dark:text-blue-200 bg-blue-100 dark:bg-blue-900/40 rounded-full">
              AI INSIGHT
            </span>
          </div>

          <h3 className="text-lg font-semibold text-blue-900 dark:text-blue-100 mb-2">
            {insight.observation}
          </h3>

          <p className="text-sm text-blue-800 dark:text-blue-200 mb-3">{insight.context}</p>

          {insight.recommendation && (
            <div className="bg-blue-100 dark:bg-blue-900/40 border border-blue-200 dark:border-blue-800 rounded p-3 mb-3">
              <div className="text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                RECOMMENDATION
              </div>
              <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                {insight.recommendation}
              </p>
            </div>
          )}

          {(insight.evidence?.length ?? 0) > 0 && (
            <button
              onClick={() => setExpanded(!expanded)}
              className="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 font-medium transition-colors"
            >
              {expanded ? 'Hide data' : 'Show data'}
            </button>
          )}

          {expanded && (insight.evidence?.length ?? 0) > 0 && (
            <div className="mt-3 pt-3 border-t border-blue-300 dark:border-blue-700">
              <div className="text-xs font-medium text-gray-600 dark:text-gray-400 mb-2">DATA</div>
              <ul className="space-y-1">
                {insight.evidence?.map((evidence, idx) => (
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
