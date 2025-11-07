import { useState } from 'react';
import type { Engineer, PerformanceScore, Role } from '../api/types';
import { ScoreBreakdown } from './score-breakdown';

interface ScoreCardProps {
  engineer: Engineer;
  score: PerformanceScore | null;
  role: Role | null;
}

export function ScoreCard({ engineer, score, role }: ScoreCardProps) {
  const [expanded, setExpanded] = useState(false);

  const getScoreColor = (scoreValue: number) => {
    if (scoreValue >= 111) return 'blue';
    if (scoreValue >= 90) return 'green';
    if (scoreValue >= 70) return 'yellow';
    return 'red';
  };

  const getScoreStatus = (scoreValue: number, targetScore: number) => {
    if (scoreValue >= targetScore + 11) return { label: 'Strong', color: 'blue' };
    if (scoreValue >= targetScore - 10) return { label: 'On Track', color: 'green' };
    if (scoreValue >= targetScore - 30) return { label: 'Below', color: 'yellow' };
    return { label: 'Needs Attention', color: 'red' };
  };

  if (!score) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
              {engineer.name}
            </h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              {role?.name || 'No role assigned'}
            </p>
          </div>
        </div>
        <div className="mt-4 text-sm text-gray-500 dark:text-gray-400">
          No performance data available yet
        </div>
      </div>
    );
  }

  const totalScore = score.total_score;
  const targetScore = role?.target_score || 100;
  const status = getScoreStatus(totalScore, targetScore);
  const scoreColor = getScoreColor(totalScore);

  const colorClasses = {
    blue: 'bg-blue-500 dark:bg-blue-400',
    green: 'bg-green-500 dark:bg-green-400',
    yellow: 'bg-yellow-500 dark:bg-yellow-400',
    red: 'bg-red-500 dark:bg-red-400',
  };

  const textColorClasses = {
    blue: 'text-blue-600 dark:text-blue-400',
    green: 'text-green-600 dark:text-green-400',
    yellow: 'text-yellow-600 dark:text-yellow-400',
    red: 'text-red-600 dark:text-red-400',
  };

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
            {engineer.name}
          </h3>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            {role?.name || 'No role assigned'}
          </p>
        </div>
        <div className="text-right">
          <div className="text-3xl font-bold text-gray-900 dark:text-gray-100">
            {(totalScore ?? 0).toFixed(0)}
            <span className="text-lg text-gray-500 dark:text-gray-400">/100</span>
          </div>
          <div className={`text-sm font-medium ${textColorClasses[scoreColor]}`}>
            {status.label}
          </div>
        </div>
      </div>

      <div className="mb-4">
        <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-4">
          <div
            className={`${colorClasses[scoreColor]} h-4 rounded-full transition-all`}
            style={{ width: `${Math.min((totalScore / 150) * 100, 100)}%` }}
          />
        </div>
        <div className="flex justify-between mt-1 text-xs text-gray-500 dark:text-gray-400">
          <span>0</span>
          <span>Target: {targetScore}</span>
          <span>150</span>
        </div>
      </div>

      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 text-left font-medium transition-colors"
      >
        {expanded ? 'Hide breakdown' : 'View breakdown'}
      </button>

      {expanded && <ScoreBreakdown score={score} />}
    </div>
  );
}
