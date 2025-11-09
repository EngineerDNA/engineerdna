import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/client';
import type { Team, Engineer, AlertInstance } from '../../api/types';

interface TeamDetailModalProps {
  team_id: string;
  isOpen: boolean;
  onClose: () => void;
}

type TabType = 'overview' | 'members' | 'activity' | 'alerts';

export function TeamDetailModal({ team_id, isOpen, onClose }: TeamDetailModalProps) {
  const [activeTab, setActiveTab] = useState<TabType>('overview');

  const { data: team, isLoading: teamLoading } = useQuery({
    queryKey: ['team', team_id],
    queryFn: () => api.getTeam(team_id),
    enabled: isOpen,
  });

  const { data: scorecard, isLoading: scorecardLoading } = useQuery({
    queryKey: ['team-scorecard', team_id],
    queryFn: () => api.getTeamScorecard(team_id),
    enabled: isOpen && activeTab === 'overview',
  });

  const { data: membersData, isLoading: membersLoading } = useQuery({
    queryKey: ['team-members', team_id],
    queryFn: () => api.getTeamMembers(team_id),
    enabled: isOpen && activeTab === 'members',
  });

  const { data: alertsData, isLoading: alertsLoading } = useQuery({
    queryKey: ['team-alerts', team_id],
    queryFn: () => api.getAlertInstances({ entity_type: 'team', entity_id: team_id }),
    enabled: isOpen && activeTab === 'alerts',
  });

  const { data: eventsData, isLoading: eventsLoading } = useQuery({
    queryKey: ['team-events', team_id],
    queryFn: () => api.getEvents({ limit: 20 }),
    enabled: isOpen && activeTab === 'activity',
  });

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      return () => document.removeEventListener('keydown', handleEscape);
    }
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const tabs: { id: TabType; label: string }[] = [
    { id: 'overview', label: 'Overview' },
    { id: 'members', label: 'Members' },
    { id: 'activity', label: 'Activity' },
    { id: 'alerts', label: 'Alerts' },
  ];

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="fixed inset-0 bg-black bg-opacity-50" onClick={onClose} />

      <div className="relative min-h-screen flex items-center justify-center p-4">
        <div className="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] flex flex-col">
          <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
            <div>
              <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                {teamLoading ? 'Loading...' : team?.name || 'Team Details'}
              </h2>
              {team?.description && (
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">{team.description}</p>
              )}
            </div>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
              aria-label="Close"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>

          <div className="border-b border-gray-200 dark:border-gray-700">
            <nav className="flex space-x-8 px-6" aria-label="Tabs">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                    activeTab === tab.id
                      ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </nav>
          </div>

          <div className="flex-1 overflow-y-auto p-6">
            {activeTab === 'overview' && (
              <OverviewTab
                team={team}
                scorecard={scorecard}
                isLoading={teamLoading || scorecardLoading}
              />
            )}
            {activeTab === 'members' && (
              <MembersTab members={membersData?.members} isLoading={membersLoading} />
            )}
            {activeTab === 'activity' && (
              <ActivityTab events={eventsData?.events} isLoading={eventsLoading} />
            )}
            {activeTab === 'alerts' && (
              <AlertsTab alerts={alertsData?.alerts} isLoading={alertsLoading} />
            )}
          </div>

          <div className="flex gap-3 justify-end p-6 border-t border-gray-200 dark:border-gray-700">
            <button
              onClick={onClose}
              className="px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

function OverviewTab({
  team,
  scorecard,
  isLoading,
}: {
  team?: Team;
  scorecard?: any;
  isLoading: boolean;
}) {
  if (isLoading) {
    return <LoadingSkeleton />;
  }

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
          Team Information
        </h3>
        <dl className="grid grid-cols-2 gap-4">
          <div>
            <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Team Name</dt>
            <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100">{team?.name || 'N/A'}</dd>
          </div>
          <div>
            <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Manager</dt>
            <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100">
              {team?.manager?.name || 'Not assigned'}
            </dd>
          </div>
          <div>
            <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Created</dt>
            <dd className="mt-1 text-sm text-gray-900 dark:text-gray-100">
              {team ? new Date(team.created_at).toLocaleDateString() : 'N/A'}
            </dd>
          </div>
        </dl>
      </div>

      {scorecard?.performance && (
        <div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
            Performance Metrics
          </h3>
          <div className="grid grid-cols-3 gap-4">
            <MetricCard
              label="Total Score"
              value={scorecard.performance.total_score.toFixed(1)}
              change={scorecard.performance.score_change}
            />
            <MetricCard
              label="Velocity (PRs/week)"
              value={scorecard.performance.velocity_prs_per_week.toFixed(1)}
              change={scorecard.performance.velocity_change}
            />
            <MetricCard
              label="Cycle Time (days)"
              value={scorecard.performance.cycle_time_days.toFixed(1)}
              change={scorecard.performance.cycle_time_change}
              invertChange
            />
          </div>
        </div>
      )}
    </div>
  );
}

