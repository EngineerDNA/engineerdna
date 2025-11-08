import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';
import { ScoreCard } from '../components/score-card';
import type { Role } from '../api/types';
import type { MetricValue } from '../types/metrics';
import { useMemo } from 'react';

export function TeamDetailPage() {
  const { id } = useParams<{ id: string }>();

  const {
    data: teamData,
    isLoading: teamLoading,
    error: teamError,
  } = useQuery({
    queryKey: ['team', id],
    queryFn: () => api.getTeam(id!),
    enabled: !!id,
    staleTime: 30000,
  });

  const {
    data: performanceData,
    isLoading: performanceLoading,
    error: performanceError,
  } = useQuery({
    queryKey: ['team-performance', id],
    queryFn: () => api.getTeamPerformance(id!),
    enabled: !!id,
    staleTime: 30000,
  });

  const {
    data: membersData,
    isLoading: membersLoading,
    error: membersError,
  } = useQuery({
    queryKey: ['team-members', id],
    queryFn: () => api.getTeamMembers(id!),
    enabled: !!id,
    staleTime: 30000,
  });

  const { data: rolesData } = useQuery({
    queryKey: ['roles'],
    queryFn: api.getRoles,
    staleTime: 60000,
  });

  const members = membersData?.members || [];
  const roles = rolesData?.roles || [];

  const memberScoresQueries = useQuery({
    queryKey: ['team-member-scores', id, members.map((m) => m.id)],
    queryFn: async () => {
      const scorePromises = members.map(async (member) => {
        try {
          const result = await api.getPerformanceScores(member.id);
          return { memberId: member.id, metrics: result.metrics || [] };
        } catch {
          return { memberId: member.id, metrics: [] };
        }
      });
      return Promise.all(scorePromises);
    },
    enabled: members.length > 0,
    staleTime: 30000,
  });

  const memberMetricsMap = useMemo(() => {
    if (!memberScoresQueries.data) return new Map<string, MetricValue[]>();
    const map = new Map<string, MetricValue[]>();
    memberScoresQueries.data.forEach(({ memberId, metrics }) => {
      map.set(memberId, metrics);
    });
    return map;
  }, [memberScoresQueries.data]);

  const rolesMap = useMemo(() => {
    const map = new Map<string, Role>();
    roles.forEach((role) => {
      map.set(role.id, role);
    });
    return map;
  }, [roles]);

  const sortedMembers = useMemo(() => {
    return [...members].sort((a, b) => {
      const metricsA = memberMetricsMap.get(a.id) || [];
      const metricsB = memberMetricsMap.get(b.id) || [];
      const totalA = metricsA.find((m) => m.metric_name === 'engineer_total_score')?.value || 0;
      const totalB = metricsB.find((m) => m.metric_name === 'engineer_total_score')?.value || 0;
      return totalB - totalA;
    });
  }, [members, memberMetricsMap]);

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

  const isLoading = teamLoading || performanceLoading || membersLoading;
  const error = teamError || performanceError || membersError;

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded">
        <p className="text-red-800 dark:text-red-200">
          Failed to load team details: {(error as Error).message}
        </p>
      </div>
    );
  }

  if (!teamData || !performanceData) {
    return (
      <div className="text-center py-12">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">Team Not Found</h2>
        <Link
          to="/teams"
          className="inline-block px-4 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded hover:bg-blue-700 dark:hover:bg-blue-800 transition-colors"
        >
          Back to Teams
        </Link>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6">
        <Link
          to="/teams"
          className="text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 mb-4 inline-block"
        >
          ← Back to Teams
        </Link>
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">{teamData.name}</h1>
            {teamData.manager && (
              <p className="text-gray-600 dark:text-gray-400 mt-1">
                Engineering Manager: {teamData.manager.name}
              </p>
            )}
          </div>
        </div>
      </div>

      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 mb-8">
        <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100 mb-4">
          Team Performance
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
          <div>
            <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Team Score</div>
            <div className="flex items-baseline gap-2">
              <span className="text-3xl font-bold text-gray-900 dark:text-gray-100">
                {(performanceData.total_score ?? 0).toFixed(0)}/100
              </span>
              <span
                className={`text-sm font-medium ${getChangeColor(performanceData.score_change)}`}
              >
                {getChangeIcon(performanceData.score_change)}
              </span>
            </div>
          </div>

          <div>
            <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Team Size</div>
            <div className="text-3xl font-bold text-gray-900 dark:text-gray-100">
              {performanceData.member_count}
            </div>
          </div>

          <div>
            <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Velocity</div>
            <div className="flex items-baseline gap-2">
              <span className="text-3xl font-bold text-gray-900 dark:text-gray-100">
                {(performanceData.velocity_prs_per_week ?? 0).toFixed(0)}
              </span>
              <span className="text-sm text-gray-600 dark:text-gray-400">PRs/week</span>
              <span
                className={`text-sm font-medium ${getChangeColor(performanceData.velocity_change)}`}
              >
                {getChangeIcon(performanceData.velocity_change)}
                {(Math.abs((performanceData.velocity_change ?? 0) * 100) ?? 0).toFixed(0)}%
              </span>
            </div>
          </div>

          <div>
            <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Cycle Time</div>
            <div className="flex items-baseline gap-2">
              <span className="text-3xl font-bold text-gray-900 dark:text-gray-100">
                {(performanceData.cycle_time_days ?? 0).toFixed(1)}
              </span>
              <span className="text-sm text-gray-600 dark:text-gray-400">days</span>
              <span
                className={`text-sm font-medium ${getChangeColor(-performanceData.cycle_time_change)}`}
              >
                {getChangeIcon(-performanceData.cycle_time_change)}
                {(Math.abs((performanceData.cycle_time_change ?? 0) * 100) ?? 0).toFixed(0)}%
              </span>
            </div>
          </div>
        </div>
      </div>

      <div>
        <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">Team Members</h2>
        {memberScoresQueries.isLoading ? (
          <div className="flex justify-center items-center h-64">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400" />
          </div>
        ) : members.length === 0 ? (
          <div className="text-center py-12 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
            <p className="text-gray-600 dark:text-gray-400">No team members yet</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-6">
            {sortedMembers.map((member) => {
              const metrics = memberMetricsMap.get(member.id) || [];
              const role = member.role_id ? rolesMap.get(member.role_id) || null : null;
              return <ScoreCard key={member.id} engineer={member} metrics={metrics} role={role} />;
            })}
          </div>
        )}
      </div>
    </div>
  );
}
