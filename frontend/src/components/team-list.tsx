import { useState } from 'react';
import { Link } from 'react-router-dom';
import type { Team, TeamPerformanceScore } from '../api/types';

interface TeamCardProps {
  team: Team;
  performance: TeamPerformanceScore;
  expanded: boolean;
  onToggle: () => void;
}

function TeamCard({ team, performance, expanded, onToggle }: TeamCardProps) {
  const isUnconfigured = performance.member_count === 0;

  const getStatusColor = (score: number, unconfigured: boolean) => {
    if (unconfigured) return 'gray';
    if (score < 80) return 'red';
    if (score < 90) return 'yellow';
    if (score <= 110) return 'green';
    return 'blue';
  };

  const getStatusIcon = (score: number, unconfigured: boolean) => {
    if (unconfigured) return 'NOT CONFIGURED';
    if (score < 80) return 'WARNING';
    if (score > 110) return 'STRONG';
    return 'OK';
  };

  const getStatusLabel = (score: number, unconfigured: boolean) => {
    if (unconfigured) return 'No team members assigned';
    if (score < 80) return 'Needs Attention';
    if (score < 90) return 'Below Target';
    if (score <= 110) return 'On Track';
    return 'Exceeding Expectations';
  };

  const getChangeIcon = (change: number) => {
    if (change > 0) return '↑';
    if (change < 0) return '↓';
    return '→';
  };

  const statusColor = getStatusColor(performance.total_score, isUnconfigured);
  const statusIcon = getStatusIcon(performance.total_score, isUnconfigured);
  const statusLabel = getStatusLabel(performance.total_score, isUnconfigured);

  const statusColorClasses = {
    gray: 'bg-gray-50 dark:bg-gray-800 border-gray-200 dark:border-gray-700',
    red: 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800',
    yellow: 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800',
    green: 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800',
    blue: 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800',
  };

  const statusTextClasses = {
    gray: 'text-gray-700 dark:text-gray-400',
    red: 'text-red-700 dark:text-red-300',
    yellow: 'text-yellow-700 dark:text-yellow-300',
    green: 'text-green-700 dark:text-green-300',
    blue: 'text-blue-700 dark:text-blue-300',
  };

  return (
    <div
      className={`border rounded-lg p-6 mb-4 ${statusColorClasses[statusColor]} ${
        performance.total_score < 80 ? 'shadow-md' : ''
      }`}
    >
      <div
        className="flex justify-between items-start cursor-pointer"
        onClick={onToggle}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            onToggle();
          }
        }}
      >
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-2">
            <span className="text-gray-600 dark:text-gray-400 font-mono text-sm">
              {expanded ? '▼' : '▶'}
            </span>
            <h3 className="text-xl font-bold text-gray-900 dark:text-gray-100">{team.name}</h3>
            {team.manager && (
              <span className="text-sm text-gray-600 dark:text-gray-400">
                (EM: {team.manager.name})
              </span>
            )}
          </div>

          <div className="ml-6 text-sm text-gray-700 dark:text-gray-300">
            {performance.member_count} engineer{performance.member_count !== 1 ? 's' : ''} ·{' '}
            {(performance.velocity_prs_per_week ?? 0).toFixed(0)} PRs/week (
            {getChangeIcon(performance.velocity_change)}
            {(Math.abs((performance.velocity_change ?? 0) * 100) ?? 0).toFixed(0)}%) ·{' '}
            {(performance.cycle_time_days ?? 0).toFixed(1)}d cycle (
            {getChangeIcon(-performance.cycle_time_change)}
            {(Math.abs((performance.cycle_time_change ?? 0) * 100) ?? 0).toFixed(0)}%)
          </div>
        </div>

        <div className="text-right">
          <div className="flex items-baseline gap-2 mb-1">
            <span className="text-3xl font-bold text-gray-900 dark:text-gray-100">
              {(performance.total_score ?? 0).toFixed(0)}/100
            </span>
            <span className="text-lg">{getChangeIcon(performance.score_change)}</span>
          </div>
          <div className={`text-sm font-medium ${statusTextClasses[statusColor]}`}>
            {statusIcon}
          </div>
        </div>
      </div>

      {expanded && (
        <div className="mt-4 ml-6 pt-4 border-t border-gray-300 dark:border-gray-600">
          <div className="mb-3">
            <span className={`font-medium ${statusTextClasses[statusColor]}`}>{statusLabel}</span>
          </div>
          <Link
            to={`/teams/${team.id}`}
            className="inline-block px-4 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded hover:bg-blue-700 dark:hover:bg-blue-800 transition-colors"
          >
            Drill into team →
          </Link>
        </div>
      )}
    </div>
  );
}

interface TeamListProps {
  teams: Team[];
  performances: Map<string, TeamPerformanceScore>;
}

export function TeamList({ teams, performances }: TeamListProps) {
  const [expandedTeams, setExpandedTeams] = useState<Set<string>>(new Set());

  const toggleTeam = (teamId: string) => {
    setExpandedTeams((prev) => {
      const next = new Set(prev);
      if (next.has(teamId)) {
        next.delete(teamId);
      } else {
        next.add(teamId);
      }
      return next;
    });
  };

  const sortedTeams = [...teams].sort((a, b) => {
    const perfA = performances.get(a.id);
    const perfB = performances.get(b.id);
    if (!perfA || !perfB) return 0;
    return perfA.total_score - perfB.total_score;
  });

  if (teams.length === 0) {
    return (
      <div className="text-center py-12 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
        <p className="text-gray-600 dark:text-gray-400">No teams found</p>
      </div>
    );
  }

  return (
    <div>
      {sortedTeams.map((team) => {
        const performance = performances.get(team.id);
        if (!performance) return null;
        return (
          <TeamCard
            key={team.id}
            team={team}
            performance={performance}
            expanded={expandedTeams.has(team.id)}
            onToggle={() => toggleTeam(team.id)}
          />
        );
      })}
    </div>
  );
}