function MembersTab({ members, isLoading }: { members?: Engineer[]; isLoading: boolean }) {
  if (isLoading) {
    return <LoadingSkeleton />;
  }

  if (!members || members.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500 dark:text-gray-400">No team members found</div>
    );
  }

  return (
    <div className="space-y-4">
      {members.map((member) => (
        <div
          key={member.id}
          className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
        >
          <div>
            <h4 className="font-medium text-gray-900 dark:text-gray-100">{member.name}</h4>
            {member.email && (
              <p className="text-sm text-gray-600 dark:text-gray-400">{member.email}</p>
            )}
          </div>
          <span
            className={`px-2 py-1 text-xs rounded ${
              member.active
                ? 'bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-400'
                : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-400'
            }`}
          >
            {member.active ? 'Active' : 'Inactive'}
          </span>
        </div>
      ))}
    </div>
  );
}

function ActivityTab({ events, isLoading }: { events?: any[]; isLoading: boolean }) {
  if (isLoading) {
    return <LoadingSkeleton />;
  }

  if (!events || events.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500 dark:text-gray-400">No recent activity</div>
    );
  }

  return (
    <div className="space-y-3">
      {events.map((event) => (
        <div key={event.id} className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <p className="font-medium text-gray-900 dark:text-gray-100">{event.type}</p>
              <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                by {event.actor} on {new Date(event.timestamp).toLocaleString()}
              </p>
            </div>
            <span className="px-2 py-1 text-xs bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400 rounded">
              {event.source}
            </span>
          </div>
        </div>
      ))}
    </div>
  );
}

function AlertsTab({ alerts, isLoading }: { alerts?: AlertInstance[]; isLoading: boolean }) {
  if (isLoading) {
    return <LoadingSkeleton />;
  }

  if (!alerts || alerts.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500 dark:text-gray-400">No active alerts</div>
    );
  }

  return (
    <div className="space-y-3">
      {alerts.map((alert) => (
        <div key={alert.id} className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <h4 className="font-medium text-gray-900 dark:text-gray-100">{alert.title}</h4>
              <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">{alert.message}</p>
              <p className="text-xs text-gray-500 dark:text-gray-500 mt-2">
                Fired at {new Date(alert.fired_at).toLocaleString()}
              </p>
            </div>
            <span
              className={`px-2 py-1 text-xs rounded ${
                alert.severity === 'critical'
                  ? 'bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-400'
                  : alert.severity === 'warning'
                    ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400'
                    : 'bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400'
              }`}
            >
              {alert.severity}
            </span>
          </div>
        </div>
      ))}
    </div>
  );
}

function MetricCard({
  label,
  value,
  change,
  invertChange = false,
}: {
  label: string;
  value: string;
  change: number;
  invertChange?: boolean;
}) {
  const isPositive = invertChange ? change < 0 : change > 0;
  const changeColor = isPositive
    ? 'text-green-600 dark:text-green-400'
    : 'text-red-600 dark:text-red-400';

  return (
    <div className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
      <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">{label}</dt>
      <dd className="mt-2 flex items-baseline">
        <span className="text-2xl font-semibold text-gray-900 dark:text-gray-100">{value}</span>
        {change !== 0 && (
          <span className={`ml-2 text-sm font-medium ${changeColor}`}>
            {change > 0 ? '+' : ''}
            {change.toFixed(1)}%
          </span>
        )}
      </dd>
    </div>
  );
}

function LoadingSkeleton() {
  return (
    <div className="space-y-4">
      {[1, 2, 3].map((i) => (
        <div key={i} className="animate-pulse">
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/4 mb-2" />
          <div className="h-8 bg-gray-200 dark:bg-gray-700 rounded w-full" />
        </div>
      ))}
    </div>
  );
}
