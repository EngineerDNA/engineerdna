import { useState } from 'react';
import type { Engineer } from '../api/types';

interface EngineerActivityMetrics {
  pull_requests: number;
  reviews: number;
  issues: number;
  commits: number;
}

interface EngineerCardProps {
  engineer: Engineer;
  onEdit: (engineer: Engineer) => void;
  activity?: EngineerActivityMetrics;
  isActivityLoading?: boolean;
  activityError?: boolean;
  onActivityToggle?: (engineerId: string, show: boolean) => void;
}

export function EngineerCard({
  engineer,
  onEdit,
  activity,
  isActivityLoading,
  activityError,
  onActivityToggle,
}: EngineerCardProps) {
  const [showActivity, setShowActivity] = useState(false);

  const handleToggle = () => {
    const newShowActivity = !showActivity;
    setShowActivity(newShowActivity);
    onActivityToggle?.(engineer.id, newShowActivity);
  };

  const metrics = activity ?? {
    pull_requests: 0,
    reviews: 0,
    issues: 0,
    commits: 0,
  };

  return (
    <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow border border-gray-200 dark:border-gray-700">
      <div className="flex items-start justify-between mb-4">
        <div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
            {engineer.name}
          </h3>
          {engineer.email && (
            <div className="text-sm text-gray-600 dark:text-gray-400">{engineer.email}</div>
          )}
          {engineer.manager && (
            <div className="text-sm text-gray-500 dark:text-gray-500">
              Manager: {engineer.manager}
            </div>
          )}
        </div>
        <button
          onClick={() => onEdit(engineer)}
          className="px-3 py-1 text-sm bg-blue-600 dark:bg-blue-700 text-white rounded hover:bg-blue-700 dark:hover:bg-blue-800 transition-colors"
        >
          Edit
        </button>
      </div>

      {Object.keys(engineer.identifiers).length > 0 && (
        <div className="mb-4">
          <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Identifiers</h4>
          <div className="space-y-1">
            {Object.entries(engineer.identifiers).map(([source, identifier]) => (
              <div key={source} className="text-sm">
                <span className="text-gray-600 dark:text-gray-400">{source}:</span>{' '}
                <span className="text-gray-900 dark:text-gray-100">{identifier}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
        <button
          onClick={handleToggle}
          className="w-full flex items-center justify-between text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 hover:text-gray-900 dark:hover:text-gray-100 transition-colors"
        >
          <span>Last 30 Days Activity</span>
          <span className="text-gray-400 dark:text-gray-500">{showActivity ? '▼' : '▶'}</span>
        </button>
        {showActivity && (
          <>
            {isActivityLoading ? (
              <div className="text-sm text-gray-500 dark:text-gray-400">Loading activity...</div>
            ) : activityError ? (
              <div className="text-sm text-red-600 dark:text-red-400">Failed to load activity</div>
            ) : (
              <div className="grid grid-cols-2 gap-2">
                <div className="text-center p-2 bg-blue-50 dark:bg-blue-900/20 rounded border border-blue-200 dark:border-blue-800">
                  <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">
                    {metrics.pull_requests}
                  </div>
                  <div className="text-xs text-gray-600 dark:text-gray-400">PRs</div>
                </div>
                <div className="text-center p-2 bg-green-50 dark:bg-green-900/20 rounded border border-green-200 dark:border-green-800">
                  <div className="text-2xl font-bold text-green-600 dark:text-green-400">
                    {metrics.reviews}
                  </div>
                  <div className="text-xs text-gray-600 dark:text-gray-400">Reviews</div>
                </div>
                <div className="text-center p-2 bg-purple-50 dark:bg-purple-900/20 rounded border border-purple-200 dark:border-purple-800">
                  <div className="text-2xl font-bold text-purple-600 dark:text-purple-400">
                    {metrics.issues}
                  </div>
                  <div className="text-xs text-gray-600 dark:text-gray-400">Issues</div>
                </div>
                <div className="text-center p-2 bg-orange-50 dark:bg-orange-900/20 rounded border border-orange-200 dark:border-orange-800">
                  <div className="text-2xl font-bold text-orange-600 dark:text-orange-400">
                    {metrics.commits}
                  </div>
                  <div className="text-xs text-gray-600 dark:text-gray-400">Commits</div>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
