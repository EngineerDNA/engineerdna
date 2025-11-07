import type { Insight } from '../api/types';

interface InsightsPanelProps {
  insights: Insight[];
}

export function InsightsPanel({ insights }: InsightsPanelProps) {
  if (insights.length === 0) {
    return (
      <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
        <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">Insights</h3>
        <p className="text-gray-600 dark:text-gray-400">
          No insights yet. Generate insights from the AI Insights plugin.
        </p>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
      <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">Insights</h3>
      <div className="space-y-4">
        {insights
          .filter((i) => !i.dismissed)
          .map((insight) => (
            <InsightCard key={insight.id} insight={insight} />
          ))}
      </div>
    </div>
  );
}

function InsightCard({ insight }: { insight: Insight }) {
  const severityColors = {
    info: 'border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20',
    warning: 'border-yellow-200 dark:border-yellow-800 bg-yellow-50 dark:bg-yellow-900/20',
    error: 'border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20',
  };

  return (
    <div className={`border-l-4 p-4 ${severityColors[insight.severity]}`}>
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <h4 className="font-semibold text-gray-900 dark:text-gray-100">{insight.title}</h4>
          <p className="text-sm mt-1 text-gray-700 dark:text-gray-300">{insight.description}</p>
          {insight.recommendation && (
            <p className="text-sm mt-2 font-medium text-gray-800 dark:text-gray-200">
              Recommendation: {insight.recommendation}
            </p>
          )}
        </div>
        <button className="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 ml-4">
          ×
        </button>
      </div>
    </div>
  );
}
