import type { OrgScorecard } from '../api/types';

interface OrgScorecardProps {
  scorecard: OrgScorecard;
}

export function OrgScorecardComponent({ scorecard }: OrgScorecardProps) {
  const getChangeIcon = (change: number) => {
    if (change > 0) return '↑';
    if (change < 0) return '↓';
    return '→';
  };

  const getChangeColor = (change: number) => {
    if (change > 0) return 'text-green-600 dark:text-green-400';
    if (change < 0) return 'text-red-600 dark:text-red-400';
    return 'text-gray-600 dark:text-gray-400';
  };

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 mb-6">
      <div className="mb-4">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
          Engineering Organization ({scorecard.total_engineers} engineers)
        </h2>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div>
          <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Overall Score</div>
          <div className="flex items-baseline gap-2">
            <span className="text-3xl font-bold text-gray-900 dark:text-gray-100">
              {(scorecard.total_score ?? 0).toFixed(0)}/100
            </span>
            <span className={`text-sm font-medium ${getChangeColor(scorecard.score_change ?? 0)}`}>
              {getChangeIcon(scorecard.score_change ?? 0)}
              {Math.abs(scorecard.score_change ?? 0).toFixed(0)} from last month
            </span>
          </div>
        </div>

        <div>
          <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Velocity</div>
          <div className="flex items-baseline gap-2">
            <span className="text-3xl font-bold text-gray-900 dark:text-gray-100">
              {(scorecard.velocity_prs_per_week ?? 0).toFixed(0)} PRs/week
            </span>
            <span
              className={`text-sm font-medium ${getChangeColor(scorecard.velocity_change ?? 0)}`}
            >
              {getChangeIcon(scorecard.velocity_change ?? 0)}
              {Math.abs((scorecard.velocity_change ?? 0) * 100).toFixed(0)}%
            </span>
          </div>
        </div>

        <div>
          <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Cycle Time</div>
          <div className="flex items-baseline gap-2">
            <span className="text-3xl font-bold text-gray-900 dark:text-gray-100">
              {(scorecard.cycle_time_days ?? 0).toFixed(1)} days
            </span>
            <span
              className={`text-sm font-medium ${getChangeColor(-(scorecard.cycle_time_change ?? 0))}`}
            >
              {getChangeIcon(-(scorecard.cycle_time_change ?? 0))} stable
            </span>
          </div>
        </div>
      </div>

      {scorecard.teams_needing_attention > 0 && (
        <div className="mt-6 p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded">
          <div className="flex items-center gap-2">
            <span className="text-xl font-bold text-yellow-800 dark:text-yellow-200">WARNING</span>
            <span className="text-yellow-800 dark:text-yellow-200 font-medium">
              {scorecard.teams_needing_attention} team
              {scorecard.teams_needing_attention > 1 ? 's' : ''} need
              {scorecard.teams_needing_attention === 1 ? 's' : ''} attention
            </span>
          </div>
        </div>
      )}
    </div>
  );
}
