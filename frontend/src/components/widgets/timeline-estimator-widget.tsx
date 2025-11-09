import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { api } from '../../api/client';
import type { WidgetProps } from './widget-registry';
import type { TimelineEstimate } from '../../api/types';

export function TimelineEstimatorWidget({ config }: WidgetProps) {
  const defaultTeamId = config.team_id as string | undefined;

  const [featureName, setFeatureName] = useState('');
  const [storyPoints, setStoryPoints] = useState('');
  const [estimate, setEstimate] = useState<TimelineEstimate | null>(null);

  const estimateMutation = useMutation({
    mutationFn: (data: { feature_name: string; estimated_points: number; team_id: string }) =>
      api.estimateTimeline(data),
    onSuccess: (data) => {
      setEstimate(data);
    },
  });

  const handleEstimate = (e: React.FormEvent) => {
    e.preventDefault();
    if (featureName && storyPoints && defaultTeamId) {
      estimateMutation.mutate({
        feature_name: featureName,
        estimated_points: parseFloat(storyPoints),
        team_id: defaultTeamId,
      });
    }
  };

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full flex flex-col">
      <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-4">
        Timeline Estimator
      </h3>

      <form onSubmit={handleEstimate} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Feature Name
          </label>
          <input
            type="text"
            value={featureName}
            onChange={(e) => setFeatureName(e.target.value)}
            placeholder="e.g., User Authentication"
            required
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Story Points
          </label>
          <input
            type="number"
            value={storyPoints}
            onChange={(e) => setStoryPoints(e.target.value)}
            placeholder="e.g., 13"
            required
            min="1"
            step="0.5"
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>

        <button
          type="submit"
          disabled={estimateMutation.isPending || !defaultTeamId}
          className="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
        >
          {estimateMutation.isPending ? 'Estimating...' : 'Estimate Timeline'}
        </button>
      </form>

      {!defaultTeamId && (
        <p className="mt-4 text-sm text-yellow-600 dark:text-yellow-400">
          Configure team_id in widget settings to enable estimation
        </p>
      )}

      {estimate && (
        <div className="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700 space-y-4">
          <div>
            <h4 className="text-sm font-medium text-gray-900 dark:text-gray-100 mb-2">
              {estimate.feature_name}
            </h4>
            <p className="text-xs text-gray-500 dark:text-gray-400">
              Based on {estimate.team_name} avg velocity: {estimate.average_velocity.toFixed(1)}{' '}
              pts/sprint
            </p>
          </div>

          <div className="grid grid-cols-3 gap-3">
            <div className="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg p-3">
              <p className="text-xs text-green-700 dark:text-green-300 mb-1">Best Case</p>
              <p className="text-lg font-bold text-green-900 dark:text-green-100">
                {estimate.best_case_weeks} <span className="text-xs">wks</span>
              </p>
              <p className="text-xs text-green-600 dark:text-green-400">
                {estimate.best_case_sprints} sprints
              </p>
            </div>

            <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-3">
              <p className="text-xs text-blue-700 dark:text-blue-300 mb-1">Likely Case</p>
              <p className="text-lg font-bold text-blue-900 dark:text-blue-100">
                {estimate.likely_case_weeks} <span className="text-xs">wks</span>
              </p>
              <p className="text-xs text-blue-600 dark:text-blue-400">
                {estimate.likely_case_sprints} sprints
              </p>
            </div>

            <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-3">
              <p className="text-xs text-red-700 dark:text-red-300 mb-1">Worst Case</p>
              <p className="text-lg font-bold text-red-900 dark:text-red-100">
                {estimate.worst_case_weeks} <span className="text-xs">wks</span>
              </p>
              <p className="text-xs text-red-600 dark:text-red-400">
                {estimate.worst_case_sprints} sprints
              </p>
            </div>
          </div>

          <div className="bg-gray-50 dark:bg-gray-900 rounded-lg p-3">
            <p className="text-xs text-gray-600 dark:text-gray-400 mb-2">Assumptions:</p>
            <ul className="text-xs text-gray-700 dark:text-gray-300 space-y-1 list-disc list-inside">
              {estimate.assumptions.map((assumption, idx) => (
                <li key={idx}>{assumption}</li>
              ))}
            </ul>
          </div>

          <p className="text-xs text-gray-500 dark:text-gray-400 text-center">
            Confidence: {estimate.confidence_level}
          </p>
        </div>
      )}
    </div>
  );
}
